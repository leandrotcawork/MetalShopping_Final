# ADR-0045: Shopping Manual URL Candidates From Catalog v1

- Status: accepted
- Date: 2026-03-20
- Updated: 2026-03-22

## Context

The Shopping manual URL panel (ADR-0039) lists data via `GET /api/v1/shopping/supplier-signals`, which reads from `shopping_supplier_product_signals`.

This works only when the signals table is already populated. After importing the full catalog (ADR-0043), it is common and expected for `shopping_supplier_product_signals` to be empty, which creates a UX dead-end:

- no rows are shown
- there is no way to add the first manual URLs

We also need to preserve the semantics from ADR-0035:

- `url_status` is the "URL health" vocabulary (`ACTIVE|STALE|INVALID`)
- discovery outcomes are represented via existing lifecycle fields (`next_discovery_at`, `not_found_count`, `last_success_at`, `last_http_status`, `last_error_message`) without expanding the `url_status` vocabulary

We need a scalable solution that:

- allows manual URL insertion even when no signal rows exist
- does not require seeding a cross product of `(all products x all suppliers)` into `shopping_supplier_product_signals`
- keeps the source-of-truth for URLs in `shopping_supplier_product_signals`

## Decision

Add a new Shopping endpoint that lists **catalog products as candidates** and overlays any existing signal state by `LEFT JOIN` to `shopping_supplier_product_signals`.

The manual URL panel will be backed by this endpoint instead of listing only persisted signals.

Key rules:

- The query is scoped to **one supplier** per request.
  - This avoids `products x suppliers` explosion and keeps paging deterministic.
- Results are paginated and filterable by catalog attributes (search, brand, taxonomy leaf0).
- The response includes:
  - catalog fields needed for the table UX
  - signal fields when a persisted signal row exists
- lifecycle fields needed for operational clarity, without introducing any new status vocabulary

## Execution (Codex $ms workflow)

Implementation and verification for this ADR follow the `$ms` orchestration workflow:

- Phase 1 (architectural thinking) and Phase 2 (plan) are captured in `tasks/todo.md` under **Feature: ADR-0045 Manual URL candidates**.
- Phase 3 (execute) is done one task at a time with a commit after each task.
- Phase 4 (review) is done before closing the ADR.

This ADR is considered **accepted** only after the acceptance evidence in this document is filled with real results and the close-out commit exists.

## Contracts (touchpoints)

- OpenAPI:
  - `contracts/api/openapi/shopping_v1.openapi.yaml`
- JSON Schemas:
  - new schema `shopping_manual_url_candidate_list_v1.schema.json`
  - reuse `shopping_supplier_signal_v1.schema.json` fields where appropriate

## API Shape (proposed)

New endpoint:

- `GET /api/v1/shopping/manual-url-candidates`

Required query params:

- `supplier_code` (single supplier)

Optional query params:

- `search` (catalog search)
- `brand_name`
- `taxonomy_leaf0_name`
- `include_existing` (default true; when false, return only rows with no stored `product_url`)
- `limit`, `offset`

Response:

- `rows[]` with:
  - `product_id`, `sku`, `name`, `brand_name`, `taxonomy_leaf0_name`
  - signal overlay fields: `product_url`, `url_status`, `manual_override`, `next_discovery_at`, `not_found_count`, `last_success_at`, `last_http_status`, `last_error_message`
- `paging` (same shape as other list endpoints)

Writes remain unchanged:

- manual URL save continues to call `PUT /api/v1/shopping/supplier-signals` (upsert creates the row when missing).

## Implementation Checklist

Source of truth for execution is `tasks/todo.md` (Feature: ADR-0045). Expected task breakdown:

1. T1: contract (OpenAPI + schemas)
2. T2: Go module (reader + handler + Postgres adapter)
3. T4: SDK generation (after contract)
4. T5: frontend binding (manual URL panel uses candidates endpoint)
5. T6: ADR close-out (this document) with real evidence

Rules:

- No task is marked `[x]` without: build passes + real data verified + commit made.
- The ADR status is updated to `accepted` only in T6 after evidence is captured.

## Acceptance Evidence (for Status: accepted)

API evidence captured on 2026-03-21 (local :8080, tenant_default, trace_id prefix `adr0045-`):

- With `shopping_supplier_product_signals` empty:
  - `GET /api/v1/shopping/supplier-signals?supplier_code=DEXCO&limit=1&offset=0` returned `total=0`
  - `GET /api/v1/shopping/manual-url-candidates?supplier_code=DEXCO&limit=10&offset=0` returned `total=3838`, `returned=10` (trace_id `adr0045-20260321-104738`)
- After saving a manual URL:
  - `PUT /api/v1/shopping/supplier-signals` with `productId=prd_0e49147e883e64a8cb07de8e` and `productUrl=https://example.com/adr-0045` returned `urlStatus=ACTIVE`, `manualOverride=true`
  - `GET /api/v1/shopping/manual-url-candidates?supplier_code=DEXCO&search=38481&limit=10&offset=0` returned row with `productUrl` set and `urlStatus=ACTIVE` (trace_id `adr0045-20260321-110652`)

Build evidence:

- `go build ./apps/server_core/...` passes
- `npm.cmd run web:typecheck` passes
- `npm.cmd run web:build` passes

UI evidence:

- Confirmed in browser on 2026-03-22 (local `:5173`, API `:8080`, tenant_default):
  - Manual URL panel lists candidate products even when signals are empty.
  - Saving a URL persists and the row reflects the signal overlay.
  - Panel refresh does not flicker/layout-jump; supplier=all and "Mostrar URLs cadastradas" behave as expected.

## Alternatives considered

- Seed "empty" signals for every `(product, supplier)`:
  - rejected because it scales as `N*M`, grows storage, and adds write amplification.
- Expand `GET /supplier-signals` to include missing rows:
  - rejected for v1 to keep existing contract semantics stable and avoid mixing persisted-only and candidate rows in one endpoint.

## Consequences

- Manual URL insertion works even when the signals table is empty.
- The solution scales by requiring supplier scoping and paging.
