# Next Execution Decision

## Decision

The next implementation area should be real local issuer bootstrap with Keycloak before the login surface.

## Why

- the backend-owned web session runtime already exists
- the next structural gap is connecting that runtime to a real issuer instead of bootstrap auth
- login UI should not start before realm, client, claims, and callback wiring are real
- the future app strategy depends on choosing the initial IdP and cross-channel identity model correctly

## Constraints

This decision is valid only if planning and implementation follow:

- `docs/LOGIN_AND_IDENTITY_ARCHITECTURE.md`
- `docs/WEB_AUTH_SESSION_IMPLEMENTATION_PLAN.md`
- `docs/adrs/ADR-0009-web-session-boundary-on-oidc-and-http-only-cookies.md`
- `docs/adrs/ADR-0010-initial-identity-provider-keycloak.md`
- `docs/adrs/ADR-0011-tenant-context-issued-as-idp-claim.md`
- `docs/adrs/ADR-0012-cross-channel-identity-model.md`

## Explicit rejection

Do not jump next to:

- implementing the login screen before Keycloak local is running
- inventing app-owned credential auth because the IdP is not configured yet
- pushing tenant resolution into the frontend
- opening the next operational surface before authenticated shell behavior is real

until the real local issuer foundation is in place.
