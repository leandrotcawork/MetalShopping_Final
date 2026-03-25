---
id: hippocampus-conventions
title: Conventions
region: hippocampus
tags: [conventions, rules, patterns, absolute-rules]
links:
  - hippocampus/architecture
  - cortex/backend/index
weight: 0.95
updated_at: 2026-03-24T10:00:00Z
---

# MetalShopping Conventions & Absolute Rules

## Go Backend — Absolute Rules (No Exceptions)

### Tenant Safety (Non-Negotiable)

1. **`pgdb.BeginTenantTx` on every Postgres adapter query** — no exceptions
   - Usage: Start every database transaction with tenant context
   - Why: Prevents cross-tenant data leaks

2. **`current_tenant_id()` in every WHERE clause on tenant-scoped tables**
   - Usage: Every SELECT/UPDATE/DELETE must filter by tenant
   - Why: Defense in depth against accidental cross-tenant queries

3. **`platformauth.PrincipalFromContext` → 401 before any handler operation**
   - Usage: First line in every HTTP handler
   - Why: Unauthenticated requests must fail fast

4. **`tenancy_runtime.TenantFromContext` → 403 before any handler operation**
   - Usage: Second line in every HTTP handler (after auth check)
   - Why: Wrong-tenant requests must fail fast

### Event Publishing

5. **`outbox.AppendInTx` inside the same transaction as INSERT — never after Commit**
   - Usage: Events must be published within the same database transaction
   - Why: Ensures atomicity; prevents lost events

### Module Registration

6. **Every new module registered in `composition_modules.go`**
   - Usage: New module = add entry to composition
   - Why: Ensures module is wired into the application

## Python Worker — Absolute Rules

1. **`set_config('app.current_tenant_id', %s, true)` before every write transaction**
   - Usage: Set tenant context before any database write
   - Why: Worker runs in separate process; must explicitly set tenant

2. **`ON CONFLICT ... DO UPDATE` on every insert (idempotency)**
   - Usage: All inserts must be idempotent (safe to retry)
   - Why: Worker is retry-safe; idempotency prevents duplicates

3. **Never call `server_core` HTTP endpoints (one-way dependency)**
   - Usage: Worker reads from database only, never calls backend
   - Why: Maintains separation; prevents circular dependencies

## Frontend (React) — Absolute Rules

1. **Data only via `sdk.*` methods from `@metalshopping/sdk-runtime` — no raw `fetch()`**
   - Usage: All data access through SDK
   - Why: Enforces contract-driven data flow

2. **Design tokens only — no hardcoded hex values**
   - Usage: Use `$metalshopping-design-system` tokens
   - Why: Maintains visual consistency

3. **Check `packages/ui/src/index.ts` before creating any component**
   - Usage: Reuse shared components before building new ones
   - Why: Prevents duplication, maintains consistency

4. **Every data-fetching component must have loading + error + empty states**
   - Usage: Three explicit states for async operations
   - Why: Users see explicit feedback; no ambiguous UI

5. **Fetch pattern: `useEffect + cancelled flag`**
   - Usage: Standard async pattern for data fetching
   - Why: Prevents memory leaks from unmounted components

## Commit Format

```
<type>(<scope>): <what>
```

**Types:** `feat | fix | docs | chore | refactor`

**Example:**
```
feat(products): add price override endpoint
fix(auth): prevent token expiry race condition
docs(readme): update setup instructions
```

## Process Rules

1. **A task is done only when:**
   - Build passes
   - Real data verified (not just happy path)
   - One commit made

2. **One commit per completed task — no uncommitted work at session end**

3. **ADR committed only after the acceptance test passes**

## Engineering Bar

**Filter:** *"Would a Stripe or Google senior engineer approve this in code review?"*

- **Names:** Self-documenting (no cryptic abbreviations)
- **Errors:** Structured codes (`MODULE_ENTITY_REASON`)
- **Logging:** Every handler logs `trace_id`, `action`, `result`, `duration_ms`
- **Idempotency:** Every write is idempotent and retry-safe
- **Tenant safety:** No query returns cross-tenant data

## Code Organization

- **Folder structure:** `apps/` for runnable applications, `packages/` for reusable libraries
- **Module naming:** `internal/modules/<domain>` for Go modules
- **Feature naming:** `packages/feature-<surface>` for feature packages
- **Layer naming:** `domain`, `application`, `ports`, `adapters`, `transport` in consistent order

---

**Created:** 2026-03-23 | **Last Updated:** 2026-03-24 | **Weight:** 0.95

See [[hippocampus/architecture.md]] for architecture overview
