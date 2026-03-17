# Pricing Implementation Plan

## Purpose

Define the full execution plan for the first `pricing` slice so implementation does not start from intuition or local convenience.

This plan exists to keep pricing aligned with:

- the frozen MetalShopping architecture
- the canonical `catalog`
- runtime governance
- contract-first evolution
- future modules such as analytics, procurement, CRM, and market intelligence

## Decision

`pricing` is the correct next implementation area.

This is the right next move because:

- `catalog` is now strong enough to be the canonical owner of product identity
- the platform now already proves auth, tenancy, governance, contract tooling, and outbox
- legacy pricing and analytics semantics are rich enough to justify explicit pricing ownership now
- many future modules depend more on a strong `pricing` model than on opening another generic domain first

## Why pricing is better than other immediate candidates

### Better than opening `inventory` first

`inventory` matters, but current future-fit value is lower than `pricing` because:

- pricing is core to commercial strategy
- analytics and market intelligence already depend heavily on price and cost semantics
- governance thresholds and policies are naturally useful in pricing from day one

### Better than opening `market_intelligence` first

`market_intelligence` should consume and compare against canonical internal price, not define it.

Without `pricing`, market observations would have no clean internal price owner to compare against.

### Better than opening `crm` first

`crm` is product-relevant, but not the most structural next step for the commercial engine.

Pricing quality will influence CRM outcomes later, while CRM does not need to own internal price truth.

### Better than opening `procurement` first

`procurement` does influence costs and replenishment, but the first pricing slice can begin with explicit cost basis ownership even before procurement workflows are complete.

## Legacy signals worth preserving

The legacy MetalShopping system strongly suggests that pricing should model more than a single current price field.

Useful signals include:

- explicit `preco_interno`
- `custo_variavel`
- `custo_medio`
- realized price and realized margin analysis
- margin leakage and pricing execution quality
- market-relative pricing signals
- reasoned pricing actionability and confidence layers

Important rule:

- reuse these as semantics
- do not copy the legacy schema or analytics payloads directly into the first pricing table

## First-slice objective

The first slice of `pricing` should solve one clean problem:

`server_core` must be able to set and read a governed internal product price for a canonical product in one tenant, with explicit replacement cost semantics, optional average cost, effective window, origin, and event publication.

## First-slice scope

### In scope

- set internal product price
- list product prices
- current effective price view
- explicit replacement cost on write
- optional average cost on write
- validity window
- runtime policy enforcement
- versioned event publication
- tenant-aware persistence and RLS

### Out of scope

- customer-specific pricing
- channel pricing
- campaign engine
- promotion engine
- approval workflows
- dynamic repricing engine
- competitor-driven automatic repricing
- negotiated CRM offers
- supplier quote orchestration

## Canonical ownership rules

### `pricing` owns

- internal price records
- price validity windows
- cost basis used for price decision
- margin floor and pricing decision context
- price reason metadata
- price write lineage and events

### `pricing` does not own

- product identity
- taxonomy ownership
- inventory state
- procurement workflow state
- competitor observations
- customer negotiations

## Required contracts before implementation

### API

- `contracts/api/openapi/pricing_v1.openapi.yaml`

### JSON Schemas

- `pricing_set_product_price_request_v1.schema.json`
- `pricing_product_price_v1.schema.json`
- `pricing_product_price_list_v1.schema.json`

### Event

- `contracts/events/v1/pricing_price_set.v1.json`
- payload schema for the event

### Governance

- `pricing.manual_price_override`

### Generated artifacts

Pricing contracts must flow through:

- `scripts/validate_contracts.ps1`
- `scripts/generate_contract_artifacts.ps1`

before module wiring is considered complete.

## Required platform behavior

The first pricing slice must use the platform already in place:

- auth and authz
- tenant runtime
- DB-backed governance
- outbox
- contract validation/generation
- request logging and trace metadata

No pricing path should bypass those platform layers.

## First persistence model

### Table

- `pricing_product_prices`

### Minimum columns

- `price_id`
- `tenant_id`
- `product_id`
- `currency_code`
- `price_amount`
- `replacement_cost_amount`
- `average_cost_amount`
- `pricing_status`
- `effective_from`
- `effective_to`
- `origin_type`
- `origin_ref`
- `reason_code`
- `updated_by`
- `created_at`
- `updated_at`

### Required data constraints

- one open active effective price window per product per tenant
- foreign key to canonical product
- non-negative monetary values
- effective window consistency
- tenant-aware unique and lookup indexes
- RLS enabled and forced

## Runtime governance plan

### Policy

Add `pricing.manual_price_override`.

First behavior:

- if manual override is not allowed for the resolved scope, reject manual-origin price writes

## Event plan

### Event name

- `pricing.price_set`

### Event meaning

An internal product price was written or replaced for a canonical product and effective window.

### Event publication rule

- append to outbox in the same transaction as the pricing write
- publish through the platform outbox path
- keep payload additive and explainable

## API plan

### First endpoints

- `POST /api/v1/pricing/products/{product_id}/prices`
- `GET /api/v1/pricing/products/{product_id}/prices`

Optional first read if useful:

- `GET /api/v1/pricing/products/{product_id}/prices/current`

### Permission model

Use existing IAM semantics:

- `pricing:read`
- `pricing:write`

## Recommended implementation order

### Phase 1. Contracts and governance

1. add policy contract for `pricing.manual_price_override`
2. add pricing OpenAPI
3. add pricing JSON Schemas
4. add pricing event contract and payload schema
5. validate and generate artifacts

### Phase 2. Core module

1. create `apps/server_core/internal/modules/pricing`
2. implement domain model and errors
3. implement governance guards
4. implement Postgres adapter and migration
5. implement HTTP transport
6. wire bootstrap in `main.go`

### Phase 3. Runtime validation

1. apply migration
2. grant least-privilege access
3. smoke read/write path
4. validate threshold enforcement
5. validate policy enforcement
6. validate outbox event publication

### Phase 4. Post-slice hardening

1. review whether a current-price read model is needed
2. review whether analytics-serving should consume the pricing event
3. review whether procurement and market intelligence need explicit follow-up contracts

## Acceptance criteria

Pricing is considered complete for the first slice only if:

- a canonical product can receive a price write
- pricing write is tenant-safe
- pricing write is governed by threshold and policy
- pricing write emits a real outbox event
- API contracts validate
- generated artifacts update cleanly
- smoke tests pass end to end

## Risks to avoid

- putting price columns in `catalog`
- adding analytics-specific output fields into canonical pricing writes
- mixing procurement semantics into first pricing write model
- skipping effective windows
- skipping governance because the first slice seems small
- treating realized sales price as the same thing as canonical internal price

## Next move after the first pricing slice

After the first pricing slice is live, the next likely follow-on should be one of:

1. pricing read model for current effective price
2. inventory
3. market intelligence comparison surfaces
4. procurement cost and supplier linkage
5. analytics-serving consumers of pricing events

Which one comes next should be decided after observing whether the first pricing slice reveals stronger pressure from operations, analytics, or supplier workflows.
