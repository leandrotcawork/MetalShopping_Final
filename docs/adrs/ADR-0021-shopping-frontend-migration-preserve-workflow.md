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

## Consequences

- We can match the legacy surface visually without importing its architectural debt.
- Backend-owned workflow contracts unlock stable UI behavior and testing over time.
- The Shopping surface becomes a scalable foundation for later modules (analytics, procurement).

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Legacy study + preserve/refactor/reject classification: `metalshopping-frontend-migration-guardrails`
- Page delivery with platform SDK: `metalshopping-page-delivery`
- SDK boundary verification: `metalshopping-sdk-generation`

