# ADR-0022: Postgres Tenant Session Key for RLS is app.tenant_id

- Status: accepted
- Date: 2026-03-19

## Context

The platform uses shared tables + RLS for tenant isolation (ADR-0002).
RLS policies depend on a stable runtime function:

- `current_tenant_id()` reads `current_setting('app.tenant_id', true)`

Go code sets tenant context through the platform helper:

- `pgdb.BeginTenantTx` -> `SetTenantContext` -> `set_config('app.tenant_id', ...)`

Workers must follow the same convention, otherwise writes and reads become unsafe and inconsistent across environments.

## Decision

The only supported Postgres session key for tenancy is:

- `app.tenant_id`

Rules:

- all Go transactions use `pgdb.BeginTenantTx` / `SetTenantContext`
- all worker write transactions must execute `set_config('app.tenant_id', <tenant>, true)` before writing
- no alternative keys (for example `app.current_tenant_id`) are allowed

## Consequences

- RLS isolation is consistent across Go and Python.
- Environment differences (superuser bypass, local shortcuts) are less likely to mask real tenancy bugs.
- Skill guidance and scaffolds remain aligned with the real platform runtime behavior.

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Platform package review: `metalshopping-platform-packages`
- Worker scaffolding rules: `metalshopping-worker-scaffold`

