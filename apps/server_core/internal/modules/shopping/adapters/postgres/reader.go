package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
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

func (r *Reader) GetRunItemStatusSummary(ctx context.Context, tenantID, runID string) (ports.RunItemStatusSummary, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.RunItemStatusSummary{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const runExistsQuery = `
SELECT 1
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
LIMIT 1
`
	var one int
	if err := tx.QueryRowContext(ctx, runExistsQuery, runID).Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.RunItemStatusSummary{}, ErrRunNotFound
		}
		return ports.RunItemStatusSummary{}, fmt.Errorf("check shopping run exists: %w", err)
	}

	const query = `
SELECT
  item_status,
  COUNT(*) AS total
FROM shopping_price_run_items
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
GROUP BY item_status
ORDER BY COUNT(*) DESC, item_status ASC
`

	rows, err := tx.QueryContext(ctx, query, runID)
	if err != nil {
		return ports.RunItemStatusSummary{}, fmt.Errorf("query shopping run item status summary: %w", err)
	}
	defer rows.Close()

	items := make([]ports.RunItemStatusCount, 0, 8)
	var totalItems int64
	for rows.Next() {
		var item ports.RunItemStatusCount
		if err := rows.Scan(&item.ItemStatus, &item.Total); err != nil {
			return ports.RunItemStatusSummary{}, fmt.Errorf("scan shopping run item status summary: %w", err)
		}
		item.ItemStatus = strings.TrimSpace(item.ItemStatus)
		items = append(items, item)
		totalItems += item.Total
	}
	if err := rows.Err(); err != nil {
		return ports.RunItemStatusSummary{}, fmt.Errorf("iterate shopping run item status summary: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.RunItemStatusSummary{}, fmt.Errorf("commit shopping run item status summary read: %w", err)
	}

	return ports.RunItemStatusSummary{
		RunID:      runID,
		TotalItems: totalItems,
		Rows:       items,
	}, nil
}

func (r *Reader) GetRunSupplierItemStatusSummary(
	ctx context.Context,
	tenantID string,
	runID string,
) (ports.RunSupplierItemStatusSummary, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.RunSupplierItemStatusSummary{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const runExistsQuery = `
SELECT 1
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
LIMIT 1
`
	var one int
	if err := tx.QueryRowContext(ctx, runExistsQuery, runID).Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.RunSupplierItemStatusSummary{}, ErrRunNotFound
		}
		return ports.RunSupplierItemStatusSummary{}, fmt.Errorf("check shopping run exists: %w", err)
	}

	const query = `
SELECT
  supplier_code,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE item_status = 'OK') AS ok,
  COUNT(*) FILTER (WHERE item_status = 'NOT_FOUND') AS not_found,
  COUNT(*) FILTER (WHERE item_status = 'AMBIGUOUS') AS ambiguous,
  COUNT(*) FILTER (WHERE item_status = 'ERROR') AS error
FROM shopping_price_run_items
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
GROUP BY supplier_code
ORDER BY supplier_code ASC
`

	rows, err := tx.QueryContext(ctx, query, runID)
	if err != nil {
		return ports.RunSupplierItemStatusSummary{}, fmt.Errorf("query shopping run supplier status summary: %w", err)
	}
	defer rows.Close()

	items := make([]ports.RunSupplierItemStatusCount, 0, 16)
	for rows.Next() {
		var item ports.RunSupplierItemStatusCount
		if err := rows.Scan(
			&item.SupplierCode,
			&item.Total,
			&item.Ok,
			&item.NotFound,
			&item.Ambiguous,
			&item.Error,
		); err != nil {
			return ports.RunSupplierItemStatusSummary{}, fmt.Errorf("scan shopping run supplier status summary: %w", err)
		}
		item.SupplierCode = strings.TrimSpace(item.SupplierCode)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.RunSupplierItemStatusSummary{}, fmt.Errorf("iterate shopping run supplier status summary: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.RunSupplierItemStatusSummary{}, fmt.Errorf("commit shopping run supplier status summary read: %w", err)
	}

	return ports.RunSupplierItemStatusSummary{
		RunID:          runID,
		TotalSuppliers: int64(len(items)),
		Rows:           items,
	}, nil
}

func (r *Reader) ListRunItems(
	ctx context.Context,
	tenantID string,
	runID string,
	filter ports.RunItemListFilter,
) (ports.RunItemList, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.RunItemList{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const runExistsQuery = `
SELECT 1
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
LIMIT 1
`
	var one int
	if err := tx.QueryRowContext(ctx, runExistsQuery, runID).Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.RunItemList{}, ErrRunNotFound
		}
		return ports.RunItemList{}, fmt.Errorf("check shopping run exists: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	supplierCode := strings.TrimSpace(filter.SupplierCode)
	itemStatus := strings.TrimSpace(filter.ItemStatus)

	const countQuery = `
SELECT COUNT(*)
FROM shopping_price_run_items
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
  AND ($2 = '' OR supplier_code = $2)
  AND ($3 = '' OR item_status = $3)
`
	var total int64
	if err := tx.QueryRowContext(ctx, countQuery, runID, supplierCode, itemStatus).Scan(&total); err != nil {
		return ports.RunItemList{}, fmt.Errorf("count shopping run items: %w", err)
	}

	const listQuery = `
SELECT
  i.run_item_id,
  i.run_id,
  i.product_id,
  p.name,
  i.supplier_code,
  i.item_status,
  i.observed_price,
  i.currency_code,
  i.observed_at,
  i.seller_name,
  i.channel,
  i.product_url,
  i.http_status,
  i.elapsed_s,
  i.chosen_seller_json->>'lookup_term',
  i.notes,
  m.config_json::text
FROM shopping_price_run_items i
JOIN catalog_products p
  ON p.tenant_id = i.tenant_id
 AND p.product_id = i.product_id
LEFT JOIN supplier_driver_manifests m
  ON m.tenant_id = i.tenant_id
 AND m.supplier_code = i.supplier_code
 AND m.is_active = TRUE
 AND m.validation_status = 'valid'
WHERE i.tenant_id = current_tenant_id()
  AND i.run_id = $1
  AND ($2 = '' OR i.supplier_code = $2)
  AND ($3 = '' OR i.item_status = $3)
ORDER BY i.observed_at DESC, i.created_at DESC
LIMIT $4 OFFSET $5
`
	rows, err := tx.QueryContext(ctx, listQuery, runID, supplierCode, itemStatus, limit, offset)
	if err != nil {
		return ports.RunItemList{}, fmt.Errorf("list shopping run items: %w", err)
	}
	defer rows.Close()

	items := make([]ports.RunItem, 0, limit)
	for rows.Next() {
		item, scanErr := scanRunItem(rows)
		if scanErr != nil {
			return ports.RunItemList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.RunItemList{}, fmt.Errorf("iterate shopping run items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.RunItemList{}, fmt.Errorf("commit shopping run items read: %w", err)
	}

	return ports.RunItemList{
		Rows:   items,
		Offset: offset,
		Limit:  limit,
		Total:  total,
	}, nil
}

func (r *Reader) ListRunItemsForExport(
	ctx context.Context,
	tenantID string,
	runID string,
	filter ports.RunExportListFilter,
) (ports.RunExportList, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.RunExportList{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const runExistsQuery = `
SELECT 1
FROM shopping_price_runs
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
LIMIT 1
`
	var one int
	if err := tx.QueryRowContext(ctx, runExistsQuery, runID).Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.RunExportList{}, ErrRunNotFound
		}
		return ports.RunExportList{}, fmt.Errorf("check shopping run exists: %w", err)
	}

	supplierCodes := make([]string, 0, len(filter.SupplierCodes))
	seenSuppliers := map[string]struct{}{}
	for _, code := range filter.SupplierCodes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		if _, exists := seenSuppliers[code]; exists {
			continue
		}
		seenSuppliers[code] = struct{}{}
		supplierCodes = append(supplierCodes, code)
	}

	whereSQL := "WHERE i.tenant_id = current_tenant_id()\n  AND i.run_id = $1"
	args := []any{runID}
	argPos := 2
	if len(supplierCodes) > 0 {
		whereSQL += fmt.Sprintf("\n  AND i.supplier_code = ANY($%d)", argPos)
		args = append(args, supplierCodes)
		argPos++
	}

	countQuery := `
SELECT COUNT(*)
FROM shopping_price_run_items i
` + whereSQL

	var total int64
	if err := tx.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return ports.RunExportList{}, fmt.Errorf("count shopping run items for export: %w", err)
	}

	limit := filter.Limit
	limitSQL := ""
	if limit > 0 {
		limitSQL = fmt.Sprintf("\nLIMIT $%d", argPos)
		args = append(args, limit)
		argPos++
	}

	const exportCTE = `
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
    MAX(CASE WHEN level = 0 THEN name END) AS taxonomy_leaf0_name
  FROM taxonomy_chain
  GROUP BY leaf_id
),
identifiers_lookup AS (
  SELECT
    product_id,
    MAX(CASE WHEN identifier_type = 'pn_interno' THEN identifier_value END) AS pn_interno,
    MAX(CASE WHEN identifier_type = 'reference' THEN identifier_value END) AS reference,
    MAX(CASE WHEN identifier_type = 'ean' THEN identifier_value END) AS ean
  FROM catalog_product_identifiers
  WHERE tenant_id = current_tenant_id()
  GROUP BY product_id
)
`

	listQuery := exportCTE + `
SELECT
  i.run_item_id,
  i.run_id,
  i.product_id,
  p.sku,
  idf.pn_interno,
  idf.reference,
  idf.ean,
  p.name,
  p.brand_name,
  txo.taxonomy_leaf0_name,
  i.supplier_code,
  i.item_status,
  i.observed_price,
  i.currency_code,
  i.observed_at,
  i.seller_name,
  i.channel,
  i.product_url,
  i.http_status,
  i.elapsed_s,
  i.chosen_seller_json->>'lookup_term',
  i.notes,
  m.config_json::text
FROM shopping_price_run_items i
JOIN catalog_products p
  ON p.tenant_id = i.tenant_id
 AND p.product_id = i.product_id
LEFT JOIN taxonomy_lookup txo
  ON txo.taxonomy_node_id = p.primary_taxonomy_node_id
LEFT JOIN identifiers_lookup idf
  ON idf.product_id = p.product_id
LEFT JOIN supplier_driver_manifests m
  ON m.tenant_id = i.tenant_id
 AND m.supplier_code = i.supplier_code
 AND m.is_active = TRUE
 AND m.validation_status = 'valid'
` + whereSQL + `
ORDER BY i.observed_at DESC, i.created_at DESC` + limitSQL

	rows, err := tx.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return ports.RunExportList{}, fmt.Errorf("list shopping run items for export: %w", err)
	}
	defer rows.Close()

	items := make([]ports.RunExportRow, 0, 128)
	for rows.Next() {
		item, scanErr := scanRunExportRow(rows)
		if scanErr != nil {
			return ports.RunExportList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.RunExportList{}, fmt.Errorf("iterate shopping run items for export: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.RunExportList{}, fmt.Errorf("commit shopping run export read: %w", err)
	}

	return ports.RunExportList{
		Rows:  items,
		Total: total,
	}, nil
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
  error_message,
  total_items,
  processed_items,
  current_supplier_code,
  current_product_id,
  current_product_label,
  progress_updated_at
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
	var totalItems sql.NullInt64
	var processedItems sql.NullInt64
	var currentSupplierCode sql.NullString
	var currentProductID sql.NullString
	var currentProductLabel sql.NullString
	var progressUpdatedAt sql.NullTime

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
		&totalItems,
		&processedItems,
		&currentSupplierCode,
		&currentProductID,
		&currentProductLabel,
		&progressUpdatedAt,
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
	if totalItems.Valid {
		value := totalItems.Int64
		result.TotalItems = &value
	}
	if processedItems.Valid {
		value := processedItems.Int64
		result.ProcessedItems = &value
	}
	if currentSupplierCode.Valid {
		value := strings.TrimSpace(currentSupplierCode.String)
		if value != "" {
			result.CurrentSupplierCode = &value
		}
	}
	if currentProductID.Valid {
		value := strings.TrimSpace(currentProductID.String)
		if value != "" {
			result.CurrentProductID = &value
		}
	}
	if currentProductLabel.Valid {
		value := strings.TrimSpace(currentProductLabel.String)
		if value != "" {
			result.CurrentProductLabel = &value
		}
	}
	if progressUpdatedAt.Valid {
		value := progressUpdatedAt.Time.UTC()
		result.ProgressUpdatedAt = &value
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
  next_discovery_at,
  not_found_count,
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

func (r *Reader) ListManualURLCandidates(ctx context.Context, tenantID string, filter ports.ManualURLCandidateFilter) (ports.ManualURLCandidateList, error) {
	if strings.TrimSpace(filter.SupplierCode) == "" {
		return ports.ManualURLCandidateList{}, fmt.Errorf("supplier_code is required")
	}

	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.ManualURLCandidateList{}, err
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

	args := make([]any, 0, 6)
	args = append(args, filter.SupplierCode)
	argPos := 2

	clauses := []string{"WHERE p.tenant_id = current_tenant_id()"}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(`AND (
  p.sku ILIKE $%d
  OR p.name ILIKE $%d
  OR p.product_id ILIKE $%d
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

	includeExisting := filter.IncludeExisting
	if !includeExisting {
		clauses = append(clauses, "AND (s.product_url IS NULL OR BTRIM(s.product_url) = '')")
	}
	if filter.OnlyWithURL {
		clauses = append(clauses, "AND (s.product_url IS NOT NULL AND BTRIM(s.product_url) <> '')")
	}

	whereSQL := strings.Join(clauses, "\n")

	const manualCandidatesCTE = `
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
    MAX(CASE WHEN level = 0 THEN name END) AS taxonomy_leaf0_name
  FROM taxonomy_chain
  GROUP BY leaf_id
),
identifiers_lookup AS (
  SELECT
    product_id,
    MAX(CASE WHEN identifier_type = 'pn_interno' THEN identifier_value END) AS pn_interno,
    MAX(CASE WHEN identifier_type = 'reference' THEN identifier_value END) AS reference,
    MAX(CASE WHEN identifier_type = 'ean' THEN identifier_value END) AS ean
  FROM catalog_product_identifiers
  WHERE tenant_id = current_tenant_id()
  GROUP BY product_id
)
`

	countSQL := manualCandidatesCTE + `
SELECT COUNT(*)
FROM catalog_products p
LEFT JOIN taxonomy_lookup txo ON txo.taxonomy_node_id = p.primary_taxonomy_node_id
LEFT JOIN identifiers_lookup idf ON idf.product_id = p.product_id
LEFT JOIN shopping_supplier_product_signals s
  ON s.tenant_id = current_tenant_id()
 AND s.product_id = p.product_id
 AND s.supplier_code = $1
` + whereSQL

	var total int64
	if err := tx.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return ports.ManualURLCandidateList{}, fmt.Errorf("count shopping manual url candidates: %w", err)
	}

	querySQL := manualCandidatesCTE + `
SELECT
  p.product_id,
  $1 AS supplier_code,
  p.sku,
  idf.pn_interno,
  idf.reference,
  idf.ean,
  p.name,
  p.brand_name,
  txo.taxonomy_leaf0_name,
  s.product_url,
  COALESCE(s.url_status, 'STALE') AS url_status,
  COALESCE(s.lookup_mode, 'REFERENCE') AS lookup_mode,
  CASE
    WHEN s.product_id IS NULL THEN 'INFERRED'
    ELSE s.lookup_mode_source
  END AS lookup_mode_source,
  COALESCE(s.manual_override, FALSE) AS manual_override,
  s.last_checked_at,
  s.last_success_at,
  s.last_http_status,
  s.last_error_message,
  s.next_discovery_at,
  COALESCE(s.not_found_count, 0) AS not_found_count,
  COALESCE(s.updated_at, p.updated_at) AS updated_at
FROM catalog_products p
LEFT JOIN taxonomy_lookup txo ON txo.taxonomy_node_id = p.primary_taxonomy_node_id
LEFT JOIN identifiers_lookup idf ON idf.product_id = p.product_id
LEFT JOIN shopping_supplier_product_signals s
  ON s.tenant_id = current_tenant_id()
 AND s.product_id = p.product_id
 AND s.supplier_code = $1
` + whereSQL + `
ORDER BY p.name ASC, p.sku ASC
LIMIT $` + fmt.Sprintf("%d", argPos) + ` OFFSET $` + fmt.Sprintf("%d", argPos+1)

	rows, err := tx.QueryContext(ctx, querySQL, append(args, limit, offset)...)
	if err != nil {
		return ports.ManualURLCandidateList{}, fmt.Errorf("list shopping manual url candidates: %w", err)
	}
	defer rows.Close()

	items := make([]ports.ManualURLCandidate, 0, limit)
	for rows.Next() {
		item, scanErr := scanManualURLCandidate(rows)
		if scanErr != nil {
			return ports.ManualURLCandidateList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.ManualURLCandidateList{}, fmt.Errorf("iterate shopping manual url candidates: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.ManualURLCandidateList{}, fmt.Errorf("commit shopping manual url candidates: %w", err)
	}

	return ports.ManualURLCandidateList{
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

func scanRunItem(s scanner) (ports.RunItem, error) {
	var item ports.RunItem
	var observedAt time.Time
	var productURL sql.NullString
	var httpStatus sql.NullInt64
	var elapsedSeconds sql.NullFloat64
	var lookupTerm sql.NullString
	var notes sql.NullString
	var runtimeConfigJSON sql.NullString
	if err := s.Scan(
		&item.RunItemID,
		&item.RunID,
		&item.ProductID,
		&item.ProductLabel,
		&item.SupplierCode,
		&item.ItemStatus,
		&item.ObservedPrice,
		&item.Currency,
		&observedAt,
		&item.SellerName,
		&item.Channel,
		&productURL,
		&httpStatus,
		&elapsedSeconds,
		&lookupTerm,
		&notes,
		&runtimeConfigJSON,
	); err != nil {
		return ports.RunItem{}, fmt.Errorf("scan shopping run item: %w", err)
	}
	item.ObservedAt = observedAt.UTC()
	item.SupplierCode = strings.TrimSpace(item.SupplierCode)
	item.ItemStatus = strings.TrimSpace(item.ItemStatus)
	if productURL.Valid {
		value := strings.TrimSpace(productURL.String)
		if value != "" {
			item.ProductURL = &value
		}
	}
	if httpStatus.Valid {
		value := httpStatus.Int64
		item.HTTPStatus = &value
	}
	if elapsedSeconds.Valid {
		value := elapsedSeconds.Float64
		item.ElapsedSeconds = &value
	}
	if lookupTerm.Valid {
		value := strings.TrimSpace(lookupTerm.String)
		if value != "" {
			item.LookupTerm = &value
		}
	}
	if notes.Valid {
		value := strings.TrimSpace(notes.String)
		if value != "" {
			item.Notes = &value
		}
	}
	if item.ProductURL == nil && item.LookupTerm != nil && runtimeConfigJSON.Valid {
		if computed := deriveSearchURLFromRuntimeConfig(runtimeConfigJSON.String, item.SupplierCode, item.ProductID, *item.LookupTerm); computed != "" {
			item.ProductURL = &computed
		}
	}
	return item, nil
}

func scanRunExportRow(s scanner) (ports.RunExportRow, error) {
	var item ports.RunExportRow
	var observedAt time.Time
	var pnInterno sql.NullString
	var reference sql.NullString
	var ean sql.NullString
	var brandName sql.NullString
	var taxonomyLeaf0 sql.NullString
	var productURL sql.NullString
	var httpStatus sql.NullInt64
	var elapsedSeconds sql.NullFloat64
	var lookupTerm sql.NullString
	var notes sql.NullString
	var runtimeConfigJSON sql.NullString

	if err := s.Scan(
		&item.RunItemID,
		&item.RunID,
		&item.ProductID,
		&item.SKU,
		&pnInterno,
		&reference,
		&ean,
		&item.ProductLabel,
		&brandName,
		&taxonomyLeaf0,
		&item.SupplierCode,
		&item.ItemStatus,
		&item.ObservedPrice,
		&item.Currency,
		&observedAt,
		&item.SellerName,
		&item.Channel,
		&productURL,
		&httpStatus,
		&elapsedSeconds,
		&lookupTerm,
		&notes,
		&runtimeConfigJSON,
	); err != nil {
		return ports.RunExportRow{}, fmt.Errorf("scan shopping run export row: %w", err)
	}

	item.ObservedAt = observedAt.UTC()
	item.SupplierCode = strings.TrimSpace(item.SupplierCode)
	item.ItemStatus = strings.TrimSpace(item.ItemStatus)

	if pnInterno.Valid {
		value := strings.TrimSpace(pnInterno.String)
		if value != "" {
			item.PNInterno = &value
		}
	}
	if reference.Valid {
		value := strings.TrimSpace(reference.String)
		if value != "" {
			item.Reference = &value
		}
	}
	if ean.Valid {
		value := strings.TrimSpace(ean.String)
		if value != "" {
			item.EAN = &value
		}
	}
	if brandName.Valid {
		value := strings.TrimSpace(brandName.String)
		if value != "" {
			item.BrandName = &value
		}
	}
	if taxonomyLeaf0.Valid {
		value := strings.TrimSpace(taxonomyLeaf0.String)
		if value != "" {
			item.TaxonomyLeaf0Name = &value
		}
	}
	if productURL.Valid {
		value := strings.TrimSpace(productURL.String)
		if value != "" {
			item.ProductURL = &value
		}
	}
	if httpStatus.Valid {
		value := httpStatus.Int64
		item.HTTPStatus = &value
	}
	if elapsedSeconds.Valid {
		value := elapsedSeconds.Float64
		item.ElapsedSeconds = &value
	}
	if lookupTerm.Valid {
		value := strings.TrimSpace(lookupTerm.String)
		if value != "" {
			item.LookupTerm = &value
		}
	}
	if notes.Valid {
		value := strings.TrimSpace(notes.String)
		if value != "" {
			item.Notes = &value
		}
	}
	if item.ProductURL == nil && item.LookupTerm != nil && runtimeConfigJSON.Valid {
		if computed := deriveSearchURLFromRuntimeConfig(runtimeConfigJSON.String, item.SupplierCode, item.ProductID, *item.LookupTerm); computed != "" {
			item.ProductURL = &computed
		}
	}

	return item, nil
}

func deriveSearchURLFromRuntimeConfig(rawConfigJSON, supplierCode, productID, lookupTerm string) string {
	trimmedLookup := strings.TrimSpace(lookupTerm)
	if trimmedLookup == "" {
		return ""
	}
	var config map[string]any
	if err := json.Unmarshal([]byte(rawConfigJSON), &config); err != nil {
		return ""
	}
	candidates := []string{
		extractString(config["debugSearchUrlTemplate"]),
		extractString(config["searchUrlTemplate"]),
		extractString(config["searchUrl"]),
		extractString(config["endpointTemplate"]),
		extractString(config["startUrl"]),
	}
	for _, candidate := range candidates {
		rendered := renderSearchURLTemplate(candidate, supplierCode, productID, trimmedLookup)
		if rendered != "" {
			return rendered
		}
	}
	if vtexURL := deriveVTEXSearchURL(config, trimmedLookup); vtexURL != "" {
		return vtexURL
	}
	baseURL := strings.TrimSpace(extractString(config["baseUrl"]))
	if strings.HasPrefix(baseURL, "http://") || strings.HasPrefix(baseURL, "https://") {
		return baseURL
	}
	return ""
}

func renderSearchURLTemplate(template, supplierCode, productID, lookupTerm string) string {
	trimmed := strings.TrimSpace(template)
	if trimmed == "" {
		return ""
	}
	encoded := url.QueryEscape(lookupTerm)
	encodedProductID := url.QueryEscape(strings.TrimSpace(productID))
	lookupMode := "REFERENCE"
	if isLikelyEAN(lookupTerm) {
		lookupMode = "EAN"
	}
	encodedLookupMode := url.QueryEscape(lookupMode)
	encodedSupplier := url.QueryEscape(strings.TrimSpace(supplierCode))
	replaced := false
	placeholders := []string{
		"{term}",
		"{lookup_term}",
		"{lookupTerm}",
		"{lookup}",
		"{ean}",
		"{reference}",
		"{sku}",
		"{product_id}",
		"{productId}",
		"{lookup_mode}",
		"{supplier_code}",
		"{{term}}",
		"{{lookup_term}}",
		"{{lookupTerm}}",
	}
	for _, placeholder := range placeholders {
		if strings.Contains(trimmed, placeholder) {
			switch placeholder {
			case "{product_id}", "{productId}":
				trimmed = strings.ReplaceAll(trimmed, placeholder, encodedProductID)
			case "{lookup_mode}":
				trimmed = strings.ReplaceAll(trimmed, placeholder, encodedLookupMode)
			case "{supplier_code}":
				trimmed = strings.ReplaceAll(trimmed, placeholder, encodedSupplier)
			default:
				trimmed = strings.ReplaceAll(trimmed, placeholder, encoded)
			}
			replaced = true
		}
	}
	if strings.Contains(trimmed, "%s") {
		trimmed = fmt.Sprintf(trimmed, encoded)
		replaced = true
	}
	if !replaced {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	return ""
}

func deriveVTEXSearchURL(config map[string]any, lookupTerm string) string {
	baseURL := strings.TrimSpace(extractString(config["baseUrl"]))
	operationName := strings.TrimSpace(extractString(config["operationName"]))
	sha256Hash := strings.TrimSpace(extractString(config["sha256Hash"]))
	if baseURL == "" || operationName == "" || sha256Hash == "" {
		return ""
	}

	skusFilter := strings.TrimSpace(extractString(config["skusFilter"]))
	if skusFilter == "" {
		skusFilter = "FIRST_AVAILABLE"
	}
	toN := extractInt(config["toN"], 11)
	if toN < 1 {
		toN = 11
	}
	includeVariant := extractBool(config["includeVariant"], false)

	values := url.Values{}
	values.Set("workspace", "master")
	values.Set("maxAge", "short")
	values.Set("appsEtag", "remove")
	values.Set("domain", "store")
	values.Set("locale", "pt-BR")
	values.Set("__bindingId", "933c1279-a878-4f45-8bb1-d84d09dba761")
	values.Set("operationName", operationName)
	variablesPayload, err := buildVTEXVariablesJSON(lookupTerm, skusFilter, toN, includeVariant)
	if err != nil {
		return ""
	}
	values.Set("variables", variablesPayload)
	values.Set("extensions", fmt.Sprintf(`{"persistedQuery":{"version":1,"sha256Hash":"%s","sender":"vtex.store-resources@0.x","provider":"vtex.search-graphql@0.x"}}`, sha256Hash))
	return fmt.Sprintf("%s?%s", baseURL, values.Encode())
}

func buildVTEXVariablesJSON(lookupTerm, skusFilter string, toN int, includeVariant bool) (string, error) {
	payload := map[string]any{
		"hideUnavailableItems": false,
		"skusFilter":           skusFilter,
		"simulationBehavior":   "default",
		"installmentCriteria":  "MAX_WITHOUT_INTEREST",
		"productOriginVtex":    false,
		"map":                  "ft",
		"query":                lookupTerm,
		"orderBy":              "OrderByScoreDESC",
		"from":                 0,
		"to":                   toN,
		"selectedFacets": []map[string]string{
			{"key": "ft", "value": lookupTerm},
		},
		"operator":             "and",
		"fuzzy":                "0",
		"searchState":          nil,
		"facetsBehavior":       "Static",
		"categoryTreeBehavior": "default",
		"withFacets":           includeVariant,
		"advertisementOptions": map[string]any{
			"showSponsored":           true,
			"sponsoredCount":          3,
			"repeatSponsoredProducts": true,
		},
		"fullText":           lookupTerm,
		"autocompleteField":  "ft",
		"source":             "autocomplete",
		"searchCorrection":   true,
		"spellCheck":         true,
		"hideSponsoredItems": false,
		"fuzzyProductSearch": false,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func extractInt(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return fallback
	}
}

func extractBool(value any, fallback bool) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return fallback
	}
}

func isLikelyEAN(term string) bool {
	trimmed := strings.TrimSpace(term)
	if len(trimmed) != 13 && len(trimmed) != 14 {
		return false
	}
	for _, char := range trimmed {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func scanSupplierSignal(s scanner) (ports.SupplierSignal, error) {
	var item ports.SupplierSignal
	var productURL sql.NullString
	var lastCheckedAt sql.NullTime
	var lastSuccessAt sql.NullTime
	var lastHTTPStatus sql.NullInt64
	var lastErrorMessage sql.NullString
	var nextDiscoveryAt sql.NullTime
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
		&nextDiscoveryAt,
		&item.NotFoundCount,
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
	if nextDiscoveryAt.Valid {
		value := nextDiscoveryAt.Time.UTC()
		item.NextDiscoveryAt = &value
	}
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

func scanManualURLCandidate(s scanner) (ports.ManualURLCandidate, error) {
	var item ports.ManualURLCandidate
	var pnInterno sql.NullString
	var reference sql.NullString
	var ean sql.NullString
	var brandName sql.NullString
	var taxonomyLeaf0Name sql.NullString
	var productURL sql.NullString
	var lastCheckedAt sql.NullTime
	var lastSuccessAt sql.NullTime
	var lastHTTPStatus sql.NullInt64
	var lastErrorMessage sql.NullString
	var nextDiscoveryAt sql.NullTime
	if err := s.Scan(
		&item.ProductID,
		&item.SupplierCode,
		&item.SKU,
		&pnInterno,
		&reference,
		&ean,
		&item.Name,
		&brandName,
		&taxonomyLeaf0Name,
		&productURL,
		&item.URLStatus,
		&item.LookupMode,
		&item.LookupModeSource,
		&item.ManualOverride,
		&lastCheckedAt,
		&lastSuccessAt,
		&lastHTTPStatus,
		&lastErrorMessage,
		&nextDiscoveryAt,
		&item.NotFoundCount,
		&item.UpdatedAt,
	); err != nil {
		return ports.ManualURLCandidate{}, fmt.Errorf("scan shopping manual url candidate: %w", err)
	}

	if pnInterno.Valid {
		value := strings.TrimSpace(pnInterno.String)
		if value != "" {
			item.PNInterno = &value
		}
	}
	if reference.Valid {
		value := strings.TrimSpace(reference.String)
		if value != "" {
			item.Reference = &value
		}
	}
	if ean.Valid {
		value := strings.TrimSpace(ean.String)
		if value != "" {
			item.EAN = &value
		}
	}
	if brandName.Valid {
		value := strings.TrimSpace(brandName.String)
		if value != "" {
			item.BrandName = &value
		}
	}
	if taxonomyLeaf0Name.Valid {
		value := strings.TrimSpace(taxonomyLeaf0Name.String)
		if value != "" {
			item.TaxonomyLeaf0Name = &value
		}
	}
	if productURL.Valid {
		value := strings.TrimSpace(productURL.String)
		if value != "" {
			item.ProductURL = &value
		}
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
		value := strings.TrimSpace(lastErrorMessage.String)
		if value != "" {
			item.LastErrorMessage = &value
		}
	}
	if nextDiscoveryAt.Valid {
		value := nextDiscoveryAt.Time.UTC()
		item.NextDiscoveryAt = &value
	}
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}
