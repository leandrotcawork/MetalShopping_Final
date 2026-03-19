# Module Scaffold Checklist

## Before writing code
- [ ] OpenAPI contract exists in `contracts/api/openapi/`
- [ ] closest reference module identified (shopping or catalog)
- [ ] `_template/README.md` has been read

## Structure
- [ ] `domain/model.go` — business types present
- [ ] `domain/errors.go` — sentinel errors defined
- [ ] `ports/repository.go` or `ports/read_models.go` — interfaces defined
- [ ] `adapters/postgres/` — implements port interface
- [ ] `application/service.go` — orchestrates without touching HTTP or DB directly
- [ ] `transport/http/handler.go` — HTTP only, calls service

## Postgres adapter
- [ ] every query uses `pgdb.BeginTenantTx`
- [ ] every query uses `current_tenant_id()` in WHERE clause
- [ ] `tx.Commit()` called even on read-only transactions
- [ ] `defer tx.Rollback()` present on every transaction path

## Handler
- [ ] `platformauth.PrincipalFromContext` checked before any operation
- [ ] `tenancy_runtime.TenantFromContext` checked before any operation
- [ ] structured log via `logRequest` defer pattern
- [ ] no DB access in handler — only service calls
- [ ] `writeJSON` used for all responses

## Governance (if applicable)
- [ ] guard implements the correct port interface
- [ ] uses `platform/governance/feature_flags`, `policy_resolver`, or `threshold_resolver`
- [ ] resolved via `bootstrap.*Key` constants

## Composition
- [ ] module registered in `composition_modules.go`
- [ ] follows same pattern as existing modules (catalog, shopping, home)

## Final
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] handler returns shape declared in the OpenAPI contract
- [ ] no hardcoded tenant IDs, DSNs, or credentials
