package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"metalshopping/server_core/internal/modules/suppliers/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type DirectoryReader struct {
	db *sql.DB
}

func NewDirectoryReader(db *sql.DB) *DirectoryReader {
	return &DirectoryReader{db: db}
}

func (r *DirectoryReader) ListDirectory(ctx context.Context, tenantID string, onlyEnabled bool) ([]ports.DirectorySupplier, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT supplier_code, supplier_label, execution_kind, lookup_policy, enabled
FROM suppliers_directory
WHERE tenant_id = current_tenant_id()
  AND ($1 = FALSE OR enabled = TRUE)
ORDER BY supplier_code ASC
`

	rows, err := tx.QueryContext(ctx, query, onlyEnabled)
	if err != nil {
		return nil, fmt.Errorf("list suppliers directory: %w", err)
	}
	defer rows.Close()

	items := []ports.DirectorySupplier{}
	for rows.Next() {
		var item ports.DirectorySupplier
		if err := rows.Scan(
			&item.SupplierCode,
			&item.SupplierLabel,
			&item.ExecutionKind,
			&item.LookupPolicy,
			&item.Enabled,
		); err != nil {
			return nil, fmt.Errorf("scan suppliers directory row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate suppliers directory rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit suppliers directory read: %w", err)
	}
	return items, nil
}
