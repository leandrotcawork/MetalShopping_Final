---
id: hippocampus-conventions
title: MetalShopping Conventions
type: hippocampus
tags: [conventions, rules, go, frontend, python, process]
updated_at: 2026-03-26
---

# MetalShopping Conventions

## Go (server_core) — Absolute Rules

- `pgdb.BeginTenantTx` on **every** Postgres adapter query — no exceptions
- `current_tenant_id()` in **every** WHERE clause on tenant-scoped tables
- `platformauth.PrincipalFromContext` → 401 before any handler operation
- `tenancy_runtime.TenantFromContext` → 403 before any handler operation
- `outbox.AppendInTx` inside the **same transaction** as INSERT — never after Commit
- Every new module must be registered in `composition_modules.go`
- Error codes: `MODULE_ENTITY_REASON` structured format
- Every handler logs: `trace_id`, `action`, `result`, `duration_ms`
- Every write is idempotent and retry-safe
- No query ever returns cross-tenant data

## Frontend — Absolute Rules

- Data only via `sdk.*` methods from `@metalshopping/sdk-runtime` — no raw `fetch()`
- Design tokens only — no hardcoded hex values (`$metalshopping-design-system`)
- Check `packages/ui/src/index.ts` before creating any component
- Every data-fetching component must have loading + error + empty states
- Fetch pattern: `useEffect + cancelled flag`

## Python Workers — Absolute Rules

- `set_config('app.current_tenant_id', %s, true)` before every write transaction
- `ON CONFLICT ... DO UPDATE` on every insert (idempotency)
- Never call `server_core` HTTP endpoints (one-way dependency)

## Contract Layer

- `contracts/` are hand-authored — never derive contracts from code
- Auto-generated SDK artifacts in `packages/` are never edited manually
- Run `scripts/generate_contract_artifacts.ps1 -Target all` after any contract change
- Run `scripts/validate_contracts.ps1 -Scope all` before committing contract changes

## Commit Format

```
<type>(<scope>): <what>
```
Types: `feat | fix | docs | chore | refactor`

## Process

- A task is done only when: build passes + real data verified + commit made
- One commit per completed task — no uncommitted work at session end
- ADR committed only after the acceptance test passes
- Names are self-documenting — no cryptic abbreviations

## Engineering Bar

Every decision passes: *"Would a Stripe or Google senior engineer approve this in code review?"*

## Module Creation Checklist

1. Create under `internal/modules/<name>/` with full layer structure
2. Register in `composition_modules.go`
3. Add OpenAPI spec in `contracts/api/openapi/<name>.yaml`
4. Add event schemas if emitting events
5. Regenerate SDK artifacts after contract changes
