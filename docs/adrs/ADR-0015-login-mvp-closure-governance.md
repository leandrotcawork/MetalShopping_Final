# ADR-0015: Login MVP Closure Governance

- Status: accepted
- Date: 2026-03-18

## Context

The login implementation evolved significantly, but closure criteria remained implicit. This repeatedly reopened the same architecture concerns in review cycles, especially around SDK/runtime boundary semantics, auth/session quality gates, and scope control.

## Decision

- The login tranche is now governed by explicit frozen references:
  - `docs/LOGIN_MVP_SCOPE.md`
  - `docs/LOGIN_DOD.md`
  - `docs/LOGIN_MVP_EXECUTION_PLAN.md`
- Login cannot be declared complete unless every DoD item is checked.
- New authenticated surfaces remain blocked until login DoD is complete.
- Out-of-scope items for this tranche stay explicitly out of scope:
  - production issuer/JWKS finalization beyond current baseline
  - full broker/consumer rollout for auth events

## Consequences

- Reviews now use one deterministic completion contract instead of ad hoc criteria.
- Scope creep is reduced and execution can move tranche-by-tranche.
- Login closure status becomes auditable in CI and documentation.

## Follow-up

- Apply T1/T2/T3 from `docs/LOGIN_MVP_EXECUTION_PLAN.md`
- Keep `PROJECT_SOT`, `IMPLEMENTATION_PLAN`, and `PROGRESS` synchronized in each structural login tranche
