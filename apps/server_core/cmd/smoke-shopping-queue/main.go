package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type payload struct {
	InputMode         string   `json:"inputMode"`
	CatalogProductIDs []string `json:"catalogProductIds,omitempty"`
	XLSXFilePath      string   `json:"xlsxFilePath,omitempty"`
}

type result struct {
	RunRequestID string `json:"run_request_id"`
	Status       string `json:"status"`
	RunID        string `json:"run_id"`
	RowsWritten  int    `json:"rows_written"`
}

func main() {
	ctx := context.Background()
	dsn := strings.TrimSpace(os.Getenv("MS_DATABASE_URL"))
	tenantID := strings.TrimSpace(os.Getenv("MS_TENANT_ID"))
	inputMode := strings.ToLower(strings.TrimSpace(os.Getenv("MS_INPUT_MODE")))
	productIDsRaw := strings.TrimSpace(os.Getenv("MS_PRODUCT_IDS"))
	xlsxPath := strings.TrimSpace(os.Getenv("MS_XLSX_FILE_PATH"))

	if dsn == "" || tenantID == "" {
		fmt.Fprintln(os.Stderr, "MS_DATABASE_URL and MS_TENANT_ID are required")
		os.Exit(2)
	}
	if inputMode == "" {
		inputMode = "catalog"
	}
	if inputMode != "catalog" && inputMode != "xlsx" {
		fmt.Fprintln(os.Stderr, "MS_INPUT_MODE must be catalog or xlsx")
		os.Exit(2)
	}

	productIDs := []string{}
	if productIDsRaw != "" {
		for _, part := range strings.Split(productIDsRaw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				productIDs = append(productIDs, part)
			}
		}
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	runRequestID := newUUID()
	runID := newUUID()

	p := payload{
		InputMode: inputMode,
	}
	if inputMode == "catalog" {
		p.CatalogProductIDs = productIDs
	} else {
		p.XLSXFilePath = xlsxPath
	}
	payloadJSON, _ := json.Marshal(p)

	if err := insertRunRequest(ctx, pool, tenantID, runRequestID, inputMode, string(payloadJSON)); err != nil {
		fmt.Fprintln(os.Stderr, "insert run_request:", err)
		os.Exit(1)
	}
	if err := claimRunRequest(ctx, pool, tenantID, runRequestID, "smoke"); err != nil {
		fmt.Fprintln(os.Stderr, "claim run_request:", err)
		os.Exit(1)
	}
	if err := markRunning(ctx, pool, tenantID, runRequestID, runID, "smoke"); err != nil {
		fmt.Fprintln(os.Stderr, "mark running:", err)
		os.Exit(1)
	}

	rowsWritten, err := writeReadModels(ctx, pool, tenantID, runID, inputMode, productIDs)
	if err != nil {
		_ = markFailed(ctx, pool, tenantID, runRequestID, err.Error())
		fmt.Fprintln(os.Stderr, "write read models:", err)
		os.Exit(1)
	}
	if err := markCompleted(ctx, pool, tenantID, runRequestID); err != nil {
		fmt.Fprintln(os.Stderr, "mark completed:", err)
		os.Exit(1)
	}

	out := result{
		RunRequestID: runRequestID,
		Status:       "completed",
		RunID:        runID,
		RowsWritten:  rowsWritten,
	}
	encoded, _ := json.Marshal(out)
	fmt.Println(string(encoded))
}

func setTenant(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID)
	return err
}

func insertRunRequest(ctx context.Context, pool *pgxpool.Pool, tenantID, runRequestID, inputMode, payloadJSON string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO shopping_price_run_requests (
  run_request_id, tenant_id, request_status, input_mode, input_payload_json, requested_by, requested_at
) VALUES (
  $1, current_tenant_id(), 'queued', $2, $3::jsonb, 'smoke', NOW()
)`, runRequestID, inputMode, payloadJSON)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func claimRunRequest(ctx context.Context, pool *pgxpool.Pool, tenantID, runRequestID, workerID string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE shopping_price_run_requests
SET request_status = 'claimed',
    claimed_at = NOW(),
    worker_id = $1,
    updated_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_request_id = $2
  AND request_status = 'queued'
`, workerID, runRequestID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func markRunning(ctx context.Context, pool *pgxpool.Pool, tenantID, runRequestID, runID, workerID string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE shopping_price_run_requests
SET request_status = 'running',
    started_at = NOW(),
    run_id = $1,
    worker_id = $2,
    updated_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_request_id = $3
  AND request_status IN ('queued','claimed')
`, runID, workerID, runRequestID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func markCompleted(ctx context.Context, pool *pgxpool.Pool, tenantID, runRequestID string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE shopping_price_run_requests
SET request_status = 'completed',
    finished_at = NOW(),
    updated_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_request_id = $1
`, runRequestID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func markFailed(ctx context.Context, pool *pgxpool.Pool, tenantID, runRequestID, msg string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE shopping_price_run_requests
SET request_status = 'failed',
    finished_at = NOW(),
    error_message = $1,
    updated_at = NOW()
WHERE tenant_id = current_tenant_id()
  AND run_request_id = $2
`, trimMsg(msg), runRequestID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func writeReadModels(ctx context.Context, pool *pgxpool.Pool, tenantID, runID, inputMode string, productIDs []string) (int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return 0, err
	}

	items, err := selectItems(ctx, tx, inputMode, productIDs)
	if err != nil {
		return 0, err
	}
	processed := int64(len(items))

	startedAt := time.Now().UTC()
	finishedAt := startedAt

	_, err = tx.Exec(ctx, `
INSERT INTO shopping_price_runs (
  run_id, tenant_id, run_status, started_at, finished_at, processed_items, total_items, notes
) VALUES (
  $1, current_tenant_id(), 'completed', $2, $3, $4, $5, 'smoke run'
) ON CONFLICT (run_id) DO UPDATE SET
  run_status = EXCLUDED.run_status,
  finished_at = EXCLUDED.finished_at,
  processed_items = EXCLUDED.processed_items,
  total_items = EXCLUDED.total_items,
  notes = EXCLUDED.notes,
  updated_at = NOW()
`, runID, startedAt, finishedAt, processed, processed)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		runItemID := newUUID()
		_, err := tx.Exec(ctx, `
INSERT INTO shopping_price_run_items (
  run_item_id, tenant_id, run_id, product_id, supplier_code, seller_name, channel,
  observed_price, currency_code, observed_at
) VALUES (
  $1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9
) ON CONFLICT (run_item_id) DO NOTHING
`, runItemID, runID, item.ProductID, "DEFAULT", item.SellerName, item.Channel, item.ObservedPrice, item.CurrencyCode, startedAt)
		if err != nil {
			return 0, err
		}

		snapshotID := fmt.Sprintf("%s:%s:%s", tenantID, item.ProductID, "DEFAULT")
		_, err = tx.Exec(ctx, `
INSERT INTO shopping_price_latest_snapshot (
  snapshot_id, tenant_id, product_id, supplier_code, run_id, seller_name, channel,
  observed_price, currency_code, observed_at
) VALUES (
  $1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9
) ON CONFLICT (tenant_id, product_id, supplier_code) DO UPDATE SET
  run_id = EXCLUDED.run_id,
  seller_name = EXCLUDED.seller_name,
  channel = EXCLUDED.channel,
  observed_price = EXCLUDED.observed_price,
  currency_code = EXCLUDED.currency_code,
  observed_at = EXCLUDED.observed_at,
  updated_at = NOW()
`, snapshotID, item.ProductID, "DEFAULT", runID, item.SellerName, item.Channel, item.ObservedPrice, item.CurrencyCode, startedAt)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(items), nil
}

type itemRow struct {
	ProductID     string
	ObservedPrice float64
	CurrencyCode  string
	SellerName    string
	Channel       string
}

func selectItems(ctx context.Context, tx pgx.Tx, inputMode string, productIDs []string) ([]itemRow, error) {
	switch inputMode {
	case "catalog":
		if len(productIDs) == 0 {
			return nil, errors.New("catalog mode requires MS_PRODUCT_IDS (comma separated)")
		}
		rows, err := tx.Query(ctx, `
SELECT
  p.product_id,
  COALESCE(pr.price_amount, 0)::double precision AS observed_price,
  COALESCE(pr.currency_code, 'BRL') AS currency_code
FROM catalog_products p
LEFT JOIN pricing_product_prices pr
  ON pr.tenant_id = current_tenant_id()
 AND pr.product_id = p.product_id
 AND pr.effective_to IS NULL
WHERE p.tenant_id = current_tenant_id()
  AND p.product_id = ANY($1::text[])
ORDER BY p.product_id
`, productIDs)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		out := []itemRow{}
		for rows.Next() {
			var row itemRow
			if err := rows.Scan(&row.ProductID, &row.ObservedPrice, &row.CurrencyCode); err != nil {
				return nil, err
			}
			row.SellerName = "smoke_catalog"
			row.Channel = "CATALOG"
			out = append(out, row)
		}
		return out, rows.Err()
	case "xlsx":
		rows, err := tx.Query(ctx, `
SELECT
  p.product_id,
  COALESCE(pr.price_amount, 0)::double precision AS observed_price,
  COALESCE(pr.currency_code, 'BRL') AS currency_code
FROM catalog_products p
LEFT JOIN pricing_product_prices pr
  ON pr.tenant_id = current_tenant_id()
 AND pr.product_id = p.product_id
 AND pr.effective_to IS NULL
WHERE p.tenant_id = current_tenant_id()
ORDER BY p.updated_at DESC
LIMIT 10
`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		out := []itemRow{}
		for rows.Next() {
			var row itemRow
			if err := rows.Scan(&row.ProductID, &row.ObservedPrice, &row.CurrencyCode); err != nil {
				return nil, err
			}
			row.SellerName = "smoke_xlsx"
			row.Channel = "XLSX"
			out = append(out, row)
		}
		return out, rows.Err()
	default:
		return nil, fmt.Errorf("unsupported input mode: %s", inputMode)
	}
}

func trimMsg(msg string) string {
	msg = strings.TrimSpace(msg)
	if len(msg) > 500 {
		return msg[:500]
	}
	return msg
}

func newUUID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	hexValue := hex.EncodeToString(buf)
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hexValue[0:8],
		hexValue[8:12],
		hexValue[12:16],
		hexValue[16:20],
		hexValue[20:32],
	)
}
