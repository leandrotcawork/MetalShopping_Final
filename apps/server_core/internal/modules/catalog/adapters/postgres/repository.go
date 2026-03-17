package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"metalshopping/server_core/internal/modules/catalog/domain"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProduct(ctx context.Context, product domain.Product) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, product.TenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const insertSQL = `
INSERT INTO catalog_products (
  product_id,
  tenant_id,
  sku,
  name,
  status,
  created_at,
  updated_at
)
VALUES (
  $1,
  current_tenant_id(),
  $2,
  $3,
  $4,
  $5,
  $6
)
`
	if _, err := tx.ExecContext(ctx, insertSQL, product.ProductID, product.SKU, product.Name, string(product.Status), product.CreatedAt, product.UpdatedAt); err != nil {
		return fmt.Errorf("insert catalog product: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit catalog product: %w", err)
	}
	return nil
}

func (r *Repository) ListProducts(ctx context.Context, tenantID string) ([]domain.Product, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT product_id, tenant_id, sku, name, status, created_at, updated_at
FROM catalog_products
ORDER BY sku ASC
`
	rows, err := tx.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, fmt.Errorf("query catalog products: %w", err)
	}
	defer rows.Close()

	products := make([]domain.Product, 0, 16)
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(
			&product.ProductID,
			&product.TenantID,
			&product.SKU,
			&product.Name,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan catalog product: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate catalog products: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit catalog product list: %w", err)
	}

	return products, nil
}
