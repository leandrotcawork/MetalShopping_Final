package tenantdb

import (
	"context"
	"database/sql"
	"errors"
	"testing"
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

func (f fakeSQLResult) LastInsertId() (int64, error) { return int64(f), nil }
func (f fakeSQLResult) RowsAffected() (int64, error) { return int64(f), nil }

func TestSetTenantContextRejectsReservedSystemTenantID(t *testing.T) {
	execer := &fakeExecContexter{}
	if err := SetTenantContext(context.Background(), execer, "*"); err == nil {
		t.Fatal("expected reserved tenant id error")
	}
}

func TestSetSystemTenantContextWritesReservedMarker(t *testing.T) {
	execer := &fakeExecContexter{}
	if err := SetSystemTenantContext(context.Background(), execer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(execer.args) != 1 || execer.args[0] != "*" {
		t.Fatalf("expected system marker arg, got %#v", execer.args)
	}
}

func TestSetSystemTenantContextPropagatesDatabaseError(t *testing.T) {
	execer := &fakeExecContexter{err: errors.New("db unavailable")}
	if err := SetSystemTenantContext(context.Background(), execer); err == nil {
		t.Fatal("expected database error")
	}
}
