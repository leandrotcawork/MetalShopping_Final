# Next Execution Decision

## Decision

The next implementation area should be `procurement` planning, not immediate procurement runtime code.

## Why

- `catalog`, `pricing`, and `inventory` now exist as real bounded modules
- the next structural risk is letting supplier-side replenishment semantics leak back into those modules
- legacy architecture notes show future `buying` depends on canonical datasets such as `inventory_position`, `sales_orders_open`, `purchase_orders_open`, and `supplier_lead_time`
- freezing procurement first keeps the next domain portable across different ERPs and integration connectors

## Constraints

This decision is valid only if planning and implementation follow:

- `docs/PROCUREMENT_CANONICAL_MODEL.md`
- `docs/PROCUREMENT_IMPLEMENTATION_PLAN.md`
- `docs/adrs/ADR-0008-procurement-birth-on-published-inputs.md`
- the existing SKU ownership rules in `docs/adrs/ADR-0007-canonical-sku-data-ownership.md`

## Explicit rejection

Do not jump next to:

- direct procurement runtime implementation on raw ERP reads
- supplier-side semantics inside `pricing`
- purchase-order semantics inside `inventory`
- procurement coupling to analytics internals

until procurement ownership and procurement birth constraints are frozen and reflected in the plan.
