# Repo Module Flow

## Read first

1. `docs/MODULE_STANDARDS.md`
2. `docs/PLATFORM_BOUNDARIES.md`
3. `docs/READMODEL_AND_EVENTS_RULES.md`
4. `docs/MODULE_CREATION_CHECKLIST.md`
5. `apps/server_core/internal/modules/_template/README.md`

## Files this skill normally touches

- `apps/server_core/internal/modules/<module_name>/`
- `contracts/` when the module needs API, event, or governance contracts

## Repo-specific rules

- business capabilities belong in `internal/modules/`
- runtime capabilities belong in `internal/platform/`
- low-semantic helpers may belong in `internal/shared/`
- module structure is fixed and not optional
- event contracts live in `contracts/events/`, not inside the module tree

## Decision questions

- is this truly a bounded business capability?
- does this module own canonical write semantics?
- does it need `readmodel/`, `events/`, or both?
- are any required contracts missing?

