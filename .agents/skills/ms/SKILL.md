---
name: ms
description: MetalShopping master orchestrator. Use for implementation tasks. Validates architecture first, uses update_plan for complex work, asks the user to run /plan manually when needed, and orchestrates specialist skills in the correct order.
---

# MetalShopping — Master Orchestrator

Read first: `tasks/lessons.md`, `tasks/todo.md`, `AGENTS.md`.

## 1) Complexity gate

Treat the task as **complex** if any condition is true:
- touches 2+ layers (`contract`, `Go`, `worker`, `SDK`, `frontend`, `DB`, `ADR`)
- adds a new page/flow or 3+ requirements
- changes migration, manifest, seed, runtime config, queue, or worker behavior
- has architectural ambiguity or the user asks for planning

For **complex** tasks:
1. Call `update_plan` before coding.
2. Tell the user explicitly: `Task complexa; rode /plan se quiser o fluxo guiado antes da implementação.`
3. Add or refresh the relevant feature block in `tasks/todo.md` before execution.
4. Wait for approval when the task is still ambiguous or when the user asked for a plan-first flow.

For **simple** tasks:
- do a short architecture check
- skip full `tasks/todo.md` rewrite
- execute directly if the scope is clear

## 2) Architecture check before code

State only what matters:
- **module type**: `read-only` | `write+events` | `CRUD+governance` | `scraping` | `frontend-only` | `docs/process`
- **exact files/folders** to change
- **risks/invariants** that can break tenant safety, runtime behavior, or UX
- **Level 1 now vs later** when the user asks for staged delivery

Use `references/folder-patterns.md` only when folder structure is unclear.

## 3) Skill order

Use only the stages that apply:
- `T1 contract` → `$metalshopping-openapi-contracts`
- `T2/T3 implementation` → `$metalshopping-implement`
- `T4 SDK` → `$metalshopping-sdk-generation` after contract changes
- legacy-first frontend migration / visual parity baseline → `$metalshopping-legacy-migration`
- `T5 frontend` → `$metalshopping-frontend`
- `T6 ADR` → `$metalshopping-adr`
- significant feature/review gate → `$metalshopping-review`

If a stage is skipped, say why.

## 4) Operational gate

Before blaming code on tasks that touch `worker`, `manifest`, `seed`, `migration`, `runtime config`, or `DB`, verify:
- the migration/seed actually ran
- the active manifest/config in DB matches the expected version
- API and worker were restarted when runtime behavior changed
- the run/test data was created after the latest change
- tenant context and `current_tenant_id()` are correct

## 5) Execution rules

- one task at a time
- keep `tasks/todo.md` as the active source of truth for complex work
- mark a task done only after validation passes
- write a lesson to `tasks/lessons.md` after every correction
- commit after each completed task
- never edit `packages/generated/` manually

## 6) Closeout

Before declaring done:
- run the smallest relevant validation first, then broader checks if needed
- confirm the browser/runtime behavior when the task is user-facing
- ensure no uncommitted work remains for the completed task
