# Procurement Implementation Plan

## Goal

Open `procurement` as a first-class module without breaking the ownership already established in `catalog`, `pricing`, and `inventory`.

## Phase 1: Freeze the procurement contract boundary

Deliverables:

- canonical procurement model frozen
- birth gate on published inputs frozen
- explicit upstream inputs listed
- explicit non-goals listed

Exit criteria:

- procurement ownership is not ambiguous
- supplier-side semantics are not leaking into `pricing` or `inventory`
- the required published inputs are known before runtime code

## Phase 2: Define upstream integration inputs

Deliverables:

- canonical contract definitions for:
  - `inventory_position`
  - `supplier_prices`
  - `sales_orders_open`
  - `purchase_orders_open`
  - `supplier_lead_time`
- decision on whether those inputs start as events, pull APIs, read surfaces, or integration-owned tables

Exit criteria:

- procurement can name its required upstream contracts explicitly
- procurement is not blocked on undocumented connector behavior

## Phase 3: Procurement contracts

Deliverables:

- governance contracts for procurement policies if needed
- `procurement_v1.openapi.yaml`
- request/response schemas
- first procurement event contract

Exit criteria:

- procurement surfaces are contract-first
- generated artifacts can include procurement without hand-maintained side types

## Phase 4: First runtime slice

Candidate first slice:

- supplier linkage to product
- open purchase order state
- lead-time state
- procurement recommendation seed surface

Exit criteria:

- procurement writes are tenant-aware
- procurement references canonical `product_id`
- procurement emits versioned outbox-backed events
- procurement does not directly read ERP internals

## Guardrails

- do not put supplier lead time into `inventory`
- do not put purchase-order state into `pricing`
- do not make procurement a direct DB integration module
- do not make procurement depend on analytics internals for operational facts
