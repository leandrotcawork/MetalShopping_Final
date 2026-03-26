---
id: cortex-database-index
title: Database Cortex Index
region: database
type: cortex-index
tags: [database, postgres, tenant-isolation, rls, migrations]
updated_at: 2026-03-26
---

# Database Cortex

## Scope

Postgres as canonical state — tenant isolation model, RLS, migrations, timeseries infra.

## Tenant Isolation Model

- **Shared database, shared tables** — `tenant_id` column on every tenant-scoped table
- **RLS** — `current_tenant_id()` enforced via Postgres row-level security
- **Every query**: must include `WHERE current_tenant_id() = tenant_id` (or equivalent RLS policy)
- **Every tx**: opened via `pgdb.BeginTenantTx(ctx, db)` — never raw `db.Begin()`

No cross-tenant data leaks — ever. This is a hard invariant.

## Historical Data Model

- No top-level `history` module — each domain owns its own history
- `internal/platform/db/timeseries` is infrastructure support only
- Large temporal tables: design with partition and retention policies from day one

## Governance / Runtime Config

- `contracts/governance/*` defines schema for feature flags, policies, thresholds
- `bootstrap/seeds/governance/*` defines initial defaults
- Effective state lives in the database — not in config files

## Migration Rules

- Migrations are additive where possible — avoid destructive schema changes
- Every migration is idempotent (re-runnable safely)
- New columns get `DEFAULT` or `NOT NULL` only if backfill is provided

## Platform Postgres Helpers

- `pgdb.BeginTenantTx` — opens tenant-scoped transaction
- `pgdb.QueryTenant` — wraps queries with tenant filter
- All in `apps/server_core/internal/platform/db/postgres/`

## Sinapses in This Region

_Add links to `.brain/sinapses/<database-topic>.md` files as they are created._
