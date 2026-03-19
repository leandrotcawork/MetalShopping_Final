---
name: metalshopping-module-implementation
description: Implement a complete MetalShopping feature end-to-end: OpenAPI contract, Go module (domain + ports + adapters + service + handler + composition), optional Python worker, SDK generation, and React page. Use as the single orchestrating skill whenever a new feature needs to be built from zero. This skill enforces the frozen delivery sequence and references the real patterns from the existing codebase.
---

# MetalShopping Module Implementation

## Overview

Use this skill to orchestrate a full feature delivery with the frozen MetalShopping sequence. Keep the work anchored to repository standards and specialist skills instead of inventing local shortcuts.

## Workflow

1. Read only the minimum frozen context:
   `ARCHITECTURE.md`
   `docs/PROJECT_SOT.md`
   `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`
   `docs/IMPLEMENTATION_PLAN.md`
2. Define the contract first with `metalshopping-openapi-contracts`:
   `contracts/api/openapi/<module>_v1.openapi.yaml`
3. Implement the Go module with `metalshopping-module-scaffold`:
   `domain`, `ports`, `adapters`, `application`, `transport`, composition wiring.
4. Enforce tenancy in every Postgres adapter using:
   `pgdb.BeginTenantTx` and `current_tenant_id()`.
5. Enforce handler guardrails in every endpoint:
   `platformauth.PrincipalFromContext` and `tenancy_runtime.TenantFromContext`.
6. Add governance adapter under `adapters/governance/` only when policy checks are required.
7. Register the module in:
   `apps/server_core/cmd/metalshopping-server/composition_modules.go`.
8. If scraping or Python-only libraries are needed, use `metalshopping-worker-scaffold`:
   set `app.current_tenant_id`, use idempotent upserts, never call server_core HTTP.
9. Generate SDK artifacts with `metalshopping-sdk-generation`:
   run `./scripts/generate_contract_artifacts.ps1`.
10. Implement the page with `metalshopping-page-delivery`:
   consume `@metalshopping/platform-sdk`, avoid direct `fetch()`.

## Delivery sequence

1. Step 1: OpenAPI contract in `contracts/api/openapi/`
2. Step 2: Go module in `server_core`
3. Step 3: Python worker in `integration_worker` only if needed
4. Step 4: SDK generation
5. Step 5: React page consuming platform SDK

Step 3 is optional. All other steps are mandatory and ordered.

## Guardrails

- never query Postgres directly from handlers
- never skip `pgdb.BeginTenantTx` in adapters
- never omit `current_tenant_id()` from tenant-scoped queries
- never skip principal or tenant context checks in handlers
- never call server_core HTTP from workers
- never write worker data without setting `app.current_tenant_id`
- never use direct `fetch()` in pages/components
- never hand-edit `packages/generated/`
- never leave the module unregistered in composition

## Output expectations

When using this skill:

- preserve contract-first delivery
- keep module boundaries explicit and clean
- preserve tenancy/auth/governance safety in all layers
- keep frontend thin and SDK-driven
- finish with verifiable completion criteria

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
- For repo touchpoints and sequence detail, read `references/implementation-flow.md`.
- For the final review pass, read `references/implementation-checklist.md`.
