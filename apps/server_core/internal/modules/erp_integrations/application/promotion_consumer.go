package application

import (
	"context"
	"errors"
	"log"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/events"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

const promotionBatchSize = 50

// PromotionConsumer polls erp_reconciliation_results for promotable records
// and promotes them into canonical domain tables.
//
// v1 limitation: the canonical module write ports are not available from this
// service layer. The consumer therefore writes a placeholder "promoted" marker:
// it claims the record, marks it as promoted (using source_id as a proxy for
// the canonical_id), and publishes the entity_promoted outbox event. The actual
// domain write (e.g., updating the canonical products table) will be added in a
// follow-up task when canonical module ports are wired into this layer.
type PromotionConsumer struct {
	reconRepo      ports.ReconciliationReader
	autoPromoGuard ports.AutoPromotionGuard
	outboxStore    *outbox.Store
	interval       time.Duration
}

// NewPromotionConsumer constructs a PromotionConsumer.
func NewPromotionConsumer(
	reconRepo ports.ReconciliationReader,
	autoPromoGuard ports.AutoPromotionGuard,
	outboxStore *outbox.Store,
) *PromotionConsumer {
	return &PromotionConsumer{
		reconRepo:      reconRepo,
		autoPromoGuard: autoPromoGuard,
		outboxStore:    outboxStore,
		interval:       5 * time.Second,
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
	// Check whether auto-promotion is permitted for this tenant.
	if err := c.autoPromoGuard.CheckAutoPromotion(ctx, result.TenantID); err != nil {
		if errors.Is(err, domain.ErrAutoPromotionDisabled) {
			// Tenant has disabled auto-promotion; leave as pending for manual resolution.
			return
		}
		log.Printf("WARN erp PromotionConsumer: check auto-promotion for tenant %s: %v", result.TenantID, err)
		return
	}

	// Atomically claim the record (promotion_status: pending → promoting).
	if err := c.reconRepo.ClaimForPromotion(ctx, result.TenantID, result.ReconciliationID); err != nil {
		log.Printf("WARN erp PromotionConsumer: claim reconciliation %s: %v", result.ReconciliationID, err)
		return
	}

	// v1 placeholder: no canonical module write in v1.
	// Use source_id as a proxy canonical_id until canonical module ports are wired.
	// TODO(follow-up): replace with actual canonical module write when ports are available.
	canonicalID := result.SourceID

	if err := c.reconRepo.MarkPromoted(ctx, result.TenantID, result.ReconciliationID, canonicalID); err != nil {
		log.Printf("WARN erp PromotionConsumer: mark promoted for reconciliation %s: %v", result.ReconciliationID, err)
		if failErr := c.reconRepo.MarkPromotionFailed(ctx, result.TenantID, result.ReconciliationID); failErr != nil {
			log.Printf("WARN erp PromotionConsumer: mark promotion failed for reconciliation %s: %v", result.ReconciliationID, failErr)
		}
		return
	}

	// Publish entity_promoted outbox event (best-effort).
	if c.outboxStore != nil {
		now := time.Now().UTC()
		// Set the canonical_id on the result for the event payload.
		result.CanonicalID = &canonicalID
		record, err := events.NewEntityPromotedOutboxRecord(result, "", now)
		if err != nil {
			log.Printf("WARN erp PromotionConsumer: build entity_promoted outbox record for %s: %v", result.ReconciliationID, err)
			return
		}
		if appendErr := c.outboxStore.Append(ctx, []outbox.Record{record}); appendErr != nil {
			log.Printf("WARN erp PromotionConsumer: append entity_promoted outbox event for %s: %v (promotion already marked)", result.ReconciliationID, appendErr)
		}
	}
}
