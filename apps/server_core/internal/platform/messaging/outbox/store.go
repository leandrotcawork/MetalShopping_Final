package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Append opens its own database transaction and appends the given records.
// Use AppendInTx when you already hold a transaction.
func (s *Store) Append(ctx context.Context, records []Record) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin outbox append tx: %w", err)
	}
	if err := s.AppendInTx(ctx, tx, records); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Store) AppendInTx(ctx context.Context, tx *sql.Tx, records []Record) error {
	if len(records) == 0 {
		return nil
	}

	const insertSQL = `
INSERT INTO outbox_events (
  event_id,
  aggregate_type,
  aggregate_id,
  event_name,
  event_version,
  tenant_id,
  trace_id,
  idempotency_key,
  payload_json,
  status,
  attempts,
  available_at,
  created_at,
  published_at,
  last_error
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

	for _, record := range records {
		if err := record.ValidateForAppend(); err != nil {
			return err
		}
		status := record.Status
		if status == "" {
			status = StatusPending
		}
		if _, err := tx.ExecContext(
			ctx,
			insertSQL,
			record.EventID,
			record.AggregateType,
			record.AggregateID,
			record.EventName,
			record.EventVersion,
			nullableText(record.TenantID),
			nullableText(record.TraceID),
			record.IdempotencyKey,
			record.PayloadJSON,
			string(status),
			record.Attempts,
			record.AvailableAt,
			record.CreatedAt,
			record.PublishedAt,
			nullableText(record.LastError),
		); err != nil {
			return fmt.Errorf("append outbox event: %w", err)
		}
	}

	return nil
}

func (s *Store) ListPending(ctx context.Context, limit int) ([]Record, error) {
	const querySQL = `
SELECT event_id, aggregate_type, aggregate_id, event_name, event_version, COALESCE(tenant_id, ''), COALESCE(trace_id, ''), idempotency_key, payload_json, status, attempts, available_at, created_at, published_at, COALESCE(last_error, '')
FROM outbox_events
WHERE status IN ('pending', 'failed')
  AND available_at <= NOW()
ORDER BY created_at ASC
LIMIT $1
`
	rows, err := s.db.QueryContext(ctx, querySQL, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending outbox events: %w", err)
	}
	defer rows.Close()

	records := make([]Record, 0, limit)
	for rows.Next() {
		var record Record
		var status string
		if err := rows.Scan(
			&record.EventID,
			&record.AggregateType,
			&record.AggregateID,
			&record.EventName,
			&record.EventVersion,
			&record.TenantID,
			&record.TraceID,
			&record.IdempotencyKey,
			&record.PayloadJSON,
			&status,
			&record.Attempts,
			&record.AvailableAt,
			&record.CreatedAt,
			&record.PublishedAt,
			&record.LastError,
		); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		record.Status = Status(status)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox events: %w", err)
	}

	return records, nil
}

func (s *Store) MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	const updateSQL = `
UPDATE outbox_events
SET status = 'published',
    attempts = attempts + 1,
    published_at = $2,
    last_error = NULL
WHERE event_id = $1
`
	if _, err := s.db.ExecContext(ctx, updateSQL, eventID, publishedAt); err != nil {
		return fmt.Errorf("mark outbox event published: %w", err)
	}
	return nil
}

func (s *Store) MarkFailed(ctx context.Context, eventID string, errText string, availableAt time.Time) error {
	const updateSQL = `
UPDATE outbox_events
SET status = 'failed',
    attempts = attempts + 1,
    available_at = $2,
    last_error = $3
WHERE event_id = $1
`
	if _, err := s.db.ExecContext(ctx, updateSQL, eventID, availableAt, errText); err != nil {
		return fmt.Errorf("mark outbox event failed: %w", err)
	}
	return nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
