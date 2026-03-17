package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"metalshopping/server_core/internal/modules/catalog/domain"
	catalogevents "metalshopping/server_core/internal/modules/catalog/events"
	"metalshopping/server_core/internal/modules/catalog/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

type Repository struct {
	db          *sql.DB
	outboxStore *outbox.Store
}

func NewRepository(db *sql.DB, outboxStore *outbox.Store) *Repository {
	return &Repository{db: db, outboxStore: outboxStore}
}

func (r *Repository) CreateProduct(ctx context.Context, product domain.Product, traceID string) error {
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
`
	if _, err := tx.ExecContext(ctx, insertSQL, product.ProductID, product.SKU, product.Name, nullableText(product.Description), nullableText(product.BrandName), nullableText(product.StockProfileCode), nullableText(product.PrimaryTaxonomyNodeID), string(product.Status), product.CreatedAt, product.UpdatedAt); err != nil {
		return fmt.Errorf("insert catalog product: %w", err)
	}

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

	if r.outboxStore != nil {
		record, err := catalogevents.NewProductCreatedOutboxRecord(product, traceID, product.CreatedAt)
		if err != nil {
			return err
		}
		if err := r.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
			return err
		}
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
SELECT product_id, tenant_id, sku, name, COALESCE(description, ''), COALESCE(brand_name, ''), COALESCE(stock_profile_code, ''), COALESCE(primary_taxonomy_node_id, ''), status, created_at, updated_at
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
			&product.Description,
			&product.BrandName,
			&product.StockProfileCode,
			&product.PrimaryTaxonomyNodeID,
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

func (r *Repository) ListProductIdentifiers(ctx context.Context, tenantID, productID string) ([]domain.ProductIdentifier, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT product_identifier_id, product_id, tenant_id, identifier_type, identifier_value, COALESCE(source_system, ''), is_primary, created_at, updated_at
FROM catalog_product_identifiers
WHERE product_id = $1
ORDER BY identifier_type ASC, identifier_value ASC
`
	rows, err := tx.QueryContext(ctx, querySQL, productID)
	if err != nil {
		return nil, fmt.Errorf("query catalog product identifiers: %w", err)
	}
	defer rows.Close()

	identifiers := make([]domain.ProductIdentifier, 0, 8)
	for rows.Next() {
		var identifier domain.ProductIdentifier
		if err := rows.Scan(
			&identifier.ProductIdentifierID,
			&identifier.ProductID,
			&identifier.TenantID,
			&identifier.IdentifierType,
			&identifier.IdentifierValue,
			&identifier.SourceSystem,
			&identifier.IsPrimary,
			&identifier.CreatedAt,
			&identifier.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan catalog product identifier: %w", err)
		}
		identifiers = append(identifiers, identifier)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate catalog product identifiers: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit catalog product identifiers: %w", err)
	}

	return identifiers, nil
}

func (r *Repository) ListTaxonomyNodes(ctx context.Context, tenantID string, filter ports.TaxonomyNodeFilter) ([]domain.TaxonomyNode, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	querySQL := `
SELECT taxonomy_node_id, tenant_id, name, COALESCE(name_norm, ''), COALESCE(code, ''), COALESCE(parent_taxonomy_node_id, ''), level, COALESCE(path, ''), is_active, created_at, updated_at
FROM catalog_taxonomy_nodes
WHERE TRUE
`
	args := make([]any, 0, 2)
	argPos := 1

	if filter.ParentTaxonomyNodeID != "" {
		querySQL += fmt.Sprintf(" AND parent_taxonomy_node_id = $%d", argPos)
		args = append(args, filter.ParentTaxonomyNodeID)
		argPos++
	}
	if filter.Level != nil {
		querySQL += fmt.Sprintf(" AND level = $%d", argPos)
		args = append(args, *filter.Level)
		argPos++
	}

	querySQL += " ORDER BY level ASC, name ASC"

	rows, err := tx.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query catalog taxonomy nodes: %w", err)
	}
	defer rows.Close()

	nodes := make([]domain.TaxonomyNode, 0, 16)
	for rows.Next() {
		var node domain.TaxonomyNode
		if err := rows.Scan(
			&node.TaxonomyNodeID,
			&node.TenantID,
			&node.Name,
			&node.NameNorm,
			&node.Code,
			&node.ParentTaxonomyNodeID,
			&node.Level,
			&node.Path,
			&node.IsActive,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan catalog taxonomy node: %w", err)
		}
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate catalog taxonomy nodes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit catalog taxonomy nodes: %w", err)
	}

	return nodes, nil
}

func (r *Repository) ListTaxonomyLevelDefs(ctx context.Context, tenantID string) ([]domain.TaxonomyLevelDef, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT tenant_id, level, label, COALESCE(short_label, ''), is_enabled, created_at, updated_at
FROM catalog_taxonomy_level_defs
ORDER BY level ASC
`
	rows, err := tx.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, fmt.Errorf("query catalog taxonomy levels: %w", err)
	}
	defer rows.Close()

	defs := make([]domain.TaxonomyLevelDef, 0, 8)
	for rows.Next() {
		var def domain.TaxonomyLevelDef
		if err := rows.Scan(
			&def.TenantID,
			&def.Level,
			&def.Label,
			&def.ShortLabel,
			&def.IsEnabled,
			&def.CreatedAt,
			&def.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan catalog taxonomy level: %w", err)
		}
		defs = append(defs, def)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate catalog taxonomy levels: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit catalog taxonomy levels: %w", err)
	}

	return defs, nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
