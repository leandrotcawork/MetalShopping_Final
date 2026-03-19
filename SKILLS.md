# MetalShopping Skills Guide

## Status update

No dedicated orchestrator skill is active.
Use the specialist skills directly based on the feature step.

## Active skill map

### Specialists (use in implementation order)

1. `metalshopping-openapi-contracts` - Step 1: OpenAPI contract
2. `metalshopping-module-scaffold` - Step 2: Go module (domain, ports, adapters, handler)
3. `metalshopping-worker-scaffold` - Step 3: Python worker (only when scraping/Python-only is required)
4. `metalshopping-sdk-generation` - Step 4: SDK generation
5. `metalshopping-page-delivery` - Step 5: React page

### Governance and review

- `metalshopping-architecture-direction-review`
- `metalshopping-adr-updates`
- `metalshopping-observability-security`
- `metalshopping-governance-contracts`
- `metalshopping-event-contracts`
- `metalshopping-contract-authoring`
- `metalshopping-platform-packages`
- `metalshopping-server-core-modules`
- `metalshopping-worker-patterns`
- `metalshopping-frontend-migration-guardrails`

## Non-negotiable implementation standards

### Go handler

- always validate `platformauth.PrincipalFromContext` before operating
- always validate `tenancy_runtime.TenantFromContext` before operating
- never access DB directly from handler, use service/repository boundary

### Postgres adapter

- always use `pgdb.BeginTenantTx` for queries
- always use `current_tenant_id()` in tenant filtering
- never hardcode tenant values in queries

### Python worker

- always execute `set_config('app.tenant_id', ...)` before writes
- never call `server_core` HTTP endpoints directly
- never write without tenant context

### React page

- always consume data through `@metalshopping/platform-sdk`
- never call `fetch()` directly in pages/components

### Composition

- every new Go module must be registered in `composition_modules.go`
- no module is considered active if it is not wired in composition

## Level 1 done criteria

- real data visible on screen (no mock)
- typecheck/build pass
- no regression in login and previously closed modules
