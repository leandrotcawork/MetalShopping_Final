package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

const promotionBatchSize = 50
const autoPromotionDisabledReasonCode = "ERP_PROMOTION_AUTO_DISABLED"
const autoPromotionDisabledReviewSummary = "auto-promotion is disabled for this tenant"
const autoPromotionDisabledRecommendedAction = "enable auto-promotion or route the item through manual review"

// PromotionConsumer polls erp_reconciliation_results for promotable records and
// promotes product rows into the canonical catalog.
type PromotionConsumer struct {
	reconRepo        ports.ReconciliationReader
	autoPromoGuard   ports.AutoPromotionGuard
	productPromotion *ProductPromotion
	interval         time.Duration
}

// NewPromotionConsumer constructs a PromotionConsumer.
func NewPromotionConsumer(
	reconRepo ports.ReconciliationReader,
	autoPromoGuard ports.AutoPromotionGuard,
	productPromotion *ProductPromotion,
) *PromotionConsumer {
	return &PromotionConsumer{
		reconRepo:        reconRepo,
		autoPromoGuard:   autoPromoGuard,
		productPromotion: productPromotion,
		interval:         5 * time.Second,
	}
}

// Start begins the promotion loop, polling on the configured interval until
// ctx is cancelled.
func (c *PromotionConsumer) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.runPromotion(ctx)
		}
	}
}

func (c *PromotionConsumer) runPromotion(ctx context.Context) {
	results, err := c.reconRepo.ListAllPendingPromotion(ctx, promotionBatchSize)
	if err != nil {
		log.Printf("WARN erp PromotionConsumer: list pending promotion: %v", err)
		return
	}

	for _, result := range results {
		if ctx.Err() != nil {
			return
		}
		c.promoteOne(ctx, result)
	}
}

func (c *PromotionConsumer) promoteOne(ctx context.Context, result *domain.ReconciliationResult) {
	if result == nil {
		log.Printf("WARN erp PromotionConsumer: skip nil reconciliation result")
		return
	}

	// Check whether auto-promotion is permitted for this tenant.
	if err := c.autoPromoGuard.CheckAutoPromotion(ctx, result.TenantID); err != nil {
		if errors.Is(err, domain.ErrAutoPromotionDisabled) {
			promotionResult := *result
			promotionResult.PromotionStatus = domain.PromotionStatusFailed
			warningDetails := buildPromotionFailureWarningDetails(&promotionResult, autoPromotionDisabledReasonCode, "auto-promotion disabled", nil)
			if reviewErr := c.reconRepo.MarkReviewRequired(
				ctx,
				result.TenantID,
				result.ReconciliationID,
				autoPromotionDisabledReasonCode,
				autoPromotionDisabledReviewSummary,
				autoPromotionDisabledRecommendedAction,
				warningDetails,
			); reviewErr != nil {
				log.Printf("WARN erp PromotionConsumer: mark review required for reconciliation %s tenant=%s reason_code=%s: %v", result.ReconciliationID, result.TenantID, autoPromotionDisabledReasonCode, reviewErr)
				failureDetails := buildPromotionFailureWarningDetails(&promotionResult, promotionFailureReasonCode, "mark review required failed", reviewErr)
				if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID, promotionFailureReasonCode, failureDetails); failErr != nil {
					log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
				}
				return
			}
			log.Printf("INFO erp PromotionConsumer: routed reconciliation %s tenant=%s to review_required reason_code=%s", result.ReconciliationID, result.TenantID, autoPromotionDisabledReasonCode)
			return
		}
		log.Printf("WARN erp PromotionConsumer: check auto-promotion for tenant %s: %v", result.TenantID, err)
		return
	}

	// Atomically claim the record (promotion_status: pending -> promoting).
	claimed, err := c.reconRepo.ClaimForPromotion(ctx, result.TenantID, result.ReconciliationID)
	if err != nil {
		log.Printf("WARN erp PromotionConsumer: claim reconciliation %s: %v", result.ReconciliationID, err)
		return
	}
	if !claimed {
		log.Printf("INFO erp PromotionConsumer: skip already claimed reconciliation %s", result.ReconciliationID)
		return
	}

	promotionResult := *result
	promotionResult.PromotionStatus = domain.PromotionStatusPromoting

	if result.EntityType != domain.EntityTypeProducts {
		warningDetails := buildPromotionFailureWarningDetails(&promotionResult, unsupportedPromotionEntityReasonCode, "unsupported entity type", nil)
		if err := c.reconRepo.MarkReviewRequired(
			ctx,
			result.TenantID,
			result.ReconciliationID,
			unsupportedPromotionEntityReasonCode,
			"unsupported ERP entity type",
			"review the entity mapping before rerunning promotion",
			warningDetails,
		); err != nil {
			log.Printf("WARN erp PromotionConsumer: mark review required for reconciliation %s tenant=%s entity_type=%s reason_code=%s: %v", result.ReconciliationID, result.TenantID, result.EntityType, unsupportedPromotionEntityReasonCode, err)
			failureDetails := buildPromotionFailureWarningDetails(&promotionResult, promotionFailureReasonCode, "mark review required failed", err)
			if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID, promotionFailureReasonCode, failureDetails); failErr != nil {
				log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
			}
		}
		log.Printf(
			"WARN erp PromotionConsumer: unsupported entity type for reconciliation %s tenant=%s entity_type=%s reason_code=%s",
			result.ReconciliationID,
			result.TenantID,
			result.EntityType,
			unsupportedPromotionEntityReasonCode,
		)
		return
	}

	canonicalID, err := c.productPromotion.PromoteProduct(ctx, result)
	if err != nil {
		log.Printf(
			"WARN erp PromotionConsumer: promote product reconciliation %s tenant=%s staging_id=%s source_id=%s entity_type=%s: %v",
			result.ReconciliationID,
			result.TenantID,
			result.StagingID,
			result.SourceID,
			result.EntityType,
			err,
		)
		failureDetails := buildPromotionFailureWarningDetails(&promotionResult, promotionFailureReasonCode, "catalog promotion failed", err)
		if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID, promotionFailureReasonCode, failureDetails); failErr != nil {
			log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
		}
		return
	}

	if err := c.reconRepo.MarkPromoted(ctx, result.TenantID, result.ReconciliationID, canonicalID); err != nil {
		log.Printf("WARN erp PromotionConsumer: mark promoted for reconciliation %s: %v", result.ReconciliationID, err)
		failureDetails := buildPromotionFailureWarningDetails(&promotionResult, promotionFailureReasonCode, "mark promoted failed", err)
		if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID, promotionFailureReasonCode, failureDetails); failErr != nil {
			log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
		}
	}
}

const promotionFailureReasonCode = "ERP_PROMOTION_FAILED"

func buildPromotionFailureWarningDetails(result *domain.ReconciliationResult, reasonCode, failureStep string, cause error) *string {
	if result == nil {
		return nil
	}

	details := map[string]any{
		"reason_code":       reasonCode,
		"failure_step":      failureStep,
		"reconciliation_id": result.ReconciliationID,
		"tenant_id":         result.TenantID,
		"run_id":            result.RunID,
		"staging_id":        result.StagingID,
		"source_id":         result.SourceID,
		"entity_type":       result.EntityType,
		"promotion_status":  result.PromotionStatus,
		"reconciled_at":     result.ReconciledAt.UTC().Format(time.RFC3339),
	}
	if cause != nil {
		details["error"] = cause.Error()
	}

	payload, err := json.Marshal(details)
	if err != nil {
		fallback := fmt.Sprintf(`{"reason_code":%q,"failure_step":%q,"error":%q}`, reasonCode, failureStep, err.Error())
		return &fallback
	}
	value := string(payload)
	return &value
}
