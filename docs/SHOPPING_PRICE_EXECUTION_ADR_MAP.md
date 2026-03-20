# Shopping Price Execution ADR Map

Status: active
Date: 2026-03-19

Purpose:
Freeze the ADR set required to implement the Shopping Price core (legacy workflow preserved, target architecture enforced) and explicitly state which skills must be used for each decision follow-up.

## How to execute (frozen sequence)

For any Shopping evolution (Level 2+), the execution sequence is frozen:

1. Contracts first
   - OpenAPI in `contracts/api/openapi/shopping_v1.openapi.yaml`
   - JSON Schemas under `contracts/api/jsonschema/shopping_*.schema.json`
   - Events under `contracts/events/v1/` only when Phase 2 is started (ADR-0018)
   - Governance under `contracts/governance/` only when runtime flags/policies are required
   Skill: `metalshopping-openapi-contracts` (and `metalshopping-contract-authoring` when multiple contract types move together)
2. Go module next (real data, tenant-safe)
   Skill: `metalshopping-module-scaffold` (review with `metalshopping-server-core-modules` when structure is at risk)
3. Worker only when needed (scraping/async execution)
   Skill: `metalshopping-worker-scaffold` (+ `metalshopping-worker-patterns` for claim loop semantics)
4. SDK generation (source of truth is contracts)
   Skill: `metalshopping-sdk-generation`
5. Page delivery (legacy workflow preserved, thin client enforced)
   Skill: `metalshopping-frontend-migration-guardrails` then `metalshopping-page-delivery`

## ADR set (binding)

- ADR-0017: Shopping Price Operational Workflow Write Surface
  Skills: `metalshopping-adr-updates`, `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`, `metalshopping-page-delivery`, `metalshopping-frontend-migration-guardrails`
  Contracts: OpenAPI + JSON Schema (`shopping_v1` + `shopping_*` schemas)
  Output after coding phase: Shopping bootstrap + run submission endpoints, UI bound to real workflow surfaces.

- ADR-0018: Shopping Run Orchestration via DB Queue (Phase 1)
  Skills: `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`, `metalshopping-event-contracts` (Phase 2)
  Contracts: OpenAPI (run submission + run-request status); Events only in Phase 2
  Output after coding phase: run request table + worker claim loop (SKIP LOCKED) + observable lifecycle.

- ADR-0019: Suppliers Directory and Driver Manifests as Tenant-Scoped Data
  Skills: `metalshopping-contract-authoring`, `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`
  Contracts: OpenAPI for supplier read/management surfaces (may start embedded in Shopping bootstrap; may evolve into `suppliers_v1`)
  Output after coding phase: supplier list + manifest storage/validation surfaces that drive Shopping "Configurar".

- ADR-0020: Shopping Price Observation Data Model v1
  Skills: `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`, `metalshopping-platform-packages`
  Contracts: OpenAPI read surfaces + JSON Schemas for runs, items, latest snapshots
  Output after coding phase: per-(run, product, supplier) observations with idempotent upserts + latest snapshots per supplier.

- ADR-0021: Shopping Frontend Migration Preserves Legacy Workflow, Not Legacy Shortcuts
  Skills: `metalshopping-frontend-migration-guardrails`, `metalshopping-page-delivery`, `metalshopping-sdk-generation`
  Contracts: No new contract type; enforces consumption rules and ownership mapping
  Output after coding phase: legacy-like Shopping UI with thin-client boundaries and reusable widgets.

- ADR-0022: Postgres Tenant Session Key for RLS is app.tenant_id
  Skills: `metalshopping-platform-packages`, `metalshopping-worker-scaffold`
  Contracts: No new contracts; this freezes the runtime tenancy session key
  Output after coding phase: consistent tenancy behavior across Go + Python and no drift in scaffolds.

- ADR-0023: Shopping Run Input Sources (XLSX Scope and Catalog Selection)
  Skills: `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-page-delivery`, `metalshopping-frontend-migration-guardrails`
  Contracts: OpenAPI + JSON Schema changes (run request payload; optional XLSX scope surfaces)
  Output after coding phase: XLSX can be used as run scope without turning into a canonical product import path.

- ADR-0024: Persisted Supplier Product URLs and Lookup Mode for Shopping Execution
  Skills: `metalshopping-platform-packages`, `metalshopping-openapi-contracts`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`
  Contracts: OpenAPI + JSON Schema for supplier signals surfaces; no events required in v1
  Output after coding phase: manual URLs workflow + worker discovery acceleration without UI coupling.

## Phase 2 ADRs (broker + management surfaces)

- ADR-0025: Shopping Run Requested Event And Broker Consumption (Phase 2)
  Skills: `metalshopping-event-contracts`, `metalshopping-platform-packages`, `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`
  Contracts: Event (`contracts/events/v1/shopping_run_requested.v1.json`), outbox/inbox wiring; OpenAPI unchanged
  Status: accepted (evidence: `docs/SHOPPING_ADR025_ACCEPTANCE.md`)

- ADR-0026: Suppliers Management API Surface Split (Phase 2)
  Skills: `metalshopping-openapi-contracts`, `metalshopping-contract-authoring`, `metalshopping-module-scaffold`, `metalshopping-server-core-modules`, `metalshopping-page-delivery`, `metalshopping-frontend-migration-guardrails`
  Contracts: New OpenAPI `suppliers_v1` + JSON Schemas
  Status: accepted (evidence: `docs/SHOPPING_ADR026_ACCEPTANCE.md`)

- ADR-0027: Shopping Driver Manifest Validation And Activation (Phase 2)
  Skills: `metalshopping-platform-packages`, `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`
  Contracts: Management OpenAPI + per-family JSON Schemas; leverages existing manifest tables
  Status: accepted (evidence: `docs/SHOPPING_ADR027_ACCEPTANCE.md`)

- ADR-0028: Shopping Worker Manifest Runtime Gating (Phase 2)
  Skills: `metalshopping-worker-scaffold`, `metalshopping-worker-patterns`, `metalshopping-observability-security`
  Contracts: No new contract type; execution eligibility enforced from existing Suppliers + Shopping surfaces
  Status: accepted (evidence: `docs/SHOPPING_ADR028_ACCEPTANCE.md`)

- ADR-0029: Shopping Driver Runtime v1 (HTTP and PLAYWRIGHT)
  Skills: `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`, `metalshopping-adr-updates`
  Contracts: No new contract type in v1; runtime semantics frozen in worker + ADR
  Status: accepted (evidence: `docs/SHOPPING_ADR029_ACCEPTANCE.md`)

## Driver framework (scale-out gate)

- ADR-0030: Shopping Driver Strategy Framework v1 (family + strategy)
  Skills: `metalshopping-contract-authoring`, `metalshopping-platform-packages`, `metalshopping-module-scaffold`, `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`, `metalshopping-adr-updates`
  Contracts: JSON Schema evolution for driver manifests + deterministic validation semantics + worker dispatcher contract
  Status: accepted (evidence: `docs/SHOPPING_ADR030_ACCEPTANCE.md`)

## Backend completion ADRs (driver framework parity)

- ADR-0031: Shopping Driver Runtime Package Layout v1 (integration_worker)
  Skills: `metalshopping-adr-updates`, `metalshopping-worker-scaffold`, `metalshopping-worker-patterns`, `metalshopping-observability-security`
  Contracts: no new external contracts; refactor is internal packaging
  Goal: strategy executors become modular/testable and entrypoint is orchestration-only.
  Status: accepted (evidence: ADR-0031 build + smoke + entrypoint boundary review)

- ADR-0032: Shopping Driver Parallelism and Rate Limits v1
  Skills: `metalshopping-contract-authoring`, `metalshopping-platform-packages`, `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`
  Contracts: JSON Schema evolution (bounded concurrency keys) + deterministic validation
  Goal: bounded, family-aware concurrency that is safe under multi-tenant execution.

- ADR-0033: Shopping Driver Smoke Suite v1 (multi-supplier)
  Skills: `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`, `metalshopping-adr-updates`
  Contracts: no new external contracts
  Goal: one deterministic command that validates the legacy parity supplier set.

- ADR-0034: Shopping Playwright Driver Runtime v1 (PDP-first, non-mock)
  Skills: `metalshopping-contract-authoring`, `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`, `metalshopping-adr-updates`
  Contracts: JSON Schema evolution + deterministic validation for Playwright selectors/runtime options
  Goal: at least one non-mock Playwright supplier (OBRA_FACIL) accepted end-to-end.

## Rule

Next gate before UI: complete backend driver framework parity ADRs (ADR-0031..ADR-0034) with objective smoke evidence for the legacy supplier set.
