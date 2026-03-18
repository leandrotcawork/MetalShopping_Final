# Procurement Canonical Model

## Purpose

Define the canonical ownership and birth constraints for `procurement` so the platform can add replenishment and supplier-side workflows without collapsing `pricing`, `inventory`, and integration concerns into a single module.

`procurement` is the target-platform successor of the legacy planning and supplier-facing semantics often described as `buying`.

## Ownership

`procurement` owns:

- supplier-side replenishment context
- supplier linkage for acquisition decisions
- purchase-side demand coverage context
- open purchase order state once intentionally modeled
- supplier lead time semantics once intentionally modeled
- procurement decisions, recommendations, and workflow state

`procurement` does not own:

- canonical product identity
- internal selling price
- live stock write ownership
- direct ERP connectivity
- analytics internals

## Required upstream inputs

`procurement` should be born on top of published canonical inputs, not direct legacy tables or ad hoc joins.

The minimum upstream inputs are:

- `products_master`
- `inventory_position`
- `supplier_prices`
- `sales_orders_open`
- `purchase_orders_open`
- `supplier_lead_time`

These inputs are part of the integration-facing target shape already documented in legacy architecture notes and must become explicit published datasets or contracts before procurement becomes a real runtime module.

## First owned semantics

The first procurement slice should own semantics such as:

- `supplier_id`
- `supplier_product_reference`
- `open_purchase_quantity`
- `expected_receipt_at`
- `lead_time_days`
- `lead_time_variability_days`
- `procurement_status`
- `recommendation_reason_code`
- `origin_type`
- `origin_ref`
- `updated_by`
- `effective_from`
- `effective_to`

These are procurement semantics because they represent acquisition and replenishment state, not product identity, commercial pricing, or live stock position.

## Explicit split from current modules

### `catalog`

`catalog` continues to own:

- `product_id`
- `sku`
- `pn_interno`, `reference`, `ean`, and future aliases
- taxonomy and stable master data

### `pricing`

`pricing` continues to own:

- `price_amount`
- `replacement_cost_amount`
- `average_cost_amount`

`pricing` may later consume procurement signals, but it does not own supplier-side replenishment context.

### `inventory`

`inventory` continues to own:

- `on_hand_quantity`
- `last_purchase_at`
- `last_sale_at`

`inventory` does not become the owner of supplier lead time or purchase-order semantics just because they affect stock coverage.

## Birth gate

`procurement` does not start as a direct reader of ERP or external databases.

Frozen gate:

1. `integration` must publish or expose the minimum canonical procurement inputs
2. `procurement` must consume those published inputs, not raw legacy internals
3. `procurement` must not import internals from analytics-serving logic to get operational facts

This keeps the future module portable across different tenant ERPs and connector strategies.

## Non-goals

This document does not authorize:

- putting supplier lead time into `inventory_product_positions`
- putting purchase-order state into `pricing_product_prices`
- connecting `procurement` directly to raw ERP tables as its long-term model
- making `procurement` depend on analytics internals for operational data
