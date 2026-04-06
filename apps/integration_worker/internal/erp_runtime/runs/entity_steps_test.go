package runs

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"metalshopping/integration_worker/internal/erp_runtime/types"
)

func TestMarkEntityStepFailedKeepsBatchCheckpoint(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 4, 5, 21, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 5, 21, 10, 0, 0, time.UTC)
	cursor := "cursor:batch-3"

	db, state := newScriptedDB(t,
		// MarkStarted
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_run_entity_steps", args: []any{"run-1", "products"}},
		scriptStep{kind: stepCommit},
		// MarkBatch
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "SET batch_ordinal = $4", args: []any{"run-1", "products", "tenant-1", 3, &cursor}},
		scriptStep{kind: stepCommit},
		// MarkFailed
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "SET status = 'failed'", args: []any{"run-1", "products", "extract failed"}},
		scriptStep{kind: stepCommit},
		// Get
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "FROM erp_run_entity_steps",
			args:  []any{"run-1", "products"},
			rows: [][]driver.Value{{
				"step-1",
				"tenant-1",
				"run-1",
				"products",
				"failed",
				int64(3),
				cursor,
				"extract failed",
				startedAt,
				completedAt,
			}},
		},
		scriptStep{kind: stepCommit},
	)

	store := NewEntityStepStore(db)
	if err := store.MarkStarted(context.Background(), "tenant-1", "run-1", types.EntityTypeProducts); err != nil {
		t.Fatalf("MarkStarted returned error: %v", err)
	}
	if err := store.MarkBatch(context.Background(), "tenant-1", "run-1", types.EntityTypeProducts, 3, &cursor); err != nil {
		t.Fatalf("MarkBatch returned error: %v", err)
	}
	if err := store.MarkFailed(context.Background(), "tenant-1", "run-1", types.EntityTypeProducts, "extract failed"); err != nil {
		t.Fatalf("MarkFailed returned error: %v", err)
	}

	step, err := store.Get(context.Background(), "tenant-1", "run-1", types.EntityTypeProducts)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if step.BatchOrdinal != 3 {
		t.Fatalf("BatchOrdinal = %d, want 3", step.BatchOrdinal)
	}
	if step.SourceCursor == nil || *step.SourceCursor != cursor {
		t.Fatalf("SourceCursor = %v, want %q", step.SourceCursor, cursor)
	}
	if step.Status != EntityStepStatusFailed {
		t.Fatalf("Status = %q, want %q", step.Status, EntityStepStatusFailed)
	}

	state.done()
}
