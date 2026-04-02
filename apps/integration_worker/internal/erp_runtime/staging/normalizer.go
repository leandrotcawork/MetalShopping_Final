package staging

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"metalshopping/integration_worker/internal/erp_runtime/raw"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// Connector is a minimal interface the normalizer needs — only used as a marker in v1.
type Connector interface {
	Type() string
}

// Normalizer converts raw records to staging records.
type Normalizer struct {
	db *sql.DB
}

// NewNormalizer constructs a Normalizer backed by the given DB connection.
func NewNormalizer(db *sql.DB) *Normalizer {
	return &Normalizer{db: db}
}

// Normalize maps raw records to staging records using a v1 passthrough strategy.
// The connector's Extract already returns structured data, so normalization
// validates the JSON and persists it in erp_staging_records.
// Returns the persisted staging records.
func (n *Normalizer) Normalize(
	ctx context.Context,
	tenantID, runID string,
	savedRaw []*raw.SavedRecord,
	_ Connector,
) ([]*StagingRecord, error) {
	tx, err := n.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
INSERT INTO erp_staging_records
  (staging_id, tenant_id, run_id, raw_id, entity_type, source_id,
   normalized_json, validation_status, validation_errors, normalized_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	now := time.Now().UTC()
	results := make([]*StagingRecord, 0, len(savedRaw))

	for _, sr := range savedRaw {
		rec := sr.Record
		stagingID := uuid.New().String()

		// v1: passthrough — validate that the payload is a JSON object
		var validationStatus ValidationStatus
		var validationErrors []string

		var probe map[string]interface{}
		if err := json.Unmarshal(rec.PayloadJSON, &probe); err != nil {
			validationStatus = ValidationStatusInvalid
			validationErrors = []string{"payload is not a valid JSON object: " + err.Error()}
		} else {
			validationStatus = ValidationStatusValid
		}

		// Encode validation_errors as a JSONB-compatible value (null or JSON array)
		var validationErrorsJSON interface{}
		if len(validationErrors) > 0 {
			errBytes, _ := json.Marshal(validationErrors)
			validationErrorsJSON = string(errBytes)
		}

		_, err := tx.ExecContext(ctx, q,
			stagingID,
			tenantID,
			runID,
			sr.RawID,
			string(rec.EntityType),
			rec.SourceID,
			rec.PayloadJSON,
			string(validationStatus),
			validationErrorsJSON,
			now,
		)
		if err != nil {
			return nil, err
		}

		stagingRec := &StagingRecord{
			StagingID:        stagingID,
			TenantID:         tenantID,
			RunID:            runID,
			RawID:            sr.RawID,
			EntityType:       types.EntityType(rec.EntityType),
			SourceID:         rec.SourceID,
			NormalizedJSON:   rec.PayloadJSON,
			ValidationStatus: validationStatus,
			NormalizedAt:     now,
		}
		if len(validationErrors) > 0 {
			stagingRec.ValidationErrors = validationErrors
		}
		results = append(results, stagingRec)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}
