package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metalshopping/server_core/internal/modules/catalog/domain"
	catalogevents "metalshopping/server_core/internal/modules/catalog/events"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

// CreateOrGetProductInTx inserts the canonical product if it does not already
// exist for the tenant, otherwise it returns the existing product ID.
func (r *Repository) CreateOrGetProductInTx(ctx context.Context, tx *sql.Tx, product domain.Product, traceID string) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("catalog product transaction is required")
	}

	inserted, err := r.insertProductIfMissingInTx(ctx, tx, product)
	if err != nil {
		return "", err
	}

	if inserted {
		if err := r.insertProductIdentifiersInTx(ctx, tx, product); err != nil {
			return "", err
		}
		if err := r.appendProductCreatedOutboxInTx(ctx, tx, product, traceID); err != nil {
			return "", err
		}
		return product.ProductID, nil
	}

	return r.findProductIDBySKUInTx(ctx, tx, product.SKU)
}

func (r *Repository) insertProductIfMissingInTx(ctx context.Context, tx *sql.Tx, product domain.Product) (bool, error) {
	const insertSQL = `
INSERT INTO catalog_products (
  product_id,
  tenant_id,
  sku,
  name,
  description,
  brand_name,
  stock_profile_code,
  primary_taxonomy_node_id,
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
  $6,
  $7,
  $8,
  $9,
  $10
)
ON CONFLICT (tenant_id, sku) DO NOTHING
RETURNING product_id
`
	var insertedProductID string
	if err := tx.QueryRowContext(ctx, insertSQL, product.ProductID, product.SKU, product.Name, nullableText(product.Description), nullableText(product.BrandName), nullableText(product.StockProfileCode), nullableText(product.PrimaryTaxonomyNodeID), string(product.Status), product.CreatedAt, product.UpdatedAt).Scan(&insertedProductID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("insert catalog product if missing: %w", err)
	}
	return true, nil
}

func (r *Repository) insertProductIdentifiersInTx(ctx context.Context, tx *sql.Tx, product domain.Product) error {
	const insertIdentifierSQL = `
INSERT INTO catalog_product_identifiers (
  product_identifier_id,
  product_id,
  tenant_id,
  identifier_type,
  identifier_value,
  source_system,
  is_primary,
  created_at,
  updated_at
)
VALUES (
  $1,
  $2,
  current_tenant_id(),
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
)
`
	for _, identifier := range product.Identifiers {
		if _, err := tx.ExecContext(
			ctx,
			insertIdentifierSQL,
			identifier.ProductIdentifierID,
			identifier.ProductID,
			identifier.IdentifierType,
			identifier.IdentifierValue,
			nullableText(identifier.SourceSystem),
			identifier.IsPrimary,
			identifier.CreatedAt,
			identifier.UpdatedAt,
		); err != nil {
			return fmt.Errorf("insert catalog product identifier: %w", err)
		}
	}
	return nil
}

func (r *Repository) appendProductCreatedOutboxInTx(ctx context.Context, tx *sql.Tx, product domain.Product, traceID string) error {
	if r.outboxStore == nil {
		return nil
	}
	record, err := catalogevents.NewProductCreatedOutboxRecord(product, traceID, product.CreatedAt)
	if err != nil {
		return err
	}
	if err := r.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
		return err
	}
	return nil
}

func (r *Repository) findProductIDBySKUInTx(ctx context.Context, tx *sql.Tx, sku string) (string, error) {
	const querySQL = `
SELECT product_id
FROM catalog_products
WHERE tenant_id = current_tenant_id()
  AND sku = $1
LIMIT 1
`
	var productID string
	if err := tx.QueryRowContext(ctx, querySQL, sku).Scan(&productID); err != nil {
		return "", fmt.Errorf("lookup catalog product by sku: %w", err)
	}
	return productID, nil
}
