# Repo SDK Flow

## Read first

1. `docs/SDK_GENERATION_STRATEGY.md`
2. `docs/CONTRACT_CONVENTIONS.md`
3. `packages/generated/README.md`
4. `scripts/generate_contract_artifacts.ps1`

## Files this skill normally touches

- `scripts/generate_contract_artifacts.ps1`
- `packages/generated/sdk_ts/`
- `packages/generated/types_ts/`
- `packages/generated/sdk_py/`

## Repo-specific rules

- generation always starts from `contracts/`
- generated outputs live only in `packages/generated/*`
- OpenAPI feeds SDK clients
- JSON Schema and event contracts feed shared type and model generation
- generated artifacts must remain reproducible and reviewable

## Output mapping

- `sdk_ts`: web and admin client bindings
- `types_ts`: shared TS models for frontend composition
- `sdk_py`: worker-facing Python models or SDKs

