package tenantdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type tenantIDKey struct{}

type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const setTenantConfigSQL = `SELECT set_config('app.tenant_id', $1, true)`

// BeginTenantTx starts a transaction and binds app.tenant_id for tenant-scoped work.
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

// SetTenantContext binds app.tenant_id on an existing transaction or executor.
func SetTenantContext(ctx context.Context, execer execContexter, tenantID string) error {
	tenantID, err := ValidateTenantID(tenantID)
	if err != nil {
		return err
	}

	if _, err := execer.ExecContext(ctx, setTenantConfigSQL, tenantID); err != nil {
		return fmt.Errorf("set tenant session: %w", err)
	}

	return nil
}

// WithTenantID stores a validated tenant id in context for downstream tenant-bound work.
func WithTenantID(ctx context.Context, tenantID string) (context.Context, error) {
	tenantID, err := ValidateTenantID(tenantID)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, tenantIDKey{}, tenantID), nil
}

// TenantIDFromContext returns the tenant id stored in context, if present.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantIDKey{}).(string)
	return tenantID, ok
}

// ValidateTenantID trims and validates a tenant identifier before it is bound
// into the postgres session.
func ValidateTenantID(tenantID string) (string, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "", fmt.Errorf("tenant id is required")
	}
	return tenantID, nil
}
