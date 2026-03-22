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
- [ ] T5b: frontend - manual URL panel UX fixes
      commit: "fix(web): stabilize manual URL panel interactions"
- [ ] T6: ADR close-out - capture evidence and accept ADR-0045
      commit: "docs(adr): ADR-0045 manual URL candidates - verified and closed"

## Acceptance tests
- [x] `go build ./...` passes
- [x] `npm.cmd run web:typecheck` passes
- [x] `npm.cmd run web:build` passes
- [x] With `shopping_supplier_product_signals` empty:
      `GET /api/v1/shopping/manual-url-candidates?supplier_code=DEXCO&limit=10&offset=0` returns catalog rows
- [ ] In browser: Manual URL panel lists products even with empty signals; saving a URL creates signal and table reflects overlay
- [ ] In browser: manual URL panel refreshes without layout jump, supplier=all behaves per spec, and toggle slider aligns

---

## Feature: Shopping run progress (polling)
Type: scraping | Events: yes (shopping.run_requested) | ADR: no

## Phase 1 - Architectural thinking

Module type:
- scraping (Python worker writes progress; Go reader exposes status)

Module location:
- `apps/server_core/internal/modules/shopping`
  - extend `ports/read_models.go`, `adapters/postgres/reader.go`, `transport/http/handler.go`
- `apps/server_core/migrations`
  - add progress columns on `shopping_price_run_requests`
- `apps/integration_worker/shopping_price_worker.py`
  - update progress per item/supplier
- `contracts/api/openapi/shopping_v1.openapi.yaml`
  - expose progress fields on run request payload

Risks to manage:
- migration + backfill defaults for existing rows
- avoid high write frequency (batch update progress)
- keep RLS + tenant_id safety

Level 1 scope:
- UI polling shows % progress and current supplier/product info

## Phase 2 - Plan (wait for approval, then execute T1..T5)

## Tasks
- [x] T1: contract - $metalshopping-openapi-contracts
      commit: "feat(shopping): add run progress fields"
- [x] T2: Go module - reader + handler + postgres adapter
      commit: "feat(shopping): expose run request progress"
- [x] T2b: Go module - run item status summary endpoint
      commit: "feat(shopping): expose run item status summary"
- [x] T3: worker - update progress during runs
      commit: "feat(worker): persist shopping run progress"
- [x] T3b: worker - add per-item logs for debugging
      commit: "feat(worker): add per-item shopping run logs"
- [ ] T3c: worker - keep alive polling mode for queue
      commit: "feat(worker): keep alive when queue is empty"
- [x] T4: SDK - $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate after shopping progress contract"
- [x] T5: frontend - $metalshopping-frontend
      commit: "feat(web): poll and display shopping run progress"
- [x] T5b: frontend - KPIs from item summary + cap history list
      commit: "feat(web): show run item KPIs and cap recent history"

## Acceptance tests
- [ ] `go build ./...` passes
- [x] `npm.cmd run web:typecheck` passes
- [ ] `GET /api/v1/shopping/run-requests/{id}` returns progress fields
- [ ] `GET /api/v1/shopping/runs/{run_id}/item-status-summary` returns grouped counts
- [ ] In browser: progress bar updates over time with worker running
- [ ] In browser: KPI cards reflect selected run item counts (OK/NOT_FOUND/AMBIGUOUS/ERROR)
- [ ] In browser: "Historico recente" shows max N with "Ver tudo"
- [ ] In dev: worker keeps running with empty queue when enabled

---

## Feature: Obra Fácil Playwright performance hardening
Type: scraping | Events: no | ADR: no

## Phase 1 - Architectural thinking

Module type:
- scraping (Python worker runtime strategy + supplier manifest tuning)

Module location:
- `apps/integration_worker/src/shopping_price_runtime/playwright/strategies.py`
- `apps/integration_worker/shopping_price_worker.py`
- `scripts/seed_tenant_default_driver_manifests.py`

Legacy references analyzed:
- `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\Nova pasta\MetalShopping\drivers\obrafacil_driver.py`
- `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\Nova pasta\MetalShopping\drivers\framework\playwright\runtime.py`
- `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\Nova pasta\MetalShopping\drivers\framework\playwright\pipelines.py`
- `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\Nova pasta\MetalShopping\drivers\common\playwright_pdp.py`

Measured evidence (2026-03-22):
- Active manifest for `OBRA_FACIL` has no `tabs` configured, so current runtime defaults to `tabs=1` (serial per supplier).
- Current strategy (sync Playwright + current selector) benchmark: ~`28.1s/item` (2 runs).
- Current strategy without selector wait (`pdpSelectors.price=''`, regex fallback): ~`7.0-9.0s/item` (2 runs).
- Reused browser/context benchmark (legacy-like tabs runtime): ~`2.2s` first item and ~`1.4-2.3s` subsequent item.
- Direct selector validation for `div.col-des div.price-box p span`: timeout at `30.0s`, confirming timeout-driven latency.

Primary bottlenecks:
- Browser/context startup per item/attempt/url in current `playwright.pdp_first.v1` path.
- Price selector mismatch causes full `timeoutSeconds` wait before regex fallback.
- `OBRA_FACIL` running with implicit `tabs=1`, unlike legacy `tabs_default=7`.
- No stage-level elapsed persisted for runtime observations, reducing performance debuggability.

Level 1 scope:
- Restore near-legacy throughput for `OBRA_FACIL` without changing business semantics (same statuses + same URL lifecycle).

## Phase 2 - Plan (wait for approval, then execute T1..T5)

## Tasks
- [x] T1: worker observability baseline (elapsed + stage notes for runtime)
      commit: "feat(worker): add playwright runtime latency telemetry"
- [x] T2: Playwright strategy fast path (avoid blocking selector timeout; parse-first fallback)
      commit: "fix(worker): remove selector-timeout bottleneck in playwright strategy"
- [x] T3: Playwright batch runtime reuse (browser/context reuse with tab workers)
      commit: "feat(worker): add tab-based playwright batch execution"
- [x] T4: supplier config parity (`OBRA_FACIL` tabs defaults + safe knobs)
      commit: "chore(shopping): tune obrafacil playwright runtime config"
- [ ] T5: validation + evidence (smoke run comparison vs HTTP and previous Playwright baseline)
      commit: "docs(perf): capture playwright run performance evidence"

## Acceptance tests
- [ ] `go build ./...` passes
- [ ] Worker smoke with `OBRA_FACIL` completes with real items (non-zero rows written)
- [ ] For equal product sample, median `OBRA_FACIL` item latency improves materially vs current baseline
- [ ] Progress UI keeps updating during run (no regression)
- [ ] Existing HTTP suppliers (DEXCO/CONDEC/ABC/LEROY/TELHA_NORTE) show no regression in smoke
