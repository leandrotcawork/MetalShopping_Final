package raw

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// Store persists raw ERP records to erp_raw_records.
type Store struct {
	db *sql.DB
}

// NewStore constructs a Store backed by the given DB connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// SavedRecord pairs a RawRecord with its assigned raw_id after persistence.
type SavedRecord struct {
	RawID  string
	Record *types.RawRecord
}

// Save inserts a batch of RawRecord rows into erp_raw_records.
// All records belong to the same run and tenant.
// Returns the list of persisted records with their assigned raw_ids.
// Note: NO current_tenant_id() RLS here — worker uses a direct DB connection
// and must set tenant_id explicitly in the INSERT.
// This is intentional: the worker runs as a system process, not within a tenant HTTP request.
func (s *Store) Save(ctx context.Context, tenantID, runID string, records []*types.RawRecord) ([]*SavedRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
INSERT INTO erp_raw_records
  (raw_id, tenant_id, run_id, connector_type, entity_type, source_id,
   payload_json, payload_hash, source_timestamp, cursor_value, extracted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (raw_id) DO NOTHING`

	extractedAt := time.Now().UTC()
	saved := make([]*SavedRecord, 0, len(records))
	for _, rec := range records {
		rawID := uuid.New().String()
		_, err := tx.ExecContext(ctx, q,
			rawID,
			tenantID,
			runID,
			rec.ConnectorType,
			string(rec.EntityType),
			rec.SourceID,
			rec.PayloadJSON,
			rec.PayloadHash,
			rec.SourceTimestamp,
			rec.CursorValue,
			extractedAt,
		)
		if err != nil {
			return nil, err
		}
		saved = append(saved, &SavedRecord{RawID: rawID, Record: rec})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return saved, nil
}
