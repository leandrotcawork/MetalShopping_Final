# Repo Worker Flow

## Read first

1. `docs/WORKER_OPERATING_MODEL.md`
2. `docs/CONTRACT_EVOLUTION_RULES.md`
3. `docs/READMODEL_AND_EVENTS_RULES.md`
4. `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`

## Repo-specific rules

- workers are async-first
- workers do not own canonical write semantics
- contract boundaries remain explicit
- observability and correlation must survive async boundaries

