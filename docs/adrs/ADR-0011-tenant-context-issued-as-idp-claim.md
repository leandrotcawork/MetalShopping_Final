# ADR-0011: Tenant Context Issued as IdP Claim

## Status

Accepted

## Context

`server_core` requires `tenant_id` in the principal so tenancy middleware can build
the effective tenant context for every authenticated request.

The login flow now depends on an external identity provider, which means the project
must define how tenant context reaches the core.

The initial options were:

- resolve tenant from an explicit token claim
- resolve tenant from an internal user-to-tenant lookup after login
- derive tenant indirectly from routing or frontend state

## Decision

The initial accepted model is that the external identity provider issues `tenant_id`
as an explicit claim in the authenticated token used by `server_core`.

`server_core` must read `tenant_id` from the validated token and populate the
authenticated principal from that value.

## Why

This is the cleanest first production path because it:

- keeps tenancy deterministic during the auth/session flow
- avoids a second hidden resolution step before principal creation
- keeps the principal model simple and explicit
- aligns well with the current multitenant runtime model

## Consequences

### Positive

- auth and tenancy remain tightly aligned
- login bootstrap is simpler to validate
- the session service can remain stateless with respect to tenant resolution at login time

### Constraints

- the Keycloak realm and client setup must include a claim mapper for `tenant_id`
- test users must carry valid tenant claim values
- issuer validation alone is not enough; claim presence and shape matter too

### Future evolution

This ADR does not forbid a future internal user-to-tenant lookup model.
It only freezes the first professional implementation path so login can ship on a
clean and deterministic basis.

### Explicit rejection

Do not derive effective tenancy from:

- browser-selected tenant state
- URL-only hints
- silent frontend storage
- undocumented post-login heuristics
