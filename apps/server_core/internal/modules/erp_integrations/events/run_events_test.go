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

func TestNewEntityPromotedOutboxRecordIncludesRunProvenance(t *testing.T) {
	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_123",
		TenantID:         "tenant_1",
		RunID:            "run_123",
		EntityType:       domain.EntityTypeProducts,
		SourceID:         "src_123",
		CanonicalID:      stringPtr("prd_123"),
		Action:           "create",
		ReconciledAt:     time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC),
	}
	run := &domain.SyncRun{
		RunID:         "run_123",
		TenantID:      "tenant_1",
		InstanceID:    "inst_123",
		ConnectorType: domain.ConnectorTypeSankhya,
	}

	record, err := NewEntityPromotedOutboxRecord(result, run, "trace_1", time.Date(2026, 4, 2, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewEntityPromotedOutboxRecord error: %v", err)
	}

	var payload struct {
		ReconciliationID string `json:"reconciliation_id"`
		TenantID         string `json:"tenant_id"`
		InstanceID       string `json:"instance_id"`
		ConnectorType    string `json:"connector_type"`
		RunID            string `json:"run_id"`
	}
	if err := json.Unmarshal(record.PayloadJSON, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.ReconciliationID != result.ReconciliationID || payload.TenantID != result.TenantID {
		t.Fatalf("unexpected payload tenant/reconciliation: %+v", payload)
	}
	if payload.InstanceID != run.InstanceID {
		t.Fatalf("expected instance_id %s, got %s", run.InstanceID, payload.InstanceID)
	}
	if payload.ConnectorType != string(run.ConnectorType) {
		t.Fatalf("expected connector_type %s, got %s", string(run.ConnectorType), payload.ConnectorType)
	}
	if payload.RunID != run.RunID {
		t.Fatalf("expected run_id %s, got %s", run.RunID, payload.RunID)
	}
}

func stringPtr(value string) *string {
	return &value
}
