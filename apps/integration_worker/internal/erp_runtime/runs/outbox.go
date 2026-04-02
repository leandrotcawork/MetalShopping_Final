package runs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const runEventVersion = "v1"

type runSnapshot struct {
	RunID          string
	TenantID       string
	InstanceID     string
	ConnectorType  string
	RunMode        string
	EntityScope    []string
	Status         string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	PromotedCount  int
	WarningCount   int
	RejectedCount  int
	ReviewCount    int
	FailureSummary *string
	CreatedAt      time.Time
}

type runCompletedPayload struct {
	RunID          string   `json:"run_id"`
	TenantID       string   `json:"tenant_id"`
	InstanceID     string   `json:"instance_id"`
	ConnectorType  string   `json:"connector_type"`
	RunMode        string   `json:"run_mode"`
	EntityScope    []string `json:"entity_scope"`
	Status         string   `json:"status"`
	PromotedCount  int      `json:"promoted_count"`
	WarningCount   int      `json:"warning_count"`
	RejectedCount  int      `json:"rejected_count"`
	ReviewCount    int      `json:"review_count"`
	StartedAt      *string  `json:"started_at,omitempty"`
	CompletedAt    *string  `json:"completed_at,omitempty"`
	FailureSummary *string  `json:"failure_summary,omitempty"`
}

type runScanner interface {
	Scan(dest ...any) error
}

func scanRunSnapshot(s runScanner) (*runSnapshot, error) {
	var run runSnapshot
	var entityScopeRaw string
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var failureSummary sql.NullString

	if err := s.Scan(
		&run.RunID,
		&run.TenantID,
		&run.InstanceID,
		&run.ConnectorType,
		&run.RunMode,
		&entityScopeRaw,
		&run.Status,
		&startedAt,
		&completedAt,
		&run.PromotedCount,
		&run.WarningCount,
		&run.RejectedCount,
		&run.ReviewCount,
		&failureSummary,
		&run.CreatedAt,
	); err != nil {
		return nil, err
	}

	run.EntityScope = parsePGTextArray(entityScopeRaw)
	if startedAt.Valid {
		value := startedAt.Time.UTC()
		run.StartedAt = &value
	}
	if completedAt.Valid {
		value := completedAt.Time.UTC()
		run.CompletedAt = &value
	}
	if failureSummary.Valid {
		value := failureSummary.String
		run.FailureSummary = &value
	}

	return &run, nil
}

func appendRunCompletedOutbox(ctx context.Context, tx *sql.Tx, run *runSnapshot) error {
	record, err := newRunCompletedOutboxRecord(run)
	if err != nil {
		return err
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
		record.Status,
		record.Attempts,
		record.AvailableAt,
		record.CreatedAt,
		nullableTime(record.PublishedAt),
		nullableText(record.LastError),
	); err != nil {
		return fmt.Errorf("insert run completed outbox: %w", err)
	}

	return nil
}

type outboxRecord struct {
	EventID        string
	AggregateType  string
	AggregateID    string
	EventName      string
	EventVersion   string
	TenantID       string
	TraceID        string
	IdempotencyKey string
	PayloadJSON    []byte
	Status         string
	Attempts       int
	AvailableAt    time.Time
	CreatedAt      time.Time
	PublishedAt    *time.Time
	LastError      string
}

func newRunCompletedOutboxRecord(run *runSnapshot) (*outboxRecord, error) {
	payload := runCompletedPayload{
		RunID:         run.RunID,
		TenantID:      run.TenantID,
		InstanceID:    run.InstanceID,
		ConnectorType: run.ConnectorType,
		RunMode:       run.RunMode,
		EntityScope:   append([]string{}, run.EntityScope...),
		Status:        run.Status,
		PromotedCount: run.PromotedCount,
		WarningCount:  run.WarningCount,
		RejectedCount: run.RejectedCount,
		ReviewCount:   run.ReviewCount,
	}
	if run.StartedAt != nil {
		value := run.StartedAt.UTC().Format(time.RFC3339)
		payload.StartedAt = &value
	}
	if run.CompletedAt != nil {
		value := run.CompletedAt.UTC().Format(time.RFC3339)
		payload.CompletedAt = &value
	}
	payload.FailureSummary = run.FailureSummary

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal run completed payload: %w", err)
	}

	eventAt := run.CreatedAt.UTC()
	if run.CompletedAt != nil {
		eventAt = run.CompletedAt.UTC()
	}

	return &outboxRecord{
		EventID:        generateEventID(),
		AggregateType:  "erp_sync_run",
		AggregateID:    run.RunID,
		EventName:      "erp_integrations.run_completed",
		EventVersion:   runEventVersion,
		TenantID:       run.TenantID,
		IdempotencyKey: "erp_run_completed:" + run.RunID,
		PayloadJSON:    payloadJSON,
		Status:         "pending",
		Attempts:       0,
		AvailableAt:    eventAt,
		CreatedAt:      eventAt,
	}, nil
}

func generateEventID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "evt_fallback"
	}
	return "evt_" + hex.EncodeToString(buf)
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}
