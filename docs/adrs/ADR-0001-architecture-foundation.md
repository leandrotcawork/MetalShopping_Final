# ADR-0001: Architecture Foundation

- Status: accepted
- Date: 2026-03-17

## Context

MetalShopping needs a platform shape that supports long-term enterprise growth without early fragmentation.

## Decision

The platform is frozen with these base rules:

- monorepo
- server-first
- modular monolith in `apps/server_core`
- specialized workers outside the core
- Postgres as canonical state
- Go in the core
- Python in workers during transition
- explicit contracts and governance outside app code
- thin clients for `web`, `desktop`, and `admin_console`

## Consequences

- canonical state ownership remains in the core
- workers stay focused on async and compute concerns
- contracts and governance become first-class repo assets
- frontend stays dependent on platform contracts, not private local logic

