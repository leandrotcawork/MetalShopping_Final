# Web Auth Session Implementation Plan

## Goal

Implement MetalShopping web authentication as a professional session boundary owned by
`server_core`, using an external OIDC identity provider and secure `HttpOnly` cookies.

## Architecture target

- OIDC Authorization Code + PKCE with an external identity provider
- `server_core` handles login start, callback, refresh, logout, and session introspection
- `apps/web` consumes generated contracts and bootstraps from `auth/session`
- authorization remains backend-owned
- audit, correlation, observability, and abuse controls remain platform-owned

## Scope freeze

This plan is for the web authentication boundary only.

It is not the phase to:

- add app-local credential storage
- build custom identity management
- move authorization logic into React components
- implement admin user management UX

## Phase 1: Design and contracts

Deliverables:

- ADR for the web session boundary
- `auth_session_v1.openapi.yaml`
- JSON Schemas for session responses
- governance artifacts for session enablement and timeouts

Exit criteria:

- session lifecycle is explicit
- cookie session model is frozen
- the frontend has a single source of truth for session contracts

## Phase 2: Platform runtime in `server_core`

Deliverables:

- OIDC runtime config package
- issuer and JWKS validation path
- cookie session package
- session middleware and principal rehydration
- login, callback, logout, refresh, and `me` handlers

Exit criteria:

- login flow works against a real issuer configuration
- `HttpOnly` session cookies are issued and rotated
- session timeout resolution comes from governance-aware runtime configuration
- auth/session endpoints are observable and auditable

## Phase 3: Frontend session bootstrap

Deliverables:

- session bootstrap provider in `apps/web`
- route guards for authenticated surfaces
- logout wiring through generated SDK
- login screen and logged-out landing state

Exit criteria:

- `apps/web` does not parse or persist tokens
- authenticated routes depend on `GET /api/v1/auth/session/me`
- login and logout are fully contract-driven

## Session lifecycle

### Start login

- browser navigates to `GET /api/v1/auth/session/login`
- `server_core` creates PKCE/state material and redirects to the IdP

### Callback

- IdP returns to `GET /api/v1/auth/session/callback`
- `server_core` validates state, exchanges code, validates issuer claims, and creates a web session
- `server_core` sets `HttpOnly` cookie and redirects to the web return target

### Session bootstrap

- `apps/web` calls `GET /api/v1/auth/session/me`
- the response defines identity, tenant context, capabilities, and session expiry metadata

### Refresh

- `POST /api/v1/auth/session/refresh` rotates the server-owned session if still valid
- refresh behavior follows governance thresholds and runtime policy

### Logout

- `POST /api/v1/auth/session/logout` revokes the active session, clears cookie state, and records the event

## Governance

Runtime behavior must be controlled by contracts, not hardcoded values.

Required artifacts:

- `auth.web_session_enabled`
- `auth.session_idle_timeout_minutes`
- `auth.session_absolute_timeout_minutes`

These artifacts must resolve through the existing governance hierarchy and be usable by the
web session runtime without frontend duplication.

## Observability and security

Required controls:

- correlation id propagation through login, callback, `me`, refresh, and logout
- structured logs without token or secret leakage
- metrics for success/failure of login, callback, refresh, and logout
- audit records for security-relevant session transitions
- environment-aware cookie configuration
- issuer, audience, expiry, and clock-skew validation
- CSRF protection for state-changing cookie-backed endpoints
- backend-owned double-submit CSRF cookie plus required request header for cookie-backed mutations
- trusted-origin validation for browser-driven cookie-backed mutations
- rate limiting and abuse protection on login and refresh paths

## Web client rules

- the web remains a thin client
- no access token persistence in browser storage
- no claim parsing inside feature packages
- no page-local auth transport wrappers
- generated SDK remains the only client contract source

## Execution order

1. author contracts and governance artifacts
2. implement OIDC runtime config and cookie session platform packages
3. implement `auth/session` endpoints in `server_core`
4. wire generated SDK support for the new surface
5. stand up the first real local issuer with Keycloak
6. configure realm, client, mapper, and test users with `tenant_id`
7. add `SessionProvider`, guards, and login/logout UI in `apps/web`
8. run smoke tests against the real local issuer configuration

## Local issuer bootstrap

The accepted first issuer bootstrap is:

- Keycloak
- realm `metalshopping`
- client `metalshopping-web`
- redirect URI `http://127.0.0.1:8080/api/v1/auth/session/callback`
- token claim `tenant_id`

This bootstrap must happen before the final login UI slice starts.

## Go / No-Go gate before UI login

Do not begin the login page implementation until:

- the `auth/session` contract is frozen
- governance artifacts are validated
- the session platform package boundaries are explicit
- observability and security controls are defined for the flow
