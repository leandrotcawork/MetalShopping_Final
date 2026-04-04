package reconciliation

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"metalshopping/integration_worker/internal/erp_runtime/staging"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

type scriptStepKind string

const (
	stepBegin    scriptStepKind = "begin"
	stepExec     scriptStepKind = "exec"
	stepCommit   scriptStepKind = "commit"
	stepRollback scriptStepKind = "rollback"
)

type scriptStep struct {
	kind  scriptStepKind
	query string
	args  []any
}

type scriptState struct {
	t     *testing.T
	mu    sync.Mutex
	steps []scriptStep
	pos   int
}

func newScriptedDB(t *testing.T, steps ...scriptStep) *sql.DB {
	t.Helper()

	state := &scriptState{t: t, steps: steps}
	db := sql.OpenDB(&scriptConnector{state: state})
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
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

func (c *scriptConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (c *scriptConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	c.state.expect(stepBegin, "", nil)
	return &scriptTx{state: c.state}, nil
}

func (c *scriptConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.expect(stepExec, query, args)
	return driver.RowsAffected(1), nil
}

func (c *scriptConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, fmt.Errorf("query not supported")
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

func (s *scriptState) expect(kind scriptStepKind, query string, args []driver.NamedValue) scriptStep {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos >= len(s.steps) {
		s.t.Fatalf("unexpected %s call: %q", kind, query)
	}

	step := s.steps[s.pos]
	s.pos++

	if step.kind != kind {
		s.t.Fatalf("expected step %s, got %s", step.kind, kind)
	}
	if step.query != "" && !strings.Contains(query, step.query) {
		s.t.Fatalf("expected query containing %q, got %q", step.query, query)
	}
	if len(step.args) != 0 {
		if len(step.args) != len(args) {
			s.t.Fatalf("expected %d args, got %d", len(step.args), len(args))
		}
		for i := range step.args {
			if fmt.Sprint(step.args[i]) != fmt.Sprint(args[i].Value) {
				s.t.Fatalf("expected arg %d to be %v, got %v", i, step.args[i], args[i].Value)
			}
		}
	}

	return step
}

func TestReconcilerPromotesUniqueProduct(t *testing.T) {
	t.Parallel()

	db := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "set_config", args: []any{"tenant-a"}},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_reconciliation_results"},
		scriptStep{kind: stepCommit},
	)

	reconciler := NewReconciler(db)
	now := time.Now().UTC()
	results, err := reconciler.Reconcile(context.Background(), "tenant-a", "run-1", []*staging.StagingRecord{
		{
			StagingID:        "stage-1",
			TenantID:         "tenant-a",
			RunID:            "run-1",
			RawID:            "raw-1",
			EntityType:       types.EntityTypeProducts,
			SourceID:         "SKU-1",
			NormalizedJSON:   []byte(`{"CODPROD":"SKU-1","REFERENCIA":"7891234567890","REFFORN":"FAB-1"}`),
			ValidationStatus: staging.ValidationStatusValid,
			NormalizedAt:     now,
		},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 reconciliation result, got %d", len(results))
	}
	if results[0].Classification != ClassificationPromotable {
		t.Fatalf("expected promotable classification, got %s", results[0].Classification)
	}
	if results[0].WarningDetails != nil {
		t.Fatalf("expected no warning details for unique product, got %s", *results[0].WarningDetails)
	}
}

func TestReconcilerFlagsDuplicateEAN(t *testing.T) {
	t.Parallel()

	db := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "set_config", args: []any{"tenant-a"}},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_reconciliation_results"},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_reconciliation_results"},
		scriptStep{kind: stepCommit},
	)

	reconciler := NewReconciler(db)
	now := time.Now().UTC()
	results, err := reconciler.Reconcile(context.Background(), "tenant-a", "run-1", []*staging.StagingRecord{
		{
			StagingID:        "stage-1",
			TenantID:         "tenant-a",
			RunID:            "run-1",
			RawID:            "raw-1",
			EntityType:       types.EntityTypeProducts,
			SourceID:         "SKU-1",
			NormalizedJSON:   []byte(`{"CODPROD":"SKU-1","REFERENCIA":"7891234567890","REFFORN":"FAB-1"}`),
			ValidationStatus: staging.ValidationStatusValid,
			NormalizedAt:     now,
		},
		{
			StagingID:        "stage-2",
			TenantID:         "tenant-a",
			RunID:            "run-1",
			RawID:            "raw-2",
			EntityType:       types.EntityTypeProducts,
			SourceID:         "SKU-2",
			NormalizedJSON:   []byte(`{"CODPROD":"SKU-2","REFERENCIA":"7891234567890","REFFORN":"FAB-2"}`),
			ValidationStatus: staging.ValidationStatusValid,
			NormalizedAt:     now,
		},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	assertDuplicateResults(t, results, "ean")
}

func TestReconcilerFlagsDuplicateManufacturerReference(t *testing.T) {
	t.Parallel()

	db := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "set_config", args: []any{"tenant-a"}},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_reconciliation_results"},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_reconciliation_results"},
		scriptStep{kind: stepCommit},
	)

	reconciler := NewReconciler(db)
	now := time.Now().UTC()
	results, err := reconciler.Reconcile(context.Background(), "tenant-a", "run-1", []*staging.StagingRecord{
		{
			StagingID:        "stage-1",
			TenantID:         "tenant-a",
			RunID:            "run-1",
			RawID:            "raw-1",
			EntityType:       types.EntityTypeProducts,
			SourceID:         "SKU-1",
			NormalizedJSON:   []byte(`{"CODPROD":"SKU-1","REFERENCIA":"7891234567890","REFFORN":"FAB-1"}`),
			ValidationStatus: staging.ValidationStatusValid,
			NormalizedAt:     now,
		},
		{
			StagingID:        "stage-2",
			TenantID:         "tenant-a",
			RunID:            "run-1",
			RawID:            "raw-2",
			EntityType:       types.EntityTypeProducts,
			SourceID:         "SKU-2",
			NormalizedJSON:   []byte(`{"CODPROD":"SKU-2","REFERENCIA":"7891234567899","REFFORN":"FAB-1"}`),
			ValidationStatus: staging.ValidationStatusValid,
			NormalizedAt:     now,
		},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	assertDuplicateResults(t, results, "manufacturer_reference")
}

func assertDuplicateResults(t *testing.T, results []*ReconciliationResult, field string) {
	t.Helper()

	if len(results) != 2 {
		t.Fatalf("expected 2 reconciliation results, got %d", len(results))
	}
	for i, result := range results {
		if result.Classification != ClassificationReviewRequired {
			t.Fatalf("expected result %d to be review_required, got %s", i, result.Classification)
		}
		if result.ReasonCode != duplicateSecondaryIdentifierReasonCode {
			t.Fatalf("expected duplicate reason code, got %s", result.ReasonCode)
		}
		if result.WarningDetails == nil {
			t.Fatalf("expected warning details for duplicate %s", field)
		}
		if !strings.Contains(*result.WarningDetails, `"blocking_scope":"product_prices_inventory"`) {
			t.Fatalf("expected blocking scope in warning details, got %s", *result.WarningDetails)
		}
		if !strings.Contains(*result.WarningDetails, fmt.Sprintf(`"field":"%s"`, field)) {
			t.Fatalf("expected duplicate field %s in warning details, got %s", field, *result.WarningDetails)
		}
		if !strings.Contains(*result.WarningDetails, `"blocked_entities":["products","prices","inventory"]`) {
			t.Fatalf("expected downstream block entities in warning details, got %s", *result.WarningDetails)
		}
	}
}
