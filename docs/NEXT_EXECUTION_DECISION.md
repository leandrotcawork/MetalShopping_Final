# Next Execution Decision

## Decision

The next implementation area is login MVP closure governance and execution (T1/T2/T3) before opening new authenticated surfaces.

## Why

- local issuer bootstrap and login flow are already operational
- the next structural risk is repeated review loops without a single closure contract
- closing scope, DoD, and SDK boundary rules now removes ad hoc re-analysis churn

## Constraints

This decision is valid only if planning and implementation follow:

- `docs/LOGIN_AND_IDENTITY_ARCHITECTURE.md`
- `docs/LOGIN_MVP_SCOPE.md`
- `docs/LOGIN_DOD.md`
- `docs/LOGIN_MVP_EXECUTION_PLAN.md`
- `docs/SDK_BOUNDARY.md`
- `docs/SDK_GENERATION_STRATEGY.md`
- `docs/PROJECT_SOT.md`
- `docs/PROGRESS.md`

## Explicit rejection

Do not jump next to:

- opening additional authenticated surfaces before `docs/LOGIN_DOD.md` is fully checked
- reintroducing SDK transport emission in generation scripts
- allowing generated/runtime boundary ambiguity in frontend auth path
- treating login closure as subjective instead of checklist-governed

until T3 closure validation is complete.
