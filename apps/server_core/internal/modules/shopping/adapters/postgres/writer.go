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
