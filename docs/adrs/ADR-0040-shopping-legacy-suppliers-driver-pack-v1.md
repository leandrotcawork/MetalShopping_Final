# ADR-0040: Shopping Legacy Suppliers Driver Pack v1 (TELHA_NORTE, LEROY, ABC)

- Status: draft
- Date: 2026-03-20

## Context

We already implemented the Shopping driver runtime and governance baseline:

- Driver selection by `(family, strategy)` (ADR-0030).
- Runtime package layout under `apps/integration_worker/src/shopping_price_runtime/*` (ADR-0031).
- Bounded parallelism and rate limits (ADR-0032).
- Multi-supplier smoke suite framework (ADR-0033).
- Non-mock Playwright PDP-first pilot for `OBRA_FACIL` (ADR-0034).

However, we still lack legacy parity on supplier coverage. In the legacy codebase we have supplier-specific drivers for:

- `TELHA_NORTE` (VTEX persisted query over GraphQL)
- `LEROY` (search HTML -> product page -> sellers JSON, region-aware)
- `ABC` (HTML search results parsing)

In the current repo, the worker runtime already supports:

- HTTP family:
  - `http.vtex_persisted_query.v1`
  - `http.html_search.v1`
  - `http.mock.v1`
- Playwright family:
  - `playwright.pdp_first.v1`
  - `playwright.mock.v1`

But the missing suppliers require either:

- new strategy implementations (framework extension), or
- clear manifest config patterns and seed paths (supplier pack).

We must do this without violating the frozen platform rules:

- Workers write to Postgres tables; they do not call `server_core` HTTP.
- `server_core` owns supplier directory + driver manifests and validates activation.
- Tenancy is enforced via RLS and `current_tenant_id()`.

## Decision

Deliver a “legacy suppliers driver pack” for tenant `tenant_default` that makes the missing suppliers runnable end-to-end under the same governance rules as existing suppliers.

### Supplier mapping (v1)

- `TELHA_NORTE`:
  - `family=http`
  - `strategy=http.vtex_persisted_query.v1`
  - required config keys:
    - `baseUrl`, `operationName`, `sha256Hash`
  - recommended config keys:
    - `skusFilter`, `toN`, `includeVariant`
    - `requireAvailableOffer=true`
    - `preferredSellerName` (optional)

- `LEROY`:
  - introduce a dedicated HTTP strategy for multi-step flow:
    - `family=http`
    - `strategy=http.leroy_search_sellers.v1`
  - config includes:
    - `searchUrlTemplate`
    - `sellersUrlTemplate`
    - `region` / `x-region` default
    - seller selection policy (`selected|min_sale`)

- `ABC`:
  - introduce an HTML DOM-oriented strategy (not regex-only):
    - `family=http`
    - `strategy=http.html_dom_first_card.v1`
  - config includes:
    - `searchUrlTemplate`
    - selectors for product card, title, and price fields
    - optional normalization rules for BRL price extraction

### Seed + activation rule

The supplier directory and the initial active driver manifests must be seeded for `tenant_default` so the Web UI bootstrap shows the suppliers and the worker can run them without “env override” hacks.

## Contracts (touchpoints)

### Manifest schema

- Extend `contracts/api/jsonschema/supplier_driver_manifest_http_v1.schema.json` to include:
  - `http.leroy_search_sellers.v1`
  - `http.html_dom_first_card.v1`

### Go validation (activation safety)

- Extend `apps/server_core/internal/platform/suppliers/driver_family_registry.go`:
  - validate required fields for the new strategies
  - validate optional bounds (timeouts, retries) consistently

## Implementation Checklist

1. Draft per-strategy ADRs for:
   - `http.leroy_search_sellers.v1`
   - `http.html_dom_first_card.v1`
   Skill: `metalshopping-adr-updates`
2. Implement the strategies in `apps/integration_worker/src/shopping_price_runtime/http/*`
   Skill: `metalshopping-module-implementation` (worker step) + `metalshopping-worker-scaffold`
3. Update schema + Go registry validation for the new strategies
   Skill: `metalshopping-module-implementation`
4. Seed suppliers directory + manifests for `tenant_default` and activate them
   Skill: `metalshopping-module-implementation`
5. Extend `scripts/smoke_shopping_driver_suite_local.ps1` config to include the new strategies without env-only behavior
   Skill: `metalshopping-module-implementation`
6. Acceptance evidence:
   - run driver suite and attach the report row set
   - show `bootstrap` returns all expected suppliers for `tenant_default`

## Acceptance Evidence (for Status: accepted)

- `GET /api/v1/shopping/bootstrap` (tenant `tenant_default`) returns enabled suppliers:
  - `DEXCO`, `TELHA_NORTE`, `CONDEC`, `OBRA_FACIL`, `LEROY`, `ABC`
- Driver suite (`scripts/smoke_shopping_driver_suite_local.ps1`) includes:
  - `TELHA_NORTE` non-mock HTTP VTEX strategy
  - `LEROY` non-mock strategy
  - `ABC` non-mock strategy
- DB evidence:
  - `shopping_price_observations` contains rows with `channel` matching the family and `item_status` not all `ERROR`.

## Consequences

- Adding suppliers becomes “manifest config first” for strategy-supported patterns.
- New strategy work stays bounded to explicit executor modules with validation and schema touchpoints.
- Shopping UI bootstrap in `tenant_default` becomes legacy-complete for the initial supplier set.

