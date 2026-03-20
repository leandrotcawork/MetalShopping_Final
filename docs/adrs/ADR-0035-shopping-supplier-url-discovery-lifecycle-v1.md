# ADR-0035: Shopping Supplier URL Discovery Lifecycle v1

- Status: accepted
- Date: 2026-03-20

## Context

Playwright suppliers depend on pre-known PDP URLs to avoid Cloudflare blocks and keep runs stable. Today we store URLs in `shopping_supplier_product_signals`, but discovery is naive: products that fail URL discovery can be retried on every run, which does not scale and wastes compute.

We already observed the operational failure mode:

- initial URL discovery finds a partial set of URLs (e.g. 300/1000)
- the remaining products are repeatedly retried on every run
- the “cost” grows with catalog size, not with new additions
- repeated discovery retries increase block risk and noise in logs

We need a deterministic lifecycle that:

- discovers URLs incrementally
- avoids reprocessing NOT_FOUND products on every run
- schedules refresh/retry windows safely
- keeps Playwright execution predictable as catalog size grows

## Decision

Introduce an explicit URL discovery lifecycle in `shopping_supplier_product_signals` with cooldown scheduling and deterministic rules for when the worker may attempt discovery.

Rules:

- The source of truth for supplier URLs remains `shopping_supplier_product_signals`.
- URL discovery only runs for `PLAYWRIGHT` suppliers.
- URL discovery is *best-effort* enrichment. It must never be a synchronous dependency of API reads.
- A product is eligible for discovery when:
  - `product_url IS NULL` (we do not have a known PDP URL), and
  - `url_status IN ('STALE','INVALID')`, and
  - `next_discovery_at IS NULL OR next_discovery_at <= now()`, and
  - `manual_override = false` (manual URLs are authoritative).
- Outcomes:
  - `FOUND`: set `product_url`, `url_status='ACTIVE'`, clear `next_discovery_at`, reset `not_found_count`.
  - `NOT_FOUND`: set `next_discovery_at = now() + discovery_retry_window`, increment `not_found_count`, keep `url_status='STALE'`.
  - `BLOCKED`: set `next_discovery_at = now() + blocked_retry_window`, keep `url_status='STALE'`.
  - `INVALIDATED` (HTTP 404/410 on a stored URL): set `product_url=NULL`, set `url_status='INVALID'` and `next_discovery_at = now() + invalid_retry_window`.
- Manual overrides (`manual_override = true`) are never altered by automatic discovery.
- Default retry windows:
  - `discovery_retry_window = 30 days`
  - `blocked_retry_window = 7 days`
  - `invalid_retry_window = 30 days`

### Data model changes

Add lifecycle fields to the existing `shopping_supplier_product_signals` table:

- `next_discovery_at TIMESTAMPTZ NULL`
  - earliest timestamp when the worker is allowed to re-attempt discovery for this row
- `not_found_count INTEGER NOT NULL DEFAULT 0`
  - monotonically increases only when discovery yields `NOT_FOUND` (used for ops/debug, not for branching logic in v1)

Constraints:

- `not_found_count >= 0`

Indexes:

- add an index to support eligibility queries:
  - `(tenant_id, supplier_code, next_discovery_at)` where `product_url IS NULL AND manual_override = false`

### Worker behavior (non-negotiable)

- Strategy executors must remain DB-agnostic:
  - they receive `SupplierSignal` and manifest config and return a `RuntimeObservation`
  - they must not query Postgres directly
- The orchestration layer (worker entrypoint) owns URL discovery scheduling decisions:
  - when `product_url IS NULL`, discovery is only attempted if eligibility rules pass
  - on `NOT_FOUND`/`BLOCKED`, the worker must persist the cooldown via `next_discovery_at` so subsequent runs skip
- Discovery must be bounded:
  - per-run cap for how many items may attempt discovery (avoid a discovery “storm”)
  - conservative default (e.g. 50) and validated upper bound (e.g. 500) once surfaced in config/governance

## Contracts (touchpoints)

- OpenAPI:
  - `contracts/api/openapi/shopping_v1.openapi.yaml`
- JSON Schemas:
  - `contracts/api/jsonschema/shopping_supplier_signal_v1.schema.json`
  - `contracts/api/jsonschema/shopping_upsert_supplier_signal_request_v1.schema.json`
- Governance (optional, if we need runtime-configurable windows):
  - `contracts/governance/policies/shopping_supplier_url_discovery_v1.policy.json`

## Implementation Checklist

Frozen execution order for this ADR:

1. Contracts (OpenAPI + Schemas)
   Skill: `metalshopping-openapi-contracts` (and `metalshopping-contract-authoring` if multiple contract types change)
2. Migration (tenant-scoped + RLS preserved)
   Skill: `metalshopping-adr-updates` (review) + normal migration flow in `apps/server_core/migrations`
3. Go module behavior (tenant-safe read/write)
   Skill: `metalshopping-module-scaffold` (review: `metalshopping-server-core-modules`)
4. Worker discovery scheduling (cooldowns enforced)
   Skill: `metalshopping-worker-scaffold` (and `metalshopping-worker-patterns`)
5. SDK generation
   Skill: `metalshopping-sdk-generation`
6. Observability/security review
   Skill: `metalshopping-observability-security`

## Acceptance Evidence (for Status: accepted)

- Migration:
  - new discovery lifecycle columns added to `shopping_supplier_product_signals`
- Build/test:
  - `go build ./apps/server_core/...` -> pass
  - `go test ./apps/server_core/...` -> pass
- Smoke:
  - Playwright supplier respects `next_discovery_at` cooldown and does not reprocess NOT_FOUND products on every run.
  - Event smoke with catalog scope executed:
    - `run_request_id=1f07ab5f-22b0-4440-bde0-e9c9eeea25e2`
    - `run_id=80a0f6a9-9d34-45a4-928c-d150d2e9dbdc`
    - `rows_written=10`

## Alternatives considered

- Encode `NOT_FOUND`/`BLOCKED` directly into `url_status`:
  - rejected because it overloads the “URL health” semantic and expands a previously accepted vocabulary (ADR-0024) in a breaking way.
- Create a separate URL discovery table:
  - rejected for v1 because it introduces joins and a second source-of-truth for a concept that is already row-scoped to `(tenant, product, supplier)`.

## Consequences

- URL discovery becomes predictable and scalable.
- Retry cost is bounded by explicit cooldown windows.
- Adds lifecycle fields that must be kept consistent by worker and API.
