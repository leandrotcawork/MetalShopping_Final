# tasks/todo.md
# Current feature state. Updated by `skills/ms/SKILL.md` during implementation.

## Feature: ADR-0045 Manual URL candidates
Type: read-only  |  Events: no  |  ADR: ADR-0045

## Phase 1 (architectural thinking)
- Module: `apps/server_core/internal/modules/shopping` (read-only list endpoint).
- API: `GET /api/v1/shopping/manual-url-candidates` (single supplier scope, paginated).
- Data model: `catalog_products` are the base rows, overlay `shopping_supplier_product_signals` by `(product_id, supplier_code)`.
- Writes: unchanged. Manual URL save remains via existing upsert path for supplier signals.

Risks to manage in implementation:
- Query scale: must remain single-supplier + paginated to avoid `products x suppliers` explosion.
- Query performance: ensure indexes exist for `(tenant_id, supplier_code, product_id)` join and filter columns where needed.
- UX correctness: when signals are empty, list still returns catalog rows; when a URL is saved, row reflects overlay fields.

## Tasks
- [ ] T1: contract — $metalshopping-openapi-contracts
      commit: "feat(shopping): add manual URL candidates endpoint contract"
- [ ] T2: Go module — reader + handler + postgres adapter
      commit: "feat(shopping): implement manual URL candidates list endpoint"
- [ ] T4: SDK — $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate after shopping contract"
- [ ] T5: frontend — bind manual URL panel to candidates endpoint
      commit: "feat(web): manual URL panel uses candidates endpoint"
- [ ] T6: ADR — $metalshopping-adr (close ADR-0045 with evidence)
      commit: "docs(adr): ADR-0045 manual URL candidates — verified and closed"

## Acceptance tests
- [ ] `go build ./...` passes
- [ ] `npm.cmd run web:typecheck` passes
- [ ] `npm.cmd run web:build` passes
- [ ] With `shopping_supplier_product_signals` empty:
      `GET /api/v1/shopping/manual-url-candidates?supplier_code=DEXCO&limit=10&offset=0` returns catalog rows
- [ ] In browser: Manual URL panel lists products even with empty signals; saving a URL creates signal and table reflects overlay
