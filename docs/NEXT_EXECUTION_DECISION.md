# Next Execution Decision

## Decision

The next implementation area is Shopping Price Level 2 under the make-it-work-first execution mode.

## Why

- login hardening and boundary guards are implemented and stable
- Home Level 1 is closed with explicit acceptance evidence in docs
- Shopping Level 1 is implemented and formally closed
- Analytics readiness depends on Shopping snapshots being produced by a real worker core, not just a scaffold
- the team should keep deterministic module throughput with `OpenAPI -> Go handler -> generated SDK -> React page`
- the Shopping Level 2 ADR set is now frozen as the binding pre-coding gate

Binding references:

- `docs/SHOPPING_PRICE_EXECUTION_ADR_MAP.md`
- ADR-0017 .. ADR-0024

## Constraints

This decision is valid only if planning and implementation follow:

- `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`
- `docs/OPERATIONAL_SURFACES_PLAN.md`
- `docs/FRONTEND_MIGRATION_MATRIX.md`
- `docs/DATA_CONTRACT_MAP.md`
- `docs/PROJECT_SOT.md`
- `docs/PROGRESS.md`

## Explicit rejection

Do not jump next to:

- changing module order without updating SoT and plan docs in the same tranche
- changing Analytics scope without freezing the screen-level data inventory and API contract map first
- introducing manual frontend transport or page-level fetch shortcuts
- reopening closed Level 1 modules without user-driven need or dependency gate
