# Next Execution Decision

## Decision

The next implementation area should be the backend-owned web auth/session foundation before the login surface.

## Why

- `Products` now exists as the first real thin-client operational surface
- the next structural gap is authenticated browser session handling, not another unauthenticated screen
- the web must not grow route-by-route on static bearer bootstrap behavior
- login must be implemented on top of a frozen session model, not as a page-local form flow

## Constraints

This decision is valid only if planning and implementation follow:

- `docs/WEB_AUTH_SESSION_IMPLEMENTATION_PLAN.md`
- `docs/adrs/ADR-0005-thin-clients-and-generated-sdks.md`
- `docs/adrs/ADR-0009-web-session-boundary-on-oidc-and-http-only-cookies.md`
- `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`

## Explicit rejection

Do not jump next to:

- implementing the login screen before the `auth/session` contract is frozen
- storing browser tokens in `localStorage` as the long-term session model
- pushing issuer, callback, or claim logic into React pages
- opening the next operational surface before authenticated shell behavior is ready

until the backend-owned web session boundary is frozen.
