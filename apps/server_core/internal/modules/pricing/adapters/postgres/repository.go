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

func (r *Repository) CreateProductPrice(ctx context.Context, price domain.ProductPrice, traceID string) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, price.TenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const closeOpenWindowSQL = `
UPDATE pricing_product_prices
SET effective_to = $2,
    updated_at = $3
WHERE tenant_id = current_tenant_id()
  AND product_id = $1
  AND effective_to IS NULL
`
	if _, err := tx.ExecContext(ctx, closeOpenWindowSQL, price.ProductID, price.EffectiveFrom, price.UpdatedAt); err != nil {
		return fmt.Errorf("close pricing open window: %w", err)
	}

	const insertSQL = `
INSERT INTO pricing_product_prices (
  price_id,
  tenant_id,
  product_id,
  currency_code,
  price_amount,
  cost_basis_amount,
  margin_floor_value,
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
  $15
)
`
	if _, err := tx.ExecContext(
		ctx,
		insertSQL,
		price.PriceID,
		price.ProductID,
		price.CurrencyCode,
		price.PriceAmount,
		price.CostBasisAmount,
		price.MarginFloorValue,
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
		return fmt.Errorf("insert pricing product price: %w", err)
	}

	if r.outboxStore != nil {
		record, err := pricingevents.NewPriceSetOutboxRecord(price, traceID, price.CreatedAt)
		if err != nil {
			return err
		}
		if err := r.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pricing product price: %w", err)
	}
	return nil
}

func (r *Repository) ListProductPrices(ctx context.Context, tenantID, productID string) ([]domain.ProductPrice, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT price_id, tenant_id, product_id, currency_code, price_amount, cost_basis_amount, margin_floor_value, pricing_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM pricing_product_prices
WHERE product_id = $1
ORDER BY effective_from DESC, created_at DESC
`
	rows, err := tx.QueryContext(ctx, querySQL, productID)
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
SELECT price_id, tenant_id, product_id, currency_code, price_amount, cost_basis_amount, margin_floor_value, pricing_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM pricing_product_prices
WHERE product_id = $1
  AND effective_from <= NOW()
  AND (effective_to IS NULL OR effective_to > NOW())
ORDER BY effective_from DESC, created_at DESC
LIMIT 1
`
	row := tx.QueryRowContext(ctx, querySQL, productID)
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
	if err := s.Scan(
		&item.PriceID,
		&item.TenantID,
		&item.ProductID,
		&item.CurrencyCode,
		&item.PriceAmount,
		&item.CostBasisAmount,
		&item.MarginFloorValue,
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
