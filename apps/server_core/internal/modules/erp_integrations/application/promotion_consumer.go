package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

const promotionBatchSize = 50
const autoPromotionDisabledReasonCode = "ERP_PROMOTION_AUTO_DISABLED"
const autoPromotionDisabledReviewSummary = "auto-promotion is disabled for this tenant"
const autoPromotionDisabledRecommendedAction = "enable auto-promotion or route the item through manual review"

// PromotionConsumer polls erp_reconciliation_results for promotable records and
// promotes product, price, and inventory rows in canonical order.
type PromotionConsumer struct {
	reconRepo          ports.ReconciliationReader
	autoPromoGuard     ports.AutoPromotionGuard
	productPromotion   productPromoter
	pricePromotion     pricePromoter
	inventoryPromotion inventoryPromoter
	interval           time.Duration
}

type productPromoter interface {
	PromoteProduct(ctx context.Context, result *domain.ReconciliationResult) (string, error)
}

type pricePromoter interface {
	PromotePrice(ctx context.Context, result *domain.ReconciliationResult) (string, error)
}

type inventoryPromoter interface {
	PromoteInventory(ctx context.Context, result *domain.ReconciliationResult) (string, error)
}

// NewPromotionConsumer constructs a PromotionConsumer.
func NewPromotionConsumer(
	reconRepo ports.ReconciliationReader,
	autoPromoGuard ports.AutoPromotionGuard,
	productPromotion productPromoter,
	pricePromotion pricePromoter,
	inventoryPromotion inventoryPromoter,
) *PromotionConsumer {
	return &PromotionConsumer{
		reconRepo:          reconRepo,
		autoPromoGuard:     autoPromoGuard,
		productPromotion:   productPromotion,
		pricePromotion:     pricePromotion,
		inventoryPromotion: inventoryPromotion,
		interval:           5 * time.Second,
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

	sort.SliceStable(results, func(i, j int) bool {
		pi, pj := promotionPriority(results[i].EntityType), promotionPriority(results[j].EntityType)
		if pi != pj {
			return pi < pj
		}
		if !results[i].ReconciledAt.Equal(results[j].ReconciledAt) {
			return results[i].ReconciledAt.Before(results[j].ReconciledAt)
		}
		return results[i].ReconciliationID < results[j].ReconciliationID
	})

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

	switch result.EntityType {
	case domain.EntityTypeProducts:
		canonicalID, err := c.productPromotion.PromoteProduct(ctx, result)
		if err != nil {
			c.handlePromotionFailure(ctx, result, &promotionResult, err, "catalog promotion failed")
			return
		}
		log.Printf("INFO erp PromotionConsumer: promoted reconciliation %s tenant=%s canonical_id=%s", result.ReconciliationID, result.TenantID, canonicalID)
	case domain.EntityTypePrices:
		canonicalID, err := c.pricePromotion.PromotePrice(ctx, result)
		if err != nil {
			c.handlePromotionFailure(ctx, result, &promotionResult, err, "price promotion failed")
			return
		}
		log.Printf("INFO erp PromotionConsumer: promoted reconciliation %s tenant=%s canonical_id=%s", result.ReconciliationID, result.TenantID, canonicalID)
	case domain.EntityTypeInventory:
		canonicalID, err := c.inventoryPromotion.PromoteInventory(ctx, result)
		if err != nil {
			c.handlePromotionFailure(ctx, result, &promotionResult, err, "inventory promotion failed")
			return
		}
		log.Printf("INFO erp PromotionConsumer: promoted reconciliation %s tenant=%s canonical_id=%s", result.ReconciliationID, result.TenantID, canonicalID)
	default:
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
}

const promotionFailureReasonCode = "ERP_PROMOTION_FAILED"
const promotionBlockedByProductReasonCode = "ERP_RELATED_PRODUCT_NOT_PROMOTED"

func promotionPriority(entityType domain.EntityType) int {
	switch entityType {
	case domain.EntityTypeProducts:
		return 0
	case domain.EntityTypePrices:
		return 1
	case domain.EntityTypeInventory:
		return 2
	default:
		return 3
	}
}

func (c *PromotionConsumer) handlePromotionFailure(ctx context.Context, result, promotionResult *domain.ReconciliationResult, err error, failureStep string) {
	reviewCode := promotionFailureReasonCode
	reviewSummary := "promotion failed"
	reviewAction := "inspect the source record and retry the promotion"

	switch {
	case errors.Is(err, ErrRelatedProductNotPromoted):
		reviewCode = promotionBlockedByProductReasonCode
		reviewSummary = relatedProductBlockedSummary
		reviewAction = relatedProductBlockedAction
	case errors.Is(err, ErrPriceContextMappingNotFound):
		reviewCode = priceContextMappingMissingReasonCode
		reviewSummary = priceContextMappingMissingSummary
		reviewAction = priceContextMappingMissingAction
	}

	warningDetails := buildPromotionFailureWarningDetails(promotionResult, reviewCode, failureStep, err)
	if reviewCode != promotionFailureReasonCode {
		if reviewErr := c.reconRepo.MarkReviewRequired(
			ctx,
			result.TenantID,
			result.ReconciliationID,
			reviewCode,
			reviewSummary,
			reviewAction,
			warningDetails,
		); reviewErr != nil {
			log.Printf("WARN erp PromotionConsumer: mark review required for reconciliation %s tenant=%s reason_code=%s: %v", result.ReconciliationID, result.TenantID, reviewCode, reviewErr)
			failureDetails := buildPromotionFailureWarningDetails(promotionResult, promotionFailureReasonCode, "mark review required failed", reviewErr)
			if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID, promotionFailureReasonCode, failureDetails); failErr != nil {
				log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
			}
			return
		}
		log.Printf("INFO erp PromotionConsumer: routed reconciliation %s tenant=%s to review_required reason_code=%s", result.ReconciliationID, result.TenantID, reviewCode)
		return
	}

	log.Printf(
		"WARN erp PromotionConsumer: promote reconciliation %s tenant=%s staging_id=%s source_id=%s entity_type=%s: %v",
		result.ReconciliationID,
		result.TenantID,
		result.StagingID,
		result.SourceID,
		result.EntityType,
		err,
	)
	failureDetails := buildPromotionFailureWarningDetails(promotionResult, promotionFailureReasonCode, failureStep, err)
	if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID, promotionFailureReasonCode, failureDetails); failErr != nil {
		log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
	}
}

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
