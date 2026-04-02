package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

type repoScriptStepKind string

const (
	repoStepBegin    repoScriptStepKind = "begin"
	repoStepExec     repoScriptStepKind = "exec"
	repoStepCommit   repoScriptStepKind = "commit"
	repoStepRollback repoScriptStepKind = "rollback"
)

type repoScriptStep struct {
	kind         repoScriptStepKind
	query        string
	args         []any
	rowsAffected *int64
	assert       func(*testing.T, string, []driver.NamedValue)
}

type repoScriptState struct {
	t     *testing.T
	mu    sync.Mutex
	steps []repoScriptStep
	pos   int
}

func newRepoScriptedDB(t *testing.T, steps ...repoScriptStep) (*sql.DB, *repoScriptState) {
	t.Helper()

	state := &repoScriptState{t: t, steps: steps}
	db := sql.OpenDB(&repoScriptConnector{state: state})
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db, state
}

type repoScriptConnector struct {
	state *repoScriptState
}

func (c *repoScriptConnector) Connect(context.Context) (driver.Conn, error) {
	return &repoScriptConn{state: c.state}, nil
}

func (c *repoScriptConnector) Driver() driver.Driver {
	return repoScriptDriver{}
}

type repoScriptDriver struct{}

func (repoScriptDriver) Open(string) (driver.Conn, error) {
	return nil, fmt.Errorf("open not supported")
}

type repoScriptConn struct {
	state *repoScriptState
}

func (c *repoScriptConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported")
}

func (c *repoScriptConn) Close() error { return nil }

func (c *repoScriptConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *repoScriptConn) Ping(context.Context) error { return nil }

func (c *repoScriptConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (c *repoScriptConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.state.expect(repoStepBegin, "", nil)
	return &repoScriptTx{state: c.state}, nil
}

func (c *repoScriptConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	step := c.state.expect(repoStepExec, query, args)
	rowsAffected := int64(1)
	if step.rowsAffected != nil {
		rowsAffected = *step.rowsAffected
	}
	return driver.RowsAffected(rowsAffected), nil
}

type repoScriptTx struct {
	state *repoScriptState
}

func (tx *repoScriptTx) Commit() error {
	tx.state.expect(repoStepCommit, "", nil)
	return nil
}

func (tx *repoScriptTx) Rollback() error {
	tx.state.expect(repoStepRollback, "", nil)
	return nil
}

func (s *repoScriptState) expect(kind repoScriptStepKind, query string, args []driver.NamedValue) repoScriptStep {
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

func (s *repoScriptState) done() {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pos != len(s.steps) {
		s.t.Fatalf("expected %d scripted operations, consumed %d", len(s.steps), s.pos)
	}
}

func TestReconciliationRepoMarkReviewRequiredCreatesReviewItem(t *testing.T) {
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{
			kind:  repoStepExec,
			query: "INSERT INTO erp_review_items",
			assert: func(t *testing.T, query string, args []driver.NamedValue) {
				t.Helper()
				if !strings.Contains(query, "UPDATE erp_reconciliation_results") {
					t.Fatal("expected reconciliation update in review-required statement")
				}
				if len(args) != 8 {
					t.Fatalf("expected 8 args, got %d", len(args))
				}
				if got := fmt.Sprint(args[0].Value); got != "ERP_PROMOTION_AUTO_DISABLED" {
					t.Fatalf("expected auto-disabled reason code, got %s", got)
				}
				if got := fmt.Sprint(args[1].Value); got != `{"reason_code":"ERP_PROMOTION_AUTO_DISABLED"}` {
					t.Fatalf("expected warning details payload, got %s", got)
				}
				if got := fmt.Sprint(args[2].Value); got != "rec_1" {
					t.Fatalf("expected reconciliation id rec_1, got %s", got)
				}
				if got := fmt.Sprint(args[3].Value); !strings.HasPrefix(got, "rev_") {
					t.Fatalf("expected generated review id, got %s", got)
				}
				if got := fmt.Sprint(args[4].Value); got != "warning" {
					t.Fatalf("expected warning severity, got %s", got)
				}
				if got := fmt.Sprint(args[5].Value); got != "auto-promotion is disabled for this tenant" {
					t.Fatalf("expected review summary, got %s", got)
				}
				if got := fmt.Sprint(args[6].Value); got != "enable auto-promotion or route the item through manual review" {
					t.Fatalf("expected review action, got %s", got)
				}
				createdAt, ok := args[7].Value.(time.Time)
				if !ok {
					t.Fatalf("expected created_at time, got %T", args[7].Value)
				}
				if createdAt.UTC().IsZero() {
					t.Fatal("expected non-zero created_at")
				}
			},
		},
		repoScriptStep{kind: repoStepCommit},
	)

	repo := &ReconciliationRepo{base{db: db}}
	warningDetails := `{"reason_code":"ERP_PROMOTION_AUTO_DISABLED"}`
	if err := repo.MarkReviewRequired(
		context.Background(),
		"tenant-1",
		"rec_1",
		"ERP_PROMOTION_AUTO_DISABLED",
		"auto-promotion is disabled for this tenant",
		"enable auto-promotion or route the item through manual review",
		&warningDetails,
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state.done()
}
