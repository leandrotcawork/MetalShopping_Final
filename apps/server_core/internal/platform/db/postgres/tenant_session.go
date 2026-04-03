package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const setTenantConfigSQL = `SELECT set_config('app.tenant_id', $1, true)`
const systemTenantMarker = "*"

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

// BeginSystemTx starts a transaction bound to the reserved system tenant marker.
// This is only for internal cross-tenant operational workflows.
func BeginSystemTx(ctx context.Context, db txBeginner, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("begin system transaction: %w", err)
	}
	if err := SetSystemTenantContext(ctx, tx); err != nil {
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

// SetSystemTenantContext binds the reserved system tenant marker.
func SetSystemTenantContext(ctx context.Context, execer execContexter) error {
	if _, err := execer.ExecContext(ctx, setTenantConfigSQL, systemTenantMarker); err != nil {
		return fmt.Errorf("set postgres system tenant context: %w", err)
	}
	return nil
}

func ValidateTenantID(tenantID string) (string, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "", fmt.Errorf("postgres tenant id is required")
	}
	if tenantID == systemTenantMarker {
		return "", fmt.Errorf("postgres tenant id %q is reserved", systemTenantMarker)
	}
	return tenantID, nil
}
