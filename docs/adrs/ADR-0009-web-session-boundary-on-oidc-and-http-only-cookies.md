# ADR-0009: Web Session Boundary on OIDC and HttpOnly Cookies

## Status

Accepted

## Context

MetalShopping now has the first real thin-client web surface running against `server_core`.
The next implementation gate is authenticated web session handling before the login screen
is implemented.

The repository already freezes the following relevant rules:

- thin clients consume generated SDKs and backend-owned read surfaces
- the core owns auth, authz, tenancy, governance, and auditability
- observability and security must stay centralized in the platform baseline

The remaining question is how the web should authenticate without pushing identity logic,
token parsing, or authorization semantics into the frontend.

## Decision

MetalShopping web authentication must use an external OIDC identity provider with
`server_core` as the web session boundary.

The accepted model is:

- external IdP using OIDC Authorization Code + PKCE
- `server_core` initiates the login redirect and handles the OIDC callback
- `server_core` creates and rotates the web session
- the browser receives a secure `HttpOnly` session cookie
- `apps/web` remains a thin client and bootstraps state from `GET /api/v1/auth/session/me`
- the frontend must not persist access tokens in `localStorage`, `sessionStorage`, or ad hoc caches

## Why

This model is the best fit for the long-term MetalShopping target because it:

- keeps identity and token handling out of the browser application code
- centralizes auth, authz, tenancy, audit, and observability in the core
- aligns with enterprise web application patterns used by large organizations
- scales better for future `web`, `desktop`, and `admin_console` clients
- avoids coupling route components to raw issuer or token details
- preserves a thin-client contract-first frontend model

## Consequences

### Positive

- the web login flow stays compatible with generated SDKs and backend-owned contracts
- session rotation, timeout, revocation, and observability stay server-owned
- CORS, cookie, correlation, and audit behavior can be governed centrally
- route guards in `apps/web` can remain simple and session-state driven

### Constraints

- the next auth implementation step must start from `auth/session` contracts, not from a page-only login form
- the production path must validate issuer, audience, expiry, and keys from a real OIDC/JWKS source
- cookie settings must be explicit and environment-aware
- login, logout, refresh, and `me` must be observable and auditable

### Explicit rejection

Do not adopt any of the following as the long-term web auth model:

- frontend-owned token parsing and claim validation
- access tokens stored in `localStorage`
- hand-maintained parallel DTOs for auth/session state
- login logic implemented only inside the React page layer
- ad hoc static bearer handling as the final web session model

## Implementation direction

The implementation must proceed in this order:

1. freeze the `auth/session` HTTP contract
2. freeze session governance artifacts for enablement and timeout thresholds
3. implement platform packages in `server_core` for OIDC settings, cookie sessions, and web-session observability
4. wire session bootstrap in `apps/web`
5. only then implement the login and logout UI
