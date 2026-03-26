---
id: cortex-backend-index
title: Backend Cortex Index
region: backend
type: cortex-index
tags: [backend, go, server_core, modules, patterns]
updated_at: 2026-03-26
---

# Backend Cortex

## Scope

Everything in `apps/server_core/` — Go 1.23 modular monolith.

## Module Catalogue

All modules under `internal/modules/`:

| Module | Domain | Status |
|--------|--------|--------|
| `alerts` | Notification triggers | Active |
| `analytics_serving` | Analytics read surfaces | Active |
| `automation` | Workflow automations | Active |
| `catalog` | Product catalog | Active |
| `crm` | Customer relationships | Active |
| `customers` | Customer accounts | Active |
| `home` | Home dashboard aggregator | Active |
| `iam` | Identity and access | Active |
| `integrations_control` | External integrations | Active |
| `inventory` | Stock management | Active |
| `market_intelligence` | Price signals, market index | Active |
| `pricing` | Pricing engine | Active |
| `procurement` | Purchase orders, supplier ops | Active |
| `sales` | Sales pipeline | Active |
| `shopping` | Shopping workflows | Active |
| `suppliers` | Supplier management | Active |
| `tenant_admin` | Tenant administration | Active |

## Layer Pattern

```
domain/       Entities, value objects, domain logic — no infrastructure imports
application/  Use-case handlers, command/query dispatchers
ports/        Input/output interfaces (repos, event bus, external services)
adapters/     Postgres implementations of port interfaces
transport/    HTTP handlers, request parsing, response serialization
events/       Event type definitions, emission
readmodel/    Denormalized views, projections for queries
```

## Critical Platform Calls

Every adapter: `pgdb.BeginTenantTx(ctx, db)` — never use raw `db.Begin()`
Every handler: `platformauth.PrincipalFromContext(ctx)` → 401 | `tenancy_runtime.TenantFromContext(ctx)` → 403
Every event emit: `outbox.AppendInTx(tx, event)` inside the same tx as the write

## New Module Checklist

1. Create directory structure with all 7 layers
2. Register in `composition_modules.go`
3. Add OpenAPI spec in `contracts/api/openapi/<name>.yaml`
4. Wire handlers via router in `transport/`
5. Add event schemas if module emits events

## Sinapses in This Region

_Add links to `.brain/sinapses/<backend-topic>.md` files as they are created._
