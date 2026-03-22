---
name: ms
description: MetalShopping master orchestrator. Use for any implementation task. Enters plan mode automatically, validates architecture and folder structure before any code, orchestrates specialist skills in the correct order, enforces big-tech engineering standards.
---

# MetalShopping — Master Orchestrator

**Never write code before Phase 1 and Phase 2 are complete and approved.**

---

## Phase 1 — Architectural thinking (always before code)

Read first: `tasks/lessons.md` → `ARCHITECTURE.md` → `docs/PROJECT_SOT.md`

Answer these questions explicitly before proceeding:

**1. Module type** — determines everything:
- `read-only` → Reader + postgres reader — ref: `modules/home/`
- `write+events` → Writer + AppendInTx + events/ — ref: `modules/shopping/`
- `CRUD+governance` → domain + Repository + gov adapters — ref: `modules/catalog/`
- `scraping` → Python worker + Go reader — ref: `integration_worker/`

**2. Exact folder structure** — state before creating any file:
```
apps/server_core/internal/modules/<n>/
  ports/
    reader.go          ← or writer.go or repository.go
  adapters/
    postgres/
      reader.go        ← BeginTenantTx + current_tenant_id()
    governance/        ← only if governance checks needed
  events/              ← only if outbox events fired
  application/
    service.go
  transport/
    http/
      handler.go
```
See `references/folder-patterns.md` for all variants and naming conventions.

**3. Risks** — identify before coding:
- N+1 query risk in list endpoints?
- Missing index on (tenant_id, filter_col)?
- Event payload missing fields the worker will need?
- Frontend needs data requiring multiple queries → combine in service?

**4. Level scope:**
Level 1 = real data on screen, build passes, no mocks.
State what is deferred to Level 2.

---

## Phase 2 — Plan (write tasks/todo.md, wait for approval)

```markdown
## Feature: <n>
Type: <module type>  |  Events: <yes: names | no>  |  ADR: <yes: number | no>

## Tasks
- [ ] T1: contract — $metalshopping-openapi-contracts
      commit: "feat(<m>): add OpenAPI contract"
- [ ] T2: Go module — implement per Phase 1 structure
      commit: "feat(<m>): implement handler and adapter"
- [ ] T3: worker — (scraping only)
      commit: "feat(worker): implement <m> worker"
- [ ] T4: SDK — $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate after <m>"
- [ ] T5: frontend — $metalshopping-frontend
      commit: "feat(<m>): implement React page"
- [ ] T6: ADR — $metalshopping-adr (if needed)
      commit: "docs(adr): ADR-XXXX <title> — verified and closed"

## Acceptance tests
- [ ] go build ./... passes
- [ ] pnpm tsc --noEmit passes
- [ ] GET /api/v1/<route> returns real data (no mock)
- [ ] data visible in browser
- [ ] smoke: <script name>
- [ ] no regression
```

**Present plan. Wait for explicit approval. Then begin T1.**

---

## Phase 3 — Execute (one task at a time, commit after each)

### T1 — OpenAPI contract → use `$metalshopping-openapi-contracts`
One file per bounded context: `contracts/api/openapi/<m>_v1.openapi.yaml`
Reuse schemas from `contracts/api/jsonschema/` when they exist.

### T2 — Go module → follow Phase 1 structure exactly

**Every adapter:**
```go
tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
defer func() { _ = tx.Rollback() }()
// queries use WHERE tenant_id = current_tenant_id()
tx.Commit()
```

**Every handler:**
```go
startedAt := time.Now()
traceID, statusCode, result := requestTraceID(r), http.StatusOK, "success"
defer logRequest("module.action", traceID, &statusCode, &result, startedAt)

if r.Method != http.MethodGet { statusCode = 405; w.WriteHeader(405); return }

_, ok := platformauth.PrincipalFromContext(r.Context())
if !ok { statusCode = 401; writeError(w, 401, "AUTH_UNAUTHORIZED", ...); return }

tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
if !ok { statusCode = 403; writeError(w, 403, "TENANCY_FORBIDDEN", ...); return }

data, err := h.service.GetX(r.Context(), tenant.ID)
if err != nil { statusCode = 500; writeError(w, 500, "INTERNAL_ERROR", ...); return }
writeJSON(w, 200, data)
```

**Write + outbox** (ref: `modules/shopping/adapters/postgres/writer.go`):
```go
tx, _ := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
defer func() { _ = tx.Rollback() }()
tx.ExecContext(ctx, `INSERT INTO t(...) VALUES(current_tenant_id(),...)`, ...)
w.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}) // BEFORE Commit
tx.Commit()
```

**Final step:** register in `cmd/metalshopping-server/composition_modules.go`

### T3 — Python worker (scraping only)
```python
db_url = os.getenv("MS_DATABASE_URL", "").strip()
if not db_url: sys.exit(2)                    # fail fast

cur.execute("BEGIN")
cur.execute("SELECT set_config('app.current_tenant_id',%s,true)", (tenant_id,))
cur.execute("INSERT INTO t(...) VALUES(current_tenant_id(),...) ON CONFLICT (...) DO UPDATE SET ...")
cur.execute("COMMIT")
```
Ref: `apps/integration_worker/shopping_price_worker.py`
Event mode: `MS_SHOPPING_WORKER_MODE=event`
Never call server_core HTTP. Never hardcode tenant_id or DSN.

### T4 — SDK → use `$metalshopping-sdk-generation`
`./scripts/generate_contract_artifacts.ps1`
Never edit `packages/generated/` manually.

### T5 — Frontend → use `$metalshopping-frontend`

### T6 — ADR → use `$metalshopping-adr`

### After every task
1. `go build ./...` or `pnpm tsc --noEmit` — must pass
2. Mark `[x]` in `tasks/todo.md`
3. `git commit -m "<type>(<scope>): <what>"`

---

## Phase 4 — Review before declaring done

Quick self-check:
- [ ] Every adapter uses `pgdb.BeginTenantTx` + `current_tenant_id()`?
- [ ] Every handler checks `PrincipalFromContext` + `TenantFromContext`?
- [ ] Outbox via `AppendInTx` before `Commit`?
- [ ] No `fetch()` in any React component?
- [ ] Real data in browser (no mocks)?
- [ ] All acceptance tests pass?

Run `$metalshopping-review` for full 10-lens review on significant features.

---

## After any correction during execution
Write lesson to `tasks/lessons.md` immediately (format in `AGENTS.md`).
Apply the lesson for the rest of the session.

## References
- `references/folder-patterns.md` — all folder structures + naming conventions
- `skills/metalshopping-implement/references/go-patterns.md` — Go code patterns
- `skills/metalshopping-frontend/references/design-tokens.md` — UI tokens
