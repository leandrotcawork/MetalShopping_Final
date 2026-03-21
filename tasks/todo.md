# tasks/todo.md
# Current feature state. Updated by `skills/ms/SKILL.md` during implementation.

## Feature: ADR-0045 Manual URL candidates
Type: read-only | Events: no | ADR: ADR-0045

## Phase 1 - Architectural thinking

Module type:
- read-only

Module location:
- `apps/server_core/internal/modules/shopping`
  - extend existing `ports/read_models.go`, `application/service.go`, `adapters/postgres/reader.go`, `transport/http/handler.go`

API:
- `GET /api/v1/shopping/manual-url-candidates`
  - required: `supplier_code` (single supplier per request)
  - optional: `search`, `brand_name`, `taxonomy_leaf0_name`, `include_existing` (default true), `limit`, `offset`

Data model:
- Base rows: `catalog_products`
- Overlay: `shopping_supplier_product_signals` (LEFT JOIN by `(product_id, supplier_code)`)
- Writes: unchanged (manual URL save remains via existing upsert on supplier signals)

Risks to manage:
- Query scale: must remain single-supplier + paginated to avoid `products x suppliers` explosion.
- Query performance: ensure indexes for join/filter columns (e.g. `(tenant_id, supplier_code, product_id)`).
- Tenant safety: every query must use `pgdb.BeginTenantTx` and `tenant_id = current_tenant_id()` on all tenant-scoped tables.

Level 1 scope:
- Endpoint returns real data (no mocks) and the manual URL panel shows rows even when signals are empty.

## Phase 2 - Plan (wait for approval, then execute T1..T6)

## Tasks
- [ ] T1: contract - $metalshopping-openapi-contracts
      commit: "feat(shopping): add manual URL candidates endpoint contract"
- [x] T2: Go module - reader + handler + postgres adapter
      commit: "feat(shopping): implement manual URL candidates list endpoint"
- [x] T2b: fix upsert supplier signals type casting
      commit: "fix(shopping): make supplier signal upsert type-safe"
- [x] T4: SDK - $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate after shopping contract"
- [x] T5: frontend - $metalshopping-frontend
      commit: "feat(web): manual URL panel uses candidates endpoint"
- [ ] T6: ADR close-out - capture evidence and accept ADR-0045
      commit: "docs(adr): ADR-0045 manual URL candidates - verified and closed"

## Acceptance tests
- [x] `go build ./...` passes
- [x] `npm.cmd run web:typecheck` passes
- [x] `npm.cmd run web:build` passes
- [x] With `shopping_supplier_product_signals` empty:
      `GET /api/v1/shopping/manual-url-candidates?supplier_code=DEXCO&limit=10&offset=0` returns catalog rows
- [ ] In browser: Manual URL panel lists products even with empty signals; saving a URL creates signal and table reflects overlay
