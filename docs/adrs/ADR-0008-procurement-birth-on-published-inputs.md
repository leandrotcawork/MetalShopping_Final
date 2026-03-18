# ADR-0008: Procurement Must Be Born on Published Inputs

- Status: accepted
- Date: 2026-03-17

## Context

After `catalog`, `pricing`, and `inventory` were split into explicit owners, the next structural pressure from the legacy system is supplier-side replenishment and buying logic.

Legacy notes show that the future `buying` context depends on canonical operational inputs such as:

- `inventory_position`
- `sales_orders_open`
- `purchase_orders_open`
- `supplier_lead_time`
- `supplier_prices`

Those same notes also warn against letting `buying` read raw ERP tables or internal analytics implementation directly.

Without a frozen rule here, `procurement` could easily be implemented as:

- direct ERP reads inside the module
- cross-module joins into `pricing` and `inventory` tables as a permanent design
- dependency on analytics internals for operational facts

That would weaken tenant portability, connector portability, and future adaptive integration support.

## Decision

- `procurement` is the canonical owner of supplier-side replenishment semantics
- `procurement` must be born on top of published canonical inputs, not raw ERP reads
- the required upstream inputs include `products_master`, `inventory_position`, `supplier_prices`, `sales_orders_open`, `purchase_orders_open`, and `supplier_lead_time`
- `procurement` must not import analytics internals to recover those inputs
- `procurement` may consume published read surfaces, contracts, or integration-owned datasets, but not private implementation details

## Consequences

- `pricing` stays focused on internal price and cost semantics
- `inventory` stays focused on live stock position and stock timing
- supplier lead time and purchase-order semantics do not leak into existing write models
- integration becomes the formal publication path for procurement inputs across heterogeneous tenant ERPs
- procurement implementation is gated on published inputs, even if legacy tables already exist
