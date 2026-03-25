---
id: hippocampus-architecture
title: Architecture
region: hippocampus
tags: [architecture, stack, design, tech-stack]
links:
  - hippocampus/conventions
  - cortex/backend/index
  - cortex/frontend/index
  - cortex/database/index
weight: 1.0
updated_at: 2026-03-24T10:00:00Z
---

# MetalShopping Architecture

## Project Identity

MetalShopping is a **server-first B2B platform** for commercial strategy, pricing, analytics, procurement, and CRM. Solo developer, production-intended v1. Current phase: **3A — Foundation Hardening**. Analytics legacy migration in progress.

## Tech Stack

### Backend
- **Language:** Go 1.23
- **Architecture:** Modular monolith (`apps/server_core/`)
- **Database:** PostgreSQL (multi-tenant, shared-database)
- **Event system:** Transactional outbox pattern
- **Auth:** JWT/OIDC via Keycloak
- **Tenancy:** Tenant-aware Postgres helpers, context middleware

### Frontend
- **Framework:** React 18 (Vite + TypeScript)
- **Pattern:** Thin-client (all data via generated SDK)
- **Features:** Feature-based packages (`packages/feature-*`)
- **UI:** Shared primitives (`packages/ui`, CSS modules)
- **State:** React Query + SDK-generated hooks

### Workers
- **Python workers:** Analytics compute/scoring (skeleton)
- **Constraint:** Never call `server_core` HTTP endpoints (one-way dependency)

### Contracts & Code Generation
- **Source of truth:** Hand-authored contracts drive code generation
- **OpenAPI specs:** `contracts/api/openapi/` (REST endpoints)
- **JSON Schema:** `contracts/api/jsonschema/` (payloads)
- **Event schemas:** `contracts/events/v1/` (versioned events)
- **Governance schemas:** `contracts/governance/` (feature flags, policies, thresholds)
- **SDK runtime:** SDK-generated artifacts (auto-generated, never edit)

## Monorepo Layout

```
MetalShopping/
├── apps/
│   ├── server_core/              Go modular monolith (canonical authority)
│   ├── web/                      React thin-client (Vite)
│   └── analytics_worker/         Python compute (skeleton)
├── contracts/
│   ├── api/openapi/              OpenAPI 3.0 specs
│   ├── api/jsonschema/           JSON Schema payloads
│   ├── events/v1/                Event schemas (versioned)
│   └── governance/               Feature flags, policies, thresholds
├── packages/
│   ├── ui/                       Shared React UI primitives
│   ├── platform-sdk/             Generated SDK runtime
│   ├── feature-analytics/        Analytics surface
│   ├── feature-products/         Products surface
│   ├── feature-auth-session/     Auth/session surface
├── docs/
│   ├── ARCHITECTURE.md           Frozen blueprint
│   ├── PROJECT_SOT.md            Current phase state
│   ├── IMPLEMENTATION_PLAN.md    Phased roadmap
│   ├── PROGRESS.md               Completion status
│   └── adrs/                     Architecture decision records
├── tasks/
│   ├── todo.md                   Sprint backlog
│   └── lessons.md                Captured patterns
├── bootstrap/seeds/              Governance + tenant seed data
├── ops/                          Docker, Kubernetes, ops
├── scripts/                      Build, contract generation
└── .brain/                       ForgeFlow Mini Brain
```

## Go Backend: Module Structure

Every module in `apps/server_core/internal/modules/<name>/` follows this layer structure:

```
domain/         business entities & logic
application/    use-case / command handlers
ports/          interfaces (in + out)
adapters/       Postgres persistence
transport/      HTTP handlers & serialization
events/         event definitions & emission
readmodel/      denormalized query views
```

**Platform infrastructure** (`internal/platform/`):
- `db/postgres/` — tenant-aware Postgres helpers (`pgdb.BeginTenantTx`, `current_tenant_id()`)
- `auth/` — JWT/OIDC, `PrincipalFromContext`, auth checks
- `tenancy_runtime/` — tenant context middleware, `TenantFromContext`
- `governance/` — feature flags, policies, thresholds resolution
- `outbox/` — transactional event publishing (`outbox.AppendInTx`)

**Critical rule:** Every new module must be registered in `composition_modules.go`.

## Frontend: Thin-Client Pattern

All data flows through the generated SDK (`@metalshopping/sdk-runtime`). Feature logic lives in `packages/feature-*` packages; `apps/web` is the shell that assembles them.

## Design Philosophy

- **Make it work → make it beautiful → make it fast** (in that order)
- **Production-grade v1** — not a prototype
- **Stripe/Google engineering bar** — senior-engineer-approved code
- **Tenant safety non-negotiable** — every query, every transaction
- **Event-driven** — outbox pattern for eventual consistency

---

**Created:** 2026-03-23 | **Last Updated:** 2026-03-24 | **Weight:** 1.0

See [[docs/ARCHITECTURE.md]] for frozen blueprint | [[docs/PROJECT_SOT.md]] for current phase
