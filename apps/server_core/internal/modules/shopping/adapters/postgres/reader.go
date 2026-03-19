package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/shopping/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

var (
	ErrRunNotFound           = errors.New("shopping run not found")
	ErrProductLatestNotFound = errors.New("shopping product latest not found")
)

type Reader struct {
	db *sql.DB
}

func NewReader(db *sql.DB) *Reader {
	return &Reader{db: db}
}

func (r *Reader) GetSummary(ctx context.Context, tenantID string) (ports.Summary, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.Summary{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT
  COUNT(*) AS total_runs,
  COUNT(*) FILTER (WHERE run_status = 'running') AS running_runs,
  COUNT(*) FILTER (WHERE run_status = 'completed') AS completed_runs,
  COUNT(*) FILTER (WHERE run_status = 'failed') AS failed_runs,
  MAX(started_at) AS last_run_at
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
`

	var summary ports.Summary
	var lastRunAt sql.NullTime
	if err := tx.QueryRowContext(ctx, query).Scan(
		&summary.TotalRuns,
		&summary.RunningRuns,
		&summary.CompletedRuns,
		&summary.FailedRuns,
		&lastRunAt,
	); err != nil {
		return ports.Summary{}, fmt.Errorf("query shopping summary: %w", err)
	}
	if lastRunAt.Valid {
		value := lastRunAt.Time.UTC()
		summary.LastRunAt = &value
	} else {
		now := time.Now().UTC()
		summary.LastRunAt = &now
	}

	if err := tx.Commit(); err != nil {
		return ports.Summary{}, fmt.Errorf("commit shopping summary read: %w", err)
	}
	return summary, nil
}

func (r *Reader) ListRuns(ctx context.Context, tenantID string, filter ports.RunListFilter) (ports.RunList, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.RunList{}, err
	}
	defer func() { _ = tx.Rollback() }()

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	status := strings.TrimSpace(filter.Status)
	if status == "" {
		status = "all"
	}

	const countQuery = `
SELECT COUNT(*)
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND ($1 = 'all' OR run_status = $1)
`
	var total int64
	if err := tx.QueryRowContext(ctx, countQuery, status).Scan(&total); err != nil {
		return ports.RunList{}, fmt.Errorf("count shopping runs: %w", err)
	}

	const listQuery = `
SELECT
  run_id,
  run_status,
  started_at,
  finished_at,
  processed_items,
  total_items,
  COALESCE(notes, '')
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND ($1 = 'all' OR run_status = $1)
ORDER BY started_at DESC, created_at DESC
LIMIT $2 OFFSET $3
`
	rows, err := tx.QueryContext(ctx, listQuery, status, limit, offset)
	if err != nil {
		return ports.RunList{}, fmt.Errorf("list shopping runs: %w", err)
	}
	defer rows.Close()

	items := make([]ports.Run, 0, limit)
	for rows.Next() {
		item, scanErr := scanRun(rows)
		if scanErr != nil {
			return ports.RunList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.RunList{}, fmt.Errorf("iterate shopping runs: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.RunList{}, fmt.Errorf("commit shopping run list read: %w", err)
	}

	return ports.RunList{
		Rows:   items,
		Offset: offset,
		Limit:  limit,
		Total:  total,
	}, nil
}

func (r *Reader) GetRun(ctx context.Context, tenantID, runID string) (ports.Run, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.Run{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT
  run_id,
  run_status,
  started_at,
  finished_at,
  processed_items,
  total_items,
  COALESCE(notes, '')
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
LIMIT 1
`
	row := tx.QueryRowContext(ctx, query, runID)
	run, err := scanRun(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.Run{}, ErrRunNotFound
		}
		return ports.Run{}, err
	}

	if err := tx.Commit(); err != nil {
		return ports.Run{}, fmt.Errorf("commit shopping run detail read: %w", err)
	}
	return run, nil
}

func (r *Reader) GetProductLatest(ctx context.Context, tenantID, productID string) (ports.ProductLatest, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.ProductLatest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT
  product_id,
  run_id,
  observed_at,
  seller_name,
  channel,
  observed_price,
  currency_code
FROM shopping_price_latest_snapshot
WHERE tenant_id = current_tenant_id()
  AND product_id = $1
LIMIT 1
`
	var item ports.ProductLatest
	if err := tx.QueryRowContext(ctx, query, productID).Scan(
		&item.ProductID,
		&item.RunID,
		&item.ObservedAt,
		&item.SellerName,
		&item.Channel,
		&item.ObservedPrice,
		&item.Currency,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.ProductLatest{}, ErrProductLatestNotFound
		}
		return ports.ProductLatest{}, fmt.Errorf("query shopping product latest: %w", err)
	}
	item.ObservedAt = item.ObservedAt.UTC()

	if err := tx.Commit(); err != nil {
		return ports.ProductLatest{}, fmt.Errorf("commit shopping product latest read: %w", err)
	}
	return item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRun(s scanner) (ports.Run, error) {
	var item ports.Run
	var finishedAt sql.NullTime
	if err := s.Scan(
		&item.RunID,
		&item.Status,
		&item.StartedAt,
		&finishedAt,
		&item.ProcessedItems,
		&item.TotalItems,
		&item.Notes,
	); err != nil {
		return ports.Run{}, err
	}
	item.StartedAt = item.StartedAt.UTC()
	if finishedAt.Valid {
		value := finishedAt.Time.UTC()
		item.FinishedAt = &value
	}
	return item, nil
}
