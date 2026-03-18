# ADR-0010: Initial Identity Provider is Keycloak

## Status

Accepted

## Context

MetalShopping now has the backend-owned web session boundary and OIDC-based login
model frozen. The next open choice is which real provider should back the first
issuer integration.

The project does not yet have:

- an existing enterprise IdP
- an existing SSO provider contract
- a pre-existing organizational identity stack to integrate with immediately

## Decision

Keycloak is the accepted initial identity provider for MetalShopping.

It must be used as the first real issuer for:

- local development
- staging identity validation
- initial login and session smoke tests

## Why

Keycloak is the best initial fit because it:

- supports OIDC correctly
- supports enterprise features such as MFA and identity brokering
- avoids building custom auth in the app
- is open source and operationally flexible
- keeps the architecture provider-agnostic for future migration or federation

## Consequences

### Positive

- the team can validate the real OIDC flow early
- the same realm and client concepts can be promoted through environments
- the app remains independent from a paid SaaS identity lock-in at the start

### Constraints

- realm and client setup become part of the local platform bootstrap
- the project must document Keycloak realm, client, mapper, and user bootstrap clearly

### Explicit rejection

Do not choose any of the following as the first long-term identity solution:

- app-owned password login
- ad hoc local JWT minting as the primary path
- a frontend-only auth library without backend session ownership

## Migration note

This ADR chooses Keycloak as the initial provider, not the permanent provider forever.
The architecture must remain portable enough that Entra ID, Auth0, Okta, or another
enterprise IdP can replace or federate with Keycloak later without changing the
thin-client and session-boundary model.
