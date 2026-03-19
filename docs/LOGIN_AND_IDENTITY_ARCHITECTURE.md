# Login and Identity Architecture

## Purpose

This document is the complete working reference for MetalShopping login, identity,
and authenticated session behavior across `web` today and future app surfaces later.

It exists so the team can implement login without reopening the identity model,
mixing browser concerns with platform concerns, or drifting into shortcuts that
will not scale for the future MetalShopping product.

## Outcome we want

MetalShopping must use a professional external identity layer, with the application
core owning session, authorization, tenancy, and auditability.

The long-term target is:

- external OIDC identity provider
- `server_core` as the authenticated session boundary
- thin clients for `web`, `desktop`, and future native surfaces
- internal IAM continuing to own MetalShopping roles and capabilities

## Core concepts

### Identity Provider

An Identity Provider, or IdP, is the system responsible for:

- authenticating users
- handling passwords and MFA
- issuing identity tokens
- supporting SSO and logout flows
- centralizing enterprise identity policy

Examples:

- Keycloak
- Microsoft Entra ID
- Auth0
- Okta
- Amazon Cognito

### OIDC

OIDC is the identity protocol used by modern applications to sign users in through
an external provider.

For MetalShopping, OIDC gives us:

- a standard login flow
- issuer and key validation
- a path for enterprise SSO later
- a model that supports both browser and native clients

## Frozen architecture

### Web

The browser does not own token parsing or authorization semantics.

The accepted model is:

- browser starts login
- `server_core` redirects to the IdP
- `server_core` handles the callback
- `server_core` creates an `HttpOnly` session cookie
- `apps/web` bootstraps with `GET /api/v1/auth/session/me`

### Core

`server_core` owns:

- login start
- callback
- session creation
- session refresh
- logout
- principal creation
- tenancy context
- authorization checks
- observability and audit behavior

### IAM

The external IdP does not replace MetalShopping IAM.

The IdP owns identity.
MetalShopping IAM still owns:

- internal roles
- internal capabilities
- backend authorization semantics

### Future app surfaces

The same identity foundation must support future clients.

Accepted channel strategy:

- `web`: OIDC plus `HttpOnly` cookie session
- `desktop` and native app later: OIDC Authorization Code + PKCE using native client semantics
- same issuer
- same identity
- same tenancy and IAM model in `server_core`

## Chosen initial provider

### Initial decision

Keycloak is the accepted initial IdP for MetalShopping.

### Why Keycloak

- open source and mature
- supports OIDC correctly
- supports SSO, MFA, identity brokering, and enterprise growth
- good local and staging story before the company has a formal identity stack
- keeps us from inventing login internally
- can later federate with Microsoft or other enterprise providers

### Why not build login ourselves

We explicitly reject:

- custom user/password auth owned by the app as the long-term solution
- localStorage token persistence
- page-local identity logic
- app-specific ad hoc login flows

## Tenant strategy

MetalShopping is tenant-aware at the core.

The current and accepted login strategy is:

- `tenant_id` must be available when building the principal for an authenticated session
- the initial production path should use an explicit claim from the IdP token
- the principal created by `server_core` must carry `tenant_id`

This is the cleanest path for the first real login rollout because it keeps tenant
resolution deterministic and avoids a second hidden lookup layer during the initial
session bootstrap.

## Authenticated session model

### Session start

- browser calls `GET /api/v1/auth/session/login`
- `server_core` creates state and PKCE verifier
- `server_core` redirects to the IdP

### Callback

- IdP redirects to `GET /api/v1/auth/session/callback`
- `server_core` validates state
- `server_core` exchanges authorization code for token
- `server_core` validates issuer, audience, expiry, keys, and tenant claim
- `server_core` creates a web session row and sets an `HttpOnly` cookie

### Session bootstrap

- `apps/web` calls `GET /api/v1/auth/session/me`
- backend returns:
  - user identity
  - tenant id
  - roles
  - capabilities
  - expiry metadata

### Refresh

- `POST /api/v1/auth/session/refresh`
- backend rotates or extends the session within governance rules

### Logout

- `POST /api/v1/auth/session/logout`
- backend invalidates the session
- backend clears cookie state

## Runtime governance

The session flow is intentionally governed.

Current governed controls:

- `auth.web_session_enabled`
- `auth.session_idle_timeout_minutes`
- `auth.session_absolute_timeout_minutes`

This keeps session behavior out of hardcoded app logic and aligned with the broader
runtime governance model.

## Security baseline

The login foundation must always preserve the following:

- no access token persistence in browser storage
- issuer validation
- audience validation
- expiry validation
- JWKS or public key based signature validation
- secure cookie behavior by environment
- no secret or token leakage in logs
- correlation id propagation
- auditable login and logout transitions
- CSRF-aware handling for cookie-backed state changes

## Environment model

### Local

Use Keycloak locally as the real issuer.

This is the preferred path because it validates the actual login architecture end to end.

### Staging

Use a dedicated staging realm and client registration.

### Production

Production may remain on Keycloak or move to another provider later, but the core
and thin-client model should not need architectural rework when that happens.

## Required local Keycloak setup

The next execution slice should create:

- realm: `metalshopping`
- web client: `metalshopping-web`
- redirect URI:
  - `http://127.0.0.1:8080/api/v1/auth/session/callback`
- logout and web return targets as needed
- user claims:
  - `sub`
  - `email`
  - `name`
  - `tenant_id`

## Environment variables explained

### JWT / issuer validation

- `MS_AUTH_MODE`
- `MS_AUTH_JWT_ALGORITHM`
- `MS_AUTH_JWT_ISSUER`
- `MS_AUTH_JWT_AUDIENCE`
- `MS_AUTH_JWT_PUBLIC_KEY_PEM`
- `MS_AUTH_JWT_JWKS_URL`

### OIDC flow

- `MS_AUTH_OIDC_AUTHORIZATION_URL`
- `MS_AUTH_OIDC_TOKEN_URL`
- `MS_AUTH_OIDC_CLIENT_ID`
- `MS_AUTH_OIDC_CLIENT_SECRET`
- `MS_AUTH_OIDC_REDIRECT_URI`
- `MS_AUTH_OIDC_SCOPES`
- `MS_AUTH_OIDC_HTTP_TIMEOUT_SECONDS`

### Web session and cookies

- `MS_AUTH_WEB_SESSION_MODE` (`required`, `optional`, `disabled`)
  - default: `optional` when `MS_AUTH_MODE=static`, otherwise `required`
- `MS_AUTH_WEB_DEFAULT_RETURN_TO`
- `MS_AUTH_WEB_LOGIN_STATE_TTL_MINUTES`
- `MS_AUTH_WEB_SESSION_COOKIE_NAME`
- `MS_AUTH_WEB_STATE_COOKIE_NAME`
- `MS_AUTH_WEB_COOKIE_DOMAIN`
- `MS_AUTH_WEB_COOKIE_PATH`
- `MS_AUTH_WEB_COOKIE_SAMESITE`

## Execution plan

### Phase 1

Freeze identity and login architecture.

This is complete.

### Phase 2

Implement `auth/session` contracts and core runtime.

This is complete enough to proceed.

### Phase 3

Stand up Keycloak locally and configure the first real issuer integration.

This is complete.

### Phase 4

Run backend smoke for:

- login
- callback
- `me`
- refresh
- logout

This is complete for the current local bootstrap.

### Phase 5

Implement the login screen and session bootstrap in `apps/web`.

## Go / No-Go before login UI

Do not start the final login screen until:

- Keycloak local is running
- issuer metadata is configured in `.env`
- tenant claim is flowing in the token
- backend session endpoints are smoke-tested with the real issuer

## Relationship to other docs

- `docs/WEB_AUTH_SESSION_IMPLEMENTATION_PLAN.md`
- `docs/adrs/ADR-0009-web-session-boundary-on-oidc-and-http-only-cookies.md`
- `docs/adrs/ADR-0010-initial-identity-provider-keycloak.md`
- `docs/adrs/ADR-0011-tenant-context-issued-as-idp-claim.md`
- `docs/adrs/ADR-0012-cross-channel-identity-model.md`
- `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
