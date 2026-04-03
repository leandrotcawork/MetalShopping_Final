package runs

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"metalshopping/integration_worker/internal/erp_runtime/tenantdb"
)

type scriptStepKind string

const (
	stepBegin    scriptStepKind = "begin"
	stepExec     scriptStepKind = "exec"
	stepQuery    scriptStepKind = "query"
	stepCommit   scriptStepKind = "commit"
	stepRollback scriptStepKind = "rollback"
)

type scriptStep struct {
	kind         scriptStepKind
	query        string
	args         []any
	rows         [][]driver.Value
	rowsAffected *int64
	assert       func(*testing.T, string, []driver.NamedValue)
}

type scriptState struct {
	t     *testing.T
	mu    sync.Mutex
	steps []scriptStep
	pos   int
}

func newScriptedDB(t *testing.T, steps ...scriptStep) (*sql.DB, *scriptState) {
	t.Helper()

	state := &scriptState{t: t, steps: steps}
	db := sql.OpenDB(&scriptConnector{state: state})
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db, state
}

type scriptConnector struct {
	state *scriptState
}

func (c *scriptConnector) Connect(context.Context) (driver.Conn, error) {
	return &scriptConn{state: c.state}, nil
}

func (c *scriptConnector) Driver() driver.Driver {
	return scriptDriver{}
}

type scriptDriver struct{}

func (scriptDriver) Open(string) (driver.Conn, error) {
	return nil, fmt.Errorf("open not supported")
}

type scriptConn struct {
	state *scriptState
}

func (c *scriptConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported")
}

func (c *scriptConn) Close() error { return nil }

func (c *scriptConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *scriptConn) Ping(context.Context) error { return nil }

func (c *scriptConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (c *scriptConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	c.state.expect(stepBegin, "", nil)
	return &scriptTx{state: c.state}, nil
}

func (c *scriptConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	step := c.state.expect(stepExec, query, args)
	rowsAffected := int64(1)
	if step.rowsAffected != nil {
		rowsAffected = *step.rowsAffected
	}
	return driver.RowsAffected(rowsAffected), nil
}

func (c *scriptConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	step := c.state.expect(stepQuery, query, args)
	return &scriptRows{rows: step.rows}, nil
}

type scriptTx struct {
	state *scriptState
}

func (tx *scriptTx) Commit() error {
	tx.state.expect(stepCommit, "", nil)
	return nil
}

func (tx *scriptTx) Rollback() error {
	tx.state.expect(stepRollback, "", nil)
	return nil
}

type scriptRows struct {
	rows [][]driver.Value
	pos  int
}

func (r *scriptRows) Columns() []string {
	if len(r.rows) == 0 {
		return nil
	}
	cols := make([]string, len(r.rows[0]))
	for i := range cols {
		cols[i] = fmt.Sprintf("col_%d", i)
	}
	return cols
}

func (r *scriptRows) Close() error { return nil }

func (r *scriptRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

func (s *scriptState) expect(kind scriptStepKind, query string, args []driver.NamedValue) scriptStep {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos >= len(s.steps) {
		s.t.Fatalf("unexpected %s operation with query %q", kind, query)
	}

	step := s.steps[s.pos]
	if step.kind != kind {
		s.t.Fatalf("step %d: expected %s, got %s", s.pos, step.kind, kind)
	}
	if step.query != "" && !strings.Contains(query, step.query) {
		s.t.Fatalf("step %d: expected query to contain %q, got %q", s.pos, step.query, query)
	}
	if step.assert != nil {
		step.assert(s.t, query, args)
	} else if len(step.args) > 0 {
		if len(args) != len(step.args) {
			s.t.Fatalf("step %d: expected %d args, got %d", s.pos, len(step.args), len(args))
		}
		for i := range step.args {
			if fmt.Sprint(args[i].Value) != fmt.Sprint(step.args[i]) {
				s.t.Fatalf("step %d: arg %d expected %v, got %v", s.pos, i, step.args[i], args[i].Value)
			}
		}
	}

	s.pos++
	return step
}

func (s *scriptState) done() {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pos != len(s.steps) {
		s.t.Fatalf("expected %d scripted operations, consumed %d", len(s.steps), s.pos)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestBeginTenantTxSetsTenantSession(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepRollback},
	)

	tx, err := tenantdb.BeginTenantTx(context.Background(), db, "tenant-1", nil)
	if err != nil {
		t.Fatalf("BeginTenantTx error: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback error: %v", err)
	}
	state.done()
}

func TestClaimPendingRunSetsTenantContextBeforeUpdate(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{
			kind:  stepBegin,
			query: "",
		},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"*"}},
		scriptStep{
			kind:  stepQuery,
			query: "FROM erp_sync_runs",
			rows: [][]driver.Value{{
				"run-1",
				"tenant-1",
				"instance-1",
				"sankhya",
				"bulk",
				"{products,prices}",
			}},
		},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "SET status = 'running'", args: []any{"run-1"}},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	claim, err := ledger.ClaimPendingRun(context.Background())
	if err != nil {
		t.Fatalf("ClaimPendingRun error: %v", err)
	}
	if claim == nil {
		t.Fatal("expected claim, got nil")
	}
	if claim.RunID != "run-1" || claim.TenantID != "tenant-1" || claim.InstanceID != "instance-1" {
		t.Fatalf("unexpected claim: %+v", claim)
	}
	if got, want := strings.Join(claim.EntityScope, ","), "products,prices"; got != want {
		t.Fatalf("unexpected entity scope: got %q want %q", got, want)
	}
	state.done()
}

func TestMarkCompletedUsesRunTenantContext(t *testing.T) {
	startedAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 2, 12, 5, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 2, 11, 55, 0, 0, time.UTC)

	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "RETURNING run_id, tenant_id, instance_id",
			args:  []any{"run-1", 7, 1, 2, 3},
			rows: [][]driver.Value{{
				"run-1",
				"tenant-1",
				"instance-1",
				"sankhya",
				"bulk",
				"{products,prices}",
				"completed",
				startedAt,
				completedAt,
				int64(7),
				int64(1),
				int64(2),
				int64(3),
				nil,
				createdAt,
			}},
		},
		scriptStep{
			kind:  stepExec,
			query: "INSERT INTO outbox_events",
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				assertRunCompletedOutboxArgs(t, args, "run-1", "tenant-1", createdAt)
			},
		},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	if err := ledger.MarkCompleted(tenantCtx, "run-1", 7, 1, 2, 3); err != nil {
		t.Fatalf("MarkCompleted error: %v", err)
	}
	state.done()
}

func TestSaveCursorUsesRunTenantContext(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "SET cursor_state = $2::jsonb", args: []any{"run-1", `{"cursor":"abc"}`}},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	if err := ledger.SaveCursor(tenantCtx, "run-1", `{"cursor":"abc"}`); err != nil {
		t.Fatalf("SaveCursor error: %v", err)
	}
	state.done()
}

func TestMarkCompletedReturnsErrorWhenNoRowsUpdated(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "RETURNING run_id, tenant_id, instance_id",
			args:  []any{"run-1", 7, 1, 2, 3},
			rows:  nil,
		},
		scriptStep{kind: stepRollback},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	err = ledger.MarkCompleted(tenantCtx, "run-1", 7, 1, 2, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no rows updated") {
		t.Fatalf("expected no rows updated error, got %v", err)
	}
	state.done()
}

func TestSaveCursorReturnsErrorForTenantMismatch(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-2"}},
		scriptStep{kind: stepExec, query: "SET cursor_state = $2::jsonb", args: []any{"run-1", `{"cursor":"abc"}`}, rowsAffected: int64Ptr(0)},
		scriptStep{kind: stepRollback},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-2")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	err = ledger.SaveCursor(tenantCtx, "run-1", `{"cursor":"abc"}`)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no rows updated") {
		t.Fatalf("expected no rows updated error, got %v", err)
	}
	state.done()
}

func TestMarkCompletedAppendsOutboxInSameTransaction(t *testing.T) {
	startedAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 2, 12, 5, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 2, 11, 55, 0, 0, time.UTC)

	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "RETURNING run_id, tenant_id, instance_id",
			args:  []any{"run-1", 7, 1, 2, 3},
			rows: [][]driver.Value{{
				"run-1",
				"tenant-1",
				"instance-1",
				"sankhya",
				"bulk",
				"{products,prices}",
				"completed",
				startedAt,
				completedAt,
				int64(7),
				int64(1),
				int64(2),
				int64(3),
				nil,
				createdAt,
			}},
		},
		scriptStep{
			kind:  stepExec,
			query: "INSERT INTO outbox_events",
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				assertRunCompletedOutboxArgs(t, args, "run-1", "tenant-1", createdAt)
			},
		},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	if err := ledger.MarkCompleted(tenantCtx, "run-1", 7, 1, 2, 3); err != nil {
		t.Fatalf("MarkCompleted error: %v", err)
	}
	state.done()
}

func TestMarkPartialAppendsOutboxInSameTransaction(t *testing.T) {
	startedAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 2, 12, 6, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 2, 11, 55, 0, 0, time.UTC)

	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "RETURNING run_id, tenant_id, instance_id",
			args:  []any{"run-1", "normalize failed", 5, 2, 1, 3},
			rows: [][]driver.Value{{
				"run-1",
				"tenant-1",
				"instance-1",
				"sankhya",
				"bulk",
				"{products}",
				"partial",
				startedAt,
				completedAt,
				int64(5),
				int64(2),
				int64(1),
				int64(3),
				"normalize failed",
				createdAt,
			}},
		},
		scriptStep{
			kind:  stepExec,
			query: "INSERT INTO outbox_events",
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				assertRunCompletedOutboxArgs(t, args, "run-1", "tenant-1", createdAt)
			},
		},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	if err := ledger.MarkPartial(tenantCtx, "run-1", "normalize failed", 5, 2, 1, 3); err != nil {
		t.Fatalf("MarkPartial error: %v", err)
	}
	state.done()
}

func TestMarkFailedAppendsOutboxInSameTransaction(t *testing.T) {
	startedAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 2, 12, 7, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 2, 11, 55, 0, 0, time.UTC)

	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "RETURNING run_id, tenant_id, instance_id",
			args:  []any{"run-1", "connector timeout"},
			rows: [][]driver.Value{{
				"run-1",
				"tenant-1",
				"instance-1",
				"sankhya",
				"bulk",
				"{products}",
				"failed",
				startedAt,
				completedAt,
				int64(0),
				int64(0),
				int64(0),
				int64(0),
				"connector timeout",
				createdAt,
			}},
		},
		scriptStep{
			kind:  stepExec,
			query: "INSERT INTO outbox_events",
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				assertRunCompletedOutboxArgs(t, args, "run-1", "tenant-1", createdAt)
			},
		},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	if err := ledger.MarkFailed(tenantCtx, "run-1", "connector timeout"); err != nil {
		t.Fatalf("MarkFailed error: %v", err)
	}
	state.done()
}

func TestMarkCompletedIgnoresDuplicateRunCompletedOutbox(t *testing.T) {
	startedAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 2, 12, 5, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 2, 11, 55, 0, 0, time.UTC)

	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{
			kind:  stepQuery,
			query: "RETURNING run_id, tenant_id, instance_id",
			args:  []any{"run-1", 7, 1, 2, 3},
			rows: [][]driver.Value{{
				"run-1",
				"tenant-1",
				"instance-1",
				"sankhya",
				"bulk",
				"{products,prices}",
				"completed",
				startedAt,
				completedAt,
				int64(7),
				int64(1),
				int64(2),
				int64(3),
				nil,
				createdAt,
			}},
		},
		scriptStep{
			kind:         stepExec,
			query:        "ON CONFLICT (idempotency_key) DO NOTHING",
			rowsAffected: int64Ptr(0),
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				assertRunCompletedOutboxArgs(t, args, "run-1", "tenant-1", createdAt)
			},
		},
		scriptStep{kind: stepCommit},
	)

	ledger := NewLedger(db)
	tenantCtx, err := tenantdb.WithTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("WithTenantID error: %v", err)
	}
	if err := ledger.MarkCompleted(tenantCtx, "run-1", 7, 1, 2, 3); err != nil {
		t.Fatalf("MarkCompleted error: %v", err)
	}
	state.done()
}

func assertRunCompletedOutboxArgs(t *testing.T, args []driver.NamedValue, runID, tenantID string, createdAt time.Time) {
	t.Helper()
	if len(args) != 15 {
		t.Fatalf("expected 15 outbox args, got %d", len(args))
	}
	if got := fmt.Sprint(args[2].Value); got != runID {
		t.Fatalf("expected aggregate_id %s, got %s", runID, got)
	}
	if got := fmt.Sprint(args[3].Value); got != "erp_integrations.run_completed" {
		t.Fatalf("expected run_completed event, got %s", got)
	}
	if got := fmt.Sprint(args[5].Value); got != tenantID {
		t.Fatalf("expected tenant %s, got %s", tenantID, got)
	}
	wantKey := "erp_run_completed:" + runID
	if got := fmt.Sprint(args[7].Value); got != wantKey {
		t.Fatalf("expected idempotency key %s, got %s", wantKey, got)
	}
	var payload struct {
		CreatedAt string `json:"created_at"`
	}
	payloadBytes, ok := args[8].Value.([]byte)
	if !ok {
		t.Fatalf("expected payload_json to be []byte, got %T", args[8].Value)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("unmarshal payload_json: %v", err)
	}
	if got, want := payload.CreatedAt, createdAt.UTC().Format(time.RFC3339); got != want {
		t.Fatalf("expected payload created_at %s, got %s", want, got)
	}
}
