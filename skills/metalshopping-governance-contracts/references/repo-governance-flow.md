# Repo Governance Flow

## Read first

1. `docs/PROJECT_SOT.md`
2. `docs/CONTRACT_CONVENTIONS.md`
3. `docs/adrs/ADR-0004-runtime-governance.md`
4. matching template in `contracts/governance/`

## Files this skill normally touches

- `contracts/governance/policies/*.json`
- `contracts/governance/thresholds/*.json`
- `contracts/governance/feature_flags/*.json`
- `contracts/api/jsonschema/*.schema.json`

## Repo-specific rules

- Governance artifacts are authored only in `contracts/governance/`.
- Defaults live in `bootstrap/seeds/governance/`, but runtime semantics are defined by governance contracts and platform resolution.
- Core and workers must be able to interpret the same scope semantics.
- Governance artifacts are downstream inputs for runtime resolution, not app-local config shortcuts.

## Naming guidance

- use lowercase snake_case file names
- keep artifact names stable and semantic
- keep version explicit in content even if the file name is simpler

