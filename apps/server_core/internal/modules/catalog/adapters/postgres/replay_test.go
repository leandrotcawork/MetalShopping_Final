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

	"metalshopping/server_core/internal/modules/catalog/domain"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

func TestCreateOrGetProductInTxReturnsExistingCanonicalProductID(t *testing.T) {
	db, state := newReplayScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "set_config('app.tenant_id'"},
		scriptStep{kind: stepQuery, query: "INSERT INTO catalog_products"},
		scriptStep{kind: stepQuery, query: "SELECT product_id", rows: [][]driver.Value{{"prd_existing"}}},
		scriptStep{kind: stepCommit},
	)
	repo := NewRepository(db, nil)

	tx, err := pgdb.BeginTenantTx(context.Background(), db, "tenant-1", nil)
	if err != nil {
		t.Fatalf("BeginTenantTx error: %v", err)
	}

	productID, err := repo.CreateOrGetProductInTx(context.Background(), tx, domain.Product{
		ProductID:        "prd_new",
		TenantID:         "tenant-1",
		SKU:              "PN-001",
		Name:             "Galvanized steel sheet",
		Description:      "Galvanized steel sheet",
		BrandName:        "Acme",
		StockProfileCode: "standard",
		Status:           domain.ProductStatusActive,
		CreatedAt:        time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC),
	}, "trace_1")
	if err != nil {
		t.Fatalf("CreateOrGetProductInTx error: %v", err)
	}
	if productID != "prd_existing" {
		t.Fatalf("expected existing canonical product id prd_existing, got %s", productID)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tenant tx: %v", err)
	}
	state.done()
}

type scriptStepKind string

const (
	stepBegin  scriptStepKind = "begin"
	stepExec   scriptStepKind = "exec"
	stepQuery  scriptStepKind = "query"
	stepCommit scriptStepKind = "commit"
)

type scriptStep struct {
	kind  scriptStepKind
	query string
	rows  [][]driver.Value
}

type scriptState struct {
	t     *testing.T
	mu    sync.Mutex
	steps []scriptStep
	pos   int
}

func newReplayScriptedDB(t *testing.T, steps ...scriptStep) (*sql.DB, *scriptState) {
	t.Helper()

	state := &scriptState{t: t, steps: steps}
	db := sql.OpenDB(&scriptReplayConnector{state: state})
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db, state
}

type scriptReplayConnector struct {
	state *scriptState
}

func (c *scriptReplayConnector) Connect(context.Context) (driver.Conn, error) {
	return &scriptReplayConn{state: c.state}, nil
}

func (c *scriptReplayConnector) Driver() driver.Driver {
	return scriptReplayDriver{}
}

type scriptReplayDriver struct{}

func (scriptReplayDriver) Open(string) (driver.Conn, error) {
	return nil, fmt.Errorf("open not supported")
}

type scriptReplayConn struct {
	state *scriptState
}

func (c *scriptReplayConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported")
}

func (c *scriptReplayConn) Close() error { return nil }

func (c *scriptReplayConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *scriptReplayConn) Ping(context.Context) error { return nil }

func (c *scriptReplayConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (c *scriptReplayConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.state.expect(stepBegin, "")
	return &scriptReplayTx{state: c.state}, nil
}

func (c *scriptReplayConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.expect(stepExec, query)
	return driver.RowsAffected(1), nil
}

func (c *scriptReplayConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	step := c.state.expect(stepQuery, query)
	return &scriptReplayRows{rows: step.rows}, nil
}

type scriptReplayTx struct {
	state *scriptState
}

func (tx *scriptReplayTx) Commit() error {
	tx.state.expect(stepCommit, "")
	return nil
}

func (tx *scriptReplayTx) Rollback() error {
	return nil
}

type scriptReplayRows struct {
	rows [][]driver.Value
	pos  int
}

func (r *scriptReplayRows) Columns() []string {
	if len(r.rows) == 0 {
		return []string{"product_id"}
	}
	cols := make([]string, len(r.rows[0]))
	for i := range cols {
		cols[i] = fmt.Sprintf("col_%d", i)
	}
	return cols
}

func (r *scriptReplayRows) Close() error { return nil }

func (r *scriptReplayRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

func (s *scriptState) expect(kind scriptStepKind, query string) scriptStep {
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
