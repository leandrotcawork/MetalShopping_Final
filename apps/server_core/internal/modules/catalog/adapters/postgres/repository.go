package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/catalog/domain"
	catalogevents "metalshopping/server_core/internal/modules/catalog/events"
	"metalshopping/server_core/internal/modules/catalog/ports"
	catalogreadmodel "metalshopping/server_core/internal/modules/catalog/readmodel"
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

func (r *Repository) ListProductsPortfolio(ctx context.Context, tenantID string, filter catalogreadmodel.ProductsPortfolioFilter) (catalogreadmodel.ProductsPortfolioResult, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return catalogreadmodel.ProductsPortfolioResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	whereSQL, args := buildProductsPortfolioWhereClause(filter, 1)

	querySQL := productsPortfolioCommonCTE + `
SELECT
  p.product_id,
  p.sku,
  p.name,
  p.description,
  p.brand_name,
  idf.pn_interno,
  idf.reference,
  idf.ean,
  txo.taxonomy_leaf_name,
  txo.taxonomy_leaf0_name,
  p.stock_profile_code,
  p.status,
  cp.price_amount,
  cp.replacement_cost_amount,
  cp.average_cost_amount,
  cp.currency_code,
  ip.on_hand_quantity,
  ip.position_status,
  GREATEST(
    p.updated_at,
    COALESCE(cp.updated_at, p.updated_at),
    COALESCE(ip.updated_at, p.updated_at)
  ) AS updated_at
FROM catalog_products p
LEFT JOIN taxonomy_lookup txo ON txo.taxonomy_node_id = p.primary_taxonomy_node_id
LEFT JOIN identifier_lookup idf ON idf.product_id = p.product_id
LEFT JOIN current_prices cp ON cp.product_id = p.product_id
LEFT JOIN current_positions ip ON ip.product_id = p.product_id
` + whereSQL + `
ORDER BY p.sku ASC
LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	rows, err := tx.QueryContext(ctx, querySQL, append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return catalogreadmodel.ProductsPortfolioResult{}, fmt.Errorf("query products portfolio: %w", err)
	}
	defer rows.Close()

	items := make([]catalogreadmodel.ProductsPortfolioItem, 0, filter.Limit)
	for rows.Next() {
		item, err := scanProductsPortfolioItem(rows)
		if err != nil {
			return catalogreadmodel.ProductsPortfolioResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return catalogreadmodel.ProductsPortfolioResult{}, fmt.Errorf("iterate products portfolio: %w", err)
	}

	countSQL := productsPortfolioCommonCTE + `
SELECT COUNT(*)
FROM catalog_products p
LEFT JOIN taxonomy_lookup txo ON txo.taxonomy_node_id = p.primary_taxonomy_node_id
LEFT JOIN identifier_lookup idf ON idf.product_id = p.product_id
` + whereSQL

	var total int
	if err := tx.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return catalogreadmodel.ProductsPortfolioResult{}, fmt.Errorf("count products portfolio: %w", err)
	}

	filters, err := loadProductsPortfolioFilters(ctx, tx)
	if err != nil {
		return catalogreadmodel.ProductsPortfolioResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return catalogreadmodel.ProductsPortfolioResult{}, fmt.Errorf("commit products portfolio: %w", err)
	}

	return catalogreadmodel.ProductsPortfolioResult{
		Rows:    items,
		Filters: filters,
		Paging: catalogreadmodel.ProductsPortfolioPaging{
			Offset:   filter.Offset,
			Limit:    filter.Limit,
			Returned: len(items),
			Total:    total,
		},
	}, nil
}

const productsPortfolioCommonCTE = `
WITH RECURSIVE taxonomy_chain AS (
  SELECT
    taxonomy_node_id,
    tenant_id,
    parent_taxonomy_node_id,
    name,
    level,
    taxonomy_node_id AS leaf_id
  FROM catalog_taxonomy_nodes
  WHERE tenant_id = current_tenant_id()
  UNION ALL
  SELECT
    parent.taxonomy_node_id,
    parent.tenant_id,
    parent.parent_taxonomy_node_id,
    parent.name,
    parent.level,
    chain.leaf_id
  FROM taxonomy_chain chain
  JOIN catalog_taxonomy_nodes parent
    ON parent.tenant_id = current_tenant_id()
   AND parent.taxonomy_node_id = chain.parent_taxonomy_node_id
),
taxonomy_lookup AS (
  SELECT
    leaf_id AS taxonomy_node_id,
    MAX(CASE WHEN taxonomy_node_id = leaf_id THEN name END) AS taxonomy_leaf_name,
    MAX(CASE WHEN level = 0 THEN name END) AS taxonomy_leaf0_name
  FROM taxonomy_chain
  GROUP BY leaf_id
),
identifier_lookup AS (
  SELECT
    product_id,
    MAX(CASE WHEN identifier_type = 'pn_interno' THEN identifier_value END) AS pn_interno,
    MAX(CASE WHEN identifier_type = 'reference' THEN identifier_value END) AS reference,
    MAX(CASE WHEN identifier_type = 'ean' THEN identifier_value END) AS ean
  FROM catalog_product_identifiers
  WHERE tenant_id = current_tenant_id()
  GROUP BY product_id
),
current_prices AS (
  SELECT DISTINCT ON (product_id)
    product_id,
    price_amount,
    replacement_cost_amount,
    average_cost_amount,
    currency_code,
    updated_at
  FROM pricing_product_prices
  WHERE tenant_id = current_tenant_id()
    AND effective_from <= NOW()
    AND (effective_to IS NULL OR effective_to > NOW())
  ORDER BY product_id, effective_from DESC, created_at DESC
),
current_positions AS (
  SELECT DISTINCT ON (product_id)
    product_id,
    on_hand_quantity,
    position_status,
    updated_at
  FROM inventory_product_positions
  WHERE tenant_id = current_tenant_id()
    AND effective_from <= NOW()
    AND (effective_to IS NULL OR effective_to > NOW())
  ORDER BY product_id, effective_from DESC, created_at DESC
)
`

func buildProductsPortfolioWhereClause(filter catalogreadmodel.ProductsPortfolioFilter, startArg int) (string, []any) {
	clauses := []string{"WHERE p.tenant_id = current_tenant_id()"}
	args := make([]any, 0, 4)
	argPos := startArg

	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(`AND (
  p.sku ILIKE $%d
  OR p.name ILIKE $%d
  OR COALESCE(p.description, '') ILIKE $%d
  OR COALESCE(idf.pn_interno, '') ILIKE $%d
  OR COALESCE(idf.reference, '') ILIKE $%d
  OR COALESCE(idf.ean, '') ILIKE $%d
)`, argPos, argPos, argPos, argPos, argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}
	if filter.BrandName != "" {
		clauses = append(clauses, fmt.Sprintf("AND p.brand_name = $%d", argPos))
		args = append(args, filter.BrandName)
		argPos++
	}
	if filter.TaxonomyLeaf0Name != "" {
		clauses = append(clauses, fmt.Sprintf("AND txo.taxonomy_leaf0_name = $%d", argPos))
		args = append(args, filter.TaxonomyLeaf0Name)
		argPos++
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("AND p.status = $%d", argPos))
		args = append(args, filter.Status)
	}

	return strings.Join(clauses, "\n"), args
}

func loadProductsPortfolioFilters(ctx context.Context, tx *sql.Tx) (catalogreadmodel.ProductsPortfolioFilters, error) {
	brands, err := queryStringList(ctx, tx, `
SELECT DISTINCT brand_name
FROM catalog_products
WHERE tenant_id = current_tenant_id()
  AND brand_name IS NOT NULL
  AND BTRIM(brand_name) <> ''
ORDER BY brand_name ASC
`)
	if err != nil {
		return catalogreadmodel.ProductsPortfolioFilters{}, fmt.Errorf("query products portfolio brands: %w", err)
	}

	statuses, err := queryStringList(ctx, tx, `
SELECT DISTINCT status
FROM catalog_products
WHERE tenant_id = current_tenant_id()
  AND BTRIM(status) <> ''
ORDER BY status ASC
`)
	if err != nil {
		return catalogreadmodel.ProductsPortfolioFilters{}, fmt.Errorf("query products portfolio status: %w", err)
	}

	taxonomyLeaf0Names, err := queryStringList(ctx, tx, `
WITH RECURSIVE taxonomy_chain AS (
  SELECT taxonomy_node_id, tenant_id, parent_taxonomy_node_id, name, level, taxonomy_node_id AS leaf_id
  FROM catalog_taxonomy_nodes
  WHERE tenant_id = current_tenant_id()
  UNION ALL
  SELECT parent.taxonomy_node_id, parent.tenant_id, parent.parent_taxonomy_node_id, parent.name, parent.level, chain.leaf_id
  FROM taxonomy_chain chain
  JOIN catalog_taxonomy_nodes parent
    ON parent.tenant_id = current_tenant_id()
   AND parent.taxonomy_node_id = chain.parent_taxonomy_node_id
),
taxonomy_lookup AS (
  SELECT
    leaf_id AS taxonomy_node_id,
    MAX(CASE WHEN level = 0 THEN name END) AS taxonomy_leaf0_name
  FROM taxonomy_chain
  GROUP BY leaf_id
)
SELECT DISTINCT txo.taxonomy_leaf0_name
FROM catalog_products p
JOIN taxonomy_lookup txo ON txo.taxonomy_node_id = p.primary_taxonomy_node_id
WHERE p.tenant_id = current_tenant_id()
  AND txo.taxonomy_leaf0_name IS NOT NULL
  AND BTRIM(txo.taxonomy_leaf0_name) <> ''
ORDER BY txo.taxonomy_leaf0_name ASC
`)
	if err != nil {
		return catalogreadmodel.ProductsPortfolioFilters{}, fmt.Errorf("query products portfolio taxonomy filters: %w", err)
	}

	return catalogreadmodel.ProductsPortfolioFilters{
		Brands:             brands,
		TaxonomyLeaf0Names: taxonomyLeaf0Names,
		Status:             statuses,
	}, nil
}

func queryStringList(ctx context.Context, tx *sql.Tx, query string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]string, 0, 16)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func scanProductsPortfolioItem(scanner interface{ Scan(dest ...any) error }) (catalogreadmodel.ProductsPortfolioItem, error) {
	var item catalogreadmodel.ProductsPortfolioItem
	var description sql.NullString
	var brandName sql.NullString
	var pnInterno sql.NullString
	var reference sql.NullString
	var ean sql.NullString
	var taxonomyLeafName sql.NullString
	var taxonomyLeaf0Name sql.NullString
	var stockProfileCode sql.NullString
	var currentPrice sql.NullFloat64
	var replacementCost sql.NullFloat64
	var averageCost sql.NullFloat64
	var currencyCode sql.NullString
	var onHandQuantity sql.NullFloat64
	var inventoryPositionStatus sql.NullString

	if err := scanner.Scan(
		&item.ProductID,
		&item.SKU,
		&item.Name,
		&description,
		&brandName,
		&pnInterno,
		&reference,
		&ean,
		&taxonomyLeafName,
		&taxonomyLeaf0Name,
		&stockProfileCode,
		&item.ProductStatus,
		&currentPrice,
		&replacementCost,
		&averageCost,
		&currencyCode,
		&onHandQuantity,
		&inventoryPositionStatus,
		&item.UpdatedAt,
	); err != nil {
		return catalogreadmodel.ProductsPortfolioItem{}, fmt.Errorf("scan products portfolio item: %w", err)
	}

	item.Description = nullStringPtr(description)
	item.BrandName = nullStringPtr(brandName)
	item.PNInterno = nullStringPtr(pnInterno)
	item.Reference = nullStringPtr(reference)
	item.EAN = nullStringPtr(ean)
	item.TaxonomyLeafName = nullStringPtr(taxonomyLeafName)
	item.TaxonomyLeaf0Name = nullStringPtr(taxonomyLeaf0Name)
	item.StockProfileCode = nullStringPtr(stockProfileCode)
	item.CurrentPriceAmount = nullFloat64Ptr(currentPrice)
	item.ReplacementCostAmount = nullFloat64Ptr(replacementCost)
	item.AverageCostAmount = nullFloat64Ptr(averageCost)
	item.CurrencyCode = nullStringPtr(currencyCode)
	item.OnHandQuantity = nullFloat64Ptr(onHandQuantity)
	item.InventoryPositionStatus = nullStringPtr(inventoryPositionStatus)
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	result := value.String
	return &result
}

func nullFloat64Ptr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	result := value.Float64
	return &result
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
