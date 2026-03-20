---
name: ms
description: MetalShopping master orchestrator. Use for any implementation task. Automatically enters plan mode, validates architecture before any code, orchestrates specialist skills in the correct order, and enforces big-tech engineering standards throughout. Say "use $ms to implement X" and this skill handles everything.
---

# MetalShopping Master Orchestrator

## This skill runs before any code is written. Always.

When activated, do not start implementing. Start thinking.

---

## Phase 1 — Architectural thinking (plan mode)

Read before anything else:
- `tasks/lessons.md` — apply every lesson
- `ARCHITECTURE.md` — frozen constraints
- `docs/PROJECT_SOT.md` — current truth

Then answer these questions explicitly in your response:

**1. What layer does this touch?**
Contract | Go module | Python worker | SDK | Frontend | Multiple

**2. What is the module type?**
- Read-only: data already in Postgres, Go reads and exposes
- Write+events: Go creates state, fires outbox event, worker consumes
- CRUD+governance: full domain model, validation, governance checks
- Scraping: Python worker ingests external data → Postgres → Go reads

**3. What is the folder structure?**

State the exact folders to be created before writing a single file.
Follow the pattern from the closest existing module:
- `modules/home/` for minimal readers
- `modules/shopping/` for write+events
- `modules/catalog/` for CRUD+governance

Example for analytics read-only module:
```
apps/server_core/internal/modules/analytics/
  ports/
    reader.go           ← Reader interface
  adapters/
    postgres/
      reader.go         ← BeginTenantTx + current_tenant_id() queries
  application/
    service.go          ← thin orchestration, no DB no HTTP
  transport/
    http/
      handler.go        ← auth + tenant + service + writeJSON
```

**4. What are the naming conventions?**
- Module package: `analytics`
- Handler struct: `Handler`
- Service struct: `Service`
- Error codes: `ANALYTICS_OVERVIEW_NOT_FOUND` (MODULE_ENTITY_REASON)
- Route: `/api/v1/analytics/overview`
- Contract file: `analytics_v1.openapi.yaml`

**5. What could go wrong architecturally?**
Identify risks before coding:
- Missing index on tenant_id + frequently queried column?
- N+1 query risk if joining multiple tables?
- Frontend needs data that requires multiple Go queries — combine in service or create readmodel?
- Event payload missing fields the worker will need later?

**6. Is this Level 1 or does it need more?**
Level 1 = real data on screen, build passes, no mocks
Level 2 = error handling, loading states, edge cases
Implement Level 1 first. State what is deferred to Level 2.

---

## Phase 2 — Write the plan to tasks/todo.md

Only after Phase 1 is complete and explicit, write tasks/todo.md.
Use the template from `skills/metalshopping-plan/references/todo-template.md`.

Include:
- Exact folder structure decided in Phase 1
- Task order: T1 contract → T2 Go → T3 worker → T4 SDK → T5 frontend
- One commit message per task
- Runnable acceptance tests (not "looks correct")

**Present the plan. Wait for approval before Phase 3.**

---

## Phase 3 — Execute (only after plan is approved)

Execute tasks in order. For each task:

**T1 Contract** → use `$metalshopping-openapi-contracts`
- Declare exact response shape decided in Phase 1
- One bounded context per file
- Reuse existing JSON schemas from `contracts/api/jsonschema/`

**T2 Go module** → use `$metalshopping-implement`
- Follow folder structure from Phase 1 exactly
- Every adapter: `pgdb.BeginTenantTx` + `current_tenant_id()`
- Every handler: PrincipalFromContext → TenantFromContext → service → writeJSON
- Write+events: `outbox.AppendInTx` inside same tx, never after Commit
- Register in `composition_modules.go`

**T3 Python worker** (only if scraping/Python libs needed)
→ use `$metalshopping-implement`
- `set_config('app.current_tenant_id', ...)` before every write tx
- `ON CONFLICT DO UPDATE` on every insert
- Never call server_core HTTP

**T4 SDK generation** → use `$metalshopping-sdk-generation`
- Run `./scripts/generate_contract_artifacts.ps1`
- Never edit `packages/generated/` manually

**T5 Frontend** → use `$metalshopping-frontend`
- Check `packages/ui/src/index.ts` before creating any component
- Data only via `@metalshopping/platform-sdk` hooks
- Design tokens only — no hardcoded values
- Loading + error + empty states required

After each task:
1. Build check: `go build ./...` or `pnpm tsc --noEmit`
2. Mark task [x] in `tasks/todo.md`
3. Commit: `git commit -m "<type>(<scope>): <what>"`

---

## Phase 4 — Review before declaring done

Run `$metalshopping-review` (targeted or full based on scope).

Quick self-check before calling review:
- [ ] Every Postgres query uses BeginTenantTx + current_tenant_id()?
- [ ] Every handler checks PrincipalFromContext + TenantFromContext?
- [ ] Outbox events inside transaction (AppendInTx before Commit)?
- [ ] No fetch() in any React component?
- [ ] Real data in browser (no mocks)?
- [ ] All acceptance tests in tasks/todo.md pass?

If all pass → ALIGNED → final commit → done.
If any fail → fix → re-run → commit.

---

## Phase 5 — Capture lessons

After any correction during execution:
- Use `$metalshopping-learn` immediately
- Write lesson to `tasks/lessons.md`
- Apply lesson for the rest of the session

---

## What this skill must never do
- Start writing code before Phase 1 is explicit and Phase 2 is approved
- Skip the folder structure decision
- Invent a new pattern when an existing module shows the correct one
- Proceed to next task without completing current task's build check and commit
- Mark [x] without real data verified and commit made

## References
- Folder patterns: `references/folder-patterns.md`
- Architecture decisions: `ARCHITECTURE.md` + `docs/adrs/`
- Go implementation patterns: `skills/metalshopping-implement/references/go-patterns.md`
- Frontend patterns: `skills/metalshopping-frontend/references/design-tokens.md`
