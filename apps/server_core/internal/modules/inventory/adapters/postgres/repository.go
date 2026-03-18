package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/inventory/domain"
	inventoryevents "metalshopping/server_core/internal/modules/inventory/events"
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

func (r *Repository) CreateProductPosition(ctx context.Context, position domain.ProductPosition, traceID string) (domain.ProductPosition, bool, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, position.TenantID, nil)
	if err != nil {
		return domain.ProductPosition{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	const currentOpenSQL = `
SELECT position_id, tenant_id, product_id, on_hand_quantity, last_purchase_at, last_sale_at, position_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM inventory_product_positions
WHERE tenant_id = current_tenant_id()
  AND product_id = $1
  AND effective_to IS NULL
ORDER BY effective_from DESC, created_at DESC
LIMIT 1
`
	currentRow := tx.QueryRowContext(ctx, currentOpenSQL, position.ProductID)
	currentOpen, err := scanProductPosition(currentRow)
	switch {
	case err == nil:
		if currentOpen.HasSameOperationalState(position) {
			if err := tx.Commit(); err != nil {
				return domain.ProductPosition{}, false, fmt.Errorf("commit inventory product position no-op: %w", err)
			}
			return currentOpen, false, nil
		}
	case errors.Is(err, sql.ErrNoRows):
	default:
		return domain.ProductPosition{}, false, fmt.Errorf("load current open inventory product position: %w", err)
	}

	const closeOpenWindowSQL = `
UPDATE inventory_product_positions
SET effective_to = $2,
    updated_at = $3
WHERE tenant_id = current_tenant_id()
  AND product_id = $1
  AND effective_to IS NULL
`
	if _, err := tx.ExecContext(ctx, closeOpenWindowSQL, position.ProductID, position.EffectiveFrom, position.UpdatedAt); err != nil {
		return domain.ProductPosition{}, false, fmt.Errorf("close inventory open window: %w", err)
	}

	const insertSQL = `
INSERT INTO inventory_product_positions (
  position_id,
  tenant_id,
  product_id,
  on_hand_quantity,
  last_purchase_at,
  last_sale_at,
  position_status,
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
  $14
)
`
	if _, err := tx.ExecContext(
		ctx,
		insertSQL,
		position.PositionID,
		position.ProductID,
		position.OnHandQuantity,
		nullableTime(position.LastPurchaseAt),
		nullableTime(position.LastSaleAt),
		string(position.PositionStatus),
		position.EffectiveFrom,
		nullableTime(position.EffectiveTo),
		string(position.OriginType),
		nullableText(position.OriginRef),
		position.ReasonCode,
		position.UpdatedBy,
		position.CreatedAt,
		position.UpdatedAt,
	); err != nil {
		return domain.ProductPosition{}, false, fmt.Errorf("insert inventory product position: %w", err)
	}

	if r.outboxStore != nil {
		record, err := inventoryevents.NewPositionUpdatedOutboxRecord(position, traceID, position.CreatedAt)
		if err != nil {
			return domain.ProductPosition{}, false, err
		}
		if err := r.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
			return domain.ProductPosition{}, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.ProductPosition{}, false, fmt.Errorf("commit inventory product position: %w", err)
	}
	return position, true, nil
}

func (r *Repository) ListProductPositions(ctx context.Context, tenantID, productID string) ([]domain.ProductPosition, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT position_id, tenant_id, product_id, on_hand_quantity, last_purchase_at, last_sale_at, position_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM inventory_product_positions
WHERE product_id = $1
ORDER BY effective_from DESC, created_at DESC
`
	rows, err := tx.QueryContext(ctx, querySQL, productID)
	if err != nil {
		return nil, fmt.Errorf("query inventory product positions: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ProductPosition, 0, 8)
	for rows.Next() {
		item, err := scanProductPosition(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory product positions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit inventory position list: %w", err)
	}
	return items, nil
}

func (r *Repository) GetCurrentProductPosition(ctx context.Context, tenantID, productID string) (domain.ProductPosition, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return domain.ProductPosition{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT position_id, tenant_id, product_id, on_hand_quantity, last_purchase_at, last_sale_at, position_status, effective_from, effective_to, origin_type, COALESCE(origin_ref, ''), reason_code, updated_by, created_at, updated_at
FROM inventory_product_positions
WHERE product_id = $1
  AND effective_from <= NOW()
  AND (effective_to IS NULL OR effective_to > NOW())
ORDER BY effective_from DESC, created_at DESC
LIMIT 1
`
	row := tx.QueryRowContext(ctx, querySQL, productID)
	item, err := scanProductPosition(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProductPosition{}, domain.ErrProductPositionNotFound
		}
		return domain.ProductPosition{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ProductPosition{}, fmt.Errorf("commit current inventory position: %w", err)
	}
	return item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProductPosition(s scanner) (domain.ProductPosition, error) {
	var item domain.ProductPosition
	var status string
	var originType string
	var lastPurchase sql.NullTime
	var lastSale sql.NullTime
	var effectiveTo sql.NullTime
	if err := s.Scan(
		&item.PositionID,
		&item.TenantID,
		&item.ProductID,
		&item.OnHandQuantity,
		&lastPurchase,
		&lastSale,
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
		return domain.ProductPosition{}, err
	}
	item.PositionStatus = domain.PositionStatus(status)
	item.OriginType = domain.OriginType(originType)
	if lastPurchase.Valid {
		value := lastPurchase.Time.UTC()
		item.LastPurchaseAt = &value
	}
	if lastSale.Valid {
		value := lastSale.Time.UTC()
		item.LastSaleAt = &value
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
