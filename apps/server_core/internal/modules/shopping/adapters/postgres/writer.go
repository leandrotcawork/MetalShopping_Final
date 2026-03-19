package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/shopping/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type Writer struct {
	db *sql.DB
}

func NewWriter(db *sql.DB) *Writer {
	return &Writer{db: db}
}

func (w *Writer) CreateRunRequest(ctx context.Context, tenantID string, input ports.CreateRunRequestInput) (ports.RunRequest, error) {
	if input.InputMode != "xlsx" && input.InputMode != "catalog" {
		return ports.RunRequest{}, fmt.Errorf("invalid shopping input_mode: %s", input.InputMode)
	}
	if input.InputMode == "catalog" && len(input.CatalogProductIDs) == 0 {
		return ports.RunRequest{}, fmt.Errorf("catalog mode requires at least one product")
	}
	if input.InputMode == "xlsx" && strings.TrimSpace(input.XLSXFilePath) == "" {
		return ports.RunRequest{}, fmt.Errorf("xlsx mode requires xlsx file path")
	}
	if strings.TrimSpace(input.RequestedBy) == "" {
		return ports.RunRequest{}, fmt.Errorf("requested_by is required")
	}

	tx, err := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
	if err != nil {
		return ports.RunRequest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	runRequestID := generateRunRequestID()
	payloadJSON, err := json.Marshal(map[string]any{
		"inputMode":         input.InputMode,
		"catalogProductIds": input.CatalogProductIDs,
		"xlsxFilePath":      input.XLSXFilePath,
		"supplierCodes":     input.SupplierCodes,
		"advanced": map[string]any{
			"timeoutSeconds":    input.Advanced.TimeoutSeconds,
			"httpWorkers":       input.Advanced.HTTPWorkers,
			"playwrightWorkers": input.Advanced.PlaywrightWorker,
			"topN":              input.Advanced.TopN,
		},
		"notes": input.Notes,
	})
	if err != nil {
		return ports.RunRequest{}, fmt.Errorf("marshal shopping run request payload: %w", err)
	}

	const query = `
INSERT INTO shopping_price_run_requests (
  run_request_id,
  tenant_id,
  request_status,
  input_mode,
  input_payload_json,
  requested_by,
  requested_at
)
VALUES (
  $1,
  current_tenant_id(),
  'queued',
  $2,
  $3::jsonb,
  $4,
  NOW()
)
RETURNING run_request_id, request_status, input_mode, requested_at, requested_by
`
	var result ports.RunRequest
	if err := tx.QueryRowContext(
		ctx,
		query,
		runRequestID,
		input.InputMode,
		string(payloadJSON),
		input.RequestedBy,
	).Scan(
		&result.RunRequestID,
		&result.Status,
		&result.InputMode,
		&result.RequestedAt,
		&result.RequestedBy,
	); err != nil {
		return ports.RunRequest{}, fmt.Errorf("insert shopping run request: %w", err)
	}
	result.RequestedAt = result.RequestedAt.UTC()

	if err := tx.Commit(); err != nil {
		return ports.RunRequest{}, fmt.Errorf("commit shopping run request: %w", err)
	}
	return result, nil
}

func generateRunRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		fallback := make([]byte, 6)
		_, _ = rand.Read(fallback)
		return "00000000-0000-4000-8000-" + hex.EncodeToString(fallback)
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
