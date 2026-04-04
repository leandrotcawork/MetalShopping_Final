package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/pricing/domain"
	pricingevents "metalshopping/server_core/internal/modules/pricing/events"
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

func (r *Repository) CreateProductPrice(ctx context.Context, price domain.ProductPrice, traceID string) (domain.ProductPrice, bool, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, price.TenantID, nil)
	if err != nil {
		return domain.ProductPrice{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	priceContextCode := domain.NormalizePriceContextCode(price.PriceContextCode)
	price.PriceContextCode = priceContextCode

	const currentOpenSQL = `
SELECT price_id, tenant_id, product_id, price_context_code, currency_code, price_amount, replacement_cost_amount, average_cost_amount, pricing_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM pricing_product_prices
WHERE tenant_id = current_tenant_id()
  AND product_id = $1
  AND price_context_code = $2
  AND effective_to IS NULL
ORDER BY effective_from DESC, created_at DESC
LIMIT 1
`
	currentRow := tx.QueryRowContext(ctx, currentOpenSQL, price.ProductID, priceContextCode)
	currentOpen, err := scanProductPrice(currentRow)
	switch {
	case err == nil:
		if currentOpen.HasSameCommercialState(price) {
			if err := tx.Commit(); err != nil {
				return domain.ProductPrice{}, false, fmt.Errorf("commit pricing product price no-op: %w", err)
			}
			return currentOpen, false, nil
		}
	case errors.Is(err, sql.ErrNoRows):
	default:
		return domain.ProductPrice{}, false, fmt.Errorf("load current open pricing product price: %w", err)
	}

	const closeOpenWindowSQL = `
UPDATE pricing_product_prices
SET effective_to = $2,
    updated_at = $3
WHERE tenant_id = current_tenant_id()
  AND product_id = $1
  AND price_context_code = $4
  AND effective_to IS NULL
`
	if _, err := tx.ExecContext(ctx, closeOpenWindowSQL, price.ProductID, price.EffectiveFrom, price.UpdatedAt, priceContextCode); err != nil {
		return domain.ProductPrice{}, false, fmt.Errorf("close pricing open window: %w", err)
	}

	const insertSQL = `
INSERT INTO pricing_product_prices (
  price_id,
  tenant_id,
  product_id,
  price_context_code,
  currency_code,
  price_amount,
  replacement_cost_amount,
  average_cost_amount,
  pricing_status,
  effective_from,
  effective_to,
  origin_type,
  origin_ref,
  reason_code,
  updated_by,
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
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16
)
`
	if _, err := tx.ExecContext(
		ctx,
		insertSQL,
		price.PriceID,
		price.ProductID,
		priceContextCode,
		price.CurrencyCode,
		price.PriceAmount,
		price.ReplacementCostAmount,
		nullableFloat(price.AverageCostAmount),
		string(price.PricingStatus),
		price.EffectiveFrom,
		nullableTime(price.EffectiveTo),
		string(price.OriginType),
		nullableText(price.OriginRef),
		price.ReasonCode,
		price.UpdatedBy,
		price.CreatedAt,
		price.UpdatedAt,
	); err != nil {
		return domain.ProductPrice{}, false, fmt.Errorf("insert pricing product price: %w", err)
	}

	if r.outboxStore != nil {
		record, err := pricingevents.NewPriceSetOutboxRecord(price, traceID, price.CreatedAt)
		if err != nil {
			return domain.ProductPrice{}, false, err
		}
		if err := r.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
			return domain.ProductPrice{}, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.ProductPrice{}, false, fmt.Errorf("commit pricing product price: %w", err)
	}
	return price, true, nil
}

func (r *Repository) ListProductPrices(ctx context.Context, tenantID, productID string) ([]domain.ProductPrice, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT price_id, tenant_id, product_id, price_context_code, currency_code, price_amount, replacement_cost_amount, average_cost_amount, pricing_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM pricing_product_prices
WHERE product_id = $1
  AND price_context_code = $2
ORDER BY effective_from DESC, created_at DESC
`
	rows, err := tx.QueryContext(ctx, querySQL, productID, domain.DefaultPriceContextCode)
	if err != nil {
		return nil, fmt.Errorf("query pricing product prices: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ProductPrice, 0, 8)
	for rows.Next() {
		item, err := scanProductPrice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pricing product prices: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pricing product price list: %w", err)
	}
	return items, nil
}

func (r *Repository) GetCurrentProductPrice(ctx context.Context, tenantID, productID string) (domain.ProductPrice, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return domain.ProductPrice{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT price_id, tenant_id, product_id, price_context_code, currency_code, price_amount, replacement_cost_amount, average_cost_amount, pricing_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM pricing_product_prices
WHERE product_id = $1
  AND price_context_code = $2
  AND effective_from <= NOW()
  AND (effective_to IS NULL OR effective_to > NOW())
ORDER BY effective_from DESC, created_at DESC
LIMIT 1
`
	row := tx.QueryRowContext(ctx, querySQL, productID, domain.DefaultPriceContextCode)
	item, err := scanProductPrice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProductPrice{}, domain.ErrProductPriceNotFound
		}
		return domain.ProductPrice{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ProductPrice{}, fmt.Errorf("commit current pricing product price: %w", err)
	}
	return item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProductPrice(s scanner) (domain.ProductPrice, error) {
	var item domain.ProductPrice
	var status string
	var originType string
	var effectiveTo sql.NullTime
	var averageCost sql.NullFloat64
	if err := s.Scan(
		&item.PriceID,
		&item.TenantID,
		&item.ProductID,
		&item.PriceContextCode,
		&item.CurrencyCode,
		&item.PriceAmount,
		&item.ReplacementCostAmount,
		&averageCost,
		&status,
		&item.EffectiveFrom,
		&effectiveTo,
		&originType,
		&item.OriginRef,
		&item.ReasonCode,
		&item.UpdatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return domain.ProductPrice{}, err
	}
	item.PricingStatus = domain.PricingStatus(status)
	item.OriginType = domain.OriginType(originType)
	if averageCost.Valid {
		value := averageCost.Float64
		item.AverageCostAmount = &value
	}
	if effectiveTo.Valid {
		value := effectiveTo.Time.UTC()
		item.EffectiveTo = &value
	}
	return item, nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func nullableFloat(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}
