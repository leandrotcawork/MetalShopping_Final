# Governance Documentation Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate MetalShopping's documentation governance so AI agents and humans operate from one explicit operational SoT, one stable architecture document, aligned agent entrypoints, and a clean execution/progress layer.

**Architecture:** Execute this as a documentation-hardening tranche in a dedicated worktree. First restore the missing repository control files (`tasks/`). Then make `docs/PROJECT_SOT.md` the explicit operational root, align `AGENTS.md`, `CLAUDE.md`, and `CODEX.md` to that hierarchy, and finally bring `ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md` into the same governance model.

**Tech Stack:** Markdown, PowerShell, git, ripgrep

**Specs:**
- `docs/superpowers/specs/2026-04-01-governance-audit-design.md`

---

## File Map

### Create
- `tasks/lessons.md`
- `tasks/todo.md`
- `CODEX.md`

### Modify
- `AGENTS.md`
- `CLAUDE.md`
- `ARCHITECTURE.md`
- `docs/PROJECT_SOT.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`

---

### Task 1: Restore the missing repository control files

**Files:**
- Create: `tasks/lessons.md`
- Create: `tasks/todo.md`

- [ ] **Step 1: Verify the `tasks/` directory is currently missing**

Run:
```powershell
Test-Path tasks
```

Expected: `False`

- [ ] **Step 2: Create the `tasks/` directory**

Run:
```powershell
New-Item -ItemType Directory -Path tasks
```

Expected: directory creation output for `tasks`

- [ ] **Step 3: Create `tasks/lessons.md` with the canonical lesson scaffold**

Write this file content exactly:

````markdown
# Lessons Learned

Use this file only for recurring engineering corrections that would cause bugs, review failures, or structural regressions if repeated.

## Rules

- Write a lesson immediately after any meaningful correction.
- Do not log one-off UI tweaks or cosmetic feedback.
- Keep each lesson concrete and reusable.

## Lesson Template

```text
## Lesson N — <title>
Date: YYYY-MM-DD | Trigger: <correction | review | build failure>
Wrong:   <exact code or decision>
Correct: <exact code or decision>
Rule:    <one sentence>
Layer:   <Go adapter | handler | worker | frontend | process | docs>
```
````

- [ ] **Step 4: Create `tasks/todo.md` with the governance tranche initialized**

Write this file content exactly:

```markdown
# Task Tracker

## Current State

- State: in-progress
- Active tranche: governance documentation consolidation
- Source spec: `docs/superpowers/specs/2026-04-01-governance-audit-design.md`

## Active Tasks

- [ ] Bootstrap repository control files (`tasks/lessons.md`, `tasks/todo.md`)
- [ ] Consolidate `docs/PROJECT_SOT.md` as the operational SoT
- [ ] Align `AGENTS.md`, `CLAUDE.md`, and `CODEX.md`
- [ ] Align `ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md`
- [ ] Verify documentation hierarchy and cross-file consistency

## Notes

- `fulldocs` is the strategic input source for this governance restructuring.
- Repository execution must follow the accepted SoTs inside this repository once consolidated.
```

- [ ] **Step 5: Verify both files exist and contain the expected headings**

Run:
```powershell
rg --no-heading -n "^# " tasks/lessons.md tasks/todo.md
```

Expected:
- `tasks/lessons.md:1:# Lessons Learned`
- `tasks/todo.md:1:# Task Tracker`

- [ ] **Step 6: Commit**

Run:
```powershell
git add tasks/lessons.md tasks/todo.md
git commit -m "docs(governance): restore task tracker and lessons files"
```

Expected: a successful commit containing only the two new `tasks/` files

---

### Task 2: Make `docs/PROJECT_SOT.md` the explicit operational root

**Files:**
- Modify: `docs/PROJECT_SOT.md`

- [ ] **Step 1: Read the current `docs/PROJECT_SOT.md`**

Run:
```powershell
Get-Content -Raw docs/PROJECT_SOT.md
```

Expected: the current operational SoT content, including `Purpose`, `Current state`, `Product identity`, and `Planning constraints`

- [ ] **Step 2: Replace the `Purpose` section with explicit operational ownership**

Replace the current `## Purpose` section with this exact content:

```markdown
## Purpose

This document is the operational source of truth for MetalShopping.

It defines:

- what MetalShopping is building
- which documentation governs the repository
- which document wins in case of conflict
- how strategy is consolidated into repository truth
- which execution model the agents and humans must follow
```

- [ ] **Step 3: Insert the new documentation-governance section immediately after `## Platform direction`**

Insert this exact section:

```markdown
## Documentation hierarchy and precedence

MetalShopping uses a strict documentation hierarchy so humans and AI agents do not operate from competing interpretations.

### Precedence order

1. `docs/PROJECT_SOT.md`
2. `ARCHITECTURE.md`
3. `AGENTS.md`
4. `CLAUDE.md` and `CODEX.md`
5. `docs/IMPLEMENTATION_PLAN.md`
6. `docs/PROGRESS.md`
7. focused specs and plans under `docs/superpowers/`

### Strategic source during this restructuring

For the governance restructuring tranche, `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\fulldocs` is the accepted strategic input source.

That strategic content does not override the repository at runtime automatically. Once a decision is accepted and consolidated here, repository execution follows the repository SoTs, not `fulldocs`.

### Document ownership

- `docs/PROJECT_SOT.md`: operational truth, hierarchy, precedence, execution direction
- `ARCHITECTURE.md`: stable architecture thesis and long-lived technical boundaries
- `AGENTS.md`: mandatory repository-level entry rules for all agents
- `CLAUDE.md` and `CODEX.md`: agent-specific instructions aligned to the same project truth
- `docs/IMPLEMENTATION_PLAN.md`: macro phase and sequence planning
- `docs/PROGRESS.md`: factual status, next gates, and blockers

### Conflict rule

If two documents disagree, the higher-precedence document wins. If the disagreement is architectural and durable, resolve it through an ADR and then update the affected SoT documents.
```

- [ ] **Step 4: Insert the new update protocol section immediately after `## Planning constraints`**

Insert this exact section:

```markdown
## Documentation update protocol

Use this update flow to prevent drift:

1. New strategic insight starts in `fulldocs` or structured planning discussion.
2. Accepted operational truth is consolidated into `docs/PROJECT_SOT.md`.
3. Stable architectural changes are frozen through ADRs and reflected in `ARCHITECTURE.md` when needed.
4. Macro execution order changes update `docs/IMPLEMENTATION_PLAN.md`.
5. Status and delivery movement update `docs/PROGRESS.md`.
6. Agent behavior changes update `AGENTS.md`, `CLAUDE.md`, and `CODEX.md`.

No document should duplicate another document in full. Each document should point to the canonical owner of its subject.
```

- [ ] **Step 5: Add governance-first constraints to the existing `## Planning constraints` list**

Append these exact bullets to the `## Planning constraints` section:

```markdown
- Governance entrypoint files must always require reading `docs/PROJECT_SOT.md`
- Do not let agent-specific files become competing sources of project truth
- Strategy may originate outside the repository, but runtime execution must follow the accepted repository SoTs
- Plan the project in this order: governance -> orchestration -> one module spec -> one implementation plan
```

- [ ] **Step 6: Verify the new SoT hierarchy and protocol are present**

Run:
```powershell
rg --no-heading -n "Documentation hierarchy and precedence|Documentation update protocol|Conflict rule|Governance entrypoint files" docs/PROJECT_SOT.md
```

Expected: four matches in `docs/PROJECT_SOT.md`

- [ ] **Step 7: Commit**

Run:
```powershell
git add docs/PROJECT_SOT.md
git commit -m "docs(governance): promote PROJECT_SOT to operational root"
```

Expected: a successful commit containing only the `docs/PROJECT_SOT.md` changes

---

### Task 3: Rewrite `AGENTS.md` as the repository-wide mandatory entrypoint

**Files:**
- Modify: `AGENTS.md`

- [ ] **Step 1: Read the current `AGENTS.md`**

Run:
```powershell
Get-Content -Raw AGENTS.md
```

Expected: the current repository rules with duplicated session-start behavior and absolute rules

- [ ] **Step 2: Replace the entire file with the new concise repository entrypoint**

Write this file content exactly:

```markdown
# AGENTS — MetalShopping

## Purpose

This file is the repository-wide mandatory entrypoint for any AI agent working in MetalShopping.

## Mandatory read order

1. Read `docs/PROJECT_SOT.md`
2. Read `ARCHITECTURE.md`
3. Read the agent-specific file (`CLAUDE.md` or `CODEX.md`)
4. Read `tasks/todo.md`
5. Read `tasks/lessons.md`

Do not start planning, implementation, or review before completing that read order.

## Documentation precedence

If documents conflict, follow this order:

1. `docs/PROJECT_SOT.md`
2. `ARCHITECTURE.md`
3. `AGENTS.md`
4. `CLAUDE.md` or `CODEX.md`
5. `docs/IMPLEMENTATION_PLAN.md`
6. `docs/PROGRESS.md`

## Product objective

MetalShopping is a long-lived enterprise platform for commercial strategy, pricing, procurement, CRM, analytics, automation, and future AI-assisted operations.

Do not optimize for one-off delivery. Always preserve future module growth, clean boundaries, governance, and multi-tenant safety.

## Absolute rules

### Go
- `pgdb.BeginTenantTx` on every Postgres adapter query
- `current_tenant_id()` in every WHERE on tenant-scoped tables
- `platformauth.PrincipalFromContext` -> 401 before handler work
- `tenancy_runtime.TenantFromContext` -> 403 before handler work
- `outbox.AppendInTx` inside the same transaction as the write
- every new module registered in `composition_modules.go`

### Python worker
- `set_config('app.current_tenant_id', %s, true)` before every write transaction
- `ON CONFLICT ... DO UPDATE` on every insert
- never call `server_core` HTTP endpoints directly

### Frontend
- data only through `sdk.*` from `@metalshopping/sdk-runtime`
- design tokens only, no hardcoded hex values
- check `packages/ui/src/index.ts` before creating a component
- every data-fetching surface has loading, error, and empty states
- fetch pattern is `useEffect + cancelled flag`

### Process
- `packages/generated/` and `packages/generated-types/` are never edited manually
- no task is done without verification and a commit
- no architectural rule changes without updating SoT documents and ADRs when needed

## Engineering bar

Every decision must survive the question:

`Would a Stripe or Google senior engineer approve this in code review?`

## Skill map

| Task | Skill |
|---|---|
| Any implementation (default) | `$ms` |
| OpenAPI contract | `$metalshopping-openapi-contracts` |
| Event contract | `$metalshopping-event-contracts` |
| Governance contract | `$metalshopping-governance-contracts` |
| SDK generation | `$metalshopping-sdk-generation` |
| ADR lifecycle | `$metalshopping-adr` |
| Frontend visual/component work | `$metalshopping-design-system` |
```

- [ ] **Step 3: Verify the new mandatory read order and precedence**

Run:
```powershell
rg --no-heading -n "Mandatory read order|Documentation precedence|docs/PROJECT_SOT.md|CODEX.md" AGENTS.md
```

Expected: matches for all four phrases in `AGENTS.md`

- [ ] **Step 4: Commit**

Run:
```powershell
git add AGENTS.md
git commit -m "docs(governance): rewrite AGENTS as mandatory repo entrypoint"
```

Expected: a successful commit containing only the `AGENTS.md` rewrite

---

### Task 4: Align `CLAUDE.md` and create `CODEX.md`

**Files:**
- Modify: `CLAUDE.md`
- Create: `CODEX.md`

- [ ] **Step 1: Read the current `CLAUDE.md`**

Run:
```powershell
Get-Content -Raw CLAUDE.md
```

Expected: the current long Claude-specific file, including the incorrect `docs/ARCHITECTURE.md` reference

- [ ] **Step 2: Replace `CLAUDE.md` with the aligned agent entrypoint**

Write this file content exactly:

````markdown
# CLAUDE.md

This file provides guidance to Claude when working in this repository.

## Mandatory startup

Before any planning, implementation, or review:

1. Read `docs/PROJECT_SOT.md`
2. Read `ARCHITECTURE.md`
3. Read `AGENTS.md`
4. Read `tasks/todo.md`
5. Read `tasks/lessons.md`

If documents conflict, follow `docs/PROJECT_SOT.md` first, then `ARCHITECTURE.md`, then `AGENTS.md`.

## Scope

Use this file for Claude-specific operating guidance only. Do not treat it as a competing source of project truth.

## Commands

### Web
```powershell
npm run web:typecheck
npm run web:build
npm run web:test
npm --workspace @metalshopping/web run dev
```

### Backend
```powershell
go test ./apps/server_core/...
go test ./apps/server_core/internal/modules/<module>/...
```

### Contracts
```powershell
./scripts/generate_contract_artifacts.ps1 -Target all
./scripts/validate_contracts.ps1 -Scope all
```

## Working rules

- Prefer updating repository SoTs instead of creating duplicate planning files.
- Never edit `packages/generated/` or `packages/generated-types/` manually.
- Keep architecture changes in ADRs and SoT documents, not only in agent notes.
- Use the skill map and workflow defined by `AGENTS.md`.

## Key references

- `docs/PROJECT_SOT.md`
- `ARCHITECTURE.md`
- `AGENTS.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- `docs/adrs/`
```
````

- [ ] **Step 3: Create `CODEX.md` as the Codex-aligned twin of `CLAUDE.md`**

Write this file content exactly:

````markdown
# CODEX.md

This file provides guidance to Codex when working in this repository.

## Mandatory startup

Before any planning, implementation, or review:

1. Read `docs/PROJECT_SOT.md`
2. Read `ARCHITECTURE.md`
3. Read `AGENTS.md`
4. Read `tasks/todo.md`
5. Read `tasks/lessons.md`

If documents conflict, follow `docs/PROJECT_SOT.md` first, then `ARCHITECTURE.md`, then `AGENTS.md`.

## Scope

Use this file for Codex-specific operating guidance only. Do not treat it as a competing source of project truth.

## Commands

### Web
```powershell
npm run web:typecheck
npm run web:build
npm run web:test
npm --workspace @metalshopping/web run dev
```

### Backend
```powershell
go test ./apps/server_core/...
go test ./apps/server_core/internal/modules/<module>/...
```

### Contracts
```powershell
./scripts/generate_contract_artifacts.ps1 -Target all
./scripts/validate_contracts.ps1 -Scope all
```

## Working rules

- Prefer updating repository SoTs instead of creating duplicate planning files.
- Never edit `packages/generated/` or `packages/generated-types/` manually.
- Keep architecture changes in ADRs and SoT documents, not only in agent notes.
- Use the skill map and workflow defined by `AGENTS.md`.

## Key references

- `docs/PROJECT_SOT.md`
- `ARCHITECTURE.md`
- `AGENTS.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- `docs/adrs/`
```
````

- [ ] **Step 4: Verify both files reference the correct architecture path and startup order**

Run:
```powershell
rg --no-heading -n "ARCHITECTURE.md|Mandatory startup|docs/PROJECT_SOT.md|AGENTS.md" CLAUDE.md CODEX.md
```

Expected:
- `CLAUDE.md` contains all four phrases
- `CODEX.md` contains all four phrases
- no `docs/ARCHITECTURE.md` reference remains

- [ ] **Step 5: Commit**

Run:
```powershell
git add CLAUDE.md CODEX.md
git commit -m "docs(governance): align Claude and Codex entrypoint files"
```

Expected: a successful commit containing the `CLAUDE.md` rewrite and new `CODEX.md`

---

### Task 5: Keep architecture architectural and align execution docs

**Files:**
- Modify: `ARCHITECTURE.md`
- Modify: `docs/IMPLEMENTATION_PLAN.md`
- Modify: `docs/PROGRESS.md`

- [ ] **Step 1: Insert the document-boundary section near the top of `ARCHITECTURE.md`**

Add this exact section immediately after `## Status`:

```markdown
## Document boundary

This file owns the stable architecture thesis of MetalShopping.

It does not own:

- day-to-day status tracking
- active execution order details
- task progress or backlog state
- agent precedence rules

Those belong to `docs/PROJECT_SOT.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md` according to the repository governance hierarchy.
```

- [ ] **Step 2: Update `docs/IMPLEMENTATION_PLAN.md` to include the governance-first tranche**

Insert this exact section immediately after `## Goal`:

```markdown
## Governance-first execution rule

Before expanding new module planning, MetalShopping must keep this sequence:

1. governance consolidation
2. master orchestration planning
3. one module spec at a time
4. one implementation plan at a time

This prevents the project from producing deep implementation plans on top of drifting operating rules.
```

Then append this exact bullet under `## Ongoing workstreams`:

```markdown
- governance documentation consolidation and master orchestration planning
```

- [ ] **Step 3: Update `docs/PROGRESS.md` with the governance hardening movement**

Append this exact bullet at the end of the `## Done` list:

```markdown
- governance audit design spec added to consolidate documentation precedence, agent entrypoints, and SoT ownership before the next planning wave
```

Insert these exact bullets at the top of the `## Next` list:

```markdown
- consolidate `PROJECT_SOT`, `AGENTS`, `CLAUDE`, `CODEX`, `ARCHITECTURE`, `IMPLEMENTATION_PLAN`, and `PROGRESS` under the accepted governance hierarchy
- create the master orchestration plan only after the governance tranche is fully aligned
```

- [ ] **Step 4: Verify architectural and execution boundaries**

Run:
```powershell
rg --no-heading -n "Document boundary|governance-first execution rule|governance documentation consolidation" ARCHITECTURE.md docs/IMPLEMENTATION_PLAN.md docs/PROGRESS.md
```

Expected:
- one match in `ARCHITECTURE.md`
- one match in `docs/IMPLEMENTATION_PLAN.md`
- one match in `docs/PROGRESS.md`

- [ ] **Step 5: Commit**

Run:
```powershell
git add ARCHITECTURE.md docs/IMPLEMENTATION_PLAN.md docs/PROGRESS.md
git commit -m "docs(governance): align architecture and execution docs"
```

Expected: a successful commit containing only those three document updates

---

### Task 6: Run the documentation consistency gate

**Files:** verification only

- [ ] **Step 1: Confirm all governance entrypoints exist**

Run:
```powershell
Test-Path AGENTS.md
Test-Path CLAUDE.md
Test-Path CODEX.md
Test-Path docs/PROJECT_SOT.md
Test-Path ARCHITECTURE.md
Test-Path tasks/lessons.md
Test-Path tasks/todo.md
```

Expected: seven lines of `True`

- [ ] **Step 2: Verify the precedence rule appears in all required entrypoints**

Run:
```powershell
rg --no-heading -n "docs/PROJECT_SOT.md|Documentation precedence|Mandatory startup|Mandatory read order" AGENTS.md CLAUDE.md CODEX.md docs/PROJECT_SOT.md
```

Expected: matches in all four files

- [ ] **Step 3: Verify the incorrect architecture path no longer exists**

Run:
```powershell
rg --no-heading -n "docs/ARCHITECTURE.md" .
```

Expected: no matches

- [ ] **Step 4: Review the final doc-only diff**

Run:
```powershell
git diff --stat HEAD~4..HEAD
```

Expected: changes only in:
- `tasks/lessons.md`
- `tasks/todo.md`
- `AGENTS.md`
- `CLAUDE.md`
- `CODEX.md`
- `ARCHITECTURE.md`
- `docs/PROJECT_SOT.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`

- [ ] **Step 5: Final commit**

Run:
```powershell
git commit --allow-empty -m "docs(governance): verify documentation hierarchy alignment"
```

Expected: either a real verification commit if extra notes were added, or an empty audit commit marking the consistency gate as complete

---

## Summary

| Task | Purpose |
|------|---------|
| 1 | Restore `tasks/` control files so repository startup rules are executable |
| 2 | Make `docs/PROJECT_SOT.md` the explicit operational root |
| 3 | Turn `AGENTS.md` into the mandatory repository entrypoint |
| 4 | Align `CLAUDE.md` and add `CODEX.md` |
| 5 | Keep `ARCHITECTURE.md` architectural and align execution docs |
| 6 | Run a consistency gate across the final governance layer |
