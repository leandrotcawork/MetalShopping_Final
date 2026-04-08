package erp_runtime

import (
	"context"
	"fmt"
	"log"
	"strings"

	"metalshopping/integration_worker/internal/erp_runtime/raw"
	"metalshopping/integration_worker/internal/erp_runtime/reconciliation"
	"metalshopping/integration_worker/internal/erp_runtime/review"
	"metalshopping/integration_worker/internal/erp_runtime/runs"
	"metalshopping/integration_worker/internal/erp_runtime/staging"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// Runner orchestrates the full per-run pipeline:
//  1. Extract raw records from connector
//  2. Save to raw store
//  3. Normalize to staging
//  4. Reconcile staging records
//  5. Create review items for non-promotable records
//  6. Finalize run (update ledger with counts)
//
// NO promotion step — that is server_core's responsibility.
type Runner struct {
	registry    *Registry
	rawStore    *raw.Store
	normalizer  *staging.Normalizer
	reconciler  *reconciliation.Reconciler
	reviewStore *review.Store
	ledger      *runs.Ledger
	entitySteps *runs.EntityStepStore
}

// NewRunner constructs a Runner wiring all pipeline components.
func NewRunner(
	registry *Registry,
	rawStore *raw.Store,
	normalizer *staging.Normalizer,
	reconciler *reconciliation.Reconciler,
	reviewStore *review.Store,
	ledger *runs.Ledger,
	entitySteps *runs.EntityStepStore,
) *Runner {
	return &Runner{
		registry:    registry,
		rawStore:    rawStore,
		normalizer:  normalizer,
		reconciler:  reconciler,
		reviewStore: reviewStore,
		ledger:      ledger,
		entitySteps: entitySteps,
	}
}

// entityCounts accumulates pipeline outcome statistics.
type entityCounts struct {
	promoted int
	warnings int
	rejected int
	reviews  int
}

// Execute runs the full pipeline for a claimed run using structured connection config.
// On per-entity error: continues with remaining independent entities and marks the run as partial.
func (r *Runner) Execute(ctx context.Context, claim *runs.RunClaim, connection types.ExtractConnection) error {
	connector, err := r.registry.Get(claim.ConnectorType)
	if err != nil {
		markErr := r.ledger.MarkFailed(ctx, claim.RunID, fmt.Sprintf("connector not found: %v", err))
		if markErr != nil {
			log.Printf("runner: failed to mark run %s as failed: %v", claim.RunID, markErr)
		}
		return err
	}

	var total entityCounts
	var entityErrors []string
	failedEntities := map[types.EntityType]bool{}
	completedEntities := 0

	for _, entity := range orderedEntities(claim.EntityScope) {
		if shouldSkipDueToDependency(entity, failedEntities) {
			if r.entitySteps != nil {
				if err := r.entitySteps.MarkSkipped(ctx, claim.TenantID, claim.RunID, entity, "dependency failed"); err != nil {
					return fmt.Errorf("mark %s skipped: %w", entity, err)
				}
			}
			entityErrors = append(entityErrors, fmt.Sprintf("%s: skipped_due_to_dependency", entity))
			continue
		}

		if r.entitySteps != nil {
			if err := r.entitySteps.MarkStarted(ctx, claim.TenantID, claim.RunID, entity); err != nil {
				return fmt.Errorf("mark %s started: %w", entity, err)
			}
		}

		counts, err := r.processEntity(ctx, claim, connection, connector, entity)
		if err != nil {
			failedEntities[entity] = true
			entityErrors = append(entityErrors, fmt.Sprintf("%s: %v", entity, err))
			if r.entitySteps != nil {
				if stepErr := r.entitySteps.MarkFailed(ctx, claim.TenantID, claim.RunID, entity, err.Error()); stepErr != nil {
					return fmt.Errorf("mark %s failed: %w", entity, stepErr)
				}
			}
			log.Printf("runner: entity %s error in run %s: %v", entity, claim.RunID, err)
			continue
		}
		if r.entitySteps != nil {
			if err := r.entitySteps.MarkCompleted(ctx, claim.TenantID, claim.RunID, entity); err != nil {
				return fmt.Errorf("mark %s completed: %w", entity, err)
			}
		}

		completedEntities++
		total.promoted += counts.promoted
		total.warnings += counts.warnings
		total.rejected += counts.rejected
		total.reviews += counts.reviews
	}

	switch {
	case completedEntities == len(orderedEntities(claim.EntityScope)):
		return r.ledger.MarkCompleted(ctx, claim.RunID, total.promoted, total.warnings, total.rejected, total.reviews)
	case completedEntities == 0:
		return r.ledger.MarkFailed(ctx, claim.RunID, strings.Join(entityErrors, "; "))
	default:
		return r.ledger.MarkPartial(ctx, claim.RunID, strings.Join(entityErrors, "; "),
			total.promoted, total.warnings, total.rejected, total.reviews)
	}
}

// processEntity runs the extract→raw→staging→reconcile→review pipeline for a single entity.
func (r *Runner) processEntity(
	ctx context.Context,
	claim *runs.RunClaim,
	connection types.ExtractConnection,
	connector Connector,
	entity types.EntityType,
) (entityCounts, error) {
	var counts entityCounts
	var cursor *string
	batchOrdinal := 1

	for {
		req := types.ExtractRequest{
			TenantID:   claim.TenantID,
			RunID:      claim.RunID,
			Entity:     entity,
			Cursor:     cursor,
			Connection: connection,
		}

		result, err := connector.Extract(ctx, req)
		if err != nil {
			return counts, fmt.Errorf("extract: %w", err)
		}

		if len(result.Records) == 0 {
			break
		}

		for _, rec := range result.Records {
			if rec.BatchOrdinal <= 0 {
				rec.BatchOrdinal = batchOrdinal
			}
		}

		if r.entitySteps != nil {
			if err := r.entitySteps.MarkBatch(ctx, claim.TenantID, claim.RunID, entity, batchOrdinal, cursor); err != nil {
				return counts, fmt.Errorf("mark batch: %w", err)
			}
		}

		// Save raw records
		savedRaw, err := r.rawStore.Save(ctx, claim.TenantID, claim.RunID, result.Records)
		if err != nil {
			return counts, fmt.Errorf("raw save: %w", err)
		}

		// Normalize to staging
		stagingRecords, err := r.normalizer.Normalize(ctx, claim.TenantID, claim.RunID, savedRaw, connector)
		if err != nil {
			return counts, fmt.Errorf("normalize: %w", err)
		}

		// Reconcile
		reconResults, err := r.reconciler.Reconcile(ctx, claim.TenantID, claim.RunID, stagingRecords)
		if err != nil {
			return counts, fmt.Errorf("reconcile: %w", err)
		}

		// Create review items for non-promotable records
		if err := r.reviewStore.CreateFromReconciliation(ctx, claim.InstanceID, claim.ConnectorType, reconResults, stagingRecords); err != nil {
			return counts, fmt.Errorf("review items: %w", err)
		}

		// Tally counts
		for _, rr := range reconResults {
			switch rr.Classification {
			case reconciliation.ClassificationPromotable:
				counts.promoted++
			case reconciliation.ClassificationPromotableWithWarning:
				counts.promoted++
				counts.warnings++
			case reconciliation.ClassificationReviewRequired:
				counts.reviews++
			case reconciliation.ClassificationRejected:
				counts.rejected++
				counts.reviews++
			}
		}

		if !result.HasMore {
			break
		}
		cursor = result.NextCursor
		batchOrdinal++
	}

	return counts, nil
}

func orderedEntities(scope []string) []types.EntityType {
	order := []types.EntityType{
		types.EntityTypeProducts,
		types.EntityTypePrices,
		types.EntityTypeInventory,
		types.EntityTypeCustomers,
		types.EntityTypeSuppliers,
		types.EntityTypeSales,
		types.EntityTypePurchases,
		types.EntityTypeCosts,
	}

	allowed := map[types.EntityType]bool{}
	for _, raw := range scope {
		entity := types.EntityType(strings.TrimSpace(raw))
		allowed[entity] = true
	}

	result := make([]types.EntityType, 0, len(order))
	for _, entity := range order {
		if allowed[entity] {
			result = append(result, entity)
		}
	}
	return result
}

func shouldSkipDueToDependency(entity types.EntityType, failed map[types.EntityType]bool) bool {
	// Prices/inventory/sales/purchases depend on product identity being available.
	if failed[types.EntityTypeProducts] {
		switch entity {
		case types.EntityTypePrices, types.EntityTypeInventory, types.EntityTypeSales, types.EntityTypePurchases:
			return true
		}
	}
	return false
}
