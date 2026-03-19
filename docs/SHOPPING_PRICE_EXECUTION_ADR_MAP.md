# Shopping Price Execution ADR Map

Status: active
Date: 2026-03-19

Purpose:
Freeze the ADR set required to implement the Shopping Price core (legacy workflow preserved, target architecture enforced) and explicitly state which skills must be used for each decision follow-up.

## ADR set (binding)

- ADR-0017: Shopping Price Operational Workflow Write Surface
  Skills: `metalshopping-adr-updates`, `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`, `metalshopping-page-delivery`, `metalshopping-frontend-migration-guardrails`
  Output after coding phase: Shopping bootstrap + run submission endpoints, UI bound to real workflow surfaces.

- ADR-0018: Shopping Run Orchestration via DB Queue (Phase 1)
  Skills: `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`, `metalshopping-event-contracts` (Phase 2)
  Output after coding phase: run request table + worker claim loop (SKIP LOCKED) + observable lifecycle.

- ADR-0019: Suppliers Directory and Driver Manifests as Tenant-Scoped Data
  Skills: `metalshopping-contract-authoring`, `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`
  Output after coding phase: supplier list + manifest storage/validation surfaces that drive Shopping "Configurar".

- ADR-0020: Shopping Price Observation Data Model v1
  Skills: `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-worker-scaffold`, `metalshopping-platform-packages`
  Output after coding phase: per-(run, product, supplier) observations with idempotent upserts + latest snapshots per supplier.

- ADR-0021: Shopping Frontend Migration Preserves Legacy Workflow, Not Legacy Shortcuts
  Skills: `metalshopping-frontend-migration-guardrails`, `metalshopping-page-delivery`, `metalshopping-sdk-generation`
  Output after coding phase: legacy-like Shopping UI with thin-client boundaries and reusable widgets.

- ADR-0022: Postgres Tenant Session Key for RLS is app.tenant_id
  Skills: `metalshopping-platform-packages`, `metalshopping-worker-scaffold`
  Output after coding phase: consistent tenancy behavior across Go + Python and no drift in scaffolds.

- ADR-0023: Shopping Run Input Sources (XLSX Scope and Catalog Selection)
  Skills: `metalshopping-openapi-contracts`, `metalshopping-module-scaffold`, `metalshopping-page-delivery`, `metalshopping-frontend-migration-guardrails`
  Output after coding phase: XLSX can be used as run scope without turning into a canonical product import path.

- ADR-0024: Persisted Supplier Product URLs and Lookup Mode for Shopping Execution
  Skills: `metalshopping-platform-packages`, `metalshopping-openapi-contracts`, `metalshopping-worker-scaffold`, `metalshopping-observability-security`
  Output after coding phase: manual URLs workflow + worker discovery acceleration without UI coupling.

## Rule

No implementation work for Shopping Price Level 2 may start until all ADRs above exist in `docs/adrs/` with Status: accepted and are referenced by the tranche plan.
