# tasks/todo.md
Last reset: 2026-03-23
Purpose: active execution backlog only (no historical archive).

## Execution policy (short)
- One feature block at a time.
- Only mark `[x]` after required validation + commit.
- Keep acceptance checks objective and runnable.
- Move obsolete/completed details to git history instead of growing this file.

---

# Program: Analytics Legacy Migration (Frontend-first)
Type: frontend-only now | read-only integration later | Events: no | ADR: no

## Track A — Shell + Home (baseline)
- [x] Home visual parity accepted manually
- [x] Top bar parity accepted manually
- [x] Navigation freeze issues fixed
- [x] `npm.cmd run web:build` passed in latest tranche

## Track B — Produtos
- [x] Legacy index/workspace structure copied
- [x] Navigation deadlock on tab return fixed
- [x] Spotlight table restored to legacy interactions (filters/sort/pagination)
- [x] Tooltip opacity adjusted for readability parity
- [x] Spotlight metrics mock aligned with legacy semantic keys
- [x] Spotlight table density/contrast refined toward legacy
- [ ] Visual parity audit against legacy (remaining diffs list)
- [ ] Manual acceptance: `Home -> Produtos -> Home -> Produtos` without interaction lock
- [ ] Manual acceptance: `/analytics/products` first fold matches legacy
- [ ] Commit gate for remaining parity fixes

## Track C — Classificações (Taxonomy)
- [x] Legacy page wired on `/analytics/taxonomy`
- [x] Chart runtime errors fixed (`Canvas reuse`, `linear scale`, snapshot wiring)
- [x] Card surface transparency fixes applied
- [ ] Manual acceptance: full page parity vs legacy (spacing/typography fine tuning)
- [ ] Manual acceptance: tab switching with no overlay/scroll lock
- [ ] Commit gate for any remaining CSS parity deltas

## Track D — Marca
- [x] Legacy `BrandHomePage` copied to new app
- [x] Tab `/analytics/brands` integrated in shell
- [x] Generic MVP block removed to prevent duplicate render paths
- [x] Build validation passed (`npm.cmd run web:build`)
- [ ] Manual acceptance: `/analytics/brands` parity (header, KPIs, map, table)
- [ ] Manual acceptance: no console/runtime errors while switching tabs
- [ ] Commit/close gate after manual validation

## Cross-track acceptance
- [ ] Navigation matrix passes: `Home -> Marca -> Produtos -> Classificacoes -> Home`
- [ ] No page shows blank render, blocked clicks, or stuck backdrop
- [ ] No critical console error in analytics routes

---

# Next feature queue (after Analytics parity sign-off)

## Q1 — Analytics Execução migration
- [ ] Inventory legacy source
- [ ] Define must-match surface
- [ ] Copy literal + mocks
- [ ] Integrate in shell route

## Q2 — Backend/SDK adaptation phase (only after visual sign-off)
- [ ] Define contract deltas required by migrated screens
- [ ] Implement read models/services
- [ ] Regenerate SDK
- [ ] Replace mocks incrementally without visual regressions

---

# Technical Debt — Code Quality

## Analytics Home (Post-Migration Refinement)
- [ ] **Enable TypeScript** — Remove 23 `@ts-nocheck` directives (Item 1 of 3)
  - Effort: 2-3 hours
  - Blocks: Type safety in API layer
  - Details: See `hippocampus/brain-task-execution.md` for execution checklist
  - Acceptance: `npm run web:typecheck` returns 0 errors
- [ ] Type the API layer (Item 2) — Replace `Record<string, unknown>` with discriminated unions
  - Effort: 4-6 hours
- [ ] Integrate SDK contracts (Item 3) — Use auto-generated types from OpenAPI
  - Effort: 2-3 hours

---

# Process maintenance
- [ ] Keep `tasks/lessons.md` structural-only (no local cosmetic lessons)
- [ ] Keep skills aligned with current workflow and guardrails
- [ ] Use `hippocampus/brain-workflow.md` for Brain execution file locations
- [ ] Use `hippocampus/brain-task-execution.md` for step-by-step scaffolding during `/brain-task`
