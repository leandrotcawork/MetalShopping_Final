# Folder Patterns

## Go module variants

### Read-only (ref: modules/home/, modules/suppliers/)
```
modules/<n>/
  ports/reader.go
  adapters/postgres/reader.go
  application/service.go
  transport/http/handler.go
```

### Write + outbox events (ref: modules/shopping/)
```
modules/<n>/
  ports/reader.go
  ports/writer.go
  adapters/postgres/reader.go
  adapters/postgres/writer.go
  events/<event_name>.go
  application/service.go
  transport/http/handler.go
```

### CRUD + governance (ref: modules/catalog/, modules/pricing/)
```
modules/<n>/
  domain/model.go
  domain/errors.go
  ports/repository.go
  adapters/postgres/repository.go
  adapters/governance/<n>_guard.go
  application/service.go
  transport/http/handler.go
  events/           (optional — if writes fire outbox events)
  readmodel/        (optional — complex read-side queries only)
```
Add `readmodel/` only when read query joins 3+ tables or read shape
differs significantly from write domain model.

## Python worker variants

### Single file (simple ingestion)
```
apps/integration_worker/<module>_worker.py
```

### Multi-strategy (ref: shopping_price_worker.py)
```
apps/integration_worker/
  <module>_worker.py
  src/<module>_runtime/
    __init__.py
    models.py
    dispatcher.py
    http/strategies.py
    playwright/strategies.py
    shared/parsing.py
```

## Frontend structure
```
apps/web/src/pages/<module>/
  index.tsx                   page composition only, no raw API calls
  <Module>Page.module.css

packages/feature-<module>/src/
  index.ts
  <Component>.tsx             feature-aware, uses platform-sdk hooks
  <Component>.module.css

packages/ui/src/
  <Widget>.tsx                pure presentational, used in 3+ pages
  <Widget>.module.css
```

## Naming conventions

| Element | Pattern | Example |
|---|---|---|
| Go package | lowercase single word | `analytics` |
| Go file | snake_case | `summary_reader.go` |
| Go struct | PascalCase | `Handler`, `Service`, `Reader` |
| Go error var | ErrXxx | `ErrRunNotFound` |
| HTTP error code | MODULE_ENTITY_REASON | `ANALYTICS_OVERVIEW_FAILED` |
| Route | /api/v1/module/resource | `/api/v1/analytics/overview` |
| Contract file | snake_case | `analytics_v1.openapi.yaml` |
| DB table | snake_case | `shopping_price_runs` |
| React component | PascalCase | `AnalyticsOverview` |
| CSS module key | camelCase | `styles.pageTitle` |
| platform-sdk hook | useXxx | `useAnalyticsOverview` |

## What senior engineers do that juniors miss

1. Errors carry context: `fmt.Errorf("query analytics overview: %w", err)`
2. Fail loud at startup: missing env = `log.Fatal`, never silent default
3. Indexes match queries: every WHERE has matching index on (tenant_id, filter_col)
4. Pagination on every list: never unbounded results
5. One function, one responsibility: handler doesn't parse + validate + query + format
