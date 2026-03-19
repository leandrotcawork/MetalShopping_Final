package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	shoppingevents "metalshopping/server_core/internal/modules/shopping/events"
	"metalshopping/server_core/internal/modules/shopping/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

type Writer struct {
	db          *sql.DB
	outboxStore *outbox.Store
}

func NewWriter(db *sql.DB, outboxStore *outbox.Store) *Writer {
	return &Writer{db: db, outboxStore: outboxStore}
}

func (w *Writer) CreateRunRequest(ctx context.Context, tenantID, traceID string, input ports.CreateRunRequestInput) (ports.RunRequest, error) {
	if input.InputMode != "xlsx" && input.InputMode != "catalog" {
		return ports.RunRequest{}, fmt.Errorf("invalid shopping input_mode: %s", input.InputMode)
	}
	if input.InputMode == "catalog" && len(input.CatalogProductIDs) == 0 {
		return ports.RunRequest{}, fmt.Errorf("catalog mode requires at least one product")
	}
	if input.InputMode == "xlsx" &&
		strings.TrimSpace(input.XLSXFilePath) == "" &&
		len(input.XLSXScopeIDs) == 0 &&
		len(input.CatalogProductIDs) == 0 {
		return ports.RunRequest{}, fmt.Errorf("xlsx mode requires xlsx file path or scope identifiers")
	}
	if strings.TrimSpace(input.RequestedBy) == "" {
		return ports.RunRequest{}, fmt.Errorf("requested_by is required")
	}

	tx, err := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
	if err != nil {
		return ports.RunRequest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	resolvedCatalogIDs := dedupeValues(input.CatalogProductIDs)
	unresolvedScopeIDs := []string{}
	ambiguousScopeIDs := []string{}

	if input.InputMode == "xlsx" && len(resolvedCatalogIDs) == 0 && len(input.XLSXScopeIDs) > 0 {
		resolution, err := resolveScopeIdentifiers(ctx, tx, input.XLSXScopeIDs)
		if err != nil {
			return ports.RunRequest{}, err
		}
		resolvedCatalogIDs = resolution.ResolvedCatalogProductIDs
		unresolvedScopeIDs = resolution.UnresolvedScopeIDs
		ambiguousScopeIDs = resolution.AmbiguousScopeIDs
	}
	if input.InputMode == "xlsx" && len(resolvedCatalogIDs) == 0 {
		return ports.RunRequest{}, fmt.Errorf("xlsx scope did not resolve to any catalog product")
	}

	runRequestID := generateRunRequestID()
	payloadJSON, err := json.Marshal(map[string]any{
		"inputMode":                  input.InputMode,
		"catalogProductIds":          resolvedCatalogIDs,
		"xlsxFilePath":               input.XLSXFilePath,
		"xlsxScopeIdentifiers":       input.XLSXScopeIDs,
		"resolvedCatalogProductIds":  resolvedCatalogIDs,
		"unresolvedScopeIdentifiers": unresolvedScopeIDs,
		"ambiguousScopeIdentifiers":  ambiguousScopeIDs,
		"supplierCodes":              input.SupplierCodes,
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

	if w.outboxStore != nil {
		record, err := shoppingevents.NewRunRequestedOutboxRecord(tenantID, result, input, traceID, result.RequestedAt)
		if err != nil {
			return ports.RunRequest{}, err
		}
		if err := w.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
			return ports.RunRequest{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return ports.RunRequest{}, fmt.Errorf("commit shopping run request: %w", err)
	}
	result.CatalogProductIDs = resolvedCatalogIDs
	if strings.TrimSpace(input.XLSXFilePath) != "" {
		value := input.XLSXFilePath
		result.XLSXFilePath = &value
	}
	result.XLSXScopeIDs = append([]string{}, input.XLSXScopeIDs...)
	result.ResolvedCatalogProductIDs = resolvedCatalogIDs
	result.UnresolvedScopeIDs = unresolvedScopeIDs
	result.AmbiguousScopeIDs = ambiguousScopeIDs
	return result, nil
}

func (w *Writer) UpsertSupplierSignal(ctx context.Context, tenantID string, input ports.UpsertSupplierSignalInput) (ports.SupplierSignal, error) {
	if strings.TrimSpace(input.ProductID) == "" {
		return ports.SupplierSignal{}, fmt.Errorf("product_id is required")
	}
	if strings.TrimSpace(input.SupplierCode) == "" {
		return ports.SupplierSignal{}, fmt.Errorf("supplier_code is required")
	}
	if strings.TrimSpace(input.UpdatedBy) == "" {
		return ports.SupplierSignal{}, fmt.Errorf("updated_by is required")
	}
	if input.URLStatus != nil {
		switch *input.URLStatus {
		case "ACTIVE", "STALE", "INVALID":
		default:
			return ports.SupplierSignal{}, fmt.Errorf("invalid url_status: %s", *input.URLStatus)
		}
	}
	if input.LookupMode != nil {
		switch *input.LookupMode {
		case "EAN", "REFERENCE":
		default:
			return ports.SupplierSignal{}, fmt.Errorf("invalid lookup_mode: %s", *input.LookupMode)
		}
	}

	tx, err := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
	if err != nil {
		return ports.SupplierSignal{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
INSERT INTO shopping_supplier_product_signals (
  tenant_id,
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
  created_by,
  created_at,
  updated_at
) VALUES (
  current_tenant_id(),
  $1,
  $2,
  $3,
  COALESCE($4, 'STALE'),
  COALESCE($5, 'REFERENCE'),
  'MANUAL',
  COALESCE($6, TRUE),
  NOW(),
  CASE WHEN $3 IS NULL OR $3 = '' THEN NULL ELSE NOW() END,
  NULL,
  NULL,
  $7,
  NOW(),
  NOW()
)
ON CONFLICT (tenant_id, product_id, supplier_code) DO UPDATE SET
  product_url = EXCLUDED.product_url,
  url_status = COALESCE($4, shopping_supplier_product_signals.url_status),
  lookup_mode = COALESCE($5, shopping_supplier_product_signals.lookup_mode),
  lookup_mode_source = 'MANUAL',
  manual_override = COALESCE($6, TRUE),
  last_checked_at = NOW(),
  last_success_at = CASE
    WHEN EXCLUDED.product_url IS NULL OR EXCLUDED.product_url = '' THEN shopping_supplier_product_signals.last_success_at
    ELSE NOW()
  END,
  updated_at = NOW()
RETURNING
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
`

	var signal ports.SupplierSignal
	var productURL sql.NullString
	var lastCheckedAt sql.NullTime
	var lastSuccessAt sql.NullTime
	var lastHTTPStatus sql.NullInt64
	var lastErrorMessage sql.NullString
	if err := tx.QueryRowContext(
		ctx,
		query,
		input.ProductID,
		input.SupplierCode,
		input.ProductURL,
		input.URLStatus,
		input.LookupMode,
		input.ManualOverride,
		input.UpdatedBy,
	).Scan(
		&signal.ProductID,
		&signal.SupplierCode,
		&productURL,
		&signal.URLStatus,
		&signal.LookupMode,
		&signal.LookupModeSource,
		&signal.ManualOverride,
		&lastCheckedAt,
		&lastSuccessAt,
		&lastHTTPStatus,
		&lastErrorMessage,
		&signal.UpdatedAt,
	); err != nil {
		return ports.SupplierSignal{}, fmt.Errorf("upsert shopping supplier signal: %w", err)
	}

	if productURL.Valid {
		value := productURL.String
		signal.ProductURL = &value
	}
	if lastCheckedAt.Valid {
		value := lastCheckedAt.Time.UTC()
		signal.LastCheckedAt = &value
	}
	if lastSuccessAt.Valid {
		value := lastSuccessAt.Time.UTC()
		signal.LastSuccessAt = &value
	}
	if lastHTTPStatus.Valid {
		value := lastHTTPStatus.Int64
		signal.LastHTTPStatus = &value
	}
	if lastErrorMessage.Valid {
		value := lastErrorMessage.String
		signal.LastErrorMessage = &value
	}
	signal.UpdatedAt = signal.UpdatedAt.UTC()

	if err := tx.Commit(); err != nil {
		return ports.SupplierSignal{}, fmt.Errorf("commit shopping supplier signal upsert: %w", err)
	}
	return signal, nil
}

type scopeResolution struct {
	ResolvedCatalogProductIDs []string
	UnresolvedScopeIDs        []string
	AmbiguousScopeIDs         []string
}

func resolveScopeIdentifiers(ctx context.Context, tx *sql.Tx, rawScopeIDs []string) (scopeResolution, error) {
	normalized := normalizeScopeIDs(rawScopeIDs)
	if len(normalized) == 0 {
		return scopeResolution{}, nil
	}

	args := make([]any, 0, len(normalized))
	valueRows := make([]string, 0, len(normalized))
	for index, value := range normalized {
		args = append(args, value)
		valueRows = append(valueRows, fmt.Sprintf("($%d)", index+1))
	}

	query := fmt.Sprintf(`
WITH wanted(identifier_norm) AS (
  VALUES %s
),
candidate AS (
  SELECT LOWER(BTRIM(sku)) AS identifier_norm, product_id
  FROM catalog_products
  WHERE tenant_id = current_tenant_id()
  UNION ALL
  SELECT LOWER(BTRIM(identifier_value)) AS identifier_norm, product_id
  FROM catalog_product_identifiers
  WHERE tenant_id = current_tenant_id()
),
matches AS (
  SELECT
    w.identifier_norm,
    COALESCE(MIN(c.product_id), '') AS product_id,
    COUNT(DISTINCT c.product_id)::bigint AS match_count
  FROM wanted w
  LEFT JOIN candidate c
    ON c.identifier_norm = w.identifier_norm
  GROUP BY w.identifier_norm
)
SELECT identifier_norm, product_id, match_count
FROM matches
`, strings.Join(valueRows, ","))

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return scopeResolution{}, fmt.Errorf("resolve xlsx scope identifiers: %w", err)
	}
	defer rows.Close()

	resolved := []string{}
	unresolved := []string{}
	ambiguous := []string{}
	seenProducts := map[string]struct{}{}

	for rows.Next() {
		var identifierNorm string
		var productID string
		var matchCount int64
		if err := rows.Scan(&identifierNorm, &productID, &matchCount); err != nil {
			return scopeResolution{}, fmt.Errorf("scan xlsx scope resolution: %w", err)
		}

		switch {
		case matchCount == 0:
			unresolved = append(unresolved, identifierNorm)
		case matchCount > 1:
			ambiguous = append(ambiguous, identifierNorm)
		default:
			if productID != "" {
				if _, exists := seenProducts[productID]; !exists {
					seenProducts[productID] = struct{}{}
					resolved = append(resolved, productID)
				}
			}
		}
	}
	if err := rows.Err(); err != nil {
		return scopeResolution{}, fmt.Errorf("iterate xlsx scope resolution: %w", err)
	}

	slices.Sort(unresolved)
	slices.Sort(ambiguous)

	return scopeResolution{
		ResolvedCatalogProductIDs: resolved,
		UnresolvedScopeIDs:        unresolved,
		AmbiguousScopeIDs:         ambiguous,
	}, nil
}

func normalizeScopeIDs(raw []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(raw))
	for _, value := range raw {
		norm := strings.ToLower(strings.TrimSpace(value))
		if norm == "" {
			continue
		}
		if _, exists := seen[norm]; exists {
			continue
		}
		seen[norm] = struct{}{}
		result = append(result, norm)
	}
	return result
}

func dedupeValues(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		norm := strings.TrimSpace(value)
		if norm == "" {
			continue
		}
		if _, exists := seen[norm]; exists {
			continue
		}
		seen[norm] = struct{}{}
		result = append(result, norm)
	}
	return result
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
