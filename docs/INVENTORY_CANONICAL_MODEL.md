# Inventory Canonical Model

## Purpose

Define the first canonical inventory model for MetalShopping so stock ownership opens cleanly after the pricing split.

## Ownership

`inventory` owns:

- current on-hand quantity
- operational timing signals directly tied to stock state
- inventory position history by meaningful change

`inventory` does not own:

- canonical product identity
- internal pricing semantics
- procurement replenishment workflows
- analytics-only stale or advisory classifications

## First slice

The first slice opens with `inventory_product_positions`.

Minimum fields:

- `position_id`
- `tenant_id`
- `product_id`
- `on_hand_quantity`
- `last_purchase_at`
- `last_sale_at`
- `position_status`
- `effective_from`
- `effective_to`
- `origin_type`
- `origin_ref`
- `reason_code`
- `updated_by`
- `created_at`
- `updated_at`

## Legacy alignment

Legacy `product_erp` signals map like this:

- `estoque_disponivel` -> `on_hand_quantity`
- `dt_compra` -> `last_purchase_at`
- `dt_venda` -> `last_sale_at`

`dias_sem_venda` does not belong to canonical inventory write ownership for the first slice.

It should remain derived by inventory-serving or analytics-serving read models later.

## Historical rule

`inventory_product_positions` is change history, not run history.

Frozen rule:

- write a new row only when the operational stock state changes materially
- repeated imports with the same quantity and timing state must not create a new row
- no-op writes must not publish new inventory events

## Event

The first inventory event is:

- `inventory.position_updated`

It is published through outbox only when a material inventory position change is accepted.
