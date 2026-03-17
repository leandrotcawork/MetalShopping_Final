# Repo Event Flow

## Read first

1. `docs/PROJECT_SOT.md`
2. `docs/CONTRACT_CONVENTIONS.md`
3. `docs/SDK_GENERATION_STRATEGY.md`
4. `contracts/events/v1/_template.event.json`

## Files this skill normally touches

- `contracts/events/v1/<event_name>.v1.json`
- `contracts/api/jsonschema/*.schema.json`

## Repo-specific rules

- Event contracts are authored only in `contracts/events/`.
- Event payload schemas should live in `contracts/api/jsonschema/` when shared or validated independently.
- Generated TS and Python artifacts are downstream outputs in `packages/generated/`.
- Event contracts must support async decoupling and traceability.

## Naming guidance

- file name: lowercase snake_case plus explicit `.v1.json`
- examples: `pricing_price_changed.v1.json`, `alerts_alert_raised.v1.json`
- event names should remain semantically stable after publication

