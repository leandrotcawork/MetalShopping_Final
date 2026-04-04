package reconciliation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

	duplicateSecondaryIdentifierReasonCode = "ERP_PRODUCT_IDENTIFIER_CONFLICT"
	duplicateBlockingScope                 = "product_prices_inventory"
)

var duplicateBlockingEntities = []string{"products", "prices", "inventory"}

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

type identifierOccurrence struct {
	stagingID string
	sourceID  string
	value     string
}

type duplicateConflict struct {
	Field      string   `json:"field"`
	Value      string   `json:"value"`
	Normalized string   `json:"normalized_value"`
	StagingIDs []string `json:"staging_ids"`
	SourceIDs  []string `json:"source_ids"`
}

type duplicateWarningDetails struct {
	ReasonCode      string              `json:"reason_code"`
	BlockingScope   string              `json:"blocking_scope"`
	BlockedEntities []string            `json:"blocked_entities"`
	EntityType      string              `json:"entity_type"`
	SourceID        string              `json:"source_id"`
	StagingID       string              `json:"staging_id"`
	RunID           string              `json:"run_id"`
	Conflicts       []duplicateConflict `json:"conflicts"`
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
	duplicateFindings := collectDuplicateSecondaryIdentifierFindings(stagingRecords)

	for _, s := range stagingRecords {
		reconID := uuid.New().String()

		var classification Classification
		var action, reasonCode string
		var warningDetails *string

		switch s.ValidationStatus {
		case staging.ValidationStatusValid:
			if s.EntityType == types.EntityTypeProducts {
				if conflicts := duplicateFindings[s.StagingID]; len(conflicts) > 0 {
					classification = ClassificationReviewRequired
					action = "skip"
					reasonCode = duplicateSecondaryIdentifierReasonCode
					warningDetails, err = buildDuplicateWarningDetails(s, conflicts)
					if err != nil {
						return nil, err
					}
					break
				}
			}
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
			warningDetails,
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
			WarningDetails:   warningDetails,
			ReconciledAt:     now,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}

func collectDuplicateSecondaryIdentifierFindings(stagingRecords []*staging.StagingRecord) map[string][]duplicateConflict {
	grouped := map[string]map[string][]identifierOccurrence{
		"ean":                    make(map[string][]identifierOccurrence),
		"manufacturer_reference": make(map[string][]identifierOccurrence),
	}

	for _, record := range stagingRecords {
		if record == nil || record.EntityType != types.EntityTypeProducts || record.ValidationStatus != staging.ValidationStatusValid {
			continue
		}

		payload, ok := decodeObject(record.NormalizedJSON)
		if !ok {
			continue
		}

		if value := readFirstStringField(payload, "ean", "REFERENCIA"); value != "" {
			key := normalizeIdentifierValue(value)
			grouped["ean"][key] = append(grouped["ean"][key], identifierOccurrence{
				stagingID: record.StagingID,
				sourceID:  record.SourceID,
				value:     strings.TrimSpace(value),
			})
		}
		if value := readFirstStringField(payload, "manufacturer_reference", "REFFORN"); value != "" {
			key := normalizeIdentifierValue(value)
			grouped["manufacturer_reference"][key] = append(grouped["manufacturer_reference"][key], identifierOccurrence{
				stagingID: record.StagingID,
				sourceID:  record.SourceID,
				value:     strings.TrimSpace(value),
			})
		}
	}

	results := make(map[string][]duplicateConflict)
	for _, field := range []string{"ean", "manufacturer_reference"} {
		values := grouped[field]
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			occurrences := values[key]
			if len(occurrences) < 2 {
				continue
			}

			conflict := duplicateConflict{
				Field:      field,
				Value:      occurrences[0].value,
				Normalized: key,
				StagingIDs: uniqueSortedStrings(extractOccurrenceStagingIDs(occurrences)),
				SourceIDs:  uniqueSortedStrings(extractOccurrenceSourceIDs(occurrences)),
			}
			for _, occurrence := range occurrences {
				results[occurrence.stagingID] = append(results[occurrence.stagingID], conflict)
			}
		}
	}

	for stagingID, conflicts := range results {
		sort.Slice(conflicts, func(i, j int) bool {
			if conflicts[i].Field == conflicts[j].Field {
				return conflicts[i].Normalized < conflicts[j].Normalized
			}
			return conflicts[i].Field < conflicts[j].Field
		})
		results[stagingID] = conflicts
	}

	return results
}

func buildDuplicateWarningDetails(record *staging.StagingRecord, conflicts []duplicateConflict) (*string, error) {
	details := duplicateWarningDetails{
		ReasonCode:      duplicateSecondaryIdentifierReasonCode,
		BlockingScope:   duplicateBlockingScope,
		BlockedEntities: append([]string(nil), duplicateBlockingEntities...),
		EntityType:      string(record.EntityType),
		SourceID:        record.SourceID,
		StagingID:       record.StagingID,
		RunID:           record.RunID,
		Conflicts:       append([]duplicateConflict(nil), conflicts...),
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("marshal duplicate warning details for staging %s: %w", record.StagingID, err)
	}
	value := string(payload)
	return &value, nil
}

func decodeObject(raw json.RawMessage) (map[string]any, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false
	}
	return payload, true
}

func readFirstStringField(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			if text := toStringValue(value); text != "" {
				return text
			}
		}
	}
	return ""
}

func toStringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []byte:
		return strings.TrimSpace(string(typed))
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	case float64:
		if typed == float64(int64(typed)) {
			return strings.TrimSpace(fmt.Sprintf("%.0f", typed))
		}
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func normalizeIdentifierValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func extractOccurrenceStagingIDs(occurrences []identifierOccurrence) []string {
	ids := make([]string, 0, len(occurrences))
	for _, occurrence := range occurrences {
		ids = append(ids, occurrence.stagingID)
	}
	return ids
}

func extractOccurrenceSourceIDs(occurrences []identifierOccurrence) []string {
	ids := make([]string, 0, len(occurrences))
	for _, occurrence := range occurrences {
		ids = append(ids, occurrence.sourceID)
	}
	return ids
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)

	out := sorted[:0]
	var prev string
	for i, value := range sorted {
		if i == 0 || value != prev {
			out = append(out, value)
			prev = value
		}
	}
	return append([]string(nil), out...)
}
