# ADR-0005: Thin Clients And Generated SDKs

- Status: accepted
- Date: 2026-03-17

## Context

The product will have multiple client surfaces. Duplicating business rules or maintaining parallel type systems in clients would create divergence and operational waste.

## Decision

- `web` and `desktop` consume the same `server_core`
- `admin_console` is the operational and governance surface
- frontend clients remain thin
- generated SDKs and generated types are the source used by clients
- separate BFF is allowed only if real client divergence appears later

## Consequences

- business logic stays out of frontend surfaces
- manual shared type drift is treated as an anti-pattern
- contract generation becomes a mandatory platform workflow

