# Go Patterns — MetalShopping

## Postgres adapter (every query)
```go
tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
if err != nil { return zero, err }
defer func() { _ = tx.Rollback() }()

row := tx.QueryRowContext(ctx, `
    SELECT col1, col2
    FROM table
    WHERE tenant_id = current_tenant_id()
    AND id = $1`, id)

if err := tx.Commit(); err != nil {
    return zero, fmt.Errorf("commit <action>: %w", err)
}
```

## Write + outbox (atomic)
```go
tx, err := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
defer func() { _ = tx.Rollback() }()

// 1. INSERT
tx.ExecContext(ctx, `INSERT INTO t(...) VALUES(current_tenant_id(),...)`, ...)

// 2. Outbox — BEFORE Commit, INSIDE same tx
record, _ := events.NewXxxOutboxRecord(tenantID, result, traceID, now)
w.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record})

// 3. Commit both atomically
tx.Commit()
```

## Error codes pattern
MODULE_ENTITY_REASON — e.g.:
  SHOPPING_RUN_NOT_FOUND
  AUTH_TENANT_MISSING
  CATALOG_PRODUCT_SKU_DUPLICATE

## Structured log defer
```go
startedAt := time.Now()
traceID   := requestTraceID(r)
statusCode := http.StatusOK
reqResult  := "success"
defer logRequest("module.action", traceID, &statusCode, &reqResult, startedAt)
```

## Composition wiring (composition_modules.go)
Follow existing pattern exactly — reader → service → handler → RegisterRoutes.
New module = new block following the same structure as catalog or shopping.

## Key imports
```go
pgdb          "metalshopping/server_core/internal/platform/db/postgres"
platformauth  "metalshopping/server_core/internal/platform/auth"
tenancy       "metalshopping/server_core/internal/platform/tenancy_runtime"
outbox        "metalshopping/server_core/internal/platform/messaging/outbox"
govbootstrap  "metalshopping/server_core/internal/platform/governance/bootstrap"
featureflags  "metalshopping/server_core/internal/platform/governance/feature_flags"
```
