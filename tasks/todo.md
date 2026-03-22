# Feature: Workflow skill compaction + manual /plan gate
Type: process  |  Events: no  |  ADR: no

## Phase 1 — Architectural thinking
Module type:
- `docs/process`: compact skill instructions and align `$ms` with the real harness behavior.

Risks:
- Skill text can drift from tool reality and create the wrong expectation about automatic plan mode.
- Overly verbose skills increase context load and make orchestration slower/noisier than needed.

Level scope:
- Level 1 (now): `$ms` asks the user to run `/plan` manually for complex work, uses `update_plan`, and keeps the workflow concise.
- Level 2 (later): archive completed feature blocks out of `tasks/todo.md` if the file grows again.

## Tasks
- [x] T1: process — compact `$ms` and add manual `/plan` gate
      - remove the false "enters plan mode automatically" claim
      - add complexity gate + operational gate
      commit: "docs(skills): compact ms workflow and add manual plan gate"
- [x] T2: process — compact `metalshopping-implement`
      - keep invariants in the skill and move concrete code to references
      commit: "docs(skills): compact implement skill"
- [x] T3: process — record lesson and keep workflow source-of-truth updated
      commit: "docs(process): align skill workflow with harness"

## Acceptance tests
- [x] Review: `.agents/skills/ms/SKILL.md` no longer claims automatic plan mode
- [x] Review: complex tasks now instruct the user to run `/plan` manually
- [x] Review: `.agents/skills/metalshopping-implement/SKILL.md` stays compact and points to repo references

---
# Feature: Shopping run observability + UX (history, filters, logs)
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 — Architectural thinking
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
- Run item list can be large → must be paginated and capped (`limit` <= 200) to avoid slow UI and heavy DB reads.
- Avoid N+1: run items endpoint must join `catalog_products` once to return `productLabel` (so UI does not fan out).
- Ensure tenant isolation: every query uses `pgdb.BeginTenantTx` + `WHERE tenant_id=current_tenant_id()`.

Level scope:
- Level 1 (now): real data in browser; scrollable history; status filter works; per-supplier breakdown (OK/NF/AMB/ERROR); log shows per-item lines using run items endpoint.
- Level 2 (defer): capture “URL efetivamente tentada” no worker (hoje só temos `product_url` do signal + `lookup_term`).

## Tasks
- [x] T1: contract — $metalshopping-openapi-contracts
      Add:
      - `GET /api/v1/shopping/runs/{run_id}/supplier-item-status-summary`
      - `GET /api/v1/shopping/runs/{run_id}/items` (paged)
      Schemas:
      - `shopping_run_supplier_item_status_summary_v1.schema.json`
      - `shopping_run_item_v1.schema.json`
      - `shopping_run_item_list_v1.schema.json`
      commit: "feat(shopping): add run item detail contracts"
- [x] T2: Go module — implement reader + handler
      - Reader: `ListRunItems`, `GetRunSupplierItemStatusSummary`
      - Handler: route suffixes under `handleRunByID`
      commit: "feat(shopping): add run item detail endpoints"
- [x] T4: SDK — $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate shopping contracts"
- [x] T5: frontend — $metalshopping-frontend
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

## Phase 1 — Architectural thinking
Module type:
- `apps/web/src/pages/ShoppingPage.*`: ajustes de layout e log (sem mudança de contrato/API).

Risks:
- Scroll dentro de grid/flex pode quebrar sem `min-height: 0` nos containers.
- `productUrl` é URL durável (manual/sinal). A URL tentada pode existir apenas em `notes` (debug).

Level scope:
- Level 1 (now): log mostra URL quando existir em `notes` (ex: `final_url=`); histórico recente sempre scrollável e sem “Ver tudo”.
- Level 2 (defer): persistir `request_url/final_url` em campo dedicado no backend/worker.

## Tasks
- [x] T5: frontend — $metalshopping-frontend
      - Histórico recente: remover “Ver tudo” e manter scroll interno com altura que acompanha o detalhe do run
      - Log detalhado: exibir URL derivada de `notes` quando `productUrl` estiver vazio
      commit: "fix(web): improve shopping run history and log url"

## Acceptance tests
- [x] npm.cmd run web:typecheck passes
- [ ] Browser: `/shopping` → selecionar run com `notes` contendo `final_url=` → URL aparece no log
- [ ] Browser: `/shopping` → run com muitos fornecedores → histórico recente não quebra e permanece scrollável

---

# Feature: Shopping run log search URL (computed, non-persistent)
Type: read-only  |  Events: no  |  ADR: no

## Phase 1 � Architectural thinking
Module type:
- `apps/server_core/internal/modules/shopping` (read-only): calcular URL de busca em leitura no endpoint `/runs/{run_id}/items`, sem persistir em DB.

Risks:
- Nem todo manifest possui template de busca (`searchUrl`/`endpointTemplate`/`startUrl`) com placeholder de termo.
- Regra de render deve ser conservadora para nao exibir URL invalida.

Level scope:
- Level 1 (now): preencher `productUrl` no item de run com URL de busca renderizada quando `product_url` vier vazio.
- Level 2 (defer): criar campo dedicado `searchUrl` no contrato para separar semantica.

## Tasks
- [x] T2: Go module � compute search URL at read-time
      - Reader `ListRunItems`: join manifest ativo e renderizar URL a partir de `lookup_term`
      - Sem gravacao em tabela; somente resposta da API
      commit: "fix(shopping): compute run item search url on read"

## Acceptance tests
- [x] go test ./apps/server_core/... passes
- [ ] Browser: `/shopping` log detalhado mostra URL de busca para fornecedores com template configurado

---

# Feature: Shopping run log search URL (driver notes)
Type: scraping  |  Events: no  |  ADR: no

## Phase 1 � Architectural thinking
Module type:
- `apps/integration_worker/src/shopping_price_runtime/http/strategies.py`: registrar `search_url=` nos `notes` do RuntimeObservation.

Risks:
- URLs longas podem truncar `notes` (limite 280 chars). Usar prefixo curto.
- Para fornecedores sem URL de busca, manter comportamento atual.

Level scope:
- Level 1 (now): VTEX/HTML/Leroy reportam `search_url=` nos notes para aparecer no log.

## Tasks
- [ ] T3: worker � append search_url to observation notes
      commit: "fix(worker): add search_url to http runtime notes"

## Acceptance tests
- [ ] Run real shopping with Dexco/Telha Norte shows `search_url=` in log
