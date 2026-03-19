# Module Implementation Checklist

## Contract (Step 1)
- [ ] `contracts/api/openapi/<module>_v1.openapi.yaml` exists
- [ ] started from `_template.openapi.yaml`
- [ ] shared JSON schemas reused where applicable
- [ ] contract finalized before Go code started

## Go module structure (Step 2)
- [ ] `domain/model.go` — types with validation methods
- [ ] `domain/errors.go` — sentinel errors
- [ ] `ports/` — interface definitions only, no implementation
- [ ] `adapters/postgres/` — implements port using `pgdb.BeginTenantTx`
- [ ] `application/service.go` — no HTTP, no direct DB access
- [ ] `transport/http/handler.go` — no DB, no direct business logic

## Tenancy (critical — every query)
- [ ] every Postgres adapter uses `pgdb.BeginTenantTx`
- [ ] every query uses `current_tenant_id()` in WHERE clause
- [ ] `tx.Commit()` called on every transaction path
- [ ] `defer tx.Rollback()` present on every transaction

## Auth and tenancy in handlers (critical — every handler)
- [ ] `platformauth.PrincipalFromContext` checked before operation
- [ ] `tenancy_runtime.TenantFromContext` checked before operation
- [ ] structured log defer pattern present
- [ ] 401 returned when principal missing
- [ ] 403 returned when tenant context missing

## Governance (if applicable)
- [ ] governance port interface defined in `ports/`
- [ ] governance adapter in `adapters/governance/` implements port
- [ ] uses `platform/governance/` resolvers (not hardcoded values)
- [ ] governance key constant added to `platform/governance/bootstrap/bootstrap.go`

## Composition (mandatory)
- [ ] module registered in `composition_modules.go`
- [ ] follows same wiring pattern as existing modules
- [ ] no business logic added to composition file

## Python worker (if applicable — Step 3)
- [ ] worker in `apps/integration_worker/`
- [ ] every write sets `app.current_tenant_id` via `set_config`
- [ ] all inserts use `ON CONFLICT ... DO UPDATE`
- [ ] logs at start, end, and on error
- [ ] no HTTP calls from worker to server_core

## SDK generation (Step 4)
- [ ] `generate_contract_artifacts.ps1` ran after contract finalized
- [ ] `packages/generated/` not edited manually
- [ ] `pnpm tsc --noEmit` passes after generation

## React page (Step 5)
- [ ] data fetched via `@metalshopping/platform-sdk` only
- [ ] no `fetch()` in page or component files
- [ ] no manual types duplicating generated types
- [ ] legacy visual preserved where it exists

## Final
- [ ] `go build ./...` passes
- [ ] real data visible in the browser
- [ ] no regression in other modules
- [ ] all queries are tenant-scoped
