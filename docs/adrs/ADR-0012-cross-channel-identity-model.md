# ADR-0012: Cross-Channel Identity Model

## Status

Accepted

## Context

MetalShopping will not remain only a browser application.
The platform direction already includes:

- `web`
- `desktop`
- future app/native surfaces

The login model chosen for the web must not trap the platform in a browser-only
identity architecture.

## Decision

MetalShopping must use a cross-channel identity model with a shared external IdP and
channel-specific session transport.

The accepted channel model is:

- `web`: OIDC plus backend-owned `HttpOnly` cookie session
- native or desktop later: OIDC Authorization Code + PKCE with channel-appropriate token/session handling
- `server_core` remains the owner of principal creation, tenancy, and authorization

## Why

This preserves:

- a single identity source
- a single issuer validation model
- a single IAM and tenancy boundary in `server_core`
- portability across UI surfaces

while still allowing the best transport for each client type.

## Consequences

### Positive

- the web can use secure browser session cookies
- future native clients do not need to inherit the browser cookie model
- the core does not need different authorization semantics per client

### Constraints

- client types must be registered separately at the IdP
- the web and native flows must share issuer rules but not necessarily identical transport behavior
- frontend and app surfaces must remain thin with respect to authorization

### Explicit rejection

Do not treat browser cookie session behavior as the universal session design for all clients.
Do not move identity semantics into each client just because transport differs by channel.
