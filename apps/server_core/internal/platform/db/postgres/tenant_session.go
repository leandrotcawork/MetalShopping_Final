package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const setTenantConfigSQL = `SELECT set_config('app.tenant_id', $1, true)`

type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func BeginTenantTx(ctx context.Context, db txBeginner, tenantID string, opts *sql.TxOptions) (*sql.Tx, error) {
	tenantID, err := ValidateTenantID(tenantID)
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("begin tenant transaction: %w", err)
	}

	if err := SetTenantContext(ctx, tx, tenantID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func SetTenantContext(ctx context.Context, execer execContexter, tenantID string) error {
	tenantID, err := ValidateTenantID(tenantID)
	if err != nil {
		return err
	}

	if _, err := execer.ExecContext(ctx, setTenantConfigSQL, tenantID); err != nil {
		return fmt.Errorf("set postgres tenant context: %w", err)
	}

	return nil
}

func ValidateTenantID(tenantID string) (string, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "", fmt.Errorf("postgres tenant id is required")
	}
	return tenantID, nil
}
