package raw

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"testing"

	"metalshopping/integration_worker/internal/erp_runtime/types"
)

type rawStepKind string

const (
	rawStepBegin    rawStepKind = "begin"
	rawStepExec     rawStepKind = "exec"
	rawStepCommit   rawStepKind = "commit"
	rawStepRollback rawStepKind = "rollback"
)

type rawStep struct {
	kind   rawStepKind
	query  string
	assert func(*testing.T, string, []driver.NamedValue)
}

type rawState struct {
	t     *testing.T
	mu    sync.Mutex
	steps []rawStep
	pos   int
}

func newRawScriptedDB(t *testing.T, steps ...rawStep) (*sql.DB, *rawState) {
	t.Helper()
	state := &rawState{t: t, steps: steps}
	db := sql.OpenDB(&rawConnector{state: state})
	t.Cleanup(func() { _ = db.Close() })
	return db, state
}

type rawConnector struct {
	state *rawState
}

func (c *rawConnector) Connect(context.Context) (driver.Conn, error) {
	return &rawConn{state: c.state}, nil
}

func (c *rawConnector) Driver() driver.Driver { return rawDriver{} }

type rawDriver struct{}

func (rawDriver) Open(string) (driver.Conn, error) {
	return nil, fmt.Errorf("open not supported")
}

type rawConn struct {
	state *rawState
}

func (c *rawConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported")
}
func (c *rawConn) Close() error { return nil }
func (c *rawConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *rawConn) CheckNamedValue(*driver.NamedValue) error {
	return nil
}

func (c *rawConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.state.expect(rawStepBegin, "", nil)
	return &rawTx{state: c.state}, nil
}

func (c *rawConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.expect(rawStepExec, query, args)
	return driver.RowsAffected(1), nil
}

type rawTx struct {
	state *rawState
}

func (tx *rawTx) Commit() error {
	tx.state.expect(rawStepCommit, "", nil)
	return nil
}

func (tx *rawTx) Rollback() error {
	tx.state.expect(rawStepRollback, "", nil)
	return nil
}

func (s *rawState) expect(kind rawStepKind, query string, args []driver.NamedValue) {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos >= len(s.steps) {
		s.t.Fatalf("unexpected %s operation: %q", kind, query)
	}
	step := s.steps[s.pos]
	if step.kind != kind {
		s.t.Fatalf("step %d: expected %s, got %s", s.pos, step.kind, kind)
	}
	if step.query != "" && !strings.Contains(query, step.query) {
		s.t.Fatalf("step %d: expected query containing %q, got %q", s.pos, step.query, query)
	}
	if step.assert != nil {
		step.assert(s.t, query, args)
	}
	s.pos++
}

func (s *rawState) done() {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pos != len(s.steps) {
		s.t.Fatalf("expected %d steps, consumed %d", len(s.steps), s.pos)
	}
}

func TestRawStorePersistsBatchOrdinal(t *testing.T) {
	t.Parallel()

	db, state := newRawScriptedDB(t,
		rawStep{kind: rawStepBegin},
		rawStep{kind: rawStepExec, query: "set_config('app.tenant_id'"},
		rawStep{
			kind:  rawStepExec,
			query: "INSERT INTO erp_raw_records",
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				if len(args) < 11 {
					t.Fatalf("expected at least 11 args, got %d", len(args))
				}
				if got := fmt.Sprint(args[7].Value); got != "3" {
					t.Fatalf("expected batch_ordinal arg to be 3, got %v", args[7].Value)
				}
			},
		},
		rawStep{kind: rawStepCommit},
	)

	store := NewStore(db)
	_, err := store.Save(context.Background(), "tenant-1", "run-1", []*types.RawRecord{
		{
			SourceID:      "1001",
			ConnectorType: "sankhya",
			EntityType:    types.EntityTypeProducts,
			PayloadJSON:   []byte(`{"CODPROD":"1001"}`),
			PayloadHash:   "hash-1",
			BatchOrdinal:  3,
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	state.done()
}
