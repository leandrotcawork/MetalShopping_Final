# ADR-0020: Shopping Price Observation Data Model v1

- Status: accepted
- Date: 2026-03-19

## Context

The existing Shopping Level 1 schema is a minimal read surface:

- runs
- run items
- latest snapshot

The legacy Shopping core produces richer per-supplier observations:

- supplier dimension
- per-item status (OK, NOT_FOUND, AMBIGUOUS, ERROR)
- optional product URL, HTTP status, timing, and selection/debug fields
- per-day dedupe and/or latest-by-entity materialization

To implement the real Shopping workflow and preserve the legacy UI usefulness, the platform needs an explicit, tenant-safe observation model that supports:

- idempotent upserts by natural keys
- progress reporting
- latest snapshot serving
- optional future history retention without redesign

## Decision

Shopping Price v1 will model observations with an explicit supplier dimension and idempotent natural keys.

The authoritative entities are:

- Run: lifecycle and progress
- Run Request: input scope and execution parameters
- Observation: (run, product, supplier) result record
- Latest Snapshot: latest observation per (product, supplier)

Minimum required semantics:

- Run IDs are UUID strings.
- Observations are keyed by `(tenant_id, run_id, product_id, supplier_code)`.
- Latest snapshots are keyed by `(tenant_id, product_id, supplier_code)`.
- Worker writes are idempotent via `ON CONFLICT ... DO UPDATE`.
- Run progress is updated by the worker (`processed_items`, `total_items`, `run_status`).

Item status vocabulary (v1):

- `OK`
- `NOT_FOUND`
- `AMBIGUOUS`
- `ERROR`

Optional debug fields are allowed, but must not leak sensitive data:

- `http_status`
- `elapsed_s`
- `notes`
- `chosen_seller_json`
- `product_url`

## Contracts (touchpoints)

- Data model (tenant-scoped + RLS):
  - `apps/server_core/migrations/0023_shopping_price_observation_model_v1.sql`
  - plus the read surface baseline in `apps/server_core/migrations/0020_shopping_price_read_surfaces.sql`
- OpenAPI: `contracts/api/openapi/shopping_v1.openapi.yaml`
  - `GET /api/v1/shopping/runs` and `GET /api/v1/shopping/runs/{run_id}` (progress + counts)
  - `GET /api/v1/shopping/products/{product_id}/latest` (latest snapshot)
- JSON Schemas (v1): `contracts/api/jsonschema/shopping_run_v1.schema.json`, `shopping_run_list_v1.schema.json`, `shopping_product_latest_v1.schema.json`
- Events: none required in v1 (worker writes tables; server reads)
- Governance: none required in v1

## Implementation checklist

- Worker upserts must be idempotent (`ON CONFLICT ... DO UPDATE`) and tenant-scoped (ADR-0022).
- Never log sensitive discovery/debug payloads; keep logs structured and redacted when needed.
- Prefer latest snapshots for UI reads; treat per-run observations as operational detail.

## Consequences

- The UI can show per-supplier results and meaningful status counts.
- The worker can be restarted safely without duplicating or corrupting results.
- The schema remains compatible with future history (time-series) tables if needed.

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Contract authoring for new surfaces (items, bootstrap): `metalshopping-openapi-contracts`
- Go read/write module changes: `metalshopping-module-scaffold`
- Worker upsert model and output tables: `metalshopping-worker-scaffold`
- Platform tenancy review (RLS, keys): `metalshopping-platform-packages` and `metalshopping-observability-security`
