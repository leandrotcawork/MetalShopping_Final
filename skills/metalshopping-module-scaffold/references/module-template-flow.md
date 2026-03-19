# Module Template Flow

## Read order

1. `apps/server_core/internal/modules/_template/README.md`
2. Reference module for your case:
   - Read-only (consuming worker tables): `internal/modules/shopping/`
   - Full CRUD with governance: `internal/modules/catalog/`
   - Minimal summary reader: `internal/modules/home/`

## Module categories

### Read-only module (like shopping)
- No `domain/` model needed if data comes from worker-written tables
- Ports: `Reader` interface only
- Adapter: postgres reader using `BeginTenantTx` + `current_tenant_id()`
- Service: thin orchestration layer
- Handler: auth check → tenant check → service call → writeJSON

### CRUD module (like catalog)
- Full `domain/model.go` with validation methods
- Ports: `Repository` interface + optional governance port interfaces
- Adapters: postgres repository + governance adapters
- Service: domain validation → governance checks → repo write
- Handler: parse request → call service → handle domain errors

## Files always touched by this skill

- `apps/server_core/internal/modules/<n>/` (new folder tree)
- `apps/server_core/cmd/metalshopping-server/composition_modules.go`
- `contracts/api/openapi/<n>_v1.openapi.yaml` (must exist before handler)

## Files sometimes touched

- `apps/server_core/internal/platform/governance/bootstrap/bootstrap.go`
  (add new governance key constants if module needs governance guards)
