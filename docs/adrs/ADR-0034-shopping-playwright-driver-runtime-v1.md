# ADR-0034: Shopping Playwright Driver Runtime v1 (PDP-first, non-mock)

- Status: accepted
- Date: 2026-03-20

## Context

The Playwright family exists as a strategy contract and dispatcher slot, but the current target implementation is still effectively mock/disabled for real browsing.

Legacy Shopping uses a Playwright PDP-first approach for OBRA_FACIL and similar suppliers. To claim "backend 100% complete", we need at least one real Playwright driver implemented end-to-end under the same governance and execution rules as HTTP suppliers.

Constraints:

- The worker must not call `server_core` HTTP.
- The worker must be tenant-safe on persistence.
- The runtime must be bounded (tabs, timeouts, circuit breaker).
- The driver must be config-driven (selectors and navigation behavior live in manifest config).

## Decision

Implement `playwright.pdp_first.v1` as a real runtime strategy for the Shopping Price worker, and deliver OBRA_FACIL as the pilot non-mock Playwright supplier.

Rules:

- Playwright execution is controlled by manifest config:
  - `headless`, `waitUntil`, `timeoutSeconds`, `pdpSelectors`
- The runtime must support:
  - persisted product URL first (ADR-0024 signal)
  - fallback search only when explicitly enabled (bounded max)
- Anti-block behavior is handled as:
  - explicit `item_status=ERROR` with a note that indicates block class (e.g. cloudflare 1015)
  - bounded retry with backoff (small, conservative)

## Contracts (touchpoints)

- JSON Schema:
  - evolve `contracts/api/jsonschema/supplier_driver_manifest_playwright_v1.schema.json` as needed (selectors + runtime options)
- Go validation:
  - evolve `apps/server_core/internal/platform/suppliers/driver_family_registry.go` to validate required keys and bounds

## Implementation Checklist

1. Contract review for Playwright config keys
   Skill: `metalshopping-contract-authoring` + `metalshopping-adr-updates`
2. Implement Playwright runtime executor for `playwright.pdp_first.v1`
   Skill: `metalshopping-worker-scaffold` + `metalshopping-worker-patterns`
3. Implement OBRA_FACIL manifest seed defaults and smoke
   Skill: `metalshopping-worker-patterns`
4. Observability/security review (browser automation risks)
   Skill: `metalshopping-observability-security`

## Implementation Snapshot (2026-03-20)

- Runtime implementation delivered for `playwright.pdp_first.v1`:
  - real browser navigation via Playwright Chromium (lazy import; explicit error when dependency missing)
  - persisted signal URL first, then `startUrl`, optional `searchUrl` fallback
  - selector-driven extraction (`pdpSelectors.price`, `seller`, `channel`)
  - anti-block markers (`cloudflare`, `1015`, `captcha`, `access denied`)
  - bounded retries/backoff and timeout-aware execution
- Contracts and validation evolved:
  - `supplier_driver_manifest_playwright_v1.schema.json`: `searchUrlTemplate`, `fallbackSearchEnabled`, `maxRetries` (plus ADR-0032 keys)
  - `driver_family_registry.go`: validates `pdpSelectors.price`, `maxRetries`, and optional template fields
- Smoke seed flow evolved:
  - `apps/server_core/cmd/smoke-shopping-event/main.go` now seeds `PLAYWRIGHT` execution/family when strategy prefix is `playwright.`
  - supports Playwright config env keys (`MS_SMOKE_START_URL`, `MS_SMOKE_SEARCH_URL`, `MS_SMOKE_PDP_*`, etc.)
- Local runtime non-mock validation:
  - Playwright + Chromium installed in `.venv`
  - direct runtime smoke produced:
    - `item_status=OK`
    - `channel=PLAYWRIGHT`
    - `observed_price=123.45`

Acceptance completed:
- DB-backed run confirms Playwright execution with `item_status=OK` and `channel=PLAYWRIGHT` for OBRA_FACIL.

## Acceptance Evidence (for Status: accepted)

- Smoke:
  - DB run: `run_id=23289388-75b5-4c9f-89e4-5dcc6003eeea` (tenant `tenant_default`) includes `item_status=OK`, `channel=PLAYWRIGHT`, `http_status=200`.
- Report:
  - included in the driver suite report (ADR-0033).

## Consequences

- Backend driver framework includes both major families (HTTP + Playwright) in non-mock form.
- Adding new Playwright suppliers becomes a config exercise (selectors + policy) rather than bespoke scripts.
