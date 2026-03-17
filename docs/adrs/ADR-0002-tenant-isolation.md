# ADR-0002: Tenant Isolation Model

- Status: accepted
- Date: 2026-03-17

## Context

The platform must support multi-tenant growth without taking on premature isolation complexity.

## Decision

The initial tenancy model is:

- shared Postgres database
- shared tables
- `tenant_id` on multitenant data
- Row-Level Security as the default isolation mechanism

Future exception:

- premium or regulated tenants may move to stronger isolation later if required

## Consequences

- the initial platform stays simpler operationally
- data access rules must be tenancy-aware from the start
- migration paths for stronger isolation must be kept possible, but not implemented now

