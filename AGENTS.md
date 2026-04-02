# AGENTS — MetalShopping

## Purpose

This file is the repository-wide mandatory entrypoint for any AI agent working in MetalShopping.

## Mandatory read order

Before planning, implementation, or review, read these files in this exact order:

1. Read `docs/PROJECT_SOT.md`
2. Read `ARCHITECTURE.md`
3. Read `AGENTS.md`
4. Read the agent-specific file (`CLAUDE.md` or `CODEX.md`)
5. Read `tasks/todo.md`
6. Read `tasks/lessons.md`

Do not start planning, implementation, or review before completing that sequence.

## Documentation precedence

If documents conflict, follow this order:

1. `docs/PROJECT_SOT.md`
2. `ARCHITECTURE.md`
3. `AGENTS.md`
4. `CLAUDE.md` or `CODEX.md`
5. `docs/IMPLEMENTATION_PLAN.md`
6. `docs/PROGRESS.md`
7. `tasks/todo.md`
8. `tasks/lessons.md`

## Product objective

MetalShopping is a long-lived enterprise platform for commercial strategy, pricing, procurement, CRM, analytics, automation, and future AI-assisted operations.

Do not optimize for one-off delivery. Always preserve future module growth, clean boundaries, governance, and multi-tenant safety.

## Absolute rules

### Go
- `pgdb.BeginTenantTx` on every Postgres adapter query
- `current_tenant_id()` in every WHERE on tenant-scoped tables
- `platformauth.PrincipalFromContext` -> 401 before handler work
- `tenancy_runtime.TenantFromContext` -> 403 before handler work
- `outbox.AppendInTx` inside the same transaction as the write
- every new module registered in `composition_modules.go`

### Python worker
- `set_config('app.current_tenant_id', %s, true)` before every write transaction
- `ON CONFLICT ... DO UPDATE` on every insert
- never call `server_core` HTTP endpoints directly

### Frontend
- data only through `sdk.*` from `@metalshopping/sdk-runtime`
- design tokens only, no hardcoded hex values
- check `packages/ui/src/index.ts` before creating a component
- every data-fetching surface has loading, error, and empty states
- fetch pattern is `useEffect + cancelled flag`

### Process
- `packages/generated/` and `packages/generated-types/` are never edited manually
- no task is done without verification and a commit
- no architectural rule changes without updating SoT documents and ADRs when needed

## Engineering bar

Every decision must survive the question:

`Would a Stripe or Google senior engineer approve this in code review?`

## Skill map

| Task | Skill |
|---|---|
| Any implementation (default) | `$ms` |
| OpenAPI contract | `$metalshopping-openapi-contracts` |
| Event contract | `$metalshopping-event-contracts` |
| Governance contract | `$metalshopping-governance-contracts` |
| SDK generation | `$metalshopping-sdk-generation` |
| ADR lifecycle | `$metalshopping-adr` |
| Frontend visual/component work | `$metalshopping-design-system` |
