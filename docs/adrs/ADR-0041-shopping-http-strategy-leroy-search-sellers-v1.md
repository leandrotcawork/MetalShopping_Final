# ADR-0041: Shopping HTTP Strategy `http.leroy_search_sellers.v1`

- Status: draft
- Date: 2026-03-20

## Context

The legacy supplier driver `LEROY` is not a VTEX persisted query and is not reliably solvable with a single HTML regex pass.

Legacy behavior (confirmed by the legacy driver implementation):

1. Search HTML by EAN/term
2. Follow the product page redirect (final URL includes an internal product id)
3. Fetch sellers JSON from `/api/v3/products/<id>/sellers` using `x-region` headers
4. Pick seller and price based on policy (`selected` preferred, otherwise min sale price)

Our current HTTP strategy set does not support multi-step flows with a second HTTP call.

## Decision

Add a dedicated HTTP strategy:

- `family=http`
- `strategy=http.leroy_search_sellers.v1`

This strategy performs the 2-step HTTP flow and returns a normalized observation:

- `item_status` in `OK|NOT_FOUND|AMBIGUOUS|ERROR`
- `observed_price` (sale) + seller name
- `note` field includes which sub-step was used (search/html/json)

## Config contract (manifest `config_json`)

Required:

- `strategy = "http.leroy_search_sellers.v1"`
- `searchUrlTemplate`
  - supports `{term}` and `{lookup_mode}`
- `sellersUrlTemplate`
  - supports `{product_id}`

Optional:

- `region` (default region slug)
- `sellerPickStrategy`:
  - `selected` (default)
  - `min_sale`
- `headers`:
  - additional headers (merged with runtime defaults)
- `timeoutSeconds`, `maxRetries`, `maxConcurrency`, `retryHttpStatuses` (existing HTTP knobs)

## Contracts (touchpoints)

- Extend `contracts/api/jsonschema/supplier_driver_manifest_http_v1.schema.json` enum + `oneOf` to include `http.leroy_search_sellers.v1` required keys.
- Extend `apps/server_core/internal/platform/suppliers/driver_family_registry.go` to validate the new strategy and required keys.

## Implementation notes

- Parsing product id:
  - extract from the final product URL (`_(\\d{5,})`) as in legacy, and fallback to HTML `ld+json` when possible.
- Region handling:
  - include `x-region` header in the sellers JSON request
  - if config has no region, default to a stable dev region (legacy used `uberlandia`)
- Retry policy:
  - retries apply per request; sellers JSON errors should not cause infinite search retries

## Implementation Checklist

1. Add strategy executor under:
   - `apps/integration_worker/src/shopping_price_runtime/http/strategies.py`
2. Extend HTTP family validation in:
   - `apps/server_core/internal/platform/suppliers/driver_family_registry.go`
3. Extend JSON schema:
   - `contracts/api/jsonschema/supplier_driver_manifest_http_v1.schema.json`
4. Add smoke suite entry and acceptance evidence for `LEROY`.

## Acceptance Evidence (for Status: accepted)

- Driver suite includes `LEROY` with `item_status=OK` for at least 1 real catalog product id in `tenant_default`.
- Observation row includes:
  - `channel=HTTP` (or `HTTP_LEROY`)
  - `http_status` set for the sellers request
  - `note` describing which extraction path was used.

## Consequences

- Adds a new bounded strategy without changing the family model.
- Enables LEROY without Playwright dependency, keeping cost and flakiness lower.

