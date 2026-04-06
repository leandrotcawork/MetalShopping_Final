package oracle

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
)

func init() {
	sql.Register("godror", oracleTestDriver{})
}

func TestNewQueryRunnerPingsConnection(t *testing.T) {
	t.Parallel()

	dataset := &oracleTestDataset{}
	runner := mustNewTestQueryRunner(t, dataset)
	t.Cleanup(func() {
		if err := runner.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})

	if got := dataset.pingCount(); got != 1 {
		t.Fatalf("PingContext count = %d, want 1", got)
	}
}

func TestQueryRunnerRejectsEmptySQL(t *testing.T) {
	t.Parallel()

	var runner QueryRunner

	err := runner.Query(context.Background(), dbsource.QuerySpec{}, func(dbsource.RowReader) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "SQL must not be empty") {
		t.Fatalf("expected empty SQL validation error, got %v", err)
	}
}

func TestQueryRunnerPropagatesCallbackError(t *testing.T) {
	t.Parallel()

	dataset := &oracleTestDataset{
		columns: []string{"ID"},
		rows: [][]driver.Value{
			{int64(1)},
		},
	}
	runner := mustNewTestQueryRunner(t, dataset)

	sentinel := errors.New("callback failed")
	err := runner.Query(context.Background(), dbsource.QuerySpec{SQL: "select id from dual"}, func(dbsource.RowReader) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Query returned %v, want wrapped sentinel error", err)
	}
	if got := dataset.queryCount(); got != 1 {
		t.Fatalf("QueryContext count = %d, want 1", got)
	}
}

func TestQueryRunnerRejectsDuplicateColumns(t *testing.T) {
	t.Parallel()

	dataset := &oracleTestDataset{
		columns: []string{"ID", "id"},
		rows: [][]driver.Value{
			{int64(1), int64(2)},
		},
	}
	runner := mustNewTestQueryRunner(t, dataset)

	called := false
	err := runner.Query(context.Background(), dbsource.QuerySpec{SQL: "select id, id from dual"}, func(dbsource.RowReader) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("expected duplicate column error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate column") {
		t.Fatalf("expected duplicate column error, got %v", err)
	}
	if called {
		t.Fatal("callback should not run when columns collide")
	}
}

type oracleTestDriver struct{}

func (oracleTestDriver) Open(name string) (driver.Conn, error) {
	return &oracleTestConn{dsn: name}, nil
}

type oracleTestConn struct {
	dsn string
}

func (c *oracleTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *oracleTestConn) Close() error {
	return nil
}

func (c *oracleTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

func (c *oracleTestConn) Ping(context.Context) error {
	dataset, err := lookupOracleTestDataset(c.dsn)
	if err != nil {
		return err
	}
	return dataset.recordPing()
}

func (c *oracleTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	dataset, err := lookupOracleTestDataset(c.dsn)
	if err != nil {
		return nil, err
	}
	return dataset.queryRows()
}

var _ driver.Pinger = (*oracleTestConn)(nil)
var _ driver.QueryerContext = (*oracleTestConn)(nil)

type oracleTestDataset struct {
	mu         sync.Mutex
	pingErr    error
	queryErr   error
	columns    []string
	rows       [][]driver.Value
	pingCalls  int
	queryCalls int
}

func (d *oracleTestDataset) recordPing() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pingCalls++
	return d.pingErr
}

func (d *oracleTestDataset) queryRows() (driver.Rows, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queryCalls++
	if d.queryErr != nil {
		return nil, d.queryErr
	}

	columns := append([]string(nil), d.columns...)
	rows := make([][]driver.Value, len(d.rows))
	for i := range d.rows {
		rows[i] = append([]driver.Value(nil), d.rows[i]...)
	}
	return &oracleTestRows{columns: columns, rows: rows}, nil
}

func (d *oracleTestDataset) pingCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.pingCalls
}

func (d *oracleTestDataset) queryCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.queryCalls
}

type oracleTestRows struct {
	columns []string
	rows    [][]driver.Value
	idx     int
}

func (r *oracleTestRows) Columns() []string {
	return append([]string(nil), r.columns...)
}

func (r *oracleTestRows) Close() error {
	return nil
}

func (r *oracleTestRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	for i := range dest {
		if i < len(row) {
			dest[i] = row[i]
			continue
		}
		dest[i] = nil
	}
	return nil
}

var oracleTestDatasetRegistry = struct {
	mu   sync.Mutex
	data map[string]*oracleTestDataset
}{
	data: map[string]*oracleTestDataset{},
}

func mustNewTestQueryRunner(t *testing.T, dataset *oracleTestDataset) *QueryRunner {
	t.Helper()

	cfg := testOracleConfig()
	dsn, err := cfg.ConnectString()
	if err != nil {
		t.Fatalf("ConnectString returned error: %v", err)
	}
	registerOracleTestDataset(t, dsn, dataset)

	runner, err := NewQueryRunner(cfg)
	if err != nil {
		t.Fatalf("NewQueryRunner returned error: %v", err)
	}
	return runner
}

func registerOracleTestDataset(t *testing.T, dsn string, dataset *oracleTestDataset) {
	t.Helper()

	oracleTestDatasetRegistry.mu.Lock()
	oracleTestDatasetRegistry.data[dsn] = dataset
	oracleTestDatasetRegistry.mu.Unlock()
	t.Cleanup(func() {
		oracleTestDatasetRegistry.mu.Lock()
		delete(oracleTestDatasetRegistry.data, dsn)
		oracleTestDatasetRegistry.mu.Unlock()
	})
}

func lookupOracleTestDataset(dsn string) (*oracleTestDataset, error) {
	oracleTestDatasetRegistry.mu.Lock()
	defer oracleTestDatasetRegistry.mu.Unlock()

	dataset, ok := oracleTestDatasetRegistry.data[dsn]
	if !ok {
		return nil, fmt.Errorf("oracle test dataset %q not registered", dsn)
	}
	return dataset, nil
}

func testOracleConfig() Config {
	serviceName := fmt.Sprintf("ORCL_%d", atomic.AddUint64(&oracleTestConfigSeq, 1))
	return Config{
		Host:              "db.example.internal",
		Port:              1521,
		ServiceName:       &serviceName,
		Username:          "erp_user",
		Password:          "erp_secret",
		ConnectTimeoutSec: 1,
	}
}

var oracleTestConfigSeq uint64
