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
