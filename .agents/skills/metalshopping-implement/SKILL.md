---
name: metalshopping-implement
description: Implement Go modules and Python workers for MetalShopping using real repo patterns. Called by $ms for T2 and T3. Covers all module types with concrete code patterns anchored to the actual codebase.
---

# MetalShopping Implement

## Before writing code
Read `tasks/lessons.md`. Apply every lesson in this task.

## Go — by module type

### Read-only (ref: modules/home/, modules/shopping/adapters/postgres/reader.go)
```go
func (r *Reader) GetSummary(ctx context.Context, tenantID string) (ports.Summary, error) {
    tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
    if err != nil { return ports.Summary{}, err }
    defer func() { _ = tx.Rollback() }()

    var s ports.Summary
    err = tx.QueryRowContext(ctx, `
        SELECT col1, col2
        FROM table
        WHERE tenant_id = current_tenant_id()
    `).Scan(&s.Col1, &s.Col2)
    if err != nil { return ports.Summary{}, fmt.Errorf("query summary: %w", err) }

    if err := tx.Commit(); err != nil { return ports.Summary{}, fmt.Errorf("commit summary: %w", err) }
    return s, nil
}
```

### Write + outbox (ref: modules/shopping/adapters/postgres/writer.go + events/run_requested.go)
```go
func (w *Writer) CreateX(ctx context.Context, tenantID, traceID string, input ports.CreateXInput) (ports.X, error) {
    tx, err := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
    if err != nil { return ports.X{}, err }
    defer func() { _ = tx.Rollback() }()

    // INSERT
    var result ports.X
    tx.QueryRowContext(ctx, `
        INSERT INTO table (tenant_id, col1) VALUES (current_tenant_id(), $1)
        RETURNING id, col1
    `, input.Col1).Scan(&result.ID, &result.Col1)

    // Outbox event — BEFORE Commit
    record, _ := events.NewXCreatedOutboxRecord(tenantID, result, traceID, time.Now())
    w.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record})

    if err := tx.Commit(); err != nil { return ports.X{}, fmt.Errorf("commit create x: %w", err) }
    return result, nil
}

// events/x_created.go
func NewXCreatedOutboxRecord(tenantID string, x ports.X, traceID string, now time.Time) (outbox.Record, error) {
    payload, _ := json.Marshal(XCreatedPayload{ID: x.ID, TenantID: tenantID})
    return outbox.Record{
        EventName:      "module.x_created",
        EventVersion:   "v1",
        AggregateType:  "x",
        AggregateID:    x.ID,
        TenantID:       tenantID,
        TraceID:        traceID,
        IdempotencyKey: "module.x_created:" + x.ID,
        PayloadJSON:    payload,
        Status:         outbox.StatusPending,
        AvailableAt:    now,
        CreatedAt:      now,
    }, nil
}
```

### CRUD + governance (ref: modules/catalog/)
```go
// adapters/governance/x_guard.go
type XGuard struct { resolver *feature_flags.Resolver; environment string }

func (g *XGuard) IsXEnabled(_ context.Context, tenantID string) (bool, error) {
    return g.resolver.Resolve(bootstrap.XEnabledKey, feature_flags.ResolutionContext{
        Environment: g.environment, TenantID: tenantID,
    })
}
// Register key: platform/governance/bootstrap/bootstrap.go
```

## Go — handler (same structure for every endpoint)
```go
func (h *Handler) handleX(w http.ResponseWriter, r *http.Request) {
    startedAt := time.Now()
    traceID, statusCode, result := requestTraceID(r), http.StatusOK, "success"
    defer logRequest("module.action", traceID, &statusCode, &result, startedAt)

    if r.Method != http.MethodGet { statusCode = 405; w.WriteHeader(405); return }

    _, ok := platformauth.PrincipalFromContext(r.Context())
    if !ok { statusCode = 401; writeError(w, 401, "MODULE_AUTH_UNAUTHORIZED", "Authentication required", traceID); return }

    tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
    if !ok { statusCode = 403; writeError(w, 403, "MODULE_TENANCY_FORBIDDEN", "Tenant context required", traceID); return }

    data, err := h.service.GetX(r.Context(), tenant.ID)
    if err != nil { statusCode = 500; writeError(w, 500, "MODULE_INTERNAL_ERROR", "Failed to load X", traceID); return }

    writeJSON(w, http.StatusOK, data)
}
```

## Go — composition (always final step)
Register in `cmd/metalshopping-server/composition_modules.go` following the exact pattern of existing modules. No module is active until registered here.

## Python worker (scraping only)
```python
# Fail fast on missing config
db_url = os.getenv("MS_DATABASE_URL", "").strip()
if not db_url: sys.stderr.write("MS_DATABASE_URL required\n"); sys.exit(2)

log("worker_start", tenant_id=tenant_id, run_id=run_id)
try:
    with psycopg.connect(db_url) as conn:
        with conn.cursor() as cur:
            cur.execute("BEGIN")
            cur.execute("SELECT set_config('app.current_tenant_id',%s,true)", (tenant_id,))
            cur.execute("""
                INSERT INTO table (tenant_id, col1) VALUES (current_tenant_id(), %s)
                ON CONFLICT (tenant_id, id) DO UPDATE SET col1 = EXCLUDED.col1, updated_at = NOW()
            """, (value,))
            cur.execute("COMMIT")
    log("worker_end", tenant_id=tenant_id, rows_written=n)
except Exception as exc:
    log("worker_error", tenant_id=tenant_id, error=str(exc)); return 1
```
Ref: `apps/integration_worker/shopping_price_worker.py`

## Key imports
```go
pgdb         "metalshopping/server_core/internal/platform/db/postgres"
platformauth "metalshopping/server_core/internal/platform/auth"
tenancy      "metalshopping/server_core/internal/platform/tenancy_runtime"
outbox       "metalshopping/server_core/internal/platform/messaging/outbox"
bootstrap    "metalshopping/server_core/internal/platform/governance/bootstrap"
featureflags "metalshopping/server_core/internal/platform/governance/feature_flags"
```

## After task
1. `go build ./...` passes
2. Mark `[x]` in `tasks/todo.md`
3. `git commit -m "feat(<scope>): <what>"`

## References
- `references/go-patterns.md` — nullable columns, pagination, ID generation
- `references/worker-patterns.md` — event mode, multi-strategy, smoke tests
