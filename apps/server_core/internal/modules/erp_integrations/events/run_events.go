package events

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

const (
	eventVersion = "v1"
)

// runRequestedPayload is the JSON payload for the run_requested event.
type runRequestedPayload struct {
	RunID         string   `json:"run_id"`
	TenantID      string   `json:"tenant_id"`
	InstanceID    string   `json:"instance_id"`
	ConnectorType string   `json:"connector_type"`
	RunMode       string   `json:"run_mode"`
	EntityScope   []string `json:"entity_scope"`
}

// runCompletedPayload is the JSON payload for the run_completed event.
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
	CreatedAt      string   `json:"created_at"`
}

// entityPromotedPayload is the JSON payload for the entity_promoted event.
type entityPromotedPayload struct {
	ReconciliationID string  `json:"reconciliation_id"`
	TenantID         string  `json:"tenant_id"`
	EntityType       string  `json:"entity_type"`
	SourceID         string  `json:"source_id"`
	CanonicalID      *string `json:"canonical_id,omitempty"`
	Action           string  `json:"action"`
	PromotedAt       string  `json:"promoted_at"`
}

// NewRunRequestedOutboxRecord creates an outbox record for the
// erp_integrations.run_requested event.
func NewRunRequestedOutboxRecord(run *domain.SyncRun, traceID string, now time.Time) (outbox.Record, error) {
	scope := make([]string, len(run.EntityScope))
	for i, e := range run.EntityScope {
		scope[i] = string(e)
	}

	p := runRequestedPayload{
		RunID:         run.RunID,
		TenantID:      run.TenantID,
		InstanceID:    run.InstanceID,
		ConnectorType: string(run.ConnectorType),
		RunMode:       string(run.RunMode),
		EntityScope:   scope,
	}
	payloadJSON, err := json.Marshal(p)
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal erp run_requested payload: %w", err)
	}

	return outbox.Record{
		EventID:        generateEventID(),
		AggregateType:  "erp_sync_run",
		AggregateID:    run.RunID,
		EventName:      "erp_integrations.run_requested",
		EventVersion:   eventVersion,
		TenantID:       run.TenantID,
		TraceID:        traceID,
		IdempotencyKey: "erp_run_requested:" + run.RunID,
		PayloadJSON:    payloadJSON,
		Status:         outbox.StatusPending,
		Attempts:       0,
		AvailableAt:    now,
		CreatedAt:      now,
	}, nil
}

// NewRunCompletedOutboxRecord creates an outbox record for the
// erp_integrations.run_completed event.
func NewRunCompletedOutboxRecord(run *domain.SyncRun, traceID string, now time.Time) (outbox.Record, error) {
	scope := make([]string, len(run.EntityScope))
	for i, e := range run.EntityScope {
		scope[i] = string(e)
	}

	p := runCompletedPayload{
		RunID:         run.RunID,
		TenantID:      run.TenantID,
		InstanceID:    run.InstanceID,
		ConnectorType: string(run.ConnectorType),
		RunMode:       string(run.RunMode),
		EntityScope:   scope,
		Status:        string(run.Status),
		PromotedCount: run.PromotedCount,
		WarningCount:  run.WarningCount,
		RejectedCount: run.RejectedCount,
		ReviewCount:   run.ReviewCount,
	}
	if run.StartedAt != nil {
		s := run.StartedAt.UTC().Format(time.RFC3339)
		p.StartedAt = &s
	}
	if run.CompletedAt != nil {
		s := run.CompletedAt.UTC().Format(time.RFC3339)
		p.CompletedAt = &s
	}
	p.FailureSummary = run.FailureSummary
	p.CreatedAt = run.CreatedAt.UTC().Format(time.RFC3339)

	payloadJSON, err := json.Marshal(p)
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal erp run_completed payload: %w", err)
	}

	return outbox.Record{
		EventID:        generateEventID(),
		AggregateType:  "erp_sync_run",
		AggregateID:    run.RunID,
		EventName:      "erp_integrations.run_completed",
		EventVersion:   eventVersion,
		TenantID:       run.TenantID,
		TraceID:        traceID,
		IdempotencyKey: "erp_run_completed:" + run.RunID,
		PayloadJSON:    payloadJSON,
		Status:         outbox.StatusPending,
		Attempts:       0,
		AvailableAt:    now,
		CreatedAt:      now,
	}, nil
}

// NewEntityPromotedOutboxRecord creates an outbox record for the
// erp_integrations.entity_promoted event.
func NewEntityPromotedOutboxRecord(result *domain.ReconciliationResult, traceID string, now time.Time) (outbox.Record, error) {
	p := entityPromotedPayload{
		ReconciliationID: result.ReconciliationID,
		TenantID:         result.TenantID,
		EntityType:       string(result.EntityType),
		SourceID:         result.SourceID,
		CanonicalID:      result.CanonicalID,
		Action:           result.Action,
		PromotedAt:       result.ReconciledAt.UTC().Format(time.RFC3339),
	}
	payloadJSON, err := json.Marshal(p)
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal erp entity_promoted payload: %w", err)
	}

	return outbox.Record{
		EventID:        generateEventID(),
		AggregateType:  "erp_reconciliation_result",
		AggregateID:    result.ReconciliationID,
		EventName:      "erp_integrations.entity_promoted",
		EventVersion:   eventVersion,
		TenantID:       result.TenantID,
		TraceID:        traceID,
		IdempotencyKey: "erp_entity_promoted:" + result.ReconciliationID,
		PayloadJSON:    payloadJSON,
		Status:         outbox.StatusPending,
		Attempts:       0,
		AvailableAt:    now,
		CreatedAt:      now,
	}, nil
}

func generateEventID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "evt_fallback"
	}
	return "evt_" + hex.EncodeToString(buf)
}
