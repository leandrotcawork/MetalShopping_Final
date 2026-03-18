# Implementation Plan

## Goal

Sequence the work so the team can move from planning to implementation without reopening base architecture decisions.

## Phase 0: Freeze the base

Deliverables:

- architecture validated and updated
- global system and engineering principles frozen
- ADRs for the critical freezes
- AGENTS files for repo guidance
- progress tracking in place

Exit criteria:

- tenant isolation rule is frozen
- history ownership rule is frozen
- runtime governance rule is frozen
- frontend thin-client rule is frozen
- async integration rule is frozen
- global operating principles are explicit

## Phase 1: Platform contracts and governance foundation

Deliverables:

- initial `contracts/api`, `contracts/events`, and `contracts/governance` skeletons
- official contract naming and ownership conventions
- initial contract templates
- governance schema strategy
- event versioning conventions
- SDK generation strategy

Exit criteria:

- contract folders have explicit ownership and naming conventions
- initial contract templates are available for repeatable authoring
- governance model is described end-to-end
- no parallel manual type source is planned
- generated targets for TS and Python are defined

## Phase 2: Core platform skeleton

Deliverables:

- `server_core` package boundaries
- module-by-module ownership map
- platform package boundaries for auth, tenancy, governance, db, messaging, observability, security, audit
- module standards and readmodel/event rules
- creation checklists and structural templates
- initial migration strategy for multitenant canonical data

Exit criteria:

- module pattern is fixed
- platform versus domain boundaries are documented
- readmodel and event usage rules are explicit
- templates exist for repeatable module and platform package creation
- no worker owns canonical state

## Phase 3: First implementation wave

Priority order:

1. tenancy and IAM foundation
2. governance runtime skeleton
3. contracts and generated SDK flow
4. first business modules with canonical writes

Exit criteria:

- core can expose basic health and governance-safe bootstrap surfaces
- workers have clear consumption boundaries
- frontend can depend on generated contracts only

Current status:

- substantially in progress
- Postgres, auth, tenancy, IAM, catalog, governance runtime foundation, and first event/outbox path are already implemented
- remaining work in this phase is focused on operating the new hardening path end-to-end and keeping `pricing`, `inventory`, and upcoming `procurement` boundaries clean while supplier-side semantics are frozen before code

## Phase 3A: Foundation hardening

Priority order:

1. contract validation and generation flow
2. governance influencing real runtime behavior
3. event and outbox publication for real mutations
4. stronger production-grade auth evolution path

Exit criteria:

- contracts are not only authored but validated and generation-ready
- at least one runtime path is governed by governance resolution
- at least one module publishes a real versioned event through a core-owned outbox path
- bootstrap auth has a clear upgrade path to non-static identity

Current status:

- substantially implemented
- `catalog` create is governed by database-backed feature flags and thresholds
- IAM admin assignment is governed by database-backed policy resolution
- `catalog` create writes a real versioned event through platform outbox infrastructure
- contract validation and generation scripts are functional for the current repo contract set
- catalog readiness for `pricing` is now explicitly reviewed and the next work is freezing the pricing contract and model boundary before implementation

## Phase 4: Domain expansion

Candidate order:

1. freeze the canonical `catalog` product model from legacy signals
2. pricing
3. inventory
4. procurement planning and integration gate
5. procurement
6. market_intelligence
7. analytics_serving
8. CRM
9. alerts and notifications

Phase 4 gate:

- `pricing` must build on the canonical `catalog`, not on the original minimal `catalog_products` slice
- the canonical ownership of product identity, taxonomy, identifiers, descriptions, and non-catalog data boundaries must be explicit first
- SKU data ownership across `catalog`, `pricing`, `inventory`, `procurement`, and analytics must be frozen before pricing semantics evolve further
- legacy `product_erp` signals must be explicitly split so `pricing` does not absorb stock, stale-stock, tax, or advisory analytics fields by convenience
- `procurement` must not start on raw ERP or direct external DB reads as its long-term model
- the first procurement slice must depend on published canonical procurement inputs such as `inventory_position`, `sales_orders_open`, `purchase_orders_open`, and `supplier_lead_time`
- `procurement` must not import analytics internals to recover operational facts

## Ongoing workstreams

- architecture governance
- contract governance
- operational observability and security
- domain sequencing
- migration planning from legacy
