package review

import (
	"context"
	"database/sql"
	"encoding/json"
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
   reconciliation_id, staging_snapshot, reconciliation_output, recommended_action, item_status, created_at)
VALUES ($1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb, $15, 'open', $16)`

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

		// Resolve raw_id and staging snapshot from staging index.
		rawID := ""
		var stagingSnapshot *string
		var reconciliationOutput *string
		if sr, ok := idx[r.StagingID]; ok {
			rawID = sr.RawID
			if len(sr.NormalizedJSON) > 0 {
				value := string(sr.NormalizedJSON)
				stagingSnapshot = &value
			}
		}

		reconciliationOutput, err = marshalReconciliationOutput(r)
		if err != nil {
			return fmt.Errorf("marshal reconciliation output: %w", err)
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
			stagingSnapshot,
			reconciliationOutput,
			recommendedAction,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func marshalReconciliationOutput(result *reconciliation_pkg.ReconciliationResult) (*string, error) {
	payload := map[string]any{
		"reconciliation_id": result.ReconciliationID,
		"entity_type":       string(result.EntityType),
		"source_id":         result.SourceID,
		"action":            result.Action,
		"classification":    string(result.Classification),
		"reason_code":       result.ReasonCode,
	}
	if result.WarningDetails != nil {
		if json.Valid([]byte(*result.WarningDetails)) {
			payload["warning_details"] = json.RawMessage(*result.WarningDetails)
		} else {
			payload["warning_details_raw"] = *result.WarningDetails
		}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	value := string(encoded)
	return &value, nil
}
