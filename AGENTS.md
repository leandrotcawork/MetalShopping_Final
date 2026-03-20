# AGENTS — MetalShopping

## On every session start
1. Read `tasks/lessons.md` — non-negotiable, before any code
2. Read `tasks/todo.md` — know current state and next task
3. Use skill map below to pick the right skill

## Engineering standard
Write code a Stripe or Google senior engineer would approve in review.
This means: self-documenting names, structured errors with codes, every
handler logged, every write idempotent, no cross-tenant data leakage,
layers that don't know about each other.

## Absolute rules — violation = broken code, stop and fix

**Go**
- `pgdb.BeginTenantTx` on every Postgres adapter query. No exceptions.
- `current_tenant_id()` in every WHERE on tenant-scoped tables
- `platformauth.PrincipalFromContext` → 401 before any handler operation
- `tenancy_runtime.TenantFromContext` → 403 before any handler operation
- `outbox.AppendInTx` inside the same tx as the INSERT, never after Commit
- Register every new module in `composition_modules.go`

**Python worker**
- `set_config('app.current_tenant_id', %s, true)` before every write tx
- `ON CONFLICT ... DO UPDATE` on every insert (idempotent)
- Never call server_core HTTP endpoints

**Frontend**
- Data only via `@metalshopping/platform-sdk` hooks — no `fetch()`
- Design tokens only — no hardcoded hex, px font-size, or spacing
- Check `packages/ui/src/index.ts` before creating any component
- Loading + error + empty state on every data-fetching component

**Process**
- No task marked [x] without: build passes + real data + commit made
- ADR done only when acceptance test passes and commit is made
- `packages/generated/` never edited manually

## Skill map

| Need | Skill |
|---|---|
| Plan a feature | `$metalshopping-plan` |
| Implement end-to-end | `$metalshopping-implement` |
| Review architecture | `$metalshopping-review` |
| Learn from correction | `$metalshopping-learn` |
| Frontend page/component | `$metalshopping-frontend` |
| ADR lifecycle | `$metalshopping-adr` |
| OpenAPI contract | `$metalshopping-openapi-contracts` |
| SDK generation | `$metalshopping-sdk-generation` |
| Governance/events/platform | existing specialist skills |

## Commit format
`<type>(<scope>): <what>` — feat | fix | docs | chore | refactor
One commit per completed task. No uncommitted work at session end.
