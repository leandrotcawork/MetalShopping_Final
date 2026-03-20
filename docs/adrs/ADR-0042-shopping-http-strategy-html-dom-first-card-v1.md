# ADR-0042: Shopping HTTP Strategy `http.html_dom_first_card.v1`

- Status: accepted
- Date: 2026-03-20

## Context

Legacy suppliers like `ABC` (and partially `CONDEC`) rely on parsing structured HTML search result pages.

The current `http.html_search.v1` strategy is regex-first and can become unstable:

- multiple unrelated numbers in a page can cause false positives
- price formatting varies by store
- we need a predictable “first card” extraction path

Legacy driver code uses a DOM-aware parser that targets a stable container (`spots-list` for ABC) and extracts:

- product title
- calculated price / sale price / list price (with priority rules)

## Decision

Add a DOM-oriented HTTP strategy:

- `family=http`
- `strategy=http.html_dom_first_card.v1`

The strategy extracts the first matching product card from a search results HTML page using config-provided selectors/markers, then extracts a price from a bounded scope.

This is intentionally not a general-purpose crawler; it is a bounded “search results first card” extractor for suppliers without stable JSON search APIs.

## Config contract (manifest `config_json`)

Required:

- `strategy = "http.html_dom_first_card.v1"`
- `searchUrlTemplate` (supports `{term}`)
- `cardRootHint`
  - a stable marker to find the results container (id/class/name token)

Optional (recommended for correctness):

- `titleHint`
- `priceHint`
- `listPriceHint`
- `calculatedPriceHint`
- `pricePriority`:
  - `calculated_first` (default)
  - `sale_first`
- `timeoutSeconds`, `maxRetries`, `maxConcurrency`, `retryHttpStatuses`

## Contracts (touchpoints)

- Extend `contracts/api/jsonschema/supplier_driver_manifest_http_v1.schema.json` to include `http.html_dom_first_card.v1`.
- Extend `apps/server_core/internal/platform/suppliers/driver_family_registry.go` to validate required keys.

## Implementation notes

- Use a lightweight parser:
  - Python `html.parser` (stdlib) is acceptable and already proven in legacy.
- Keep extraction deterministic:
  - search within container only
  - cap text lengths and parsing work
- Price decoding:
  - reuse existing BRL decoding helpers from runtime shared parsing when possible.

## Implementation Checklist

1. Implement executor under:
   - `apps/integration_worker/src/shopping_price_runtime/http/strategies.py`
2. Update registry validation + schema.
3. Seed `ABC` manifest using this strategy and validate via driver suite.

## Acceptance Evidence (for Status: accepted)

- Smoke (tenant `tenant_default`) produced `OK` for `ABC` using real HTTP fetch:
  - `run_request_id`: `aba7655c-d422-4960-8765-8627638fad47`
  - `run_id`: `91e4dc66-9660-414d-8072-566fbe82d690`
  - sample: `product_id=prd_smoke_abc_001`, `observed_price=198.9000`, `http_status=200`, `lookup_term=deca`
  - `notes`: `html_dom_first_card attempt=1 hint_priority:calculated_first`
- No regression in `http.html_search.v1` and `http.vtex_persisted_query.v1`.

## Consequences

- Reduces fragility for HTML-only suppliers.
- Keeps strategy surface bounded and config-driven for future HTML-only suppliers.
