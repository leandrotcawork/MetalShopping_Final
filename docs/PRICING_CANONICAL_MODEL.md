# Pricing Canonical Model

## Purpose

Define the first canonical pricing model for MetalShopping before implementation starts.

This document freezes ownership, boundaries, and the first slice shape so pricing does not begin as a convenience table or a copy of legacy fields.

## Why this exists now

Pricing is one of the core product capabilities of MetalShopping.

It must support future modules such as:

- market intelligence
- procurement
- CRM
- analytics
- automation

That means the first slice must optimize for explicit semantics and safe evolution, not only for today's write path.

## Ownership boundaries

### `pricing` owns

- internal price values
- price validity windows
- price origin and reason metadata
- cost basis used for internal pricing decisions
- margin-aware pricing decisions
- price write history as pricing-owned state or pricing-owned event lineage
- governed thresholds and policies that shape pricing behavior

### `pricing` does not own

- canonical product identity
- taxonomy ownership
- stock quantities
- supplier operational lead time
- competitor observations
- CRM customer state

Those remain in:

- `catalog`
- `inventory`
- `procurement`
- `market_intelligence`
- `crm`

## First pricing slice

The first slice should answer:

- what is the current internal price for one canonical product in one tenant
- what cost basis was used
- what margin floor applies
- why and from where was the price set
- when does this price become effective

It does not need to solve every pricing scenario yet.

## Canonical aggregate

The first canonical pricing aggregate is `product_price`.

Minimum target fields:

- `price_id`
- `tenant_id`
- `product_id`
- `currency_code`
- `price_amount`
- `replacement_cost_amount`
- `average_cost_amount` (optional in the first slice)
- `pricing_status`
- `effective_from`
- `effective_to`
- `origin_type`
- `origin_ref`
- `reason_code`
- `updated_by`
- `created_at`
- `updated_at`

## Field semantics

### `product_id`

- foreign key to canonical `catalog` product identity
- pricing never creates an alternative product key

### `currency_code`

- required from day one
- avoids hidden single-currency assumptions in the model

### `price_amount`

- canonical current internal commercial price for the effective window

### `replacement_cost_amount`

- explicit replacement cost aligned to legacy `custo_variavel`
- required because future pricing logic, analytics, and explainability depend on it

### `average_cost_amount`

- optional average cost aligned to legacy `custo_medio`
- should remain explicit if the product needs to compare replacement cost and average cost without collapsing both meanings

### `pricing_status`

Suggested first values:

- `draft`
- `active`
- `inactive`

This should remain simple until approval workflows become real.

### `effective_from` and `effective_to`

- pricing is time-aware from the first slice
- no pricing model should assume eternal present-state rows only

### `origin_type`

Suggested first values:

- `manual`
- `policy`
- `import`

This keeps future automation and explainability open.

### `origin_ref`

- optional reference to the source operation, policy execution, import batch, or integration payload

### `reason_code`

- stable reason identifier for why the price was written
- better than free-text-only justification

## First supporting governance

The first pricing slice must use runtime governance for:

### Policy

First target policy:

- `pricing.manual_price_override`

Purpose:

- decide whether manual override is permitted for the resolved scope

This is a better first policy than an approval workflow because it introduces control without prematurely modeling a large approval system.

## First contract scope

The first contract set for pricing should include:

### API

- `pricing_v1.openapi.yaml`

### JSON Schemas

- `pricing_set_product_price_request_v1.schema.json`
- `pricing_product_price_v1.schema.json`
- `pricing_product_price_list_v1.schema.json`

### Event

- `pricing_price_set.v1.json`
- payload schema for `pricing_price_set`

### Governance

- add policy contract for `pricing.manual_price_override`

## First migration target

The first table should be `pricing_product_prices`.

It should be:

- tenant-aware
- linked to `catalog_products(product_id)`
- constrained to one open effective price window per product and tenant
- protected by RLS
- treated as canonical change history, not run history

## Historical write rule

`pricing_product_prices` is the canonical history table for internal pricing changes.

Frozen rule:

- write a new row only when commercial state changes materially
- repeated imports or reruns with the same semantic price state must not create a new history row
- no-op writes must not publish new pricing events

Material change for the first slice means a change in one or more of:

- `currency_code`
- `price_amount`
- `replacement_cost_amount`
- `average_cost_amount`
- `pricing_status`
- `effective_to`

Operational fields such as actor, origin reference, and request execution time do not justify a new canonical history row by themselves

## Explicit non-goals for the first slice

Do not add all of the following now:

- customer-specific price lists
- regional channel pricing
- approval workflow engine
- promotion engine
- competitor-driven automatic repricing
- supplier quote comparison
- price simulation UI model

Those can come later, but they should not distort the first canonical model.

## Legacy guidance

Legacy price and cost fields are useful as input signals, not as the target schema itself.

Reuse posture:

- preserve useful semantics like internal price, variable cost, and average cost
- do not copy legacy naming or mixed ownership blindly
- model price decisions so analytics and explainability remain possible later

## Implementation order

1. freeze this model
2. create missing pricing policy contract
3. create pricing API and event contracts
4. implement migration and module
5. validate real governed write path
6. publish real pricing event through outbox

## Conclusion

The first pricing slice should be small, but semantically strong.

It must begin as a governed, time-aware, tenant-aware pricing capability attached to canonical product identity, not as an ad hoc `price` column on `catalog`.
