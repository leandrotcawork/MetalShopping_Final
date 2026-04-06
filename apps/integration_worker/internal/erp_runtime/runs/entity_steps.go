package runs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"metalshopping/integration_worker/internal/erp_runtime/tenantdb"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// EntityStepStatus is the lifecycle state of an entity step inside one run.
type EntityStepStatus string

const (
	EntityStepStatusRunning EntityStepStatus = "running"
	EntityStepStatusFailed  EntityStepStatus = "failed"
	EntityStepStatusDone    EntityStepStatus = "completed"
	EntityStepStatusSkipped EntityStepStatus = "skipped_due_to_dependency"
)

// EntityStep holds persisted checkpoint state for one run+entity pair.
type EntityStep struct {
	StepID         string
	TenantID       string
	RunID          string
	EntityType     types.EntityType
	Status         EntityStepStatus
	BatchOrdinal   int
	SourceCursor   *string
	FailureSummary *string
	StartedAt      *time.Time
	CompletedAt    *time.Time
}

// EntityStepStore persists entity checkpoints in erp_run_entity_steps.
type EntityStepStore struct {
	db *sql.DB
}

func NewEntityStepStore(db *sql.DB) *EntityStepStore {
	return &EntityStepStore{db: db}
}

func (s *EntityStepStore) MarkStarted(ctx context.Context, tenantID, runID string, entity types.EntityType) error {
	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
INSERT INTO erp_run_entity_steps
  (step_id, tenant_id, run_id, entity_type, status, batch_ordinal, source_cursor, failure_summary, started_at, completed_at)
VALUES (gen_random_uuid()::text, current_tenant_id(), $1, $2, 'running', 0, NULL, NULL, NOW(), NULL)
ON CONFLICT (tenant_id, run_id, entity_type) DO UPDATE
SET status = 'running',
    started_at = COALESCE(erp_run_entity_steps.started_at, NOW()),
    completed_at = NULL`

	if _, err := tx.ExecContext(ctx, q, runID, string(entity)); err != nil {
		return fmt.Errorf("mark entity step started: %w", err)
	}
	return tx.Commit()
}

func (s *EntityStepStore) MarkBatch(ctx context.Context, tenantID, runID string, entity types.EntityType, batchOrdinal int, cursor *string) error {
	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if batchOrdinal <= 0 {
		batchOrdinal = 1
	}

	const q = `
UPDATE erp_run_entity_steps
SET batch_ordinal = $4,
    source_cursor = $5
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
  AND entity_type = $2
  AND status IN ('running', 'failed', 'completed', 'skipped_due_to_dependency')`
	if _, err := tx.ExecContext(ctx, q, runID, string(entity), tenantID, batchOrdinal, cursor); err != nil {
		return fmt.Errorf("mark entity batch: %w", err)
	}
	return tx.Commit()
}

func (s *EntityStepStore) MarkFailed(ctx context.Context, tenantID, runID string, entity types.EntityType, failure string) error {
	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
UPDATE erp_run_entity_steps
SET status = 'failed',
    failure_summary = $3,
    completed_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
  AND entity_type = $2`
	if _, err := tx.ExecContext(ctx, q, runID, string(entity), failure); err != nil {
		return fmt.Errorf("mark entity failed: %w", err)
	}
	return tx.Commit()
}

func (s *EntityStepStore) MarkCompleted(ctx context.Context, tenantID, runID string, entity types.EntityType) error {
	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
UPDATE erp_run_entity_steps
SET status = 'completed',
    failure_summary = NULL,
    completed_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
  AND entity_type = $2`
	if _, err := tx.ExecContext(ctx, q, runID, string(entity)); err != nil {
		return fmt.Errorf("mark entity completed: %w", err)
	}
	return tx.Commit()
}

func (s *EntityStepStore) MarkSkipped(ctx context.Context, tenantID, runID string, entity types.EntityType, reason string) error {
	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
UPDATE erp_run_entity_steps
SET status = 'skipped_due_to_dependency',
    failure_summary = $3,
    completed_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
  AND entity_type = $2`
	if _, err := tx.ExecContext(ctx, q, runID, string(entity), reason); err != nil {
		return fmt.Errorf("mark entity skipped: %w", err)
	}
	return tx.Commit()
}

func (s *EntityStepStore) Get(ctx context.Context, tenantID, runID string, entity types.EntityType) (*EntityStep, error) {
	tx, err := tenantdb.BeginTenantTx(ctx, s.db, tenantID, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
SELECT step_id, tenant_id, run_id, entity_type, status, batch_ordinal, source_cursor, failure_summary, started_at, completed_at
FROM erp_run_entity_steps
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
  AND entity_type = $2`

	var (
		step           EntityStep
		entityType     string
		status         string
		sourceCursor   sql.NullString
		failureSummary sql.NullString
		startedAt      sql.NullTime
		completedAt    sql.NullTime
	)

	err = tx.QueryRowContext(ctx, q, runID, string(entity)).Scan(
		&step.StepID,
		&step.TenantID,
		&step.RunID,
		&entityType,
		&status,
		&step.BatchOrdinal,
		&sourceCursor,
		&failureSummary,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}
	step.EntityType = types.EntityType(entityType)
	step.Status = EntityStepStatus(status)
	if sourceCursor.Valid {
		step.SourceCursor = &sourceCursor.String
	}
	if failureSummary.Valid {
		step.FailureSummary = &failureSummary.String
	}
	if startedAt.Valid {
		started := startedAt.Time
		step.StartedAt = &started
	}
	if completedAt.Valid {
		completed := completedAt.Time
		step.CompletedAt = &completed
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &step, nil
}
