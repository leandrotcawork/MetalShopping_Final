package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
)

type repoScriptStepKind string

const (
	repoStepBegin    repoScriptStepKind = "begin"
	repoStepExec     repoScriptStepKind = "exec"
	repoStepQuery    repoScriptStepKind = "query"
	repoStepCommit   repoScriptStepKind = "commit"
	repoStepRollback repoScriptStepKind = "rollback"
)

type repoScriptStep struct {
	kind         repoScriptStepKind
	query        string
	args         []any
	columns      []string
	rows         [][]driver.Value
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

func (c *repoScriptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	step := c.state.expect(repoStepQuery, query, args)
	return &repoScriptRows{columns: step.columns, rows: step.rows}, nil
}

type repoScriptRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (r *repoScriptRows) Columns() []string {
	return r.columns
}

func (r *repoScriptRows) Close() error {
	return nil
}

func (r *repoScriptRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.index]
	for i := range dest {
		if i < len(row) {
			dest[i] = row[i]
			continue
		}
		dest[i] = nil
	}
	r.index++
	return nil
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

func assertMarkReviewRequiredExec(expectedReviewID string) func(*testing.T, string, []driver.NamedValue) {
	return func(t *testing.T, query string, args []driver.NamedValue) {
		t.Helper()
		if !strings.Contains(query, "UPDATE erp_reconciliation_results") {
			t.Fatal("expected reconciliation update in review-required statement")
		}
		if !strings.Contains(query, "promotion_status IN ('pending', 'promoting')") {
			t.Fatal("expected pending/promoting guard in review-required statement")
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
		if got := fmt.Sprint(args[3].Value); got != expectedReviewID {
			t.Fatalf("expected deterministic review id %s, got %s", expectedReviewID, got)
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
	}
}

func TestReconciliationRepoMarkReviewRequiredCreatesReviewItem(t *testing.T) {
	expectedReviewID := generateReviewID("tenant-1", "rec_1")
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{
			kind:   repoStepExec,
			query:  "INSERT INTO erp_review_items",
			assert: assertMarkReviewRequiredExec(expectedReviewID),
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

func TestReconciliationRepoMarkReviewRequiredIsIdempotentOnRepeatedCalls(t *testing.T) {
	expectedReviewID := generateReviewID("tenant-1", "rec_1")
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{kind: repoStepExec, query: "INSERT INTO erp_review_items", rowsAffected: int64Ptr(1), assert: assertMarkReviewRequiredExec(expectedReviewID)},
		repoScriptStep{kind: repoStepCommit},
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{kind: repoStepExec, query: "INSERT INTO erp_review_items", rowsAffected: int64Ptr(0), assert: assertMarkReviewRequiredExec(expectedReviewID)},
		repoScriptStep{kind: repoStepRollback},
	)

	repo := &ReconciliationRepo{base{db: db}}
	warningDetails := `{"reason_code":"ERP_PROMOTION_AUTO_DISABLED"}`
	for i := 0; i < 2; i++ {
		if err := repo.MarkReviewRequired(
			context.Background(),
			"tenant-1",
			"rec_1",
			"ERP_PROMOTION_AUTO_DISABLED",
			"auto-promotion is disabled for this tenant",
			"enable auto-promotion or route the item through manual review",
			&warningDetails,
		); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i+1, err)
		}
	}

	state.done()
}

func TestReconciliationRepoMarkReviewRequiredNoopsOnZeroRows(t *testing.T) {
	expectedReviewID := generateReviewID("tenant-1", "rec_1")
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{kind: repoStepExec, query: "INSERT INTO erp_review_items", rowsAffected: int64Ptr(0), assert: assertMarkReviewRequiredExec(expectedReviewID)},
		repoScriptStep{kind: repoStepRollback},
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

func TestReconciliationRepoResolvePriceContextCodeReturnsCanonicalContext(t *testing.T) {
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{kind: repoStepQuery, query: "FROM erp_source_price_context_mappings", args: []any{"sankhya", "17"}, columns: []string{"canonical_context_code"}, rows: [][]driver.Value{{"PROMOTION"}}},
		repoScriptStep{kind: repoStepCommit},
	)

	repo := &ReconciliationRepo{base{db: db}}
	got, found, err := repo.ResolvePriceContextCode(context.Background(), "tenant-1", "sankhya", "17")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected mapping to be found")
	}
	if got != "promotion" {
		t.Fatalf("expected normalized canonical context promotion, got %q", got)
	}

	state.done()
}

func TestReconciliationRepoResolvePriceContextCodeReturnsNotFoundWithoutError(t *testing.T) {
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{kind: repoStepQuery, query: "FROM erp_source_price_context_mappings", args: []any{"sankhya", "99"}, columns: []string{"canonical_context_code"}, rows: nil},
		repoScriptStep{kind: repoStepCommit},
	)

	repo := &ReconciliationRepo{base{db: db}}
	got, found, err := repo.ResolvePriceContextCode(context.Background(), "tenant-1", "sankhya", "99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected mapping lookup to miss")
	}
	if got != "" {
		t.Fatalf("expected empty canonical context on miss, got %q", got)
	}

	state.done()
}

func TestInstanceRepoListParsesEnabledEntitiesTextArray(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{
			kind:    repoStepQuery,
			query:   "FROM erp_integration_instances",
			args:    []any{50, 0},
			columns: []string{
				"instance_id",
				"tenant_id",
				"connector_type",
				"display_name",
				"connection_kind",
				"db_host",
				"db_port",
				"db_service_name",
				"db_sid",
				"db_username",
				"db_password_secret_ref",
				"connect_timeout_seconds",
				"fetch_batch_size",
				"entity_batch_size",
				"enabled_entities",
				"sync_schedule",
				"status",
				"created_at",
				"updated_at",
			},
			rows: [][]driver.Value{{
				"inst_1",
				"tenant-1",
				"oracle",
				"Main Instance",
				"oracle",
				"db.local",
				int64(1521),
				"ORCL",
				nil,
				"user",
				"secret-ref",
				int64(10),
				int64(500),
				int64(250),
				"{products,prices}",
				nil,
				"active",
				now,
				now,
			}},
		},
		repoScriptStep{kind: repoStepCommit},
	)

	repo := &InstanceRepo{base{db: db}}
	items, err := repo.List(context.Background(), "tenant-1", 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(items))
	}
	if got := items[0].EnabledEntities; len(got) != 2 || got[0] != domain.EntityTypeProducts || got[1] != domain.EntityTypePrices {
		t.Fatalf("unexpected enabled_entities: %#v", got)
	}

	state.done()
}

func TestRunRepoListParsesEntityScopeTextArray(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	db, state := newRepoScriptedDB(t,
		repoScriptStep{kind: repoStepBegin},
		repoScriptStep{kind: repoStepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		repoScriptStep{
			kind:    repoStepQuery,
			query:   "FROM erp_sync_runs",
			args:    []any{"inst_1", 50, 0},
			columns: []string{
				"run_id",
				"tenant_id",
				"instance_id",
				"connector_type",
				"run_mode",
				"entity_scope",
				"status",
				"started_at",
				"completed_at",
				"promoted_count",
				"warning_count",
				"rejected_count",
				"review_count",
				"failure_summary",
				"cursor_state",
				"created_at",
			},
			rows: [][]driver.Value{{
				"run_1",
				"tenant-1",
				"inst_1",
				"oracle",
				"full",
				"{products}",
				"pending",
				nil,
				nil,
				int64(0),
				int64(0),
				int64(0),
				int64(0),
				nil,
				nil,
				now,
			}},
		},
		repoScriptStep{kind: repoStepCommit},
	)

	repo := &RunRepo{base{db: db}}
	items, err := repo.List(context.Background(), "tenant-1", "inst_1", 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 run, got %d", len(items))
	}
	if got := items[0].EntityScope; len(got) != 1 || got[0] != domain.EntityTypeProducts {
		t.Fatalf("unexpected entity_scope: %#v", got)
	}

	state.done()
}

func int64Ptr(v int64) *int64 {
	return &v
}
