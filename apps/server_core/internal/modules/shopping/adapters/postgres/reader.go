package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/shopping/ports"
	suppliersapp "metalshopping/server_core/internal/modules/suppliers/application"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

var (
	ErrRunNotFound           = errors.New("shopping run not found")
	ErrProductLatestNotFound = errors.New("shopping product latest not found")
	ErrRunRequestNotFound    = errors.New("shopping run request not found")
)

type Reader struct {
	db               *sql.DB
	suppliersService *suppliersapp.Service
}

func NewReader(db *sql.DB, suppliersService *suppliersapp.Service) *Reader {
	return &Reader{
		db:               db,
		suppliersService: suppliersService,
	}
}

func (r *Reader) GetBootstrap(ctx context.Context, tenantID string) (ports.Bootstrap, error) {
	suppliers := []ports.BootstrapSupplier{}
	if r.suppliersService != nil {
		directory, err := r.suppliersService.ListDirectory(ctx, tenantID, true)
		if err != nil {
			return ports.Bootstrap{}, fmt.Errorf("list suppliers bootstrap directory: %w", err)
		}
		suppliers = make([]ports.BootstrapSupplier, 0, len(directory))
		for _, item := range directory {
			suppliers = append(suppliers, ports.BootstrapSupplier{
				SupplierCode:  item.SupplierCode,
				SupplierLabel: item.SupplierLabel,
				ExecutionKind: item.ExecutionKind,
				LookupPolicy:  item.LookupPolicy,
				Enabled:       item.Enabled,
			})
		}
	}

	return ports.Bootstrap{
		InputModes:     []string{"xlsx", "catalog"},
		RunStatuses:    []string{"queued", "running", "completed", "failed"},
		SupportsManual: true,
		AdvancedDefaults: ports.AdvancedDefaults{
			TimeoutSeconds:   60,
			HTTPWorkers:      10,
			PlaywrightWorker: 7,
			TopN:             5,
		},
		Suppliers: suppliers,
	}, nil
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
ORDER BY observed_at DESC, updated_at DESC
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

func (r *Reader) GetRunRequest(ctx context.Context, tenantID, runRequestID string) (ports.RunRequest, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.RunRequest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT
  run_request_id,
  request_status,
  input_mode,
  input_payload_json,
  requested_at,
  requested_by,
  claimed_at,
  started_at,
  finished_at,
  worker_id,
  run_id,
  error_message
FROM shopping_price_run_requests
WHERE tenant_id = current_tenant_id()
  AND run_request_id = $1
LIMIT 1
`

	var result ports.RunRequest
	var payloadRaw string
	var claimedAt sql.NullTime
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var workerID sql.NullString
	var runID sql.NullString
	var errorMessage sql.NullString

	if err := tx.QueryRowContext(ctx, query, runRequestID).Scan(
		&result.RunRequestID,
		&result.Status,
		&result.InputMode,
		&payloadRaw,
		&result.RequestedAt,
		&result.RequestedBy,
		&claimedAt,
		&startedAt,
		&finishedAt,
		&workerID,
		&runID,
		&errorMessage,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.RunRequest{}, ErrRunRequestNotFound
		}
		return ports.RunRequest{}, fmt.Errorf("query shopping run request: %w", err)
	}
	result.RequestedAt = result.RequestedAt.UTC()
	if claimedAt.Valid {
		value := claimedAt.Time.UTC()
		result.ClaimedAt = &value
	}
	if startedAt.Valid {
		value := startedAt.Time.UTC()
		result.StartedAt = &value
	}
	if finishedAt.Valid {
		value := finishedAt.Time.UTC()
		result.FinishedAt = &value
	}
	if workerID.Valid {
		value := workerID.String
		result.WorkerID = &value
	}
	if runID.Valid {
		value := runID.String
		result.RunID = &value
	}
	if errorMessage.Valid {
		value := errorMessage.String
		result.ErrorMessage = &value
	}

	if strings.TrimSpace(payloadRaw) != "" {
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
			return ports.RunRequest{}, fmt.Errorf("unmarshal shopping run request payload: %w", err)
		}
		result.CatalogProductIDs = extractStringArray(payload["catalogProductIds"])
		result.XLSXScopeIDs = extractStringArray(payload["xlsxScopeIdentifiers"])
		result.ResolvedCatalogProductIDs = extractStringArray(payload["resolvedCatalogProductIds"])
		result.UnresolvedScopeIDs = extractStringArray(payload["unresolvedScopeIdentifiers"])
		result.AmbiguousScopeIDs = extractStringArray(payload["ambiguousScopeIdentifiers"])
		if value := strings.TrimSpace(extractString(payload["xlsxFilePath"])); value != "" {
			result.XLSXFilePath = &value
		}
	}

	if err := tx.Commit(); err != nil {
		return ports.RunRequest{}, fmt.Errorf("commit shopping run request read: %w", err)
	}
	return result, nil
}

func (r *Reader) ListSupplierSignals(ctx context.Context, tenantID string, filter ports.SupplierSignalListFilter) (ports.SupplierSignalList, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.SupplierSignalList{}, err
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
	supplierCode := strings.TrimSpace(filter.SupplierCode)
	productID := strings.TrimSpace(filter.ProductID)

	const countQuery = `
SELECT COUNT(*)
FROM shopping_supplier_product_signals
WHERE tenant_id = current_tenant_id()
  AND ($1 = '' OR supplier_code = $1)
  AND ($2 = '' OR product_id = $2)
`
	var total int64
	if err := tx.QueryRowContext(ctx, countQuery, supplierCode, productID).Scan(&total); err != nil {
		return ports.SupplierSignalList{}, fmt.Errorf("count shopping supplier signals: %w", err)
	}

	const listQuery = `
SELECT
  product_id,
  supplier_code,
  product_url,
  url_status,
  lookup_mode,
  lookup_mode_source,
  manual_override,
  last_checked_at,
  last_success_at,
  last_http_status,
  last_error_message,
  updated_at
FROM shopping_supplier_product_signals
WHERE tenant_id = current_tenant_id()
  AND ($1 = '' OR supplier_code = $1)
  AND ($2 = '' OR product_id = $2)
ORDER BY updated_at DESC, supplier_code, product_id
LIMIT $3 OFFSET $4
`
	rows, err := tx.QueryContext(ctx, listQuery, supplierCode, productID, limit, offset)
	if err != nil {
		return ports.SupplierSignalList{}, fmt.Errorf("list shopping supplier signals: %w", err)
	}
	defer rows.Close()

	items := make([]ports.SupplierSignal, 0, limit)
	for rows.Next() {
		item, scanErr := scanSupplierSignal(rows)
		if scanErr != nil {
			return ports.SupplierSignalList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.SupplierSignalList{}, fmt.Errorf("iterate shopping supplier signals: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.SupplierSignalList{}, fmt.Errorf("commit shopping supplier signal list read: %w", err)
	}

	return ports.SupplierSignalList{
		Rows:   items,
		Offset: offset,
		Limit:  limit,
		Total:  total,
	}, nil
}

func extractString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func extractStringArray(value any) []string {
	rawValues, ok := value.([]any)
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(rawValues))
	for _, item := range rawValues {
		text, ok := item.(string)
		if !ok {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		result = append(result, text)
	}
	return result
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

func scanSupplierSignal(s scanner) (ports.SupplierSignal, error) {
	var item ports.SupplierSignal
	var productURL sql.NullString
	var lastCheckedAt sql.NullTime
	var lastSuccessAt sql.NullTime
	var lastHTTPStatus sql.NullInt64
	var lastErrorMessage sql.NullString
	if err := s.Scan(
		&item.ProductID,
		&item.SupplierCode,
		&productURL,
		&item.URLStatus,
		&item.LookupMode,
		&item.LookupModeSource,
		&item.ManualOverride,
		&lastCheckedAt,
		&lastSuccessAt,
		&lastHTTPStatus,
		&lastErrorMessage,
		&item.UpdatedAt,
	); err != nil {
		return ports.SupplierSignal{}, err
	}
	if productURL.Valid {
		value := productURL.String
		item.ProductURL = &value
	}
	if lastCheckedAt.Valid {
		value := lastCheckedAt.Time.UTC()
		item.LastCheckedAt = &value
	}
	if lastSuccessAt.Valid {
		value := lastSuccessAt.Time.UTC()
		item.LastSuccessAt = &value
	}
	if lastHTTPStatus.Valid {
		value := lastHTTPStatus.Int64
		item.LastHTTPStatus = &value
	}
	if lastErrorMessage.Valid {
		value := lastErrorMessage.String
		item.LastErrorMessage = &value
	}
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}
