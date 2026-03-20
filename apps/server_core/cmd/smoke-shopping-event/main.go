package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type payload struct {
	InputMode         string   `json:"inputMode"`
	CatalogProductIDs []string `json:"catalogProductIds,omitempty"`
	XLSXFilePath      string   `json:"xlsxFilePath,omitempty"`
	SupplierCodes     []string `json:"supplierCodes,omitempty"`
}

type result struct {
	RunRequestID string `json:"run_request_id"`
	EventID      string `json:"event_id"`
	Status       string `json:"status"`
}

func runtimeKindFromStrategy(strategy string) (executionKind string, family string) {
	normalized := strings.ToLower(strings.TrimSpace(strategy))
	if strings.HasPrefix(normalized, "playwright.") {
		return "PLAYWRIGHT", "playwright"
	}
	return "HTTP", "http"
}

func main() {
	ctx := context.Background()
	dsn := strings.TrimSpace(os.Getenv("MS_DATABASE_URL"))
	tenantID := strings.TrimSpace(os.Getenv("MS_TENANT_ID"))
	inputMode := strings.ToLower(strings.TrimSpace(os.Getenv("MS_INPUT_MODE")))
	supplierCode := strings.TrimSpace(os.Getenv("MS_SMOKE_SUPPLIER_CODE"))
	smokeStrategy := strings.ToLower(strings.TrimSpace(os.Getenv("MS_SMOKE_STRATEGY")))
	rawCatalogIDs := strings.TrimSpace(os.Getenv("MS_CATALOG_PRODUCT_IDS"))

	if dsn == "" || tenantID == "" {
		fmt.Fprintln(os.Stderr, "MS_DATABASE_URL and MS_TENANT_ID are required")
		os.Exit(2)
	}
	if inputMode == "" {
		inputMode = "xlsx"
	}
	if inputMode != "catalog" && inputMode != "xlsx" {
		fmt.Fprintln(os.Stderr, "MS_INPUT_MODE must be catalog or xlsx")
		os.Exit(2)
	}
	if supplierCode == "" {
		supplierCode = "OBRA_FACIL"
	}
	if smokeStrategy == "" {
		smokeStrategy = "http.mock.v1"
	}

	catalogIDs := []string{}
	if rawCatalogIDs != "" {
		for _, part := range strings.Split(rawCatalogIDs, ",") {
			value := strings.TrimSpace(part)
			if value != "" {
				catalogIDs = append(catalogIDs, value)
			}
		}
	}
	if inputMode == "catalog" && len(catalogIDs) == 0 {
		fmt.Fprintln(os.Stderr, "MS_CATALOG_PRODUCT_IDS is required when MS_INPUT_MODE=catalog")
		os.Exit(2)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := seedSupplier(ctx, pool, tenantID, supplierCode, smokeStrategy); err != nil {
		fmt.Fprintln(os.Stderr, "seed supplier:", err)
		os.Exit(1)
	}

	runRequestID := newUUID()
	eventID := newUUID()
	payloadJSON, _ := json.Marshal(payload{
		InputMode:         inputMode,
		CatalogProductIDs: catalogIDs,
		XLSXFilePath:      "",
		SupplierCodes:     []string{supplierCode},
	})

	if err := insertRunRequest(ctx, pool, tenantID, runRequestID, inputMode, string(payloadJSON)); err != nil {
		fmt.Fprintln(os.Stderr, "insert run_request:", err)
		os.Exit(1)
	}
	if err := insertOutboxEvent(ctx, pool, tenantID, eventID, runRequestID, inputMode); err != nil {
		fmt.Fprintln(os.Stderr, "insert outbox:", err)
		os.Exit(1)
	}

	out := result{
		RunRequestID: runRequestID,
		EventID:      eventID,
		Status:       "published",
	}
	encoded, _ := json.Marshal(out)
	fmt.Println(string(encoded))
}

func setTenant(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID)
	return err
}

func seedSupplier(ctx context.Context, pool *pgxpool.Pool, tenantID, supplierCode, strategy string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}

	executionKind, family := runtimeKindFromStrategy(strategy)

	_, err = tx.Exec(ctx, `
INSERT INTO suppliers_directory (
  supplier_id, tenant_id, supplier_code, supplier_label, execution_kind, lookup_policy, enabled
) VALUES (
  $1, current_tenant_id(), $2, $3, $4, 'EAN_FIRST', TRUE
)
ON CONFLICT (tenant_id, supplier_code) DO UPDATE SET
  supplier_label = EXCLUDED.supplier_label,
  execution_kind = EXCLUDED.execution_kind,
  lookup_policy = EXCLUDED.lookup_policy,
  enabled = TRUE,
  updated_at = NOW()
`, "sup_"+newUUID(), supplierCode, supplierCode, executionKind)
	if err != nil {
		return err
	}

	configJSON, err := json.Marshal(buildSmokeManifestConfig(strategy, supplierCode))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO supplier_driver_manifests (
  manifest_id, tenant_id, supplier_code, version_number, family, config_json,
  validation_status, validation_errors_json, is_active, created_by
) VALUES (
  $1, current_tenant_id(), $2, 1, $4, $3::jsonb,
  'valid', '[]'::jsonb, TRUE, 'smoke'
)
ON CONFLICT (tenant_id, supplier_code, version_number) DO UPDATE SET
  config_json = EXCLUDED.config_json,
  validation_status = 'valid',
  validation_errors_json = '[]'::jsonb,
  is_active = TRUE,
  updated_at = NOW()
`, "manifest_"+newUUID(), supplierCode, string(configJSON), family)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func buildSmokeManifestConfig(strategy, supplierCode string) map[string]any {
	sellerName := strings.ToLower(strings.TrimSpace(supplierCode)) + "_marketplace"
	base := map[string]any{
		"strategy":       strategy,
		"timeoutSeconds": 8,
		"maxRetries":     2,
		"sellerName":     sellerName,
	}

	switch strategy {
	case "http.vtex_persisted_query.v1":
		baseURL := strings.TrimSpace(os.Getenv("MS_SMOKE_BASE_URL"))
		operationName := strings.TrimSpace(os.Getenv("MS_SMOKE_OPERATION_NAME"))
		sha256Hash := strings.TrimSpace(os.Getenv("MS_SMOKE_SHA256_HASH"))
		preferredSeller := strings.TrimSpace(os.Getenv("MS_SMOKE_PREFERRED_SELLER_NAME"))
		allowFallback := strings.TrimSpace(os.Getenv("MS_SMOKE_ALLOW_FALLBACK"))
		requireAvailableOffer := strings.TrimSpace(os.Getenv("MS_SMOKE_REQUIRE_AVAILABLE_OFFER"))
		lookupVariableName := strings.TrimSpace(os.Getenv("MS_SMOKE_LOOKUP_VAR"))
		pricePath := strings.TrimSpace(os.Getenv("MS_SMOKE_PRICE_PATH"))
		sellerPath := strings.TrimSpace(os.Getenv("MS_SMOKE_SELLER_PATH"))
		channelPath := strings.TrimSpace(os.Getenv("MS_SMOKE_CHANNEL_PATH"))
		if lookupVariableName == "" {
			lookupVariableName = "term"
		}
		if pricePath == "" {
			pricePath = "data.products.0.price"
		}
		if sellerPath == "" {
			sellerPath = "data.products.0.seller"
		}
		if channelPath == "" {
			channelPath = "data.products.0.channel"
		}
		base["baseUrl"] = baseURL
		base["operationName"] = operationName
		base["sha256Hash"] = sha256Hash
		base["lookupVariableName"] = lookupVariableName
		base["pricePath"] = pricePath
		base["sellerPath"] = sellerPath
		base["channelPath"] = channelPath
		if preferredSeller != "" {
			base["preferredSellerName"] = preferredSeller
		}
		if allowFallback != "" {
			base["allowFallbackFirstProduct"] = strings.EqualFold(allowFallback, "1") || strings.EqualFold(allowFallback, "true") || strings.EqualFold(allowFallback, "yes") || strings.EqualFold(allowFallback, "on")
		}
		if requireAvailableOffer != "" {
			base["requireAvailableOffer"] = strings.EqualFold(requireAvailableOffer, "1") || strings.EqualFold(requireAvailableOffer, "true") || strings.EqualFold(requireAvailableOffer, "yes") || strings.EqualFold(requireAvailableOffer, "on")
		}
	case "http.html_search.v1":
		searchURLTemplate := strings.TrimSpace(os.Getenv("MS_SMOKE_SEARCH_URL_TEMPLATE"))
		priceRegex := strings.TrimSpace(os.Getenv("MS_SMOKE_PRICE_REGEX"))
		sellerRegex := strings.TrimSpace(os.Getenv("MS_SMOKE_SELLER_REGEX"))
		if priceRegex == "" {
			priceRegex = `(\\d{1,3}(?:\\.\\d{3})*,\\d{2}|\\d+(?:\\.\\d{2})?)`
		}
		base["baseUrl"] = strings.TrimSpace(os.Getenv("MS_SMOKE_BASE_URL"))
		base["searchUrlTemplate"] = searchURLTemplate
		base["priceRegex"] = priceRegex
		if sellerRegex != "" {
			base["sellerRegex"] = sellerRegex
		}
	case "http.leroy_search_sellers.v1":
		base["searchUrlTemplate"] = strings.TrimSpace(os.Getenv("MS_SMOKE_SEARCH_URL_TEMPLATE"))
		base["sellersUrlTemplate"] = strings.TrimSpace(os.Getenv("MS_SMOKE_SELLERS_URL_TEMPLATE"))
		region := strings.TrimSpace(os.Getenv("MS_SMOKE_REGION"))
		sellerPickStrategy := strings.TrimSpace(os.Getenv("MS_SMOKE_SELLER_PICK_STRATEGY"))
		if region != "" {
			base["region"] = region
		}
		if sellerPickStrategy != "" {
			base["sellerPickStrategy"] = sellerPickStrategy
		}
	case "http.html_dom_first_card.v1":
		base["searchUrlTemplate"] = strings.TrimSpace(os.Getenv("MS_SMOKE_SEARCH_URL_TEMPLATE"))
		base["cardRootHint"] = strings.TrimSpace(os.Getenv("MS_SMOKE_CARD_ROOT_HINT"))
		cardItemHint := strings.TrimSpace(os.Getenv("MS_SMOKE_CARD_ITEM_HINT"))
		titleHint := strings.TrimSpace(os.Getenv("MS_SMOKE_TITLE_HINT"))
		priceHint := strings.TrimSpace(os.Getenv("MS_SMOKE_PRICE_HINT"))
		listPriceHint := strings.TrimSpace(os.Getenv("MS_SMOKE_LIST_PRICE_HINT"))
		calculatedPriceHint := strings.TrimSpace(os.Getenv("MS_SMOKE_CALCULATED_PRICE_HINT"))
		pricePriority := strings.TrimSpace(os.Getenv("MS_SMOKE_PRICE_PRIORITY"))
		if cardItemHint != "" {
			base["cardItemHint"] = cardItemHint
		}
		if titleHint != "" {
			base["titleHint"] = titleHint
		}
		if priceHint != "" {
			base["priceHint"] = priceHint
		}
		if listPriceHint != "" {
			base["listPriceHint"] = listPriceHint
		}
		if calculatedPriceHint != "" {
			base["calculatedPriceHint"] = calculatedPriceHint
		}
		if pricePriority != "" {
			base["pricePriority"] = pricePriority
		}
	case "playwright.pdp_first.v1":
		startURL := strings.TrimSpace(os.Getenv("MS_SMOKE_START_URL"))
		searchURL := strings.TrimSpace(os.Getenv("MS_SMOKE_SEARCH_URL"))
		waitUntil := strings.TrimSpace(os.Getenv("MS_SMOKE_WAIT_UNTIL"))
		priceRegex := strings.TrimSpace(os.Getenv("MS_SMOKE_PRICE_REGEX"))
		headless := strings.TrimSpace(os.Getenv("MS_SMOKE_HEADLESS"))
		fallbackSearchEnabled := strings.TrimSpace(os.Getenv("MS_SMOKE_FALLBACK_SEARCH_ENABLED"))
		selectorsJSON := strings.TrimSpace(os.Getenv("MS_SMOKE_PDP_SELECTORS_JSON"))
		maxRetries := strings.TrimSpace(os.Getenv("MS_SMOKE_MAX_RETRIES"))

		base["startUrl"] = startURL
		base["searchUrl"] = searchURL
		if waitUntil != "" {
			base["waitUntil"] = waitUntil
		}
		if priceRegex != "" {
			base["priceRegex"] = priceRegex
		}
		if headless != "" {
			base["headless"] = strings.EqualFold(headless, "1") || strings.EqualFold(headless, "true") || strings.EqualFold(headless, "yes") || strings.EqualFold(headless, "on")
		}
		if fallbackSearchEnabled != "" {
			base["fallbackSearchEnabled"] = strings.EqualFold(fallbackSearchEnabled, "1") || strings.EqualFold(fallbackSearchEnabled, "true") || strings.EqualFold(fallbackSearchEnabled, "yes") || strings.EqualFold(fallbackSearchEnabled, "on")
		}
		if maxRetries != "" {
			if parsed, err := parseInt(maxRetries); err == nil {
				base["maxRetries"] = parsed
			}
		}

		pdpSelectors := map[string]any{
			"price":  strings.TrimSpace(os.Getenv("MS_SMOKE_PDP_PRICE_SELECTOR")),
			"seller": strings.TrimSpace(os.Getenv("MS_SMOKE_PDP_SELLER_SELECTOR")),
		}
		if channelSel := strings.TrimSpace(os.Getenv("MS_SMOKE_PDP_CHANNEL_SELECTOR")); channelSel != "" {
			pdpSelectors["channel"] = channelSel
		}
		if selectorsJSON != "" {
			tmp := map[string]any{}
			if err := json.Unmarshal([]byte(selectorsJSON), &tmp); err == nil && len(tmp) > 0 {
				pdpSelectors = tmp
			}
		}
		base["pdpSelectors"] = pdpSelectors
	default:
		base["strategy"] = "http.mock.v1"
		base["endpointTemplate"] = "mock://pricing/{supplier_code}/{product_id}"
		base["mockMultiplier"] = 1.04
	}
	return base
}

func parseInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty")
	}
	var out int
	_, err := fmt.Sscanf(value, "%d", &out)
	if err != nil {
		return 0, err
	}
	return out, nil
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

func insertOutboxEvent(ctx context.Context, pool *pgxpool.Pool, tenantID, eventID, runRequestID, inputMode string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := setTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	payloadJSON, _ := json.Marshal(map[string]any{
		"run_request_id": runRequestID,
		"tenant_id":      tenantID,
		"input_mode":     inputMode,
	})
	idempotencyKey := fmt.Sprintf("shopping_run_requested:%s", runRequestID)
	_, err = tx.Exec(ctx, `
INSERT INTO outbox_events (
  event_id, aggregate_type, aggregate_id, event_name, event_version,
  tenant_id, trace_id, idempotency_key, payload_json, status, available_at, published_at
) VALUES (
  $1, 'shopping_run_request', $2, 'shopping.run_requested', 'v1',
  $3, 'smoke', $4, $5::jsonb, 'published', NOW(), NOW()
) ON CONFLICT (idempotency_key) DO NOTHING
`, eventID, runRequestID, tenantID, idempotencyKey, string(payloadJSON))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
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
