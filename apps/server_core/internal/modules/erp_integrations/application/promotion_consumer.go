package application

import (
	"context"
	"errors"
	"log"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

const promotionBatchSize = 50

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
			// Tenant has disabled auto-promotion; leave as pending for manual resolution.
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

	if result.EntityType != domain.EntityTypeProducts {
		if err := c.reconRepo.MarkReviewRequired(ctx, result.TenantID, result.ReconciliationID, unsupportedPromotionEntityReasonCode); err != nil {
			log.Printf("WARN erp PromotionConsumer: mark review required for reconciliation %s: %v", result.ReconciliationID, err)
			if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID); failErr != nil {
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
		if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID); failErr != nil {
			log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
		}
		return
	}

	if err := c.reconRepo.MarkPromoted(ctx, result.TenantID, result.ReconciliationID, canonicalID); err != nil {
		log.Printf("WARN erp PromotionConsumer: mark promoted for reconciliation %s: %v", result.ReconciliationID, err)
		if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID); failErr != nil {
			log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
		}
	}
}
