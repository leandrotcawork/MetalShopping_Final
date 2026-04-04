package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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

		severity, recommendedAction, problemSummary := reviewContextForResult(r)

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

type duplicateReviewWarningDetails struct {
	BlockingScope   string                    `json:"blocking_scope"`
	BlockedEntities []string                  `json:"blocked_entities"`
	Conflicts       []duplicateReviewConflict `json:"conflicts"`
}

type duplicateReviewConflict struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

func reviewContextForResult(result *reconciliation_pkg.ReconciliationResult) (severity, recommendedAction, problemSummary string) {
	switch result.Classification {
	case reconciliation_pkg.ClassificationRejected:
		return "error", "fix in source ERP and reprocess", "Record rejected during reconciliation: " + result.ReasonCode
	case reconciliation_pkg.ClassificationReviewRequired:
		if details, ok := parseDuplicateReviewWarningDetails(result.WarningDetails); ok {
			return "warning", "review duplicate secondary identifiers in the source ERP and reprocess", duplicateProblemSummary(details)
		}
		return "warning", "review in MetalShopping", "Record requires manual review: " + result.ReasonCode
	default:
		return "info", "review in MetalShopping", "Record requires manual review: " + result.ReasonCode
	}
}

func parseDuplicateReviewWarningDetails(raw *string) (duplicateReviewWarningDetails, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return duplicateReviewWarningDetails{}, false
	}

	var details duplicateReviewWarningDetails
	if err := json.Unmarshal([]byte(*raw), &details); err != nil {
		return duplicateReviewWarningDetails{}, false
	}
	if len(details.Conflicts) == 0 || strings.TrimSpace(details.BlockingScope) == "" {
		return duplicateReviewWarningDetails{}, false
	}
	return details, true
}

func duplicateProblemSummary(details duplicateReviewWarningDetails) string {
	blockedTargets := "products, prices, and inventory"
	if len(details.BlockedEntities) > 0 {
		blockedTargets = joinWithOxfordComma(details.BlockedEntities)
	}

	switch len(details.Conflicts) {
	case 0:
		return "Duplicate secondary identifiers block " + blockedTargets + " promotion"
	case 1:
		conflict := details.Conflicts[0]
		field := duplicateFieldLabel(conflict.Field)
		if strings.TrimSpace(conflict.Value) != "" {
			return fmt.Sprintf("Duplicate %s value %q blocks %s promotion", field, conflict.Value, blockedTargets)
		}
		return fmt.Sprintf("Duplicate %s blocks %s promotion", field, blockedTargets)
	default:
		fields := make([]string, 0, len(details.Conflicts))
		seen := map[string]struct{}{}
		for _, conflict := range details.Conflicts {
			label := duplicateFieldLabel(conflict.Field)
			if _, ok := seen[label]; ok {
				continue
			}
			seen[label] = struct{}{}
			fields = append(fields, label)
		}
		if len(fields) == 0 {
			return "Duplicate secondary identifiers block " + blockedTargets + " promotion"
		}
		if len(fields) == 1 {
			return fmt.Sprintf("Duplicate %s values block %s promotion", fields[0], blockedTargets)
		}
		return fmt.Sprintf("Duplicate %s and %s values block %s promotion", fields[0], fields[1], blockedTargets)
	}
}

func duplicateFieldLabel(field string) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "ean":
		return "EAN"
	case "manufacturer_reference":
		return "manufacturer reference"
	default:
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			return trimmed
		}
		return "secondary identifier"
	}
}

func joinWithOxfordComma(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	switch len(parts) {
	case 0:
		return "product, prices, and inventory"
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " and " + parts[1]
	default:
		return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
	}
}
