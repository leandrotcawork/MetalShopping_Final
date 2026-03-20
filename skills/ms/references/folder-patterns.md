# Folder Patterns — MetalShopping

## Go module structure (canonical)

Every module follows this exact shape.
Depth varies by complexity — do not add layers that aren't needed.

### Minimal reader (e.g. home, analytics summary)
```
modules/<name>/
  ports/
    reader.go           ← Reader interface only
  adapters/
    postgres/
      reader.go         ← implements Reader via BeginTenantTx
  application/
    service.go          ← calls reader, returns domain types
  transport/
    http/
      handler.go        ← auth + tenant + service + writeJSON
```

### Write + outbox events (e.g. shopping run_request)
```
modules/<name>/
  ports/
    reader.go           ← Reader interface
    writer.go           ← Writer interface
  adapters/
    postgres/
      reader.go
      writer.go         ← INSERT + outbox.AppendInTx in same tx
  events/
    <event_name>.go     ← NewXxxOutboxRecord builder
  application/
    service.go
  transport/
    http/
      handler.go
```

### CRUD + governance (e.g. catalog, pricing)
```
modules/<name>/
  domain/
    model.go            ← business types + ValidateForCreate()
    errors.go           ← sentinel errors (ErrXNotFound, ErrXRequired)
  ports/
    repository.go       ← Repository interface
  adapters/
    postgres/
      repository.go     ← full CRUD via BeginTenantTx
    governance/
      <n>_guard.go      ← implements governance port
  application/
    service.go          ← domain validation + governance + repo
  transport/
    http/
      handler.go
  events/               ← optional, if writes fire events
  readmodel/            ← optional, for complex read-side queries
```

## When to use readmodel/
Only when: a read query requires joining 3+ tables, or the read shape
differs significantly from the write domain model.
Not needed for: simple list or detail reads of the same entity.

## Frontend structure
```
apps/web/src/
  pages/
    <module>/
      index.tsx         ← page composition only, imports from feature-*
      <Module>Page.module.css

packages/
  feature-<module>/
    src/
      index.ts
      <Component>.tsx   ← feature-aware components (uses platform-sdk)
      <Component>.module.css

  ui/
    src/
      <Widget>.tsx      ← pure presentational, 3+ pages use it
      <Widget>.module.css
      index.ts          ← re-exports all widgets
```

## Python worker structure
```
apps/integration_worker/
  <module>_worker.py          ← entry point, env config, main()
  src/
    <module>_runtime/
      __init__.py
      models.py               ← dataclasses for inputs/outputs
      dispatcher.py           ← routes to strategy by family
      http/
        strategies.py         ← HTTP-based lookup implementations
      playwright/
        strategies.py         ← Playwright-based implementations
      shared/
        parsing.py            ← safe_float, safe_str helpers
```
Use this structure only when multiple scraping strategies exist.
Single-strategy workers can be a single file.

## Naming conventions (big tech standard)

**Go packages:** lowercase, single word — `analytics`, `shopping`, `catalog`
**Go files:** snake_case — `summary_reader.go`, `run_requested.go`
**Go types:** PascalCase — `Handler`, `Service`, `Reader`, `RunRequest`
**Go errors:** Err prefix + PascalCase — `ErrRunNotFound`, `ErrTenantMissing`
**Error codes:** MODULE_ENTITY_REASON — `SHOPPING_RUN_NOT_FOUND`, `AUTH_TENANT_MISSING`
**Routes:** /api/v1/<module>/<resource> — `/api/v1/analytics/overview`
**Contract files:** snake_case — `analytics_v1.openapi.yaml`
**DB tables:** snake_case — `shopping_price_runs`, `catalog_products`
**React components:** PascalCase — `AnalyticsOverview`, `MetricCard`
**CSS modules:** camelCase — `styles.pageTitle`, `styles.metricGrid`
**platform-sdk hooks:** use prefix + PascalCase — `useAnalyticsOverview`

## What big tech gets right that junior engineers miss

1. **Errors carry context** — not `return err` but `return fmt.Errorf("query analytics overview: %w", err)`
2. **Names don't need comments** — `getOverviewWithTenantScope()` needs no explanation
3. **One thing per function** — handler doesn't parse AND validate AND query AND format
4. **Fail loudly at startup** — missing env var = `log.Fatal`, not silent default
5. **Indexes match queries** — every WHERE has a matching index on tenant_id + filter column
6. **Pagination on every list** — never return unbounded results
7. **Idempotency by default** — every write is safe to retry
