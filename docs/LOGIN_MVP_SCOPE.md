# Login MVP Scope

## Purpose

Freeze the exact scope used to close the login tranche without reopening unrelated platform work.

## In scope

- OIDC login flow owned by `server_core`
- `HttpOnly` web session cookie model
- CSRF protection for cookie-backed mutation endpoints
- frontend thin-client auth bootstrap and route guard behavior
- stable frontend SDK runtime boundary for auth/session consumption
- login visual consistency across React fallback and Keycloak login theme
- CI and smoke gates required to declare login complete

## Out of scope for this tranche

- production IdP migration strategy beyond current Keycloak baseline
- full broker/consumer rollout for all auth events
- broader post-login product surfaces unrelated to auth/session closure
- advanced SSO and federation scenarios

## Scope guardrails

- no feature may redefine auth/session contracts outside `contracts/`
- no page-level manual auth transport logic
- no new authenticated surface should open before this tranche is closed

## Completion contract

Login is considered complete only when every item in `docs/LOGIN_DOD.md` is checked and validated in CI plus local smoke.
