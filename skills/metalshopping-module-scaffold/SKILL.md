---
name: metalshopping-module-scaffold
description: Scaffold or review a Go business module under `apps/server_core/internal/modules/` following the MetalShopping frozen module structure. Use when adding a new module such as Home, Shopping, Analytics, CRM, or any new bounded-context. The canonical reference modules are `catalog` (full structure with governance) and `shopping` (read-only module consuming worker-written tables). Always start from the `_template` folder and follow the existing patterns exactly.
---

# MetalShopping Module Scaffold

## Overview

Every Go module in this repo follows the same frozen structure.
Do not invent alternative shapes. Read the template and the
reference modules before writing any code.

## Reference modules

| Module | What it shows |
|---|---|
| `internal/modules/catalog/` | full CRUD + domain + governance adapters + events + readmodel |
| `internal/modules/shopping/` | read-only module consuming worker-written Postgres tables |
| `internal/modules/home/` | minimal summary reader, no domain model needed |
| `internal/modules/pricing/` | governance policy guard in adapter layer |

## Frozen module structure

```
internal/modules/<name>/
  domain/
    model.go      ← business types and validation (ValidateForCreate etc.)
    errors.go     ← sentinel errors (ErrXNotFound, ErrXRequired etc.)
  application/
    service.go    ← orchestrates domain + ports, no DB or HTTP
  ports/
    repository.go ← interface definitions (Repository, Reader, etc.)
  adapters/
    postgres/
      repository.go  ← implements ports using pgdb.BeginTenantTx
    governance/       ← only if module uses governance guards
      <name>_guard.go ← implements port using platform/governance/*
  transport/
    http/
      handler.go  ← HTTP only: parse, call service, write JSON
  events/
    doc.go or <event_name>.go  ← event types aligned with contracts/events/
  readmodel/
    doc.go or <name>.go  ← read-side service if needed
```

At Level 1, the minimum required files are:
- `domain/model.go` and `domain/errors.go` (even if minimal)
- `ports/repository.go` (interface)
- `adapters/postgres/repository.go` or `reader.go`
- `application/service.go`
- `transport/http/handler.go`

## Postgres adapter pattern — mandatory

Every Postgres adapter must use `pgdb.BeginTenantTx` for tenant isolation:

```go
// Read-only query
tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
if err != nil {
    return zero, err
}
defer func() { _ = tx.Rollback() }()

// Query uses current_tenant_id() — set by BeginTenantTx
const query = `SELECT ... FROM table WHERE tenant_id = current_tenant_id() AND ...`

// Always commit even on read-only
if err := tx.Commit(); err != nil {
    return zero, fmt.Errorf("commit <action>: %w", err)
}
```

Never query without `pgdb.BeginTenantTx`. Never hardcode tenant_id
in SQL — always use `current_tenant_id()`.

## Handler pattern — mandatory

Every handler follows this exact structure:

```go
func (h *Handler) handleSomething(w http.ResponseWriter, r *http.Request) {
    startedAt := time.Now()
    traceID   := requestTraceID(r)
    statusCode := http.StatusOK
    reqResult  := "success"
    defer logRequest("module.action", traceID, &statusCode, &reqResult, startedAt)

    if r.Method != http.MethodGet { // or POST, etc.
        statusCode = http.StatusMethodNotAllowed
        reqResult  = "method_not_allowed"
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    _, ok := platformauth.PrincipalFromContext(r.Context())
    if !ok { /* write 401, return */ }

    tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
    if !ok { /* write 403, return */ }

    result, err := h.service.DoSomething(r.Context(), tenant.ID)
    if err != nil {
        statusCode = http.StatusInternalServerError
        reqResult  = "internal_error"
        writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "...", traceID)
        return
    }

    writeJSON(w, http.StatusOK, result)
}
```

Reference: `internal/modules/shopping/transport/http/handler.go`

## Governance adapter pattern — use when module needs a policy/flag

```go
// adapters/governance/my_guard.go
type MyGuard struct {
    resolver    *feature_flags.Resolver  // or policy_resolver / threshold_resolver
    environment string
}

func (g *MyGuard) IsEnabled(_ context.Context, tenantID string) (bool, error) {
    return g.resolver.Resolve(bootstrap.MyFlagKey, feature_flags.ResolutionContext{
        Environment: g.environment,
        TenantID:    tenantID,
    })
}
```

Reference: `internal/modules/catalog/adapters/governance/product_creation_guard.go`

## Composition — mandatory

After implementing the module, register it in:
`apps/server_core/cmd/metalshopping-server/composition_modules.go`

Follow the pattern already used for catalog, shopping, home, pricing.
Do not add composition logic anywhere else.

## Workflow

1. Read `apps/server_core/internal/modules/_template/README.md`
2. Identify the closest reference module (shopping for read-only, catalog for CRUD)
3. Confirm the OpenAPI contract exists in `contracts/api/openapi/`
4. Create module folder structure
5. Implement domain model and errors
6. Define port interfaces
7. Implement Postgres adapter using `pgdb.BeginTenantTx`
8. Implement application service
9. Implement HTTP handler
10. Register in `composition_modules.go`
11. Run `go build ./...` — must pass
12. Finish with `references/scaffold-checklist.md`

## References

- Template: `apps/server_core/internal/modules/_template/`
- Read-only reference: `apps/server_core/internal/modules/shopping/`
- Full CRUD reference: `apps/server_core/internal/modules/catalog/`
- Governance reference: `apps/server_core/internal/modules/pricing/adapters/governance/`
- Composition: `apps/server_core/cmd/metalshopping-server/composition_modules.go`
- For the final review pass, read `references/scaffold-checklist.md`.
