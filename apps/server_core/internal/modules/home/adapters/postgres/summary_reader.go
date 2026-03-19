package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/home/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type SummaryReader struct {
	db *sql.DB
}

func NewSummaryReader(db *sql.DB) *SummaryReader {
	return &SummaryReader{db: db}
}

func (r *SummaryReader) GetSummary(ctx context.Context, tenantID string) (ports.Summary, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.Summary{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT
  (SELECT COUNT(*) FROM catalog_products WHERE tenant_id = current_tenant_id()) AS product_count,
  (SELECT COUNT(*) FROM catalog_products WHERE tenant_id = current_tenant_id() AND status = 'active') AS active_product_count,
  (SELECT COUNT(*) FROM pricing_product_prices WHERE tenant_id = current_tenant_id() AND effective_to IS NULL) AS priced_product_count,
  (SELECT COUNT(*) FROM inventory_product_positions WHERE tenant_id = current_tenant_id() AND effective_to IS NULL) AS inventory_tracked_count,
  (
    SELECT MAX(updated_at)
    FROM (
      SELECT MAX(updated_at) AS updated_at FROM catalog_products WHERE tenant_id = current_tenant_id()
      UNION ALL
      SELECT MAX(updated_at) AS updated_at FROM pricing_product_prices WHERE tenant_id = current_tenant_id()
      UNION ALL
      SELECT MAX(updated_at) AS updated_at FROM inventory_product_positions WHERE tenant_id = current_tenant_id()
    ) all_updates
  ) AS last_updated
`

	var summary ports.Summary
	var lastUpdated sql.NullTime
	if err := tx.QueryRowContext(ctx, query).Scan(
		&summary.ProductCount,
		&summary.ActiveProductCount,
		&summary.PricedProductCount,
		&summary.InventoryTrackedCount,
		&lastUpdated,
	); err != nil {
		return ports.Summary{}, fmt.Errorf("query home summary: %w", err)
	}
	summary.LastUpdated = time.Now().UTC()
	if lastUpdated.Valid {
		summary.LastUpdated = lastUpdated.Time.UTC()
	}

	if err := tx.Commit(); err != nil {
		return ports.Summary{}, fmt.Errorf("commit home summary read: %w", err)
	}
	return summary, nil
}
