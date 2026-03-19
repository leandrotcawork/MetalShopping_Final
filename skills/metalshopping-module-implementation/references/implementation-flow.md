# Implementation Flow

## Read order

1. `ARCHITECTURE.md` - frozen principles
2. `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md` - delivery philosophy
3. `apps/server_core/internal/modules/_template/README.md` - module shape
4. Reference module for the category:
   - `internal/modules/shopping/` -> read-only from worker tables
   - `internal/modules/catalog/` -> CRUD with governance

## Module category decision tree

```
Does the feature need scraping, Playwright, or Python-only libs?
  YES -> Python worker (Step 3) + Go reader module (Step 2)
  NO  ->
    Does it write business data?
      YES -> CRUD module (domain + ports + repo + service + handler)
      NO  -> Read-only module (ports + reader + service + handler)
```

## Files touched by a full feature delivery

### Contract
- `contracts/api/openapi/<module>_v1.openapi.yaml`
- `contracts/api/jsonschema/*.schema.json` (if new shared schemas)

### Go module
- `apps/server_core/internal/modules/<module>/domain/model.go`
- `apps/server_core/internal/modules/<module>/domain/errors.go`
- `apps/server_core/internal/modules/<module>/ports/repository.go`
- `apps/server_core/internal/modules/<module>/adapters/postgres/repository.go`
- `apps/server_core/internal/modules/<module>/adapters/governance/<n>_guard.go` (if needed)
- `apps/server_core/internal/modules/<module>/application/service.go`
- `apps/server_core/internal/modules/<module>/transport/http/handler.go`
- `apps/server_core/cmd/metalshopping-server/composition_modules.go`
- `apps/server_core/internal/platform/governance/bootstrap/bootstrap.go` (if new keys)

### Python worker (if needed)
- `apps/integration_worker/<module>_worker.py`
- `apps/integration_worker/requirements.txt` (if new deps)

### Frontend
- `apps/web/src/pages/<module>/` (page component)
- `packages/feature-<module>/src/` (adapter if needed)
- `packages/ui/src/` (widgets if 3+ uses)

## Key rules - never violate

1. Contract before code - always
2. Postgres adapter always uses `pgdb.BeginTenantTx` + `current_tenant_id()`
3. Handler always checks `PrincipalFromContext` and `TenantFromContext`
4. Worker sets `app.tenant_id` before every write transaction
5. Worker never calls server_core HTTP endpoints
6. React page never uses `fetch()` directly - always platform-sdk
7. Module always registered in `composition_modules.go`
8. `packages/generated/` is never edited manually
