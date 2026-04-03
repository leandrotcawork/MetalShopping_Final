package unit

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type fakeExecContexter struct {
	query string
	args  []any
	err   error
}

func (f *fakeExecContexter) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.query = query
	f.args = args
	if f.err != nil {
		return nil, f.err
	}
	return fakeSQLResult(1), nil
}

type fakeSQLResult int64

func (f fakeSQLResult) LastInsertId() (int64, error) {
	return int64(f), nil
}

func (f fakeSQLResult) RowsAffected() (int64, error) {
	return int64(f), nil
}

func TestSetTenantContextRejectsEmptyTenantID(t *testing.T) {
	execer := &fakeExecContexter{}

	err := pgdb.SetTenantContext(context.Background(), execer, "")
	if err == nil {
		t.Fatal("expected tenant context error")
	}
}

func TestSetTenantContextRejectsReservedSystemTenantID(t *testing.T) {
	execer := &fakeExecContexter{}

	err := pgdb.SetTenantContext(context.Background(), execer, "*")
	if err == nil {
		t.Fatal("expected reserved tenant id error")
	}
}

func TestSetTenantContextWritesTenantSetting(t *testing.T) {
	execer := &fakeExecContexter{}

	err := pgdb.SetTenantContext(context.Background(), execer, "tenant-abc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if execer.query == "" {
		t.Fatal("expected query to be captured")
	}
	if len(execer.args) != 1 || execer.args[0] != "tenant-abc" {
		t.Fatalf("expected tenant-abc arg, got %#v", execer.args)
	}
}

func TestSetTenantContextPropagatesDatabaseError(t *testing.T) {
	execer := &fakeExecContexter{err: errors.New("db unavailable")}

	err := pgdb.SetTenantContext(context.Background(), execer, "tenant-abc")
	if err == nil {
		t.Fatal("expected database error")
	}
}

func TestSetSystemTenantContextWritesReservedMarker(t *testing.T) {
	execer := &fakeExecContexter{}

	err := pgdb.SetSystemTenantContext(context.Background(), execer)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if execer.query == "" {
		t.Fatal("expected query to be captured")
	}
	if len(execer.args) != 1 || execer.args[0] != "*" {
		t.Fatalf("expected system marker arg, got %#v", execer.args)
	}
}
