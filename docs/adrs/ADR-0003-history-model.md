# ADR-0003: Historical Data Ownership

- Status: accepted
- Date: 2026-03-17

## Context

Historical data is essential for pricing, market intelligence, sales analysis, and CRM, but a generic cross-cutting history module would quickly become a dumping ground.

## Decision

- there is no top-level `history` module
- each domain owns its own historical data
- `platform/db/timeseries` is infrastructure support only
- large temporal tables must be designed with partitioning and retention policies

## Consequences

- ownership stays aligned with business meaning
- temporal scaling concerns stay visible at the domain boundary
- audit and time-series support remain platform concerns, not a business super-module

