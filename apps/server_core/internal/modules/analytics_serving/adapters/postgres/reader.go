package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/analytics_serving/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type Reader struct {
	db *sql.DB
}

func NewReader(db *sql.DB) *Reader {
	return &Reader{db: db}
}

func (r *Reader) GetHome(ctx context.Context, tenantID string, requestedSnapshotID string) (ports.Home, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.Home{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const operationalQuery = `
SELECT
  (SELECT COUNT(*) FROM catalog_products WHERE tenant_id = current_tenant_id()) AS products_registered,
  (SELECT COUNT(*) FROM suppliers_directory WHERE tenant_id = current_tenant_id() AND enabled = true) AS suppliers_available,
  (
    SELECT r.run_id
    FROM shopping_price_runs r
    WHERE r.tenant_id = current_tenant_id() AND r.run_status = 'completed'
    ORDER BY r.started_at DESC
    LIMIT 1
  ) AS latest_completed_run_id,
  (
    SELECT MAX(i.observed_at)
    FROM shopping_price_run_items i
    JOIN shopping_price_runs r ON r.run_id = i.run_id
    WHERE i.tenant_id = current_tenant_id()
      AND r.tenant_id = current_tenant_id()
      AND r.run_status = 'completed'
  ) AS as_of,
  (
    SELECT COUNT(*)
    FROM shopping_price_run_items i
    JOIN shopping_price_runs r ON r.run_id = i.run_id
    WHERE i.tenant_id = current_tenant_id()
      AND r.tenant_id = current_tenant_id()
      AND r.run_status = 'completed'
      AND r.run_id = (
        SELECT rr.run_id
        FROM shopping_price_runs rr
        WHERE rr.tenant_id = current_tenant_id() AND rr.run_status = 'completed'
        ORDER BY rr.started_at DESC
        LIMIT 1
      )
  ) AS skus_last_run,
  (
    SELECT COUNT(*)
    FROM shopping_price_run_items i
    JOIN shopping_price_runs r ON r.run_id = i.run_id
    WHERE i.tenant_id = current_tenant_id()
      AND r.tenant_id = current_tenant_id()
      AND r.run_status = 'completed'
      AND r.run_id = (
        SELECT rr.run_id
        FROM shopping_price_runs rr
        WHERE rr.tenant_id = current_tenant_id() AND rr.run_status = 'completed'
        ORDER BY rr.started_at DESC
        LIMIT 1
      )
      AND i.item_status = 'OK'
  ) AS ok_last_run
`

	var productsRegistered int64
	var suppliersAvailable int64
	var latestCompletedRunID sql.NullString
	var asOf sql.NullTime
	var skusLastRun int64
	var okLastRun int64
	if err := tx.QueryRowContext(ctx, operationalQuery).Scan(
		&productsRegistered,
		&suppliersAvailable,
		&latestCompletedRunID,
		&asOf,
		&skusLastRun,
		&okLastRun,
	); err != nil {
		return ports.Home{}, fmt.Errorf("query analytics home operational block: %w", err)
	}

	const productsMetricsQuery = `
SELECT
  (SELECT COUNT(*) FROM catalog_products WHERE tenant_id = current_tenant_id() AND status = 'active') AS products_active_count,
  (SELECT COALESCE(SUM(price_amount), 0) FROM pricing_product_prices WHERE tenant_id = current_tenant_id() AND effective_to IS NULL) AS capital_brl_total,
  (SELECT COALESCE(AVG(price_amount), 0) FROM pricing_product_prices WHERE tenant_id = current_tenant_id() AND effective_to IS NULL) AS weighted_margin_pct_total
`

	var productsActiveCount int64
	var capitalTotal float64
	var weightedMargin float64
	if err := tx.QueryRowContext(ctx, productsMetricsQuery).Scan(
		&productsActiveCount,
		&capitalTotal,
		&weightedMargin,
	); err != nil {
		return ports.Home{}, fmt.Errorf("query analytics home products block: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.Home{}, fmt.Errorf("commit analytics home read: %w", err)
	}

	var resolvedID *string
	if latestCompletedRunID.Valid {
		resolvedID = &latestCompletedRunID.String
	}

	var asOfValue *time.Time
	if asOf.Valid {
		value := asOf.Time.UTC()
		asOfValue = &value
	}

	successRate := 0.0
	if skusLastRun > 0 {
		successRate = float64(okLastRun) / float64(skusLastRun)
	}

	operationalData := map[string]any{
		"productsRegistered": productsRegistered,
		"suppliersAvailable": suppliersAvailable,
		"successRateLastRun": successRate,
		"skusLastRun":        skusLastRun,
	}
	productsData := map[string]any{
		"productsActiveCount":      productsActiveCount,
		"capitalBrlTotal":          capitalTotal,
		"weightedMarginPctTotal":   weightedMargin,
		"potentialRevenueBrlTotal": nil,
	}

	return ports.Home{
		SchemaVersion: "1.0",
		Snapshot: ports.HomeSnapshot{
			RequestedID: requestedSnapshotIDOrDefault(requestedSnapshotID),
			ResolvedID:  resolvedID,
			AsOf:        asOfValue,
			ServedAt:    time.Now().UTC(),
		},
		Blocks: ports.HomeBlocks{
			KpisOperational:       okBlock(operationalData),
			KpisAnalytics:         notReadyBlock("ANALYTICS_KPI_NOT_READY", "Analytics KPI block is not ready"),
			KpisProducts:          okBlock(productsData),
			ActionsToday:          notReadyBlock("ANALYTICS_ACTIONS_NOT_READY", "Actions block is not ready"),
			AlertsPrioritarios:    notReadyBlock("ANALYTICS_ALERTS_NOT_READY", "Alerts block is not ready"),
			PortfolioDistribution: notReadyBlock("ANALYTICS_PORTFOLIO_NOT_READY", "Portfolio distribution block is not ready"),
			Timeline:              notReadyBlock("ANALYTICS_TIMELINE_NOT_READY", "Timeline block is not ready"),
		},
	}, nil
}

func requestedSnapshotIDOrDefault(requestedSnapshotID string) string {
	if requestedSnapshotID == "" {
		return "current"
	}
	return requestedSnapshotID
}

func okBlock(data map[string]any) ports.HomeBlock {
	return ports.HomeBlock{
		Status: ports.BlockStatusOK,
		Data:   data,
		Error:  nil,
	}
}

func notReadyBlock(code string, message string) ports.HomeBlock {
	return ports.HomeBlock{
		Status: ports.BlockStatusNotReady,
		Data:   nil,
		Error: &ports.BlockError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
		},
	}
}
