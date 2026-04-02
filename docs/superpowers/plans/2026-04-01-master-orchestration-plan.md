# Master Orchestration Plan Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the repository-level master orchestration document for MetalShopping, align core SoTs to recognize it, and update progress tracking so the next front can be selected from one live coordination index.

**Architecture:** Implement this as a documentation tranche centered on a new canonical file in `docs/MASTER_ORCHESTRATION_PLAN.md`. First create the orchestration document with the approved hybrid structure and normalized front template. Then align `docs/PROJECT_SOT.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md` so the new artifact is visible in the repository hierarchy and current execution state. Finish by updating the active tracker and running a consistency gate across all affected planning docs.

**Tech Stack:** Markdown, PowerShell, git, ripgrep

**Specs:**
- `docs/superpowers/specs/2026-04-01-master-orchestration-plan-design.md`

---

## File Map

### Create
- `docs/MASTER_ORCHESTRATION_PLAN.md`

### Modify
- `docs/PROJECT_SOT.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- `tasks/todo.md`

---

### Task 1: Create the canonical master orchestration document

**Files:**
- Create: `docs/MASTER_ORCHESTRATION_PLAN.md`

- [ ] **Step 1: Verify the orchestration doc does not already exist**

Run:
```powershell
Test-Path docs/MASTER_ORCHESTRATION_PLAN.md
```

Expected: `False`

- [ ] **Step 2: Create `docs/MASTER_ORCHESTRATION_PLAN.md` with the approved top-level structure**

Write this file content exactly:

````markdown
# Master Orchestration Plan

## Purpose

This document is the live execution index for MetalShopping.

It maps the major product fronts and transversal fronts, records their current status, shows cross-front dependencies, and identifies which front should open the next detailed spec.

This document does not replace `docs/PROJECT_SOT.md`, `ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, or `docs/PROGRESS.md`.

## How to use this document

- Use this document to decide which front should open the next detailed spec.
- Use this document to understand dependency direction between product and transversal fronts.
- Use this document to see which fronts are done, in progress, ready for spec, waiting on dependency, or blocked.
- Do not use this document as a step-by-step implementation plan.
- Do not restate full architecture or full progress evidence here; link to the canonical owner instead.

## Execution statuses

- `done`: the front already has its current tranche completed and does not need orchestration attention now
- `in progress`: the front is actively being specified or implemented
- `ready for spec`: the front is the next acceptable candidate to open a detailed spec
- `waiting on dependency`: the front is valid but should not open yet because another front must move first
- `blocked`: the front cannot move because a prerequisite is missing or unresolved

## Product fronts

### Home

- `Objective`: own the top-level operational dashboard and executive summary surfaces
- `Current status`: `done`
- `Why it matters now`: proves the thin-client delivery path and anchors future cross-module navigation
- `Depends on`: contracts, sdk generation, frontend migration guardrails
- `Unblocks`: future home expansions and shared dashboard conventions
- `Existing artifacts`: `docs/PROGRESS.md`, `docs/HOME_LEVEL1_ACCEPTANCE.md`
- `Next artifact to create`: none until a new Home tranche is explicitly opened

### Shopping

- `Objective`: own shopping intelligence, supplier price capture, operator review flows, and sourcing support
- `Current status`: `in progress`
- `Why it matters now`: already has active backend, worker, and frontend movement and still carries important parity and driver follow-up work
- `Depends on`: procurement boundaries, workers/integrations, frontend migration, contracts
- `Unblocks`: procurement signal quality, future sourcing workflows, analytics consumption of shopping outputs
- `Existing artifacts`: `docs/SHOPPING_LEVEL1_ACCEPTANCE.md`, `docs/SHOPPING_DRIVER_SUITE_ACCEPTANCE.md`, `docs/adrs/ADR-0021-frontend-migration-closure.md`
- `Next artifact to create`: the next approved Shopping spec after the current ADR-driven parity and supplier follow-up is chosen

### Analytics

- `Objective`: own analytical surfaces, intelligence layers, decision support, and future AI-assisted operator insight
- `Current status`: `ready for spec`
- `Why it matters now`: it is one of the main product fronts in the accepted module order and requires orchestration before deep planning opens
- `Depends on`: contracts, governance, read models, frontend migration, sdk generation
- `Unblocks`: analytics surfaces, campaigns, intelligence, and AI-adjacent operator workflows
- `Existing artifacts`: `.agents/skills/analytics-orchestrator/SKILL.md`, `.agents/skills/analytics-ai/SKILL.md`, `.agents/skills/analytics-campaigns/SKILL.md`, `.agents/skills/analytics-intelligence/SKILL.md`, `.agents/skills/analytics-surfaces/SKILL.md`
- `Next artifact to create`: analytics master spec

### CRM

- `Objective`: own customer relationship, operator follow-up flows, and future commercial action surfaces
- `Current status`: `waiting on dependency`
- `Why it matters now`: it is in the accepted module order but depends on upstream product and platform decisions to avoid shallow planning
- `Depends on`: auth/session, events, frontend migration, analytics and shopping signal clarity
- `Unblocks`: customer workflows, commercial follow-up, future automation and campaign integration
- `Existing artifacts`: `docs/PROGRESS.md`, `ARCHITECTURE.md`
- `Next artifact to create`: CRM master spec after analytics and upstream dependency review

### Catalog

- `Objective`: own canonical product identity, taxonomy, identifiers, and shared product master data
- `Current status`: `done`
- `Why it matters now`: it is already the canonical product foundation and remains a dependency for several downstream fronts
- `Depends on`: governance, contracts
- `Unblocks`: pricing, inventory, procurement, analytics, CRM
- `Existing artifacts`: `docs/CATALOG_CANONICAL_MODEL.md`, `docs/SKU_CANONICAL_DATA_MODEL.md`
- `Next artifact to create`: none until a new catalog expansion tranche is explicitly chosen

### Pricing

- `Objective`: own price semantics, commercial calculations, and pricing write/read flows
- `Current status`: `in progress`
- `Why it matters now`: semantics are being realigned and still require validation and follow-up migration work
- `Depends on`: catalog, governance, contracts, outbox discipline
- `Unblocks`: procurement, analytics, CRM, commercial decision support
- `Existing artifacts`: `docs/PRICING_CANONICAL_MODEL.md`, `docs/PRICING_IMPLEMENTATION_PLAN.md`, `docs/PRICING_READINESS_REVIEW.md`
- `Next artifact to create`: a focused follow-up pricing spec only if the remaining semantic or migration work cannot stay under existing artifacts

### Inventory

- `Objective`: own live stock position and inventory timing semantics
- `Current status`: `in progress`
- `Why it matters now`: it is already implemented at the first slice and remains an input to procurement and analytics
- `Depends on`: catalog, contracts, governance
- `Unblocks`: procurement, analytics, shopping context quality
- `Existing artifacts`: `docs/INVENTORY_CANONICAL_MODEL.md`, `docs/PROGRESS.md`
- `Next artifact to create`: inventory follow-up spec only when the next inventory tranche is explicitly selected

### Procurement

- `Objective`: own replenishment and supplier-side operational decisions without leaking those semantics into other modules
- `Current status`: `ready for spec`
- `Why it matters now`: the repository already states procurement as the next gate after catalog, pricing, and inventory boundaries are frozen
- `Depends on`: pricing, inventory, contracts, workers/integrations, shopping outputs
- `Unblocks`: supplier-side replenishment, operational buying workflows, analytics and shopping consolidation
- `Existing artifacts`: `docs/PROCUREMENT_CANONICAL_MODEL.md`, `docs/PROCUREMENT_IMPLEMENTATION_PLAN.md`
- `Next artifact to create`: procurement master spec refresh or procurement next-tranche spec, depending on whether the current canonical model already covers the intended scope

## Transversal fronts

### Governance

- `Objective`: keep runtime governance, repository governance, and control semantics explicit and aligned
- `Current status`: `done`
- `Why it matters now`: base governance consolidation is complete and now serves as the foundation for orchestration and future front planning
- `Depends on`: `docs/PROJECT_SOT.md`, `ARCHITECTURE.md`, agent entrypoints
- `Unblocks`: every future spec and implementation plan
- `Existing artifacts`: `docs/PROJECT_SOT.md`, `AGENTS.md`, `CLAUDE.md`, `CODEX.md`
- `Next artifact to create`: none unless governance rules materially change

### Contracts

- `Objective`: own API, event, and governance contract discipline across all fronts
- `Current status`: `in progress`
- `Why it matters now`: every thin-client and async front depends on strong contract sequencing
- `Depends on`: governance, module boundaries
- `Unblocks`: shopping, analytics, CRM, procurement, sdk generation
- `Existing artifacts`: `docs/CONTRACT_CONVENTIONS.md`, `docs/CONTRACT_EVOLUTION_RULES.md`, `contracts/`
- `Next artifact to create`: front-specific contract specs as each new front is selected

### SDK generation

- `Objective`: keep generated client/runtime artifacts aligned with the contract-first model
- `Current status`: `in progress`
- `Why it matters now`: every product surface depends on stable generated access to backend contracts
- `Depends on`: contracts, CI/quality gates
- `Unblocks`: Home, Shopping, Analytics, CRM, future desktop and admin surfaces
- `Existing artifacts`: `docs/SDK_GENERATION_STRATEGY.md`, `docs/SDK_BOUNDARY.md`
- `Next artifact to create`: no standalone spec unless generation strategy changes materially

### Auth/session

- `Objective`: own authentication, identity provider integration, and cookie-session delivery semantics for thin clients
- `Current status`: `in progress`
- `Why it matters now`: several future fronts depend on stable authenticated user context
- `Depends on`: governance, contracts, frontend migration, CI/quality gates
- `Unblocks`: CRM, analytics operator workflows, post-login product expansion
- `Existing artifacts`: `docs/LOGIN_AND_IDENTITY_ARCHITECTURE.md`, `docs/LOGIN_MVP_EXECUTION_PLAN.md`, `docs/LOGIN_DOD.md`
- `Next artifact to create`: a focused auth/session follow-up spec only if the next tranche falls outside the already frozen login closure scope

### Frontend migration

- `Objective`: preserve legacy visual value while enforcing modern package, API, and ownership boundaries
- `Current status`: `in progress`
- `Why it matters now`: Home, Shopping, Analytics, and CRM all depend on this guardrail to avoid regression into weak legacy patterns
- `Depends on`: governance, contracts, sdk generation, design system discipline
- `Unblocks`: all product surfaces
- `Existing artifacts`: `docs/FRONTEND_MIGRATION_CHARTER.md`, `docs/FRONTEND_MIGRATION_PLAYBOOK.md`, `docs/FRONTEND_MIGRATION_MATRIX.md`
- `Next artifact to create`: front-specific frontend specs as each surface is selected

### Workers/integrations

- `Objective`: own asynchronous ingestion, supplier runtime execution, connector discipline, and non-core compute paths
- `Current status`: `in progress`
- `Why it matters now`: Shopping and future procurement and analytics flows depend on governed worker and connector evolution
- `Depends on`: contracts, governance, outbox/event discipline
- `Unblocks`: Shopping follow-up work, procurement inputs, analytics enrichment
- `Existing artifacts`: `docs/WORKER_OPERATING_MODEL.md`, `docs/SHOPPING_DRIVER_SUITE_ACCEPTANCE.md`, `apps/integration_worker/`
- `Next artifact to create`: a focused worker/integration spec only when a new connector family or runtime capability is selected

### CI/quality gates

- `Objective`: enforce validation, build, drift, and verification discipline across the repository
- `Current status`: `in progress`
- `Why it matters now`: every front depends on reliable gates to avoid invisible drift
- `Depends on`: contract validation, generated artifact checks, repository structure
- `Unblocks`: safe scaling of all future module work
- `Existing artifacts`: `.github/workflows/`, `docs/PROGRESS.md`
- `Next artifact to create`: targeted quality-gate spec only if the CI scope or acceptance model changes materially

### Observability/security

- `Objective`: own baseline tracing, logging, security, and operational guardrails across the platform
- `Current status`: `waiting on dependency`
- `Why it matters now`: it remains foundational but is not yet the recommended next detailed planning front
- `Depends on`: platform maturity, auth/session, worker/runtime growth
- `Unblocks`: higher-confidence production hardening across multiple fronts
- `Existing artifacts`: `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
- `Next artifact to create`: observability/security spec when product and runtime breadth justify a dedicated tranche

## Cross-front dependency rules

- `Catalog` is canonical upstream data foundation for `Pricing`, `Inventory`, `Procurement`, `Analytics`, and parts of `CRM`.
- `Pricing` and `Inventory` must stay semantically narrow so `Procurement` can open without inheritance drift.
- `Procurement` depends on `Pricing`, `Inventory`, and `Workers/integrations`, and must not open on vague ERP semantics.
- `Shopping` and `Procurement` share supplier-side concerns, but `Shopping` does not replace procurement ownership.
- `Analytics` depends on contracts, governance, read models, and frontend migration decisions before deep planning opens.
- `CRM` depends on identity, events, and upstream commercial signals; it should not open before those dependencies are visible.
- `SDK generation` and `Frontend migration` affect every product front and must be checked before each new surface spec.
- `Auth/session` must stay visible as a transversal dependency for any front that assumes authenticated operator workflows.

## Recommended next fronts

1. `Analytics`
2. `Procurement`
3. `CRM`

Reasoning:

- `Analytics` is already in the accepted module order, has dedicated skill structure, and needs orchestration before detailed planning fragments.
- `Procurement` is explicitly called out in repository docs as the next gate after upstream boundary freezing.
- `CRM` remains important, but it should follow once analytics and upstream dependency clarity are stronger.
````

- [ ] **Step 3: Verify the new file contains the required orchestration sections**

Run:
```powershell
rg --no-heading -n "^## Purpose|^## Product fronts|^## Transversal fronts|^## Recommended next fronts" docs/MASTER_ORCHESTRATION_PLAN.md
```

Expected:
- one match for `## Purpose`
- one match for `## Product fronts`
- one match for `## Transversal fronts`
- one match for `## Recommended next fronts`

- [ ] **Step 4: Commit**

Run:
```powershell
git add docs/MASTER_ORCHESTRATION_PLAN.md
git commit -m "docs(orchestration): add master orchestration plan"
```

Expected: a successful commit containing only the new orchestration document

---

### Task 2: Align `docs/PROJECT_SOT.md` to recognize the orchestration document

**Files:**
- Modify: `docs/PROJECT_SOT.md`

- [ ] **Step 1: Add the orchestration document to the planning deliverables**

Insert this exact bullet immediately before `- explicit decision record for the next implementation area`:

```markdown
- explicit master orchestration plan that maps fronts, dependencies, and recommended spec order
```

- [ ] **Step 2: Add the orchestration document to the key planning docs list**

Insert this exact bullet immediately before `- docs/NEXT_EXECUTION_DECISION.md`:

```markdown
- `docs/MASTER_ORCHESTRATION_PLAN.md`
```

- [ ] **Step 3: Verify `docs/PROJECT_SOT.md` now points to the orchestration document**

Run:
```powershell
rg --no-heading -n "master orchestration plan|docs/MASTER_ORCHESTRATION_PLAN.md" docs/PROJECT_SOT.md
```

Expected:
- one match in `Planning deliverables`
- one match in `Key planning docs`

- [ ] **Step 4: Commit**

Run:
```powershell
git add docs/PROJECT_SOT.md
git commit -m "docs(orchestration): register master orchestration in PROJECT_SOT"
```

Expected: a successful commit containing only the `docs/PROJECT_SOT.md` update

---

### Task 3: Align `docs/IMPLEMENTATION_PLAN.md` and `docs/PROGRESS.md` to the new orchestration state

**Files:**
- Modify: `docs/IMPLEMENTATION_PLAN.md`
- Modify: `docs/PROGRESS.md`

- [ ] **Step 1: Update the ongoing workstream wording in `docs/IMPLEMENTATION_PLAN.md`**

Replace this bullet:

```markdown
- master orchestration planning on top of the accepted governance hierarchy
```

With this exact bullet:

```markdown
- master orchestration execution on top of the accepted governance hierarchy
```

- [ ] **Step 2: Add the orchestration document to the `Done` list in `docs/PROGRESS.md`**

Append this exact bullet at the end of the `## Done` list:

```markdown
- master orchestration document added as the live execution index across product and transversal fronts
```

- [ ] **Step 3: Update the `## Next` list in `docs/PROGRESS.md`**

Replace this bullet:

```markdown
- create the master orchestration plan on top of the accepted governance hierarchy now that the consolidation tranche is complete
```

With this exact bullet:

```markdown
- open the next detailed front spec from `docs/MASTER_ORCHESTRATION_PLAN.md`, starting with the recommended front unless dependencies change
```

- [ ] **Step 4: Verify progress and implementation docs now reflect orchestration completion**

Run:
```powershell
rg --no-heading -n "master orchestration" docs/IMPLEMENTATION_PLAN.md docs/PROGRESS.md
```

Expected:
- `docs/IMPLEMENTATION_PLAN.md` contains the execution wording
- `docs/PROGRESS.md` contains one `Done` bullet and one `Next` bullet for the orchestration document

- [ ] **Step 5: Commit**

Run:
```powershell
git add docs/IMPLEMENTATION_PLAN.md docs/PROGRESS.md
git commit -m "docs(orchestration): align planning and progress state"
```

Expected: a successful commit containing only the `docs/IMPLEMENTATION_PLAN.md` and `docs/PROGRESS.md` updates

---

### Task 4: Update the active task tracker for the orchestration tranche

**Files:**
- Modify: `tasks/todo.md`

- [ ] **Step 1: Replace the current tracker content with the orchestration tranche state**

Write this file content exactly:

```markdown
# Task Tracker

## Current State

- State: completed
- Active tranche: master orchestration planning
- Source spec: `docs/superpowers/specs/2026-04-01-master-orchestration-plan-design.md`

## Completed Tasks

- [x] Create `docs/MASTER_ORCHESTRATION_PLAN.md` as the live orchestration index
- [x] Register the orchestration document in `docs/PROJECT_SOT.md`
- [x] Align `docs/IMPLEMENTATION_PLAN.md` and `docs/PROGRESS.md` to the new orchestration state
- [x] Verify orchestration consistency and identify the recommended next front

## Notes

- The orchestration layer now sits between repository governance and front-specific spec work.
- The next step should be one detailed front spec selected from `docs/MASTER_ORCHESTRATION_PLAN.md`.
```

- [ ] **Step 2: Verify the tracker now points to the orchestration tranche**

Run:
```powershell
rg --no-heading -n "master orchestration planning|MASTER_ORCHESTRATION_PLAN.md|recommended next front" tasks/todo.md
```

Expected:
- one match for the active tranche
- one match for the orchestration document
- one match for the next-step note

- [ ] **Step 3: Commit**

Run:
```powershell
git add tasks/todo.md
git commit -m "docs(orchestration): update task tracker for orchestration tranche"
```

Expected: a successful commit containing only the `tasks/todo.md` update

---

### Task 5: Run the orchestration consistency gate

**Files:** verification only

- [ ] **Step 1: Confirm all core planning entrypoints exist**

Run:
```powershell
Test-Path docs/MASTER_ORCHESTRATION_PLAN.md
Test-Path docs/PROJECT_SOT.md
Test-Path docs/IMPLEMENTATION_PLAN.md
Test-Path docs/PROGRESS.md
Test-Path tasks/todo.md
```

Expected: five lines of `True`

- [ ] **Step 2: Verify the orchestration document is referenced from the expected repository docs**

Run:
```powershell
rg --no-heading -n "MASTER_ORCHESTRATION_PLAN.md|master orchestration" docs/PROJECT_SOT.md docs/IMPLEMENTATION_PLAN.md docs/PROGRESS.md tasks/todo.md
```

Expected:
- `docs/PROJECT_SOT.md` references the document directly
- `docs/IMPLEMENTATION_PLAN.md` references orchestration execution
- `docs/PROGRESS.md` references orchestration completion and next-front selection
- `tasks/todo.md` references the orchestration tranche and next-step note

- [ ] **Step 3: Review the final doc-only diff**

Run:
```powershell
git diff --stat HEAD~4..HEAD
```

Expected: changes only in:
- `docs/MASTER_ORCHESTRATION_PLAN.md`
- `docs/PROJECT_SOT.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- `tasks/todo.md`

- [ ] **Step 4: Final commit**

Run:
```powershell
git commit --allow-empty -m "docs(orchestration): verify master orchestration alignment"
```

Expected: either a real verification commit if notes were added, or an empty audit commit marking the consistency gate as complete

---

## Summary

| Task | Purpose |
|------|---------|
| 1 | Create the canonical orchestration document with the approved hybrid structure |
| 2 | Register the orchestration document in `docs/PROJECT_SOT.md` |
| 3 | Align planning and progress state to show orchestration completion |
| 4 | Update the active task tracker to the orchestration tranche |
| 5 | Run a consistency gate across the final orchestration layer |
