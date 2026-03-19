# ADR-0021: Shopping Frontend Migration Preserves Legacy Workflow, Not Legacy Shortcuts

- Status: accepted
- Date: 2026-03-19

## Context

The repo freezes:

- thin clients
- contract-driven pages
- generated SDK + authored runtime boundary
- legacy visual language preservation during migration

The legacy Shopping page is valuable because it is a workflow surface with:

- a clear 3-step wizard (Upload, Configurar, Executar)
- dense operational tables and filters
- pragmatic "manual URLs" and supplier selection patterns

But the legacy implementation contains shortcuts we must not reintroduce:

- page-local transport logic
- ad hoc DTOs
- implicit runtime dependencies
- weak boundaries between UI and integration execution

## Decision

The Shopping page migration will preserve the legacy workflow and visual language while enforcing the target frontend architecture:

- pages consume `@metalshopping/platform-sdk` (no `fetch()` in pages)
- contracts in `contracts/api/openapi/*` remain the source of truth
- UI widgets are promoted to `packages/ui` only when reused 3+ times
- Shopping-specific composition may live in a feature package if complexity warrants it, but migration starts from `apps/web` + stable UI primitives

Workflow UX is frozen as:

- Step 1: Input preparation (XLSX or catalog selection)
- Step 2: Configuration (supplier selection, execution params)
- Step 3: Execution (progress, result, history)

## Legacy mapping (frozen)

The migrated surface must preserve the legacy workflow intent and operational density:

- Tabs/steps remain visible and explicit (Upload, Configurar, Executar).
- Catalog selection mode stays available and is not replaced by "XLSX import as canonical data".
- Filters remain workflow-friendly (dense, multi-filter, quick clear/reset) and use existing shared widgets.
- Manual URL workflow remains possible, but persists through backend-owned tables and contracts (ADR-0024), not page-local state.

## Contracts (touchpoints)

- OpenAPI: `contracts/api/openapi/shopping_v1.openapi.yaml`
- JSON Schemas: `contracts/api/jsonschema/shopping_*.schema.json`
- Governance: none required for Level 1/2 baseline unless a runtime gate is introduced later
- Events: none required in v1; future async upgrades follow ADR-0018 Phase 2

## Implementation checklist (Step 5)

- Run `metalshopping-frontend-migration-guardrails` first: preserve/refactor/reject classification against the legacy surface.
- Use `metalshopping-page-delivery` to bind the page to `@metalshopping/platform-sdk` only.
- If a widget appears in 3+ places, promote it to `packages/ui`; otherwise keep it feature-local.
- Do not introduce manual DTOs, page-local transport parsing, or direct `fetch()` in pages/components.

## Consequences

- We can match the legacy surface visually without importing its architectural debt.
- Backend-owned workflow contracts unlock stable UI behavior and testing over time.
- The Shopping surface becomes a scalable foundation for later modules (analytics, procurement).

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Legacy study + preserve/refactor/reject classification: `metalshopping-frontend-migration-guardrails`
- Page delivery with platform SDK: `metalshopping-page-delivery`
- SDK boundary verification: `metalshopping-sdk-generation`
