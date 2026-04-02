package catalog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	catalogdomain "metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

type recordedCatalogWrite struct {
	traceID string
	product catalogdomain.Product
	tx      *sql.Tx
}

type recordingCatalogWriter struct {
	call recordedCatalogWrite
	err  error
}

func (w *recordingCatalogWriter) CreateProductInTx(_ context.Context, tx *sql.Tx, product catalogdomain.Product, traceID string) error {
	w.call = recordedCatalogWrite{traceID: traceID, product: product, tx: tx}
	return w.err
}

func TestProductWriterPromoteProductUsesCatalogRepositoryBoundary(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "set_config('app.tenant_id'"},
		scriptStep{kind: stepExec, query: "INSERT INTO outbox_events"},
		scriptStep{kind: stepCommit},
	)
	writer := &recordingCatalogWriter{}
	productWriter := NewProductWriter(db, outbox.NewStore(db), writer)

	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
		RunID:            "run_1",
		StagingID:        "stg_1",
		EntityType:       domain.EntityTypeProducts,
		SourceID:         "src_1",
		Action:           "create",
		ReconciledAt:     time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC),
	}
	run := &domain.SyncRun{
		RunID:         "run_1",
		TenantID:      "tenant-1",
		InstanceID:    "inst_1",
		ConnectorType: domain.ConnectorTypeSankhya,
	}
	input := ports.ProductPromotionInput{
		SKU:                   "PN-001",
		Name:                  "Galvanized steel sheet",
		Description:           "Galvanized steel sheet",
		BrandName:             "Acme",
		StockProfileCode:      "standard",
		PrimaryTaxonomyNodeID: "txn_leaf_1",
		Status:                "active",
		Identifiers: []ports.ProductPromotionIdentifierInput{
			{IdentifierType: "reference", IdentifierValue: "REF-001", SourceSystem: "erp"},
		},
	}

	productID, err := productWriter.PromoteProduct(context.Background(), "trace_1", result, run, input)
	if err != nil {
		t.Fatalf("PromoteProduct error: %v", err)
	}
	if productID == "" {
		t.Fatal("expected product id")
	}
	if writer.call.tx == nil {
		t.Fatal("expected catalog repository to receive tenant tx")
	}
	if writer.call.traceID != "trace_1" {
		t.Fatalf("expected trace id trace_1, got %s", writer.call.traceID)
	}
	if writer.call.product.SKU != input.SKU {
		t.Fatalf("expected sku %s, got %s", input.SKU, writer.call.product.SKU)
	}
	if writer.call.product.TenantID != result.TenantID {
		t.Fatalf("expected tenant %s, got %s", result.TenantID, writer.call.product.TenantID)
	}
	if len(writer.call.product.Identifiers) == 0 {
		t.Fatal("expected catalog product identifiers to be built")
	}
	state.done()
}

type scriptStepKind string

const (
	stepBegin  scriptStepKind = "begin"
	stepExec   scriptStepKind = "exec"
	stepCommit scriptStepKind = "commit"
)

type scriptStep struct {
	kind  scriptStepKind
	query string
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

func (c *scriptConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.state.expect(stepBegin, "")
	return &scriptTx{state: c.state}, nil
}

func (c *scriptConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.expect(stepExec, query)
	return driver.RowsAffected(1), nil
}

func (c *scriptConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return nil, fmt.Errorf("unexpected query: %s", query)
}

type scriptTx struct {
	state *scriptState
}

func (tx *scriptTx) Commit() error {
	tx.state.expect(stepCommit, "")
	return nil
}

func (tx *scriptTx) Rollback() error {
	return nil
}

func (s *scriptState) expect(kind scriptStepKind, query string) {
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
}

func (s *scriptState) done() {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pos != len(s.steps) {
		s.t.Fatalf("expected %d scripted operations, consumed %d", len(s.steps), s.pos)
	}
}
