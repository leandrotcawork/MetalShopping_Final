# SDK Generation Strategy

## Purpose

Define how generated SDKs and generated types will be produced from `contracts/` and consumed by clients and workers.

## Goals

- one canonical source: `contracts/`
- generated artifacts for TypeScript and Python
- no manual parallel type systems
- reproducible generation workflow
- clear ownership for outputs

## Output targets

### `packages/generated/sdk_ts`

Primary consumers:

- `apps/web`
- `apps/admin_console`
- future TypeScript tooling in the repo

Expected contents:

- API client bindings
- event contract helpers if needed
- typed request and response models

### `packages/generated/types_ts`

Primary consumers:

- frontend type usage that does not need a full client
- UI composition code

Expected contents:

- shared TypeScript types derived from schemas
- event payload type definitions
- governance type definitions

### `packages/generated/sdk_py`

Primary consumers:

- `apps/analytics_worker`
- `apps/integration_worker`
- `apps/automation_worker`
- `apps/notifications_worker`

Expected contents:

- Python models for events and governance
- optional API client surfaces for worker-to-core interactions where contract-driven access is allowed

## Generation rules

- generation always starts from `contracts/`
- generated directories are write targets, not authoring targets
- generated outputs should be reproducible from scripts
- generation scripts belong in `scripts/`
- generation should be all-or-nothing per target to avoid drift

## Initial generation split

### From OpenAPI

Generate:

- TypeScript SDK client
- optional Python API client later if needed

### From JSON Schema

Generate:

- shared TypeScript types
- Python data models where useful for worker payload validation

### From event contracts

Generate:

- TypeScript payload and envelope types
- Python payload and envelope models

## Consumer rules

- frontend code uses generated TS SDKs and types
- workers use generated Python models or SDKs where contract interaction exists
- app code must not handcraft alternative contract definitions

## Versioning strategy

- generation outputs track contract versions explicitly
- breaking contract versions produce explicit output changes
- old generated outputs can coexist temporarily only when needed for migration

## Change workflow

1. update or add contract files
2. review contract changes
3. run generation
4. review generated output deltas
5. update consuming code only after generated artifacts are accepted

## Non-goals for now

- publishing external packages
- multi-language SDK matrix beyond TS and Python
- runtime codegen inside applications

