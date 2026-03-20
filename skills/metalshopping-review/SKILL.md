---
name: metalshopping-review
description: Architecture and implementation review against MetalShopping standards. 10 lenses covering tenancy, auth, contracts, outbox, frontend, observability, idempotency, scalability, governance, SDK. Produces severity-ordered findings and Go/No-Go verdict. Use after any significant implementation or before declaring a feature done.
---

# MetalShopping Review

## When to run
- After any module implementation (before declaring done)
- When something feels wrong but root cause is unclear
- Any change to platform, auth, or tenancy packages

## Workflow
1. Read relevant ADRs + `docs/PROJECT_SOT.md` (targeted, not full repo scan)
2. Apply all 10 lenses — see `references/lenses.md`
3. Output findings by severity: Critical → High → Medium → Low
4. Issue verdict: ALIGNED | PARTIALLY ALIGNED | MISALIGNED

## Finding format
```
CRITICAL — <title>
  File: <path>:<function>
  Rule: <ADR or ARCHITECTURE.md principle>
  Impact: <what breaks and when>
  Fix: <specific action>
```

## Verdict and next action
- **ALIGNED** → commit: `"feat(<m>): <feature> — review passed"`
- **PARTIALLY ALIGNED** → create tickets, document, proceed with awareness
- **MISALIGNED** → stop. Fix before any new work. Do not commit.

## Quick critical checklist (fastest pass)
- [ ] Every Postgres adapter uses `pgdb.BeginTenantTx`?
- [ ] Every query uses `current_tenant_id()` in WHERE?
- [ ] Every handler checks `PrincipalFromContext` → 401?
- [ ] Every handler checks `TenantFromContext` → 403?
- [ ] Outbox events inside tx via `AppendInTx`?
- [ ] Worker sets `app.current_tenant_id` before every write?
- [ ] No `fetch()` in any React page/component?
- [ ] `packages/generated/` not manually edited?
- [ ] Module registered in `composition_modules.go`?

If any fails: CRITICAL finding. Stop.

## References
- `references/lenses.md` — full 10-lens checklist
