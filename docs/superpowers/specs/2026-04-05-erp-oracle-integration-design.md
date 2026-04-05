# ERP Oracle Integration Design

Date: 2026-04-05
Status: approved design baseline
Scope: ERP Integration 0.1 Oracle connectivity and ingestion runtime for Sankhya-shaped databases

## Purpose

Define the production-grade Oracle integration architecture for MetalShopping ERP ingestion.

This document freezes:

- where Oracle connectivity lives
- how Sankhya-shaped extraction is modeled
- how secrets and connection metadata are handled
- how extraction, staging, reconciliation, and canonical promotion are separated
- how runs fail, retry, checkpoint, and report status

This document does not define every entity-specific Sankhya SQL statement. It defines the architecture and execution model those entity extractors must follow.

## Goals

- Connect MetalShopping to Oracle-backed ERP sources using a production-grade Go driver
- Preserve a generic ERP ingestion architecture instead of hardcoding Sankhya assumptions into shared runtime layers
- Keep Oracle connectivity isolated to `apps/integration_worker`
- Support replay, auditability, safe retries, and clear operational status
- Support future non-Sankhya ERP connectors without reworking the ingestion core

## Non-Goals

- Writing back to the ERP database
- Allowing `apps/server_core` to open Oracle connections
- Building multiple Oracle adapters now
- Implementing generalized live ERP query features in `server_core`
- Defining every final canonical entity mapping in this document

## Frozen Decisions

### 1. Oracle access boundary

Oracle access is allowed only inside `apps/integration_worker`.

`apps/server_core` governs ERP integrations, run visibility, review flows, and canonical application state, but it does not import Oracle client code or open Oracle connections.

### 2. Generic runtime boundary

MetalShopping will implement a generic ERP database source layer under the integration worker. That layer is query-runner-only.

The generic layer owns:

- connection lifecycle
- DSN construction from structured config
- secret resolution hook usage
- query execution
- timeout handling
- row streaming
- typed row reading

The generic layer does not own ERP semantics such as products, prices, inventory, or sales.

### 3. Sankhya ownership model

MetalShopping owns all Sankhya integration code.

There is no dependency on Sankhya application code, SDKs, or runtime components. The Sankhya-specific connector exists only because the source database follows Sankhya naming, table structure, and business conventions.

The Sankhya connector owns:

- SQL statements
- source table understanding
- row-to-raw extraction logic
- normalization rules from Sankhya schema to MetalShopping staging schema

### 4. Oracle adapter choice

Use `godror` as the only Oracle adapter for now.

Rationale:

- production-oriented Oracle connectivity
- aligned with Oracle-native runtime expectations
- stronger long-term fit than a pure-Go shortcut

### 5. Runtime pattern

Use a run-scoped Oracle client/pool.

One sync run:

- resolves the secret once
- builds the `godror` DSN once
- opens one Oracle client/pool
- reuses it across all entities in that run
- closes it at the end of the run

This is preferred over long-lived tenant pools and over per-entity connection lifecycles.

### 6. Pipeline architecture

Use a four-step ingestion pipeline:

1. raw landing
2. normalized staging
3. reconciliation/classification
4. canonical promotion

This is the required architecture for Oracle ERP ingestion.

## Package and Module Boundaries

### Integration worker

Primary package layout:

```text
apps/integration_worker/internal/erp_runtime/
  dbsource/
  dbsource/oracle/
  connectors/sankhya/
  raw/
  staging/
  reconciliation/
  runs/
```

#### `dbsource`

Owns generic query-runner contracts only.

Expected responsibilities:

- query runner interfaces
- typed row-reader interfaces
- execution contracts
- generic source error categories

#### `dbsource/oracle`

Owns the `godror` implementation.

Expected responsibilities:

- structured config to Oracle connection string construction
- runtime secret injection
- `godror` client creation
- run-scoped pool lifecycle
- query execution
- typed row access implementation

#### `connectors/sankhya`

Owns all Sankhya-shaped extraction rules.

Expected responsibilities:

- entity query definitions
- parameter binding strategy
- source cursor semantics
- row-to-raw payload extraction
- normalization rules into MetalShopping staging shape

#### `raw`

Owns raw landing persistence and retrieval.

#### `staging`

Owns normalized staging persistence and retrieval.

#### `reconciliation`

Owns:

- duplicate detection
- dependency-aware entity gating
- classification and review routing
- batch checkpoint semantics

#### `runs`

Owns:

- run header state
- entity-step state
- batch checkpoint state
- final run outcome calculation

### Server core

`apps/server_core/internal/modules/erp_integrations` continues to own:

- integration governance
- operator actions
- review queues and review resolution
- canonical promotion orchestration state exposed to the rest of the platform

It does not open Oracle connections.

## Runtime Flow

### End-to-end run lifecycle

1. A governed ERP sync run is triggered by scheduler or operator action
2. `integration_worker` loads ERP instance metadata
3. `integration_worker` resolves the password through `db_password_secret_ref`
4. The Oracle adapter builds the `godror` connection configuration
5. One run-scoped Oracle client/pool is opened
6. Entities execute sequentially in governed order
7. For each entity:
   - extract rows from Oracle
   - write raw landing records
   - normalize into staging
   - reconcile and classify
   - promote canonical writes in batches
   - record entity and batch results
8. Dependent entities are skipped if a prerequisite entity failed
9. The overall run is marked `completed`, `partial`, or `failed`
10. The Oracle client/pool is closed

### Entity execution order

Initial required order:

1. products
2. prices
3. inventory
4. customers
5. suppliers
6. sales
7. purchases

Rationale:

- product identity must exist before dependent product-linked entities are promoted safely
- pricing and inventory depend on stable product keys
- customers and suppliers are lower-risk but still belong after product footing
- sales and purchases are heavier and come later

### Execution mode

Entity execution is sequential by default.

Parallel execution is explicitly deferred. It may be introduced later only after dependency safety, observability, and Oracle load behavior are proven.

## Connection Configuration and Secrets

### Structured configuration

Oracle connection configuration is stored as structured fields, not as a full DSN string.

Required fields:

- `tenant_id`
- `instance_id`
- `erp_type`
- `db_host`
- `db_port`
- `db_service_name` or `db_sid`
- `db_username`
- `db_password_secret_ref`

Optional fields:

- `connect_timeout_seconds`
- `fetch_batch_size`
- `entity_batch_size`
- `enabled`

Operational fields:

- `last_tested_at`
- `last_sync_at`
- `status`

### Secret handling

MetalShopping stores only the password secret reference.

The actual Oracle password is resolved at runtime by the integration worker through a local environment/config-backed secret resolver for now. The resolver must be interface-based so it can be replaced later by an external secret manager without changing connector logic.

### Security rules

- do not store resolved Oracle passwords in canonical tables
- do not store full resolved DSNs in canonical tables
- do not log passwords
- do not log fully resolved DSNs
- keep Oracle access isolated to the integration worker
- keep sessions run-scoped and short-lived

## Query Runner and Row Reader

### Query runner shape

The generic ERP DB layer is query-runner-only.

It does not expose product- or price-specific APIs. Connectors submit their own SQL and consume typed rows.

### Typed row-reader API

The Oracle adapter returns query results through a typed row-reader abstraction instead of `map[string]any`.

Rationale:

- centralized Oracle type conversion
- reduced per-connector parsing duplication
- safer handling of strings, numerics, nullable values, and timestamps
- clearer connector code

Expected access patterns include:

- `String`
- `NullString`
- `Float64`
- `NullFloat64`
- `Time`
- `NullTime`

The exact method names may vary, but the design intent is fixed: row access is typed and centralized.

## Ingestion Data Model

### Raw landing

Raw landing preserves what Oracle returned at sync time.

Each raw record should contain:

- `tenant_id`
- `run_id`
- `entity_type`
- `source_id`
- `source_cursor`
- `batch_ordinal`
- `extracted_at`
- `raw_payload_json`
- `raw_hash`

Raw landing exists for:

- auditability
- replay
- source-truth inspection
- debugging extraction issues without hitting Oracle again

### Normalized staging

Normalized staging converts ERP-shaped raw data into MetalShopping-shaped staging payloads.

Each staging record should contain:

- `tenant_id`
- `run_id`
- `entity_type`
- `source_id`
- `normalized_payload_json`
- `normalization_status`
- `normalization_errors`
- `derived_keys`
- `batch_ordinal`
- `normalized_at`

The normalized stage is the required handoff point into reconciliation and canonical promotion.

### Payload shape policy

Raw and normalized records store payloads as JSON plus indexed control columns.

MetalShopping should not over-flatten source payloads too early. Values should become first-class columns only when they are needed for:

- orchestration
- filtering
- indexes
- joins
- uniqueness
- canonical write logic

## Reconciliation and Canonical Promotion

### Reconciliation role

Reconciliation sits between normalized staging and canonical promotion.

It owns:

- duplicate detection
- dependency-aware classification
- review routing
- idempotency decisions
- promotion eligibility

### Canonical promotion rule

Canonical writes are committed per entity in bounded batches.

The system must not use one giant transaction for the entire run.

Within an entity:

- large entities use batched transactions with checkpoints
- small entities may use one bounded transaction if size is predictably low

### Entity failure semantics

Failure is entity-scoped.

If one entity fails:

- successful prior entities remain committed
- dependent entities may be marked `skipped_due_to_dependency`
- unrelated completed entities are preserved
- the overall run becomes `partial` unless no meaningful progress occurred

This is the required production behavior.

## Run Status, Checkpoints, and Retry

### Run status

Allowed overall run states:

- `completed`
- `partial`
- `failed`

### Entity status

Allowed entity states:

- `completed`
- `failed`
- `skipped_due_to_dependency`

### Checkpointing

Checkpointing must use both:

- source cursor
- batch ordinal

Source cursor supports correct extraction resume semantics.

Batch ordinal supports run-level auditability, checkpoint replay, and debugging.

### Retry model

Retries must be safe because:

- raw landing is preserved
- normalized staging is preserved
- canonical promotion is batched
- entity and batch state are recorded
- source cursor and batch ordinal are recorded

Retry may resume from:

- extraction
- normalization
- reconciliation
- canonical promotion

depending on the failure boundary.

## Observability

Each run must emit structured operational visibility for:

- run start and finish
- entity start and finish
- batch start and finish
- rows extracted
- rows normalized
- rows reconciled
- rows promoted
- rows rejected or routed to review
- cursor positions
- failure reason category

The system must not collapse partial success into a generic success status.

## Testing Strategy

### Unit tests

Required for:

- Oracle config builder
- typed row-reader conversion behavior
- Sankhya row mapping helpers
- normalization logic
- dependency classification
- checkpoint logic
- retry and idempotency behavior

### Fixture-based connector tests

Use deterministic Sankhya-shaped fixtures to validate:

- product extraction
- price extraction
- inventory extraction
- customer extraction
- supplier extraction
- sales extraction
- purchase extraction

These tests must not require a live Oracle database.

### Integration tests in MetalShopping

Required for:

- raw landing persistence
- normalized staging persistence
- reconciliation outputs
- canonical promotion outputs
- batched entity commit behavior
- run and entity status transitions

### Live connectivity verification

Live Oracle verification is a separate operator-level capability, not the basis of the routine automated test suite.

It may include:

- connection validation
- lightweight query validation
- small extraction smoke verification

## Implementation Constraints

- Oracle code remains isolated to `apps/integration_worker`
- `server_core` must not import Oracle client packages
- the generic DB layer must remain query-runner-only
- Sankhya SQL and mappings must remain in the Sankhya connector
- raw landing and normalized staging are both required
- canonical writes must be batched and checkpointed
- run status must support `completed`, `partial`, and `failed`

## Result

This design establishes a production-oriented Oracle ingestion architecture that is:

- secure
- replayable
- auditable
- dependency-aware
- compatible with the current MetalShopping architecture
- scalable toward future ERP connectors without weakening boundaries
