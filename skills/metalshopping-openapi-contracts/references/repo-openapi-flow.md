# Repo OpenAPI Flow

## Read first

1. `docs/PROJECT_SOT.md`
2. `docs/CONTRACT_CONVENTIONS.md`
3. `docs/SDK_GENERATION_STRATEGY.md`
4. `contracts/api/openapi/_template.openapi.yaml`

## Files this skill normally touches

- `contracts/api/openapi/<bounded_context>_v1.openapi.yaml`
- `contracts/api/jsonschema/*.schema.json`

## Repo-specific rules

- OpenAPI is authored only in `contracts/api/openapi/`.
- Shared payload shapes belong in `contracts/api/jsonschema/`.
- Generated SDKs are downstream artifacts in `packages/generated/`.
- The contract must support thin clients and generated consumption.

## Naming guidance

- file name: lowercase snake_case
- examples: `iam_v1.openapi.yaml`, `tenant_admin_v1.openapi.yaml`
- operation IDs should be stable and explicit

