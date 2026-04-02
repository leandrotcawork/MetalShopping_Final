package events

import (
	"encoding/json"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
)

func TestNewRunCompletedOutboxRecordIncludesCreatedAt(t *testing.T) {
	createdAt := time.Date(2026, 4, 2, 11, 55, 0, 0, time.FixedZone("BRT", -3*60*60))
	run := &domain.SyncRun{
		RunID:         "run_123",
		TenantID:      "tenant_1",
		InstanceID:    "inst_1",
		ConnectorType: domain.ConnectorTypeSankhya,
		RunMode:       domain.RunModeBulk,
		EntityScope:   []domain.EntityType{domain.EntityTypeProducts},
		Status:        domain.RunStatusCompleted,
		PromotedCount: 7,
		WarningCount:  1,
		RejectedCount: 2,
		ReviewCount:   3,
		CreatedAt:     createdAt,
	}

	record, err := NewRunCompletedOutboxRecord(run, "trace_1", time.Date(2026, 4, 2, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewRunCompletedOutboxRecord error: %v", err)
	}
	if got, want := record.IdempotencyKey, "erp_run_completed:"+run.RunID; got != want {
		t.Fatalf("expected idempotency key %s, got %s", want, got)
	}

	var payload struct {
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(record.PayloadJSON, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got, want := payload.CreatedAt, createdAt.UTC().Format(time.RFC3339); got != want {
		t.Fatalf("expected created_at %s, got %s", want, got)
	}
}
