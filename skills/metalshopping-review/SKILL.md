---
name: metalshopping-review
description: Architecture and implementation review using 10 lenses. Covers tenancy, auth, contracts, outbox events, frontend, observability, idempotency, scalability, governance, SDK. Produces severity-ordered findings and Go/No-Go verdict. Called by $ms Phase 4 or used directly after any significant implementation.
---

# MetalShopping Review

## Quick critical check (do this first — fastest path)
- [ ] Every Postgres adapter uses `pgdb.BeginTenantTx`?
- [ ] Every query uses `current_tenant_id()` in WHERE?
- [ ] Every handler checks `PrincipalFromContext` → 401?
- [ ] Every handler checks `TenantFromContext` → 403?
- [ ] Outbox via `AppendInTx` before `Commit` (never after)?
- [ ] Worker sets `app.current_tenant_id` before every write?
- [ ] No `fetch()` in React pages/components?
- [ ] `packages/generated/` not manually edited?
- [ ] Module registered in `composition_modules.go`?

If any fail → CRITICAL finding. Stop. Fix before proceeding.

## Full 10-lens review
See `references/lenses.md` for complete checklist per lens.

## Finding format
```
CRITICAL — <title>
  File: <path>:<function>
  Rule: <ADR number or ARCHITECTURE.md principle>
  Impact: <what breaks and when>
  Fix: <specific action>
```

## Verdict
- **ALIGNED** → `git commit -m "feat(<m>): <feature> — review passed"`
- **PARTIALLY ALIGNED** → document findings, create tickets, proceed with awareness
- **MISALIGNED** → stop. Fix before any new work. Do not commit.

## References
- `references/lenses.md` — full 10-lens checklist
- `references/observability-security.md` — logging and security baseline
