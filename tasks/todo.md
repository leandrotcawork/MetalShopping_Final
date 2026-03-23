# Feature: Products market report XLSX (pivot)
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 Ã¢â‚¬â€ Architectural thinking
Module type:
- `apps/server_core/internal/modules/shopping` (read-only + side-effect file write): read run items + pricing and write a pivoted XLSX report under an export root.
- `packages/feature-products`: product selection + run/supplier selection UI.
- `apps/web/src/app/App.tsx`: pass shopping api to products surface.
- `apps/web/src/pages/ShoppingPage.tsx`: remove export UI (moved to Products).

Exact folder structure (extensions only):
- `contracts/api/openapi/shopping_v1.openapi.yaml` (+ new request/response schemas)
- `contracts/api/jsonschema/shopping_market_report_export_xlsx_request_v1.schema.json`
- `contracts/api/jsonschema/shopping_market_report_export_xlsx_response_v1.schema.json`
- `apps/server_core/internal/modules/shopping/ports/read_models.go`
- `apps/server_core/internal/modules/shopping/adapters/postgres/reader.go`
- `apps/server_core/internal/modules/shopping/application/market_report_export_xlsx.go`
- `apps/server_core/internal/modules/shopping/application/service.go`
- `apps/server_core/internal/modules/shopping/transport/http/handler.go`
- `packages/platform-sdk/src/index.ts`
- `packages/feature-products/src/*` + `packages/feature-products/src/ProductsPortfolioPage.module.css`

Risks:
- Export can be large; cap rows and bound memory.
- Writing to arbitrary paths is unsafe; enforce export root + .xlsx path.
- Supplier label collisions in header need deterministic disambiguation.

Level scope:
- Level 1 (now): 1 row per product, dynamic supplier columns with prices, include our price + average cost + replacement cost, export from Products UI.
- Level 2 (later): download-from-browser and/or background export job for very large runs.

## Tasks
- [x] T1: contract Ã¢â‚¬â€ $metalshopping-openapi-contracts
      - `POST /api/v1/shopping/runs/{run_id}/export-market-report-xlsx`
      - schemas: request/response for market report export
      commit: "feat(shopping): add market report export contract"
- [x] T2: Go module Ã¢â‚¬â€ implement export endpoint
      - validate output path under export root
      - query product meta + pricing and run items
      - pivot rows (1 row per product) with supplier price columns
      commit: "feat(shopping): export market report xlsx"
- [x] T4: SDK Ã¢â‚¬â€ $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate shopping market report contract"
- [x] T5: frontend Ã¢â‚¬â€ $metalshopping-frontend
      - Products UI modal to select run + suppliers + export path
      - collect product ids for explicit/filtered selections
      - remove export panel from Shopping
      commit: "feat(web): add products market report export"

## Acceptance tests
- [x] go test ./apps/server_core/... passes
- [x] npm.cmd run web:typecheck passes
- [ ] Browser: `/products` export generates XLSX with pivot columns and costs
- [ ] Browser: `/shopping` no longer shows export panel

---

# Feature: Shopping export report XLSX
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/server_core/internal/modules/shopping` (read-only + side-effect file write): read run items and write an XLSX report to a configured export directory.
- `apps/web/src/pages/ShoppingPage.tsx`: add UI to pick run + suppliers + output path and trigger export.

Exact folder structure (extensions only):
- `contracts/api/openapi/shopping_v1.openapi.yaml` (+ new request/response schemas)
- `apps/server_core/internal/modules/shopping/ports/read_models.go` (new input/output models)
- `apps/server_core/internal/modules/shopping/adapters/postgres/reader.go` (list run items for export)
- `apps/server_core/internal/modules/shopping/application/service.go` (ExportRunReportXlsx)
- `apps/server_core/internal/modules/shopping/transport/http/handler.go` (POST route under `/runs/{run_id}/...`)
- `apps/server_core/internal/platform/runtime_config/*` (export root config, if needed)
- `apps/web/src/pages/ShoppingPage.tsx` (+ minor CSS if needed)

Risks:
- Export can be large; must cap rows and keep memory bounded.
- Writing to arbitrary paths is unsafe; must restrict to an export root and sanitize relative paths.

Level scope:
- Level 1 (now): export XLSX to server filesystem under configured root; user selects run + suppliers; show success with output path.
- Level 2 (later): download-from-browser and/or background export job for very large runs.

## Tasks
- [ ] T1: contract â€” $metalshopping-openapi-contracts
      - `POST /api/v1/shopping/runs/{run_id}/export-xlsx`
      - schemas: request/response for export
      commit: "feat(shopping): add run export xlsx contract"
- [ ] T2: Go module â€” implement export endpoint
      - validate output path under export root
      - query run items for selected suppliers
      - write XLSX and return metadata
      commit: "feat(shopping): export run report to xlsx"
- [ ] T4: SDK â€” $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate shopping export contract"
- [ ] T5: frontend â€” $metalshopping-frontend
      - UI: select run + suppliers + output file name/path; trigger export
      commit: "feat(web): add shopping run export"

## Acceptance tests
- [x] go test ./apps/server_core/... passes
- [x] npm.cmd run web:typecheck passes
- [ ] Browser: `/shopping` export generates XLSX file on disk and shows the saved path

---
# Feature: Shopping catalog select all
Type: frontend  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/web/src/pages/ShoppingPage.tsx`: adicionar seleÃ§Ã£o total no upload por catÃ¡logo.

Risks:
- SeleÃ§Ã£o total pode exigir mÃºltiplas pÃ¡ginas; precisa paginar com limite mÃ¡ximo do endpoint.

Level scope:
- Level 1 (now): botÃ£o "Selecionar todos" carrega todos os IDs do filtro atual e seleciona.

## Tasks
- [x] T5: frontend â€” $metalshopping-frontend
      - adicionar botÃ£o "Selecionar todos" com loading
      - buscar IDs em batches usando `productsApi.listProductsPortfolio`
      commit: "fix(web): add catalog select all"

## Acceptance tests
- [x] npm.cmd run web:typecheck passes
- [ ] Browser: `/shopping` â†’ Produtos Cadastrados â†’ botÃ£o "Selecionar todos" seleciona todos os itens do filtro

---
# Feature: Shopping manual filters show brand/group
Type: frontend  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/web/src/pages/ShoppingPage.tsx`: carregar filtros de marca/grupo mesmo quando input mode nÃ£o Ã© catÃ¡logo.

Risks:
- Filtros de URL manual dependem indevidamente do carregamento do catÃ¡logo e ficam vazios.

Level scope:
- Level 1 (now): buscar filtros do portfolio ao abrir o painel manual quando ainda nÃ£o carregados.

## Tasks
- [x] T5: frontend â€” $metalshopping-frontend
      - carregar filtros via `productsApi.listProductsPortfolio` quando o painel manual Ã© aberto
      commit: "fix(web): load manual brand/group filters"

## Acceptance tests
- [x] npm.cmd run web:typecheck passes
- [ ] Browser: `/shopping` â†’ Configurar URLs manuais â†’ filtros Marca/Grupo aparecem

---
# Feature: Shopping manual URL save CORS preflight
Type: backend  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/server_core/internal/platform/observability/cors.go`: ajustar preflight para mÃ©todos de escrita usados pelo Shopping (`PUT`).

Risks:
- Se CORS nÃ£o expÃµe `PUT`, o browser bloqueia preflight e o SDK retorna erro genÃ©rico de interceptor.

Level scope:
- Level 1 (now): incluir `PUT` no `Access-Control-Allow-Methods` e validar com teste unitÃ¡rio de CORS.

## Tasks
- [x] T2: Go module â€” liberar `PUT` no CORS
      - adicionar `http.MethodPut` na lista de mÃ©todos permitidos
      - reforÃ§ar teste para garantir `PUT` no header de preflight
      commit: "fix(server): allow cors preflight for shopping put"

## Acceptance tests
- [x] go test ./apps/server_core/internal/platform/observability/... passes
- [ ] Browser: `/shopping` â†’ salvar URL manual nÃ£o retorna mais erro de interceptor

---
# Feature: Shopping manual URL save button
Type: frontend  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/web/src/pages/ShoppingPage.tsx`: corrigir o fluxo de salvar URLs manuais sem alterar contrato ou backend.

Risks:
- O footer pode continuar mostrando um CTA inerte se nÃ£o houver rastreamento correto de drafts alterados.
- Salvar em lote nÃ£o pode sobrescrever `lookupMode` com valor fixo quando o candidato jÃ¡ traz o modo correto.

Level scope:
- Level 1 (now): habilitar o botÃ£o Salvar, salvar apenas linhas alteradas da pÃ¡gina visÃ­vel e recarregar a lista manual.
- Level 2 (later): adicionar feedback de sucesso mais rico ou persistir drafts entre pÃ¡ginas, se necessÃ¡rio.

## Tasks
- [x] T5: frontend â€” $metalshopping-frontend
      - detectar drafts realmente alterados
      - habilitar o CTA de salvar em lote
      - adicionar salvar por linha e Enter no campo URL
      - reutilizar `lookupMode` do candidato ao persistir a URL manual
      commit: "fix(web): enable shopping manual url save"

## Acceptance tests
- [x] npm.cmd run web:typecheck passes
- [x] npm.cmd run web:test passes
- [ ] Browser: `/shopping` â†’ Configurar URLs manuais â†’ editar URL vÃ¡lida â†’ botÃ£o Salvar habilita e persiste

---
# Feature: Workflow skill compaction + manual /plan gate
Type: process  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `docs/process`: compact skill instructions and align `$ms` with the real harness behavior.

Risks:
- Skill text can drift from tool reality and create the wrong expectation about automatic plan mode.
- Overly verbose skills increase context load and make orchestration slower/noisier than needed.

Level scope:
- Level 1 (now): `$ms` asks the user to run `/plan` manually for complex work, uses `update_plan`, and keeps the workflow concise.
- Level 2 (later): archive completed feature blocks out of `tasks/todo.md` if the file grows again.

## Tasks
- [x] T1: process â€” compact `$ms` and add manual `/plan` gate
      - remove the false "enters plan mode automatically" claim
      - add complexity gate + operational gate
      commit: "docs(skills): compact ms workflow and add manual plan gate"
- [x] T2: process â€” compact `metalshopping-implement`
      - keep invariants in the skill and move concrete code to references
      commit: "docs(skills): compact implement skill"
- [x] T3: process â€” record lesson and keep workflow source-of-truth updated
      commit: "docs(process): align skill workflow with harness"

## Acceptance tests
- [x] Review: `.agents/skills/ms/SKILL.md` no longer claims automatic plan mode
- [x] Review: complex tasks now instruct the user to run `/plan` manually
- [x] Review: `.agents/skills/metalshopping-implement/SKILL.md` stays compact and points to repo references

---
# Feature: Shopping run observability + UX (history, filters, logs)
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/server_core/internal/modules/shopping` (read-only): add new read endpoints for run item details + per-supplier status breakdown.
- `apps/web/src/pages/ShoppingPage.*`: UX tweaks (history scroll/height), fix status filter, show richer run details and log.

Exact folder structure (existing module; extensions only):
- `apps/server_core/internal/modules/shopping/ports/read_models.go` (new read models + reader methods)
- `apps/server_core/internal/modules/shopping/adapters/postgres/reader.go` (new SQL reads)
- `apps/server_core/internal/modules/shopping/application/service.go` (wire methods)
- `apps/server_core/internal/modules/shopping/transport/http/handler.go` (new GET routes under `/runs/{run_id}/...`)
- `contracts/api/openapi/shopping_v1.openapi.yaml` + `contracts/api/jsonschema/*` (new schemas + operations)
- `packages/platform-sdk` regenerated (T4)
- `apps/web/src/pages/ShoppingPage.tsx` + `apps/web/src/pages/ShoppingPage.module.css`

Risks:
- Run item list can be large â†’ must be paginated and capped (`limit` <= 200) to avoid slow UI and heavy DB reads.
- Avoid N+1: run items endpoint must join `catalog_products` once to return `productLabel` (so UI does not fan out).
- Ensure tenant isolation: every query uses `pgdb.BeginTenantTx` + `WHERE tenant_id=current_tenant_id()`.

Level scope:
- Level 1 (now): real data in browser; scrollable history; status filter works; per-supplier breakdown (OK/NF/AMB/ERROR); log shows per-item lines using run items endpoint.
- Level 2 (defer): capture â€œURL efetivamente tentadaâ€ no worker (hoje sÃ³ temos `product_url` do signal + `lookup_term`).

## Tasks
- [ ] T1: contract â€” $metalshopping-openapi-contracts
      Add:
      - `GET /api/v1/shopping/runs/{run_id}/supplier-item-status-summary`
      - `GET /api/v1/shopping/runs/{run_id}/items` (paged)
      Schemas:
      - `shopping_run_supplier_item_status_summary_v1.schema.json`
      - `shopping_run_item_v1.schema.json`
      - `shopping_run_item_list_v1.schema.json`
      commit: "feat(shopping): add run item detail contracts"
- [x] T2: Go module â€” implement reader + handler
      - Reader: `ListRunItems`, `GetRunSupplierItemStatusSummary`
      - Handler: route suffixes under `handleRunByID`
      commit: "feat(shopping): add run item detail endpoints"
- [x] T4: SDK â€” $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate shopping contracts"
- [x] T5: frontend â€” $metalshopping-frontend
      - Historico recente: menor altura + scroll interno (sem layout jump)
      - Filtro Todos/Queued/Running/etc: corrigir para recarregar lista corretamente
      - Detalhe do run: tabela por fornecedor (OK/NF/AMB/ERROR)
      - Log detalhado: listar itens (produto, fornecedor, status, lookup_term, product_url, notes/http_status/elapsed)
      - Historico recente: limitar visual + scroll (mantem "Ver tudo" opcional)
      commit: "feat(web): improve shopping run UX and observability"

## Acceptance tests
- [x] go build ./... passes
- [x] go test ./apps/server_core/... passes
- [x] npm.cmd run web:typecheck passes
- [x] Browser: `/shopping` shows real history + filter works + per-supplier breakdown + per-item log for a selected run

---

# Feature: Shopping run UI bugfixes (history layout + log URL)
Type: frontend  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `apps/web/src/pages/ShoppingPage.*`: ajustes de layout e log (sem mudanÃ§a de contrato/API).

Risks:
- Scroll dentro de grid/flex pode quebrar sem `min-height: 0` nos containers.
- `productUrl` Ã© URL durÃ¡vel (manual/sinal). A URL tentada pode existir apenas em `notes` (debug).

Level scope:
- Level 1 (now): log mostra URL quando existir em `notes` (ex: `final_url=`); histÃ³rico recente sempre scrollÃ¡vel e sem â€œVer tudoâ€.
- Level 2 (defer): persistir `request_url/final_url` em campo dedicado no backend/worker.

## Tasks
- [x] T5: frontend â€” $metalshopping-frontend
      - HistÃ³rico recente: remover â€œVer tudoâ€ e manter scroll interno com altura que acompanha o detalhe do run
      - Log detalhado: exibir URL derivada de `notes` quando `productUrl` estiver vazio
      commit: "fix(web): improve shopping run history and log url"

## Acceptance tests
- [x] npm.cmd run web:typecheck passes
- [ ] Browser: `/shopping` â†’ selecionar run com `notes` contendo `final_url=` â†’ URL aparece no log
- [ ] Browser: `/shopping` â†’ run com muitos fornecedores â†’ histÃ³rico recente nÃ£o quebra e permanece scrollÃ¡vel

---

# Feature: Shopping run log search URL (computed, non-persistent)
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 ï¿½ Architectural thinking
Module type:
- `apps/server_core/internal/modules/shopping` (read-only): calcular URL de busca em leitura no endpoint `/runs/{run_id}/items`, sem persistir em DB.

Risks:
- Nem todo manifest possui template de busca (`searchUrl`/`endpointTemplate`/`startUrl`) com placeholder de termo.
- Regra de render deve ser conservadora para nao exibir URL invalida.

Level scope:
- Level 1 (now): preencher `productUrl` no item de run com URL de busca renderizada quando `product_url` vier vazio.
- Level 2 (defer): criar campo dedicado `searchUrl` no contrato para separar semantica.

## Tasks
- [x] T2: Go module ï¿½ compute search URL at read-time
      - Reader `ListRunItems`: join manifest ativo e renderizar URL a partir de `lookup_term`
      - Sem gravacao em tabela; somente resposta da API
      commit: "fix(shopping): compute run item search url on read"

## Acceptance tests
- [x] go test ./apps/server_core/... passes
- [ ] Browser: `/shopping` log detalhado mostra URL de busca para fornecedores com template configurado

---

# Feature: Shopping run log search URL (driver notes)
Type: scraping  |  Events: no  |  ADR: no

## Phase 1 ï¿½ Architectural thinking
Module type:
- `apps/integration_worker/src/shopping_price_runtime/http/strategies.py`: registrar `search_url=` nos `notes` do RuntimeObservation.

Risks:
- URLs longas podem truncar `notes` (limite 280 chars). Usar prefixo curto.
- Para fornecedores sem URL de busca, manter comportamento atual.

Level scope:
- Level 1 (now): VTEX/HTML/Leroy reportam `search_url=` nos notes para aparecer no log.

## Tasks
- [ ] T3: worker ï¿½ append search_url to observation notes
      commit: "fix(worker): add search_url to http runtime notes"

## Acceptance tests
- [ ] Run real shopping with Dexco/Telha Norte shows `search_url=` in log

---

# Feature: Analytics New â€” Home tranche
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 â€” Architectural thinking
Module type:
- `contracts/api/*`: freeze initial Analytics Home read contract.
- `apps/server_core/internal/modules/analytics_serving` (read-only): tenant-safe analytics home payload.
- `packages/platform-sdk`: expose `analytics.getHome()`.
- `packages/feature-analytics` + `apps/web`: Analytics Home route and UI binding.

Risks:
- New feature package import resolution fails without `apps/web/tsconfig.json` path/include wiring.
- Contract generation drift if `analytics_v1` is not included in `scripts/generate_contract_artifacts.ps1`.
- Partial blocks in first slice need explicit `NOT_READY` state to avoid fake completeness.

Level scope:
- Level 1 (now): `/analytics` no longer placeholder; Home shows real operational/product metrics and explicit pending blocks.
- Level 2 (defer): advanced analytics blocks (actions, alerts, portfolio, timeline) move from `NOT_READY` to real data.

## Tasks
- [x] T1: contract â€” $metalshopping-openapi-contracts
      - add `contracts/api/openapi/analytics_v1.openapi.yaml`
      - add `analytics_home_v1`, `analytics_home_block_v1`, `analytics_block_error_v1` schemas
      commit: "feat(analytics): add analytics home read contracts"
- [ ] T2: Go module â€” $metalshopping-implement
      - implement `analytics_serving` reader/service/handler (`GET /api/v1/analytics/home`)
      - register module in `composition_modules.go`
      commit: "feat(analytics): add analytics home read endpoint"
- [ ] T4: SDK â€” $metalshopping-sdk-generation
      - include analytics in generation script and regenerate artifacts
      - extend `sdk-runtime` with `analytics.getHome()`
      commit: "chore(sdk): regenerate with analytics contract"
- [ ] T5: frontend â€” $metalshopping-frontend
      - create `packages/feature-analytics` Home page
      - wire route `/analytics` in `apps/web/src/app/App.tsx`
      - add workspace + tsconfig wiring for new feature package
      commit: "feat(web): implement analytics home route"

## Acceptance tests
- [x] go test ./apps/server_core/... passes
- [x] npm.cmd run web:typecheck passes
- [x] npm.cmd run web:test passes
- [ ] Browser: `/analytics` renders Analytics Home with real API payload and pending block list

---

# Feature: Analytics New â€” Legacy visual parity (tabs + cards)
Type: frontend  |  Events: no  |  ADR: no

## Tasks
- [ ] T5: frontend â€” $metalshopping-frontend
      - copy legacy `.tsx/.css` references into `packages/feature-analytics/legacy_snapshot/*`
      - deliver runnable analytics shell with legacy-style navigation tabs + card layout
      - keep data partially mocked when integration is not ready
      commit: "feat(web): align analytics visual shell with legacy"

## Acceptance tests
- [x] npm.cmd run web:build passes
- [ ] Browser: `/analytics` visually matches legacy tab navigation + home cards baseline

---

# Feature: Analytics New ï¿½ Legacy-first full phase plan (source of truth)
Type: frontend-first migration  |  Events: no  |  ADR: no

## Phase 1 ï¿½ Architectural thinking
Module type:
- `frontend-only` now (visual parity first, no real backend binding).
- `read-only` later (contract/backend/sdk wiring after visual sign-off).

Exact folder structure (target):
- `packages/feature-analytics/src/*` (runnable visual shell used by web app)
- `packages/feature-analytics/legacy_snapshot/analytics/*` (literal copy reference)
- `packages/feature-analytics/legacy_snapshot/analytics_home/*` (literal copy reference)
- `apps/web/src/pages/AnalyticsPage.tsx` + `apps/web/src/app/App.tsx` (route wiring)
- `apps/web/vite.config.ts` + `apps/web/tsconfig.json` (workspace resolution)
- deferred integration layer: `contracts/api/*`, `apps/server_core/internal/modules/analytics_serving/*`, `packages/platform-sdk/src/*`

Risks:
- Legacy home is split across many components/viewmodels and depends on app-level session/hooks.
- Copying only CSS or only one TSX file will never match legacy layout.
- Vite/TS alias drift can break build even when UI files are present.
- Premature backend binding can distort visual parity and slow migration.

Level scope:
- Level 1 (now): visual parity only (tabs, hero, command bar, bento/cards, spotlight/drawer shell, actions list, heatmap/sample blocks) with mocks.
- Level 2 (after visual sign-off): keep same visual, replace mocks with contract-backed data progressively.
- Level 3 (final): remove temporary adapters/shims and close contract/backend/sdk tranche.

## Tasks
- [x] T5-A: frontend ï¿½ inventory + freeze visual baseline from legacy
      - map all files under legacy `analytics` + `analytics_home` and lock them as migration baseline
      - define "must-match" sections (top nav tabs, hero gradient, command panel, first fold cards, spotlight shell)
      commit: "docs(analytics): freeze legacy visual parity baseline"

- [x] T5-B: frontend ï¿½ literal copy tranche (home shell)
      - copy legacy TSX/CSS structure into runnable package surface (without real backend)
      - preserve class names/layout primitives; avoid redesign
      commit: "feat(web): copy legacy analytics home shell"

- [x] T5-C: frontend ï¿½ compatibility adapters for compile/runtime
      - add temporary mock/session adapters required by copied components
      - keep route `/analytics` stable and loadable
      commit: "refactor(web): add analytics legacy compatibility adapters"

- [ ] T5-D: frontend ï¿½ visual parity pass (pixel/structure)
      - align spacing, typography, chip states, cards and first fold composition with legacy
      - ensure tabs (`Home`, `Produtos`, `Taxonomia`, `Marcas`, `Aï¿½ï¿½es`) mirror legacy behavior
      commit: "fix(web): align analytics legacy visual parity"

- [ ] T5-E: frontend ï¿½ remaining legacy sections (still mocked)
      - wire spotlight/drawer shell and sample blocks used by legacy home flow
      - keep backend integration disabled; mock-only for missing data
      commit: "feat(web): complete analytics legacy visual sections with mocks"

- [ ] T1 (defer): contract ï¿½ $metalshopping-openapi-contracts
      - only start after T5 visual sign-off
      commit: "feat(analytics): finalize read contracts after visual parity"

- [ ] T2 (defer): Go module ï¿½ $metalshopping-implement
      - only start after T1
      commit: "feat(analytics): implement analytics serving reads"

- [ ] T4 (defer): SDK ï¿½ $metalshopping-sdk-generation
      - only start after T1/T2
      commit: "chore(sdk): regenerate analytics sdk"

## Acceptance tests
- [ ] Browser: `/analytics` first fold is visually equivalent to legacy (tabs + hero + command panel + cards)
- [ ] Browser: tabs switch sections with same IA labels and ordering as legacy
- [ ] Browser: no blank/unstyled blocks on initial load (mock data allowed)
- [x] npm.cmd run web:build passes
- [x] npm.cmd run web:typecheck passes
- [ ] Gate: only after visual sign-off, unlock T1/T2/T4 integration tasks



---

# Feature: Analytics Produtos - Legacy migration (full flow)
Type: frontend-first migration  |  Events: no  |  ADR: no

## Tasks
- [ ] T5: frontend - copy legacy Produtos + workspace with mocks/shims and new routes
      - Produtos index/density view
      - Workspace tabs (overview/insights/history/simulator)
      - Local mocks + DTO shims + assets
      commit: "feat(web): migrate analytics products legacy flow"

## Acceptance tests
- [x] Browser: navegar entre tabs e voltar para /analytics/products nao trava (scroll ok + cliques ok)
- [ ] Browser: /analytics/products matches legacy Produtos index/density
- [ ] Browser: /analytics/products/:pn/overview matches legacy workspace header + hero
- [ ] Browser: /analytics/products/:pn/insights loads with empty insights state (no crash)
- [ ] Browser: /analytics/products/:pn/history renders series placeholders
- [ ] Browser: /analytics/products/:pn/simulator renders simulator controls
- [x] npm.cmd run web:typecheck passes
- [x] npm.cmd run web:build passes

---

# Feature: Analytics Classificacoes - Legacy migration (Taxonomy)
Type: frontend-first migration  |  Events: no  |  ADR: no

## Tasks
- [x] T5-A: frontend ï¿½ inventory + baseline (legacy snapshot)
      - source of truth: `packages/feature-analytics/legacy_snapshot/analytics/TaxonomyHomePage.tsx` + `taxonomy_home.module.css` + `legacy_snapshot/analytics/components/*`

- [x] T5-B: frontend ï¿½ literal copy + wiring (runnable)
      - copy page + CSS + components into `packages/feature-analytics/src/pages/analytics/*`
      - add missing helper `taxonomy_visuals`

- [x] T5-C: frontend ï¿½ dependencies (charts)
      - add `chart.js`, `react-chartjs-2`, `chartjs-chart-treemap`, `recharts`, `@tanstack/react-table`

- [x] T5-D: frontend ï¿½ shell integration
      - render `<TaxonomyHomePage />` on tab/route `/analytics/taxonomy`
      - remove MVP taxonomy state/effects to avoid redundant loads

- [x] T5-E: frontend ï¿½ mock payload parity
      - expand `AnalyticsTaxonomyScopeOverviewV1Dto` keys used by legacy
      - enrich `buildMockTaxonomyScopeOverview()` with non-empty panels (treemap + charts + tables)
      commit: "feat(web): migrate analytics taxonomy legacy page"

## Acceptance tests
- [ ] Browser: `/analytics/taxonomy` renders charts + tables filled (no console errors)
- [ ] Browser: tab switch `Home -> Classificacoes -> Produtos -> Classificacoes` works (no freeze)
- [ ] Browser: spotlights open/close restores body scroll (no stuck overlay)
- [x] npm.cmd run web:typecheck passes
- [x] npm.cmd run web:build passes

---

# Feature: Analytics Marca - Legacy migration (Brand)
Type: frontend-first migration  |  Events: no  |  ADR: no

## Tasks
- [ ] T5-A: frontend - inventory + baseline (legacy snapshot)
      - source of truth: `packages/feature-analytics/legacy_snapshot/analytics/BrandHomePage.tsx`
      - source of truth: `packages/feature-analytics/legacy_snapshot/analytics/brand_home.module.css`

- [ ] T5-B: frontend - literal copy + runnable wiring
      - copy page + CSS into `packages/feature-analytics/src/pages/analytics/*`
      - preserve markup/classes hierarchy from legacy

- [ ] T5-C: frontend - shell integration
      - render `<BrandHomePage />` on tab/route `/analytics/brands`
      - remove MVP Marca block from `AnalyticsPage` to avoid duplicate renders

- [ ] T5-D: frontend - visual parity pass (CSS)
      - ensure cards/surfaces are not transparent
      - validate spacing and breakpoints in the existing shell

- [ ] T5-E: process + validation
      - update tasks + run build + validate navigation and console behavior manually
      commit: "feat(web): migrate analytics brand legacy page"

## Acceptance tests
- [ ] Browser: `/analytics/brands` renders full legacy Marca surface (header, KPIs, panels, map, table)
- [ ] Browser: tab switch `Home -> Marca -> Produtos -> Marca -> Classificacoes` works (no freeze)
- [ ] Browser: no stuck overlay/backdrop when switching tabs
- [ ] Browser: no console errors on Marca render
- [ ] npm.cmd run web:build passes

---

# Feature: Shopping CONDEC driver hardening
Type: scraping  |  Events: no  |  ADR: no

## Phase 1 — Architectural thinking
Module type:
- `scraping`: Python worker runtime strategy/parser hardening for `CONDEC`.

Exact folder structure (target):
- `apps/integration_worker/src/shopping_price_runtime/http/strategies.py`
- `apps/integration_worker/src/shopping_price_runtime/shared/*` (if parser helpers are extracted)
- optional tests under `apps/integration_worker/tests/*` if adjacent coverage exists

Risks:
- Overfitting `CONDEC` as a one-off branch would weaken the runtime framework.
- Regex-only parsing is unsafe for HTML search pages with unrelated numeric tokens.
- Lookup policy drift can reintroduce false positives if `CONDEC` is not kept reference/search-card based.

Level scope:
- Level 1 (now): port the legacy card-based parsing behavior into the runtime in a reusable framework shape.
- Level 2 (later): generalize the same structural card parser for other HTML-search suppliers if needed.

## Tasks
- [ ] T2: worker — $metalshopping-implement
      - extract reusable HTML search-card parsing helpers
      - add `CONDEC`-specific card mapping/selection using the runtime framework
      - preserve structured runtime notes/status semantics
      commit: "fix(worker): harden condec html search parsing"

## Acceptance tests
- [ ] Real run: `CONDEC` no longer collapses all `observed_price` values to `2.00`
- [ ] Real run: `CONDEC` returns `OK/AMBIGUOUS/NOT_FOUND` based on card structure, not first numeric token

---

# Feature: Legacy DB migration for pricing and inventory
Type: write+migration  |  Events: no  |  ADR: no

## Phase 1 — Architectural thinking
Module type:
- `write+migration`: import legacy `metalshopping_db` price/cost/stock data into current `metalshopping`.

Exact folder structure (target):
- migration/import script under repo (`scripts/*` or module-local tooling)
- optional SQL mapping notes under `docs/` or `tasks/`
- target writes into current pricing/inventory tables only

Risks:
- Wrong business-key mapping can attach legacy values to the wrong current product.
- Inventory/pricing writes must stay tenant-scoped and idempotent.
- Legacy schema may not map 1:1 to current semantic fields (`price`, `replacement cost`, `average cost`, stock positions).

Level scope:
- Level 1 (now): inspect both schemas and define exact source→target field mapping.
- Level 2 (now if mapping is safe): implement import tooling for price/cost/stock with idempotent upsert semantics.

## Tasks
- [ ] T2-A: analysis — inspect `metalshopping_db` and current `metalshopping`
      - identify source tables/columns for price, cost and stock
      - identify reliable join key to current catalog

- [ ] T2-B: implementation — $metalshopping-implement
      - create migration/import tooling with tenant-safe upserts
      - validate imported counts and spot-check products
      commit: "feat(data): import legacy pricing and inventory"

## Acceptance tests
- [ ] SQL validation: current `metalshopping` has non-zero rows in pricing/inventory target tables after import
- [ ] Browser: Products/Shopping surfaces show imported own price/cost/stock for migrated products

---

# Feature: Shopping run observability polish
Type: frontend + worker  |  Events: no  |  ADR: no

## Phase 1 — Architectural thinking
Module type:
- `frontend + worker`: improve run diagnostics without changing contracts.

Exact folder structure (target):
- `apps/web/src/pages/ShoppingPage.tsx`
- `apps/integration_worker/shopping_price_worker.py`

Risks:
- UI polish must reflect real persisted timestamps, not compute fake runtime values.
- Detailed log must display the found supplier price without degrading scanability.

Level scope:
- Level 1 (now): show found price in detailed log and persist real run start/end so total duration is meaningful.

## Tasks
- [ ] T2/T5: worker + frontend — $metalshopping-implement + $metalshopping-frontend
      - persist real run `started_at` instead of stamping completion time twice
      - show total run duration in details
      - show found price for each processed log item
      commit: "fix(shopping): polish run observability"

## Acceptance tests
- [ ] Browser: detailed log shows found price per processed item
- [ ] Browser: run details show different start/end timestamps and total duration for new runs

