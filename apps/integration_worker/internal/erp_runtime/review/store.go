package review

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	reconciliation_pkg "metalshopping/integration_worker/internal/erp_runtime/reconciliation"
	staging_pkg "metalshopping/integration_worker/internal/erp_runtime/staging"
	"metalshopping/integration_worker/internal/erp_runtime/tenantdb"
)

// Store writes review items to erp_review_items.
type Store struct {
	db *sql.DB
}

// NewStore constructs a Store backed by the given DB connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// stagingIndex builds a lookup map from staging_id → StagingRecord.
func stagingIndex(stagingRecords []*staging_pkg.StagingRecord) map[string]*staging_pkg.StagingRecord {
	m := make(map[string]*staging_pkg.StagingRecord, len(stagingRecords))
	for _, s := range stagingRecords {
		m[s.StagingID] = s
	}
	return m
}

// CreateFromReconciliation writes review items for rejected or review_required records.
// instanceID and connectorType identify the integration instance that produced this run.
// stagingRecords are needed to resolve raw_id for each reconciliation result.
func (s *Store) CreateFromReconciliation(
	ctx context.Context,
	instanceID, connectorType string,
	results []*reconciliation_pkg.ReconciliationResult,
	stagingRecords []*staging_pkg.StagingRecord,
) error {
	// Filter to only actionable results
	var actionable []*reconciliation_pkg.ReconciliationResult
	for _, r := range results {
		if r.Classification == reconciliation_pkg.ClassificationRejected ||
			r.Classification == reconciliation_pkg.ClassificationReviewRequired {
			actionable = append(actionable, r)
		}
	}
	if len(actionable) == 0 {
		return nil
	}

	idx := stagingIndex(stagingRecords)
	tenantID := actionable[0].TenantID
	if tenantID == "" {
		return fmt.Errorf("review items: actionable reconciliation %s missing tenant id", actionable[0].ReconciliationID)
	}
	for _, r := range actionable[1:] {
		if r.TenantID != tenantID {
			return fmt.Errorf("review items: actionable results span multiple tenants: %s and %s", tenantID, r.TenantID)
		}
	}

	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
INSERT INTO erp_review_items
  (review_id, tenant_id, instance_id, connector_type, entity_type, source_id,
   run_id, severity, reason_code, problem_summary, raw_id, staging_id,
   reconciliation_id, recommended_action, item_status, created_at)
VALUES ($1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'open', $14)`

	now := time.Now().UTC()
	for _, r := range actionable {
		reviewID := uuid.New().String()

		var severity, recommendedAction, problemSummary string
		switch r.Classification {
		case reconciliation_pkg.ClassificationRejected:
			severity = "error"
			recommendedAction = "fix in source ERP and reprocess"
			problemSummary = "Record rejected during reconciliation: " + r.ReasonCode
		case reconciliation_pkg.ClassificationReviewRequired:
			severity = "warning"
			recommendedAction = "review in MetalShopping"
			problemSummary = "Record requires manual review: " + r.ReasonCode
		}

		// Resolve raw_id from staging index
		rawID := ""
		if sr, ok := idx[r.StagingID]; ok {
			rawID = sr.RawID
		}

		_, err := tx.ExecContext(ctx, q,
			reviewID,
			instanceID,
			connectorType,
			string(r.EntityType),
			r.SourceID,
			r.RunID,
			severity,
			r.ReasonCode,
			problemSummary,
			rawID,
			r.StagingID,
			r.ReconciliationID,
			recommendedAction,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
