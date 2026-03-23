# AGENTS — MetalShopping

## On every session start
1. Read `tasks/lessons.md` — apply every rule before touching code
2. Read `tasks/todo.md` — know current state
3. After any correction: write lesson to `tasks/lessons.md` immediately

## Engineering bar
Every decision passes this filter:
*"Would a Stripe or Google senior engineer approve this in code review?"*
- Names are self-documenting — no comment needed to understand them
- Errors carry structured codes: `MODULE_ENTITY_REASON`
- Every handler logs `trace_id`, `action`, `result`, `duration_ms`
- Every write is idempotent and retry-safe
- No query ever returns cross-tenant data

## Absolute rules — violation = stop and fix immediately

**Go**
- `pgdb.BeginTenantTx` on every Postgres adapter query — no exceptions
- `current_tenant_id()` in every WHERE on tenant-scoped tables
- `platformauth.PrincipalFromContext` → 401 before any handler operation
- `tenancy_runtime.TenantFromContext` → 403 before any handler operation
- `outbox.AppendInTx` inside the same tx as INSERT — never after Commit
- Every new module registered in `composition_modules.go`

**Python worker**
- `set_config('app.current_tenant_id', %s, true)` before every write tx
- `ON CONFLICT ... DO UPDATE` on every insert
- Never call server_core HTTP endpoints

**Frontend**
- Data only via `@metalshopping/platform-sdk` hooks — no `fetch()`
- Design tokens only — no hardcoded hex, px font-size, or spacing
- Check `packages/ui/src/index.ts` before creating any component
- Loading + error + empty state on every data-fetching component

**Process**
- No task marked [x] without: build passes + real data verified + commit made
- ADR done only when acceptance test passes and commit is made
- `packages/generated/` never edited manually
- One commit per completed task — no uncommitted work at session end

## Skill map

| Task | Skill |
|---|---|
| Any implementation (default) | `$ms` |
| OpenAPI contract | `$metalshopping-openapi-contracts` |
| Event contract | `$metalshopping-event-contracts` |
| Governance contract | `$metalshopping-governance-contracts` |
| SDK generation | `$metalshopping-sdk-generation` |
| Legacy-first visual migration | `$metalshopping-legacy-migration` |
| ADR lifecycle | `$metalshopping-adr` |

## Lesson format (write to tasks/lessons.md after every correction)
```
## Lesson N — <title>
Date: YYYY-MM-DD | Trigger: <correction | review | build failure>
Wrong:   <exact code or decision>
Correct: <exact code or decision>
Rule:    <one sentence>
Layer:   <Go adapter | handler | worker | frontend | process>
```

## Commit format
`<type>(<scope>): <what>` — feat | fix | docs | chore | refactor
