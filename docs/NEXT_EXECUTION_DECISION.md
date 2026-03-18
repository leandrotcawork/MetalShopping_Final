# Next Execution Decision

## Decision

The next implementation area should be the first operational product surface: `Products`.

## Why

- `catalog`, `pricing`, and `inventory` now exist as real bounded modules
- the product still lacks the first thin-client surface that makes those modules operable
- the legacy application already proved the importance of `Home`, `Products`, and `Shopping` as the first operational surfaces
- `Products` is the closest surface to the backend we already have, while `Shopping` and `Home` depend on more readmodel and integration work

## Constraints

This decision is valid only if planning and implementation follow:

- `docs/OPERATIONAL_SURFACES_PLAN.md`
- the existing SKU ownership rules in `docs/adrs/ADR-0007-canonical-sku-data-ownership.md`
- the thin-client rule already frozen in `docs/adrs/ADR-0005-thin-clients-and-generated-sdks.md`

## Explicit rejection

Do not jump next to:

- rebuilding `Shopping` UI before its runtime foundation is clarified
- rebuilding `Home` as a static shell before trusted read models exist
- copying legacy frontend DTOs or API wrappers as a second source of truth
- putting business logic directly inside page components

until the first operational frontend slice is frozen as a thin-client feature.
