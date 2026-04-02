# Design Spec: ERP Integration Subplatform

**Date:** 2026-04-02
**Status:** Approved
**Output:** A `0.1 ERP Integration` design that establishes a professional, scalable, Sankhya-first ERP connector subplatform for MetalShopping

---

## Overview

`0.1 ERP Integration` is not a one-off Sankhya importer. It is the foundation that allows MetalShopping to ingest canonical business data from external ERPs in a controlled, auditable, and future-proof way.

The design should create a true multi-ERP architecture from day one, while implementing only one connector in `v1`: `Sankhya`.

This subplatform exists to move ERP-origin data into MetalShopping through an explicit staged pipeline:

- `raw`
- `staging/normalized`
- `reconciliation`
- `canonical promotion`

That structure is mandatory because MetalShopping will depend on ERP-origin products, prices, costs, inventory, sales, purchases, customers, and suppliers across analytics, procurement, CRM, and operational intelligence.

---

## Problem Statement

MetalShopping needs a strong ERP ingestion foundation, but the wrong shape would create long-term problems:

- a Sankhya-only importer would block future ERP growth
- direct writes into canonical tables would weaken auditability and replay
- a separate ERP app too early would fragment the approved architecture
- manual overrides of ERP-authoritative fields in MetalShopping would create competing sources of truth

The design must solve for two realities at once:

1. `Sankhya` is the only ERP connector that needs to exist in `v1`
2. the platform must be structurally ready for additional ERP connectors later

---

## Goals

1. Define `0.1 ERP Integration` as a real multi-ERP-capable subplatform
2. Keep the runtime inside `apps/integration_worker`, aligned with the approved repository architecture
3. Keep `apps/server_core` as the control plane for configuration, governance, scheduling, and review operations
4. Enforce `raw -> staging -> canonical` as the permanent ingestion model
5. Preserve canonical ownership inside existing domain modules
6. Support the full `v1` entity foundation:
   - `products`
   - `prices`
   - `costs`
   - `inventory`
   - `sales`
   - `purchases`
   - `customers`
   - `suppliers`
7. Make replay, auditability, observability, and review handling first-class platform capabilities

---

## Non-Goals

- building a Sankhya-only importer with no extensibility model
- creating a dedicated `apps/erp_sync_worker` in `v1`
- supporting multiple active ERPs simultaneously in the same tenant in `v1`
- allowing manual override inside MetalShopping for ERP-authoritative imported fields
- making event-driven or near-real-time sync the primary `v1` model
- pushing ERP connectors to write directly into canonical tables
- using no-code or config-only abstractions for all future ERPs

---

## Recommended Approach

Use a dedicated ERP connector subplatform inside `apps/integration_worker`.

This is the strongest option because it preserves the approved architecture while still creating a serious integration platform:

- `server_core` remains the control plane
- `integration_worker` remains the execution plane
- ERP connectors are adapters inside the integration layer
- the canonical model stays owned by `catalog`, `pricing`, `inventory`, `sales`, `purchases`, `customers`, and `suppliers`

This is superior to:

- a minimal importer approach, because `0.1` is too foundational to keep weak boundaries
- a fully separate ERP app, because that would be premature specialization without proven operational need

---

## Core Decisions

The design freezes these decisions:

1. `Sankhya-first`, but structurally multi-ERP from day one
2. `single active ERP per tenant` in `v1`
3. `ERP` is the authoritative source for imported fields
4. data corrections happen in the ERP, then are re-synced into MetalShopping
5. the ingestion model is mandatory:
   - `raw`
   - `staging/normalized`
   - `reconciliation`
   - `canonical promotion`
6. promotion is automatic by default, but relevant divergences go to `review queue`
7. `v1` sync modes are:
   - `bulk initial load`
   - `scheduled incremental sync`
   - `manual rerun`
8. pipeline execution is per-entity, not one monolithic ERP-wide transaction

---

## Architecture And Ownership

### `apps/server_core`

`server_core` owns the control plane for ERP integrations:

- integration instance lifecycle
- tenant-scoped ERP type selection
- secret and connection references
- enabled entities
- schedules and policies
- operational status reads
- review queue APIs
- governance and audit boundaries

### `apps/integration_worker`

`integration_worker` owns the execution plane:

- ERP connector registry
- connector runtime
- run orchestration
- raw ingestion
- staging normalization
- reconciliation
- canonical promotion
- retry handling
- run ledger
- observability

### Canonical domain modules

The ERP subplatform never becomes the owner of business domains.

Canonical ownership remains in:

- `catalog`
- `pricing`
- `inventory`
- `sales`
- `purchases`
- `customers`
- `suppliers`

The ERP subplatform adapts to those boundaries. It does not redefine them.

---

## Data Flow And Runtime Stages

The ingestion flow must operate in explicit stages.

### 1. Run initialization

Each run is created with:

- tenant
- integration instance
- connector type
- run mode
- entity scope
- effective configuration
- starting cursor or snapshot reference

### 2. Extract to raw

The ERP connector reads source records and stores raw payloads together with source metadata such as:

- source identifiers
- connector type
- run id
- extraction time
- source timestamps
- cursor details
- payload hash

Raw storage must preserve replayability and traceability.

### 3. Normalize to staging

The raw ERP payload is converted into a stable internal staging model:

- parsed
- type-coerced
- flattened where needed
- structurally validated
- detached from source-specific wire shape

### 4. Reconciliation

The runtime reconciles staging records against canonical state:

- identifier matching
- relationship resolution
- duplicate detection
- ambiguity detection
- promotability decision

Each record is classified as:

- `promotable`
- `promotable_with_warning`
- `review_required`
- `rejected`

### 5. Canonical promotion

Approved records are promoted into the canonical module-owned tables.

Promotion must preserve source provenance:

- connector type
- source id
- sync timestamp
- promoting run id

Relevant writes may publish outbox events where module semantics require them.

### 6. Review queue

Relevant divergences must not silently promote. They become review items with explicit operational follow-up.

### 7. Cursor and run finalization

Each run must persist:

- per-entity progress
- promoted count
- warning count
- rejected count
- review count
- failure summary
- final cursor state
- final run status

---

## Canonical Boundaries And Entity Ownership

The `v1` ERP integration scope includes these entities:

- `products`
- `prices`
- `costs`
- `inventory positions`
- `sales transactions`
- `purchase transactions`
- `customers`
- `suppliers`

Their ownership remains:

- `catalog`: product identity, taxonomy, identifiers, master attributes
- `pricing`: imported prices, imported costs, relevant histories
- `inventory`: current positions and position history
- `sales`: canonical sales transactions and items
- `purchases`: canonical purchase transactions and items
- `customers`: canonical customer records
- `suppliers`: canonical supplier records

The ERP subplatform may ingest and promote into those domains, but it must not absorb their business semantics.

Fields imported from the ERP remain ERP-authoritative in `v1`.

What is not ERP-authoritative:

- MetalShopping analytical classifications
- recommendations
- derived KPIs
- commercial intelligence outputs
- inventory intelligence outputs
- pricing intelligence outputs

---

## Connector Model And Extensibility

The connector model must be real, but justified by practical future growth rather than vague abstraction.

### Base concepts

- `connector type`
  - the ERP family, such as `sankhya`
- `integration instance`
  - a tenant-scoped configured connector
- `entity capability`
  - the entities a connector supports
- `sync strategy`
  - the extraction mode per entity, such as snapshot or incremental cursor

### Minimum connector contract

Each ERP connector must implement:

- capability discovery
- connection validation
- entity extraction
- cursor handling
- raw metadata emission
- source-to-staging mapping
- error classification

### Explicit limitations

Connectors must not:

- write directly into canonical tables
- define MetalShopping domain semantics
- compute analytics or intelligence
- bypass review policy
- hide raw payload provenance

### `Sankhya` in `v1`

`Sankhya` is the only implemented connector in `v1`, but it must still be registered through the same generic connector model used for future ERPs.

Future ERP support should mean:

- add a new connector type
- implement a new adapter
- reuse the shared runtime, review, promotion, run, and observability layers

---

## Review Queue, Error Policy, And Observability

### Review queue

Every review item must record at least:

- tenant
- integration instance
- connector type
- entity
- source id
- run id
- severity
- reason code
- problem summary
- raw payload reference
- staging snapshot
- reconciliation output
- recommended action
- item status

The operational rule is simple:

- source problem: fix in ERP, then reprocess
- mapping or reconciliation problem: fix in the subplatform, then reprocess

### Error classification

The runtime should classify errors into at least:

- `transient`
- `source_data_error`
- `mapping_error`
- `reconciliation_error`
- `platform_error`

### Promotion policy

- `promotable`: promote
- `promotable_with_warning`: promote and log warning
- `review_required`: do not promote, open review item
- `rejected`: do not promote and mark explicit failure

### Observability requirements

The subplatform must provide:

- run ledger per execution
- counts by stage and entity
- duration by stage
- error rate by type
- review volume
- last successful sync age
- sync lag by entity
- integration instance health

Two operating principles are mandatory:

- `partial success is valid`
- `replay is first-class`

---

## V1 Scope

`0.1 v1` includes:

- ERP connector subplatform inside `apps/integration_worker`
- `Sankhya` as the first connector type
- `single active ERP per tenant`
- all eight required entities
- `bulk initial load`
- `scheduled incremental sync`
- `manual rerun`
- full staged pipeline
- review queue
- run ledger and observability
- source provenance preserved in canonical promotion

---

## Recommended Repo Shape

The repository should evolve in this direction:

```text
apps/
  server_core/
    internal/
      modules/
        erp_integrations/
          domain/
          application/
          ports/
          adapters/
          transport/
  integration_worker/
    internal/
      erp_runtime/
        connectors/
          sankhya/
        raw/
        staging/
        reconciliation/
        promotion/
        review/
        runs/
        observability/
```

This shape keeps:

- control and policy in the core
- execution in the worker
- connectors isolated by adapter
- runtime concerns separated by stage

---

## Acceptance Criteria

- `0.1` is framed as an ERP connector subplatform, not a one-off importer
- the design is multi-ERP by architecture and `Sankhya-first` by implementation
- `apps/integration_worker` is the approved runtime home
- `apps/server_core` is the approved control plane
- `single active ERP per tenant` is explicit for `v1`
- `raw -> staging -> canonical` is mandatory
- review handling is part of the design, not an afterthought
- canonical ownership remains in existing domain modules
- `bulk`, `incremental`, and `manual rerun` modes are all part of `v1`
- observability, provenance, and replay are first-class requirements

---

## Notes

- This design intentionally sets a stronger base than the easiest implementation path.
- The next step after approving this spec is not implementation yet. The next step is to produce the implementation plan for `0.1`.
- The plan should sequence control-plane contracts, runtime scaffolding, raw/staging layers, reconciliation, promotion, review handling, and operational verification in that order.
