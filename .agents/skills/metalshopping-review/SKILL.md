---
name: metalshopping-review
description: "Architecture and implementation review for MetalShopping. Two levels: (1) compliance — does code follow current patterns, (2) quality — are the patterns themselves professional and scalable at big-tech level. Always runs both. Produces severity-ordered findings and verdict."
---

# MetalShopping Review

## Two levels — always run both

**Level 1 — Compliance:** does the code follow current repo patterns?
**Level 2 — Quality:** are the patterns themselves professional and scalable?

A senior engineer at Stripe or Google runs both.
Compliance without quality is just consistent mediocrity.

---

## Level 1 — Compliance (quick check first)

If any fail → CRITICAL → stop:
- [ ] Every Postgres adapter uses `pgdb.BeginTenantTx`?
- [ ] Every query uses `current_tenant_id()` in WHERE?
- [ ] Every handler checks `PrincipalFromContext` → 401?
- [ ] Every handler checks `TenantFromContext` → 403?
- [ ] Outbox via `AppendInTx` before `Commit`?
- [ ] Worker sets `app.current_tenant_id` before every write?
- [ ] No `fetch()` in React pages/components?
- [ ] `packages/generated/` not manually edited?
- [ ] Module registered in `composition_modules.go`?

Full compliance lenses: `references/compliance-lenses.md`

---

## Level 2 — Quality (are the patterns themselves good?)

For each item, flag if the code being reviewed introduces or perpetuates the weak pattern.

### Q1 — Duplicate platform helpers
Does this module define its own `writeJSON`, `writeError`, or error envelope struct?
- **Problem:** `writeJSON` is copied in 7 modules. One change = 7 files to edit.
- **Target:** `internal/platform/httputil/` — one place, used everywhere.
- **Flag:** any new `func writeJSON` or `func write*Error` in a transport package.

### Q2 — Inconsistent error format
Does this module use a different error struct than other modules?
- **Problem:** `homeErrorEnvelope`, `apiErrorEnvelope`, `writeShoppingError` are all different shapes. Frontend can't parse errors generically.
- **Target:** one error format everywhere: `{"error":{"code":"...","message":"...","details":{},"trace_id":"..."}}`.
- **Flag:** any new error envelope type that isn't the shared platform one.

### Q3 — Untyped response shapes
Does the handler use `map[string]any` inline in `writeJSON`?
- **Problem:** 16 occurrences in the repo. Compiler never catches shape drift vs OpenAPI.
- **Target:** typed response structs in `transport/http/` — compiler-verified, contract-aligned.
- **Flag:** `writeJSON(w, 200, map[string]any{...})` in any handler.

### Q4 — Handler imports infrastructure directly
Does the handler import anything from `adapters/` or `adapters/postgres/`?
- **Problem:** `shopping` handler imports `adapters/postgres` directly — breaks layer isolation.
- **Target:** handler knows only `application.Service`. Zero knowledge of infrastructure.
- **Flag:** any `import "...adapters/postgres"` inside a `transport/` package.

### Q5 — No unit tests
Does this module have any `*_test.go` files?
- **Problem:** zero unit tests across all modules. Refactoring is invisible to the build.
- **Target:** every domain invariant and service use case tested with a memory repo.
- **Flag:** any new feature shipped without at least one unit test.

### Q6 — No memory test double
Does this module have an in-memory repo implementation?
- **Problem:** every test requires a live Postgres. CI is slow and brittle.
- **Target:** every postgres repo has a parallel `memory/` implementation for unit tests.
- **Flag:** new module with only a postgres adapter and no memory double.

### Q7 — Missing pagination
Does any list endpoint return unbounded results?
- **Problem:** `handleListProducts` returns all products with no limit. Breaks at 10k rows.
- **Target:** every list endpoint has `limit` + `offset` with a documented max page size.
- **Flag:** any `SELECT ... FROM table WHERE tenant_id = current_tenant_id()` without `LIMIT`.

### Q8 — Architecture layer direction
Is this code moving toward or away from clean separation?
- **Current state:** `pgdb.BeginTenantTx` is coupled into adapter layer, making domain testing hard.
- **Better direction** (ref: MetalDocs): `domain/` is pure Go with no IO. Infrastructure implements domain interfaces. Delivery only translates HTTP ↔ service commands.
- **Flag:** new code that deepens the coupling (e.g. domain importing platform packages, handler containing business conditions).

---

## Finding format

```
CRITICAL — <title>
  File: <path>:<function>
  Rule: <ADR or architectural principle>
  Impact: <what breaks and when>
  Fix: <specific action required now>

QUALITY — <title>
  File: <path>:<function>
  Pattern: <what weak pattern this perpetuates>
  Target: <what professional code looks like here>
  Priority: fix now | next cycle | track as debt
```

## Verdict

- **ALIGNED** — compliant + no new quality debt introduced → commit
- **COMPLIANT BUT WEAK** — passes compliance, adds quality issues → commit + log to `tasks/quality-debt.md`
- **MISALIGNED** — compliance failures → stop, fix, do not commit

## References
- `references/compliance-lenses.md` — full 10-lens compliance checklist
- `tasks/quality-debt.md` — running list of known quality issues to address
