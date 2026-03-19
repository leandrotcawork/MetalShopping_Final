---
name: metalshopping-module-implementation
description: Implement a complete MetalShopping feature end-to-end: OpenAPI contract, Go module (domain + ports + adapters + service + handler + composition), optional Python worker, SDK generation, and React page. Use as the single orchestrating skill whenever a new feature needs to be built from zero. This skill enforces the frozen delivery sequence and references the real patterns from the existing codebase.
---

# MetalShopping Module Implementation

## Overview

This skill orchestrates a complete feature delivery using the exact
patterns already established in the repository. It calls the right
specialist skills in the right order and prevents the most common
mistakes: wrong tenancy, direct DB access from handlers, workers
calling server_core, fetch() in React pages.

## Frozen delivery sequence — never reorder

```
Step 1 → OpenAPI contract in contracts/api/openapi/
Step 2 → Go module in server_core (domain → ports → adapters → service → handler → composition)
Step 3 → Python worker in integration_worker (only if scraping/Python-only needed)
Step 4 → SDK generation (make generate_contract_artifacts)
Step 5 → React page consuming sdk-runtime (no fetch() direct)
```

Steps 3 is optional. All others are mandatory, in order.

## Step 1 — OpenAPI contract

Use skill: `metalshopping-openapi-contracts`

- contract file: `contracts/api/openapi/<module>_v1.openapi.yaml`
- start from `contracts/api/openapi/_template.openapi.yaml`
- reuse JSON schemas from `contracts/api/jsonschema/` when they exist
- do not write Go code before the contract is finalized

## Step 2 — Go module in server_core

Use skill: `metalshopping-module-scaffold`

The complete Go module follows this structure and set of patterns:

### 2a. Identify module category

| Category | Example | What to build |
|---|---|---|
| Read-only (worker feeds data) | shopping | ports.Reader + postgres reader + service + handler |
| CRUD with governance | catalog, pricing | domain + ports + pg repo + governance adapters + service + handler |
| Minimal summary | home | ports + pg reader + thin service + handler |

### 2b. Postgres adapter — tenancy is mandatory

Every single query must use `pgdb.BeginTenantTx` and `current_tenant_id()`.
There are no exceptions. Reference:
`internal/modules/shopping/adapters/postgres/reader.go`

```go
tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
defer func() { _ = tx.Rollback() }()
// queries use WHERE tenant_id = current_tenant_id()
tx.Commit()
```

### 2c. Handler — auth and tenancy checks are mandatory

Every handler must validate auth and tenant before any operation.
Reference: `internal/modules/shopping/transport/http/handler.go`

```go
_, ok := platformauth.PrincipalFromContext(r.Context())
if !ok { /* 401 */ }
tenant, ok := tenancy_runtime.TenantFromContext(r.Context())
if !ok { /* 403 */ }
```

Every handler must use the structured log defer pattern:
```go
startedAt := time.Now()
traceID := requestTraceID(r)
statusCode := http.StatusOK
reqResult := "success"
defer logRequest("module.action", traceID, &statusCode, &reqResult, startedAt)
```

### 2d. Governance adapter — only if module needs policies/flags

When a module needs to check a governance rule, add an adapter in
`adapters/governance/` implementing the port interface.
Reference: `internal/modules/catalog/adapters/governance/product_creation_guard.go`
Reference: `internal/modules/pricing/adapters/governance/manual_override_guard.go`

### 2e. Composition — mandatory final step

Register the new module in:
`apps/server_core/cmd/metalshopping-server/composition_modules.go`

Follow the exact same pattern as catalog, shopping, home, pricing.
This is the only place where modules are wired together.

## Step 3 — Python worker (only when needed)

Use skill: `metalshopping-worker-scaffold`

Required when: scraping (Playwright), Python-only libraries, batch ingestion.
NOT required when: data is already in Postgres, Go can query it directly.

**The shopping module is the reference**: worker writes to Postgres,
Go reads from those same tables. The worker never calls server_core.

Mandatory patterns for any new worker:
1. Set tenant context before every write:
   `cur.execute("SELECT set_config('app.current_tenant_id', %s, true)", (tenant_id,))`
2. Idempotent upserts: `ON CONFLICT ... DO UPDATE SET updated_at = NOW()`
3. Structured logs at start, end, and error
4. Env vars for all config — fail fast if missing

## Step 4 — SDK generation

Use skill: `metalshopping-sdk-generation`

```powershell
.\scripts\generate_contract_artifacts.ps1
```

Do not write any React code before this step completes.
Do not edit `packages/generated/` manually — ever.

## Step 5 — React page

Use skill: `metalshopping-page-delivery`

- import data via `@metalshopping/platform-sdk` only
- no `fetch()` in page or component files
- preserve legacy visual layout from v2 where it exists
- extract widget to `packages/ui` only if 3+ pages use it

## What this skill must never produce

- Go handler that queries Postgres directly (must go through service + adapter)
- Postgres adapter that does not use `pgdb.BeginTenantTx`
- Query without `current_tenant_id()` in WHERE clause
- Handler that skips `platformauth.PrincipalFromContext` check
- Handler that skips `tenancy_runtime.TenantFromContext` check
- Python worker that calls server_core HTTP endpoints
- Python worker that writes without setting `app.current_tenant_id`
- `fetch()` in React page or component files
- Manual types in `packages/generated/` (generated only)
- Business logic added to composition_modules.go
- New module NOT registered in composition_modules.go

## Level 1 completion criteria

- real data visible in the browser
- `go build ./...` passes
- `pnpm tsc --noEmit` passes
- no regression in previously working modules
- every DB query is tenant-scoped via `current_tenant_id()`

## References

- Canonical read-only module: `internal/modules/shopping/`
- Canonical CRUD module: `internal/modules/catalog/`
- Canonical worker: `apps/integration_worker/shopping_price_worker.py`
- Composition: `apps/server_core/cmd/metalshopping-server/composition_modules.go`
- For the final review pass, read `references/implementation-checklist.md`.
