package reconciliation

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"metalshopping/integration_worker/internal/erp_runtime/staging"
	"metalshopping/integration_worker/internal/erp_runtime/tenantdb"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// Classification represents the reconciliation outcome for a staging record.
type Classification string

const (
	ClassificationPromotable            Classification = "promotable"
	ClassificationPromotableWithWarning Classification = "promotable_with_warning"
	ClassificationReviewRequired        Classification = "review_required"
	ClassificationRejected              Classification = "rejected"
)

// ReconciliationResult holds the outcome of reconciling a single staging record.
type ReconciliationResult struct {
	ReconciliationID string
	TenantID         string
	RunID            string
	StagingID        string
	EntityType       types.EntityType
	SourceID         string
	Action           string // "create", "update", "skip"
	Classification   Classification
	ReasonCode       string
	WarningDetails   *string
	ReconciledAt     time.Time
}

// Reconciler classifies staging records and writes results to erp_reconciliation_results.
type Reconciler struct {
	db *sql.DB
}

// NewReconciler constructs a Reconciler backed by the given DB connection.
func NewReconciler(db *sql.DB) *Reconciler {
	return &Reconciler{db: db}
}

// Reconcile classifies each staging record as promotable/review_required/rejected
// and persists the result. Returns the reconciliation results.
//
// v1 logic:
//   - valid staging records   → promotable,  action = "create"
//   - invalid staging records → rejected,    action = "skip"
func (r *Reconciler) Reconcile(
	ctx context.Context,
	tenantID, runID string,
	stagingRecords []*staging.StagingRecord,
) ([]*ReconciliationResult, error) {
	tx, err := tenantdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
INSERT INTO erp_reconciliation_results
  (reconciliation_id, tenant_id, run_id, staging_id, entity_type, source_id,
   action, classification, reason_code, warning_details, reconciled_at, promotion_status)
VALUES ($1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending')`

	now := time.Now().UTC()
	results := make([]*ReconciliationResult, 0, len(stagingRecords))

	for _, s := range stagingRecords {
		reconID := uuid.New().String()

		var classification Classification
		var action, reasonCode string

		switch s.ValidationStatus {
		case staging.ValidationStatusValid:
			classification = ClassificationPromotable
			action = "create"
			reasonCode = "valid_record"
		default:
			classification = ClassificationRejected
			action = "skip"
			reasonCode = "validation_failed"
		}

		_, err := tx.ExecContext(ctx, q,
			reconID,
			runID,
			s.StagingID,
			string(s.EntityType),
			s.SourceID,
			action,
			string(classification),
			reasonCode,
			nil, // warning_details — unused in v1
			now,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, &ReconciliationResult{
			ReconciliationID: reconID,
			TenantID:         tenantID,
			RunID:            runID,
			StagingID:        s.StagingID,
			EntityType:       s.EntityType,
			SourceID:         s.SourceID,
			Action:           action,
			Classification:   classification,
			ReasonCode:       reasonCode,
			ReconciledAt:     now,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}
