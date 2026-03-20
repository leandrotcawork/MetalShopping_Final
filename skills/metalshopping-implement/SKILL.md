---
name: metalshopping-implement
description: Implement any MetalShopping feature layer: Go module (handler, adapter, service, events), Python worker, or SDK generation. Uses real repo patterns as reference. Called by tasks in tasks/todo.md. Enforces all architectural rules automatically.
---

# MetalShopping Implement

## Before writing any code
Read `tasks/lessons.md`. Apply every lesson in the current task.

## Go module — pattern by type

**Read-only** (reference: `modules/home/` or `modules/shopping/reader.go`)
- ports/reader.go → Reader interface
- adapters/postgres/reader.go → BeginTenantTx + current_tenant_id() queries
- application/service.go → thin orchestration
- transport/http/handler.go → auth + tenant + service + writeJSON

**Write + events** (reference: `modules/shopping/writer.go` + `events/run_requested.go`)
- ports/writer.go → Writer interface
- adapters/postgres/writer.go → BeginTenantTx + INSERT + AppendInTx before Commit
- events/<n>.go → NewXxxOutboxRecord with idempotency_key="event:aggregate_id"
- Wire writer and outboxStore in composition_modules.go

**CRUD + governance** (reference: `modules/catalog/`)
- domain/model.go → types + ValidateForCreate()
- domain/errors.go → sentinel errors
- adapters/governance/<n>_guard.go → platform/governance/* resolvers
- Bootstrap key in platform/governance/bootstrap/bootstrap.go

## Go handler — always this structure
```go
func (h *Handler) handleX(w http.ResponseWriter, r *http.Request) {
    startedAt := time.Now()
    traceID, statusCode, result := requestTraceID(r), http.StatusOK, "success"
    defer logRequest("module.action", traceID, &statusCode, &result, startedAt)

    if r.Method != http.MethodGet { statusCode=405; w.WriteHeader(405); return }

    _, ok := platformauth.PrincipalFromContext(r.Context())
    if !ok { statusCode=401; writeError(w,401,"AUTH_UNAUTHORIZED",...); return }

    tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
    if !ok { statusCode=403; writeError(w,403,"TENANCY_FORBIDDEN",...); return }

    data, err := h.service.GetX(r.Context(), tenant.ID)
    if err != nil { statusCode=500; writeError(w,500,"INTERNAL_ERROR",...); return }

    writeJSON(w, 200, data)
}
```

## Python worker — always this structure
```python
# 1. Fail fast on missing env
db_url = os.getenv("MS_DATABASE_URL","").strip()
if not db_url: sys.exit(2)

# 2. Set tenant context before EVERY write transaction
cur.execute("BEGIN")
cur.execute("SELECT set_config('app.current_tenant_id',%s,true)",(tenant_id,))
cur.execute("INSERT INTO t(...) VALUES(current_tenant_id(),...)")
# ON CONFLICT ... DO UPDATE — always idempotent
cur.execute("COMMIT")

# 3. Structured logs
log("worker_start", tenant_id=tenant_id, run_id=run_id)
log("worker_end", tenant_id=tenant_id, rows_written=n)
```

Reference: `apps/integration_worker/shopping_price_worker.py`
Runtime: `apps/integration_worker/src/shopping_price_runtime/`

## After every task
1. `go build ./...` or `pnpm tsc --noEmit` — must pass
2. Smoke test (see `scripts/smoke_*.ps1`)
3. `git commit -m "<type>(<scope>): <what>"`
4. Update tasks/todo.md: mark [x]

## References
- Go patterns: `references/go-patterns.md`
- Python patterns: `references/worker-patterns.md`
- Composition: `apps/server_core/cmd/metalshopping-server/composition_modules.go`
