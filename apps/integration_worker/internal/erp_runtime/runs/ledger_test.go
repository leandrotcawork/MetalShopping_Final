package runs

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

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
	if len(step.args) > 0 {
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
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "SET status = 'completed'", args: []any{"run-1", 7, 1, 2, 3}},
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
		scriptStep{kind: stepExec, query: "SET status = 'completed'", args: []any{"run-1", 7, 1, 2, 3}, rowsAffected: int64Ptr(0)},
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
