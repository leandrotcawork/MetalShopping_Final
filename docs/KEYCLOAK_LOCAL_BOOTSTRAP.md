# Keycloak Local Bootstrap

## Purpose

This runbook bootstraps the first real local OIDC issuer for MetalShopping.

It exists so the team can move from the architectural login model to a real
issuer-backed flow without improvising realm settings, client registration, or
tenant claim behavior.

## Local target

- Keycloak URL: `http://127.0.0.1:18081`
- realm: `metalshopping`
- OIDC client: `metalshopping-web`
- callback handled by `server_core`
- user claim: `tenant_id`

## What this bootstrap creates

The local Keycloak setup provisions:

- one Keycloak realm: `metalshopping`
- one confidential OIDC client: `metalshopping-web`
- one `tenant_id` claim mapper
- two users with fixed subject ids:
  - `ms_admin`
  - `ms_viewer`

## Included local users

### Admin user

- username: `ms_admin`
- password: `ChangeMe123!`
- subject id: `11111111-1111-1111-1111-111111111111`
- tenant_id: `bootstrap-local`

### Viewer user

- username: `ms_viewer`
- password: `ChangeMe123!`
- subject id: `22222222-2222-2222-2222-222222222222`
- tenant_id: `bootstrap-local`

## Why fixed subject ids matter

MetalShopping IAM stores role assignments by internal subject id.
The fixed user ids in the imported realm allow local role bootstrap to be repeatable
without querying Keycloak dynamically after each container reset.

## Files

- compose: `ops/keycloak/docker-compose.yml`
- realm import: `ops/keycloak/import/metalshopping-realm.json`
- login theme: `ops/keycloak/themes/metalshopping/login`
- start script: `scripts/start_keycloak_local.ps1`
- stop script: `scripts/stop_keycloak_local.ps1`
- theme apply script: `scripts/apply_keycloak_theme_local.ps1`
- IAM bootstrap script: `scripts/bootstrap_keycloak_local_iam.ps1`

## Bootstrap order

### 1. Apply the auth session migrations completely

Before starting the real issuer flow, ensure all auth session migrations are applied:

- `0016_auth_web_sessions.sql`
- `0017_auth_session_governance_defaults.sql`
- `0018_auth_web_session_runtime_grants.sql`
- `0019_auth_session_feature_flag_repair.sql`

The grants migration is required when the tables were created by a superuser but
`server_core` runs with the application role `metalshopping_app`.
The repair migration is required because the initial auth session governance seed
used the wrong feature-flag value column for the runtime resolver.

### 2. Start Keycloak

Run:

```powershell
.\scripts\start_keycloak_local.ps1
```

The admin console becomes available at:

- `http://127.0.0.1:18081/admin`

Admin credentials:

- username: `admin`
- password: `admin`

## 3. Confirm realm import

Open the admin console and confirm that the realm `metalshopping` exists.

The realm import is startup-based. If the realm already exists, Keycloak skips import.

The imported realm is configured to use the `metalshopping` login theme. The
theme files are mounted from the repository into the Keycloak container.

If the realm already existed before the theme was added or changed, run:

```powershell
.\scripts\apply_keycloak_theme_local.ps1
```

If the container was already running before the theme volume changed, restart
Keycloak first:

```powershell
.\scripts\stop_keycloak_local.ps1
.\scripts\start_keycloak_local.ps1
```

## 4. Bootstrap IAM role assignments while `server_core` is still in static mode

Before switching `server_core` to JWT/OIDC mode, use the existing static bootstrap
token to create internal IAM role assignments for the imported Keycloak subjects.

Run:

```powershell
.\scripts\bootstrap_keycloak_local_iam.ps1
```

This creates:

- `admin` role for the Keycloak admin subject
- `viewer` role for the Keycloak viewer subject

These assignments are stored in MetalShopping Postgres and remain valid after the
backend switches from static auth to JWT mode.

## 5. Switch `.env` to OIDC/JWT mode

Use these values in `.env`:

```env
MS_AUTH_MODE=jwt
MS_AUTH_JWT_ALGORITHM=RS256
MS_AUTH_JWT_ISSUER=http://127.0.0.1:18081/realms/metalshopping
MS_AUTH_JWT_AUDIENCE=metalshopping-web
MS_AUTH_JWT_JWKS_URL=http://127.0.0.1:18081/realms/metalshopping/protocol/openid-connect/certs

MS_AUTH_OIDC_AUTHORIZATION_URL=http://127.0.0.1:18081/realms/metalshopping/protocol/openid-connect/auth
MS_AUTH_OIDC_TOKEN_URL=http://127.0.0.1:18081/realms/metalshopping/protocol/openid-connect/token
MS_AUTH_OIDC_CLIENT_ID=metalshopping-web
MS_AUTH_OIDC_CLIENT_SECRET=metalshopping-web-secret-change-me
MS_AUTH_OIDC_REDIRECT_URI=http://127.0.0.1:8080/api/v1/auth/session/callback
MS_AUTH_OIDC_SCOPES=openid profile email
MS_AUTH_OIDC_HTTP_TIMEOUT_SECONDS=10

MS_AUTH_WEB_DEFAULT_RETURN_TO=/products
MS_AUTH_WEB_LOGIN_STATE_TTL_MINUTES=10
MS_AUTH_WEB_SESSION_COOKIE_NAME=ms_web_session
MS_AUTH_WEB_STATE_COOKIE_NAME=ms_web_login_state
MS_AUTH_WEB_CSRF_COOKIE_NAME=ms_web_csrf
MS_AUTH_WEB_CSRF_HEADER_NAME=X-CSRF-Token
MS_AUTH_WEB_COOKIE_DOMAIN=
MS_AUTH_WEB_COOKIE_PATH=/
MS_AUTH_WEB_COOKIE_SAMESITE=lax
```

## 6. Restart `server_core`

After updating `.env`, restart the backend.

## 7. Smoke the session surface

Validate:

- `GET /api/v1/auth/session/login`
- `GET /api/v1/auth/session/callback`
- `GET /api/v1/auth/session/me`
- `POST /api/v1/auth/session/refresh`
- `POST /api/v1/auth/session/logout`

## Notes

- `server_core` listens on `:8080`
- Keycloak listens on `:18081`
- the browser app remains on `:5173`

## Important rule

Do not skip the IAM bootstrap step.
Without internal role assignments, the Keycloak login may succeed but the authenticated
user will not have the capabilities required to use the operational surfaces.
