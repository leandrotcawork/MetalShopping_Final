# Platform Boundaries

## Purpose

Define the boundary between platform infrastructure, business modules, and shared neutral helpers.

## Boundary model

Inside `apps/server_core/internal/` there are three categories:

- `platform/`
- `modules/`
- `shared/`

These categories must not blur over time.

## `platform/`

Use `platform/` for infrastructure and runtime capabilities that are not business domains.

Examples:

- auth
- tenancy runtime
- runtime config
- governance runtime
- db
- messaging
- jobs
- cache
- files
- delivery
- observability
- security
- auditlog

Rules:

- platform provides capabilities to the system
- platform does not become a business feature bucket
- platform code may be reused by multiple modules

## `modules/`

Use `modules/` for business capabilities with owned semantics.

Examples:

- pricing
- procurement
- market_intelligence
- analytics_serving
- crm
- alerts

Rules:

- module code owns business meaning
- module code should not be relocated into platform just because it is reused
- reuse does not erase business ownership

## `shared/`

Use `shared/` for small, neutral helpers with low semantic weight.

Examples:

- ids
- money
- clock
- pagination
- small error helpers

Rules:

- keep `shared/` small
- if something has domain meaning, it does not belong here
- if something is runtime infrastructure, it belongs in `platform/`

## Boundary tests

Ask these questions:

1. Is this business meaning?
   If yes, it belongs in `modules/`.
2. Is this runtime infrastructure or platform capability?
   If yes, it belongs in `platform/`.
3. Is this a small neutral helper with no business ownership?
   If yes, it may belong in `shared/`.

## Boundary anti-patterns

- moving domain rules into platform for convenience
- creating generic shared folders for semantically heavy code
- splitting one business capability across module and platform without a clear reason

