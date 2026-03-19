# Next Execution Decision

## Decision

The next implementation area is Home Level 1 closure under the make-it-work-first execution mode, then Shopping readiness.

## Why

- login hardening and boundary guards are already implemented
- Home is now the first surface in the current delivery strategy
- the team needs deterministic module throughput with `OpenAPI -> Go handler -> generated SDK -> React page`

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
- implementing Shopping or Analytics without frozen data contract maps
- introducing manual frontend transport or page-level fetch shortcuts
- reopening closed Level 1 modules without user-driven need or dependency gate
