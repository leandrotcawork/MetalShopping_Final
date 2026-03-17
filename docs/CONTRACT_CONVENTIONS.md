# Contract Conventions

## Purpose

Define how `contracts/` must be organized so APIs, events, and governance remain the single source of truth.

## Global rules

- `contracts/` is canonical
- contract files must be human-reviewable
- naming must be stable and explicit
- versioning must be visible in paths, not hidden in prose
- generated SDKs and generated types derive from these files
- app code must not become a competing contract source

## Folder ownership

### `contracts/api/openapi`

Use this folder for public HTTP contract definitions.

Rules:

- one bounded-context file per public API surface, unless a shared root spec is clearly justified
- filenames use lowercase snake_case
- path versioning is explicit in the contract
- request and response payloads should reference shared JSON Schemas when possible

Recommended naming:

- `iam_v1.openapi.yaml`
- `tenant_admin_v1.openapi.yaml`
- `pricing_v1.openapi.yaml`

### `contracts/api/jsonschema`

Use this folder for shared payload schemas, DTO validation shapes, event payload schemas, and governance schemas that need standalone validation.

Rules:

- filenames use lowercase snake_case
- schema filenames include semantic scope, not implementation detail
- schema ids should be stable
- avoid duplicating the same shape in OpenAPI and JSON Schema by hand when one can reference the other in generation flow

Recommended naming:

- `common_pagination_v1.schema.json`
- `pricing_price_rule_v1.schema.json`
- `market_observation_v1.schema.json`

### `contracts/events/v1`

Use this folder for the first stable generation of domain events.

Rules:

- all event contracts are immutable once published, except for additive compatible evolution inside the same version policy
- breaking changes require a new version path or new event version name according to the event strategy
- events must define producer, trigger, payload, and semantic meaning
- events are integration contracts, not internal code shortcuts

Recommended naming:

- `pricing_price_changed.v1.json`
- `catalog_product_created.v1.json`
- `alerts_alert_raised.v1.json`

### `contracts/events/v2`

Reserved for future incompatible event evolution. Do not use until `v1` has real published contracts that justify a new line.

### `contracts/governance/policies`

Use for policy schemas and policy contract definitions.

Examples:

- authorization policy
- sync policy
- automation eligibility policy

### `contracts/governance/thresholds`

Use for threshold schemas and threshold descriptors.

Examples:

- repricing threshold
- alert severity threshold
- freshness threshold

### `contracts/governance/feature_flags`

Use for feature flag definitions and rollout schemas.

Examples:

- tenant feature gates
- environment gates
- targeted rollout definitions

## Contract metadata minimum

Each contract artifact should make these fields explicit when the format allows it:

- contract name
- version
- owner bounded context
- status
- source folder
- compatibility expectations

## Status vocabulary

Use a small stable vocabulary:

- `draft`
- `proposed`
- `accepted`
- `deprecated`

## Compatibility policy

### API

- additive fields are preferred
- breaking payload or route changes require a new version strategy
- clients should move through generated SDK updates, not ad hoc manual patches

### Events

- published event meaning must remain stable
- adding optional fields is preferred over mutating semantics
- if meaning changes, publish a new version

### Governance

- schema changes must preserve runtime explainability
- override hierarchy semantics must not vary between core and workers

## Relationship to code

- `apps/server_core` implements contracts
- workers consume or publish contracts
- clients consume generated SDKs
- `packages/generated/*` is downstream only

