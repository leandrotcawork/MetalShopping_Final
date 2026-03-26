---
id: hippocampus-architecture
title: MetalShopping Architecture
type: hippocampus
tags: [architecture, platform, stack, modules]
updated_at: 2026-03-26
---

# MetalShopping Architecture

## Platform Identity

MetalShopping is an enterprise B2B platform — not traditional e-commerce. Domains: commercial strategy, pricing, market monitoring, procurement, CRM, automations, operational and strategic analytics.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23 modular monolith (`apps/server_core`) |
| Frontend | React 18 thin-client (`apps/web`) + TypeScript |
| Workers | Python (`analytics_worker`, `automation_worker`, `integration_worker`, `notifications_worker`) |
| Database | Postgres (canonical state, RLS, multi-tenant shared schema) |
| Contracts | OpenAPI 3.0 + JSON Schema + versioned event schemas |

## Architecture Model

- **Monorepo** — single repo, npm workspaces + Go workspace
- **Server-first** — Go monolith is canonical state authority
- **Multi-tenant shared database** — `tenant_id` + `current_tenant_id()` RLS on every tenant-scoped table
- **Contract-driven** — `contracts/` are hand-authored source of truth; auto-generated SDK artifacts under `packages/` are NEVER edited manually

## Source of Truth Hierarchy

```
contracts/   (hand-authored: OpenAPI + JSON Schema + events)
    |
    v  generate via scripts/generate_contract_artifacts.ps1
    |
packages/generated-sdk/  and  packages/generated-types-artifacts/
(auto-generated — never edit these directly)
```

## Go Module Structure (server_core)

Every module under `internal/modules/<name>/` follows:
```
domain/       business entities & logic
application/  use-case / command handlers
ports/        interfaces (in + out)
adapters/     Postgres persistence
transport/    HTTP handlers & serialization
events/       event definitions & emission
readmodel/    denormalized query views
```

Platform infrastructure at `internal/platform/`:
- `db/postgres/` — tenant-aware Postgres helpers
- `auth/` — JWT/OIDC, principal context
- `tenancy_runtime/` — tenant context middleware
- `governance/` — feature flags / policies / thresholds resolution
- `outbox/` — transactional event publishing

## Active Backend Modules (18)

`alerts`, `analytics_serving`, `automation`, `catalog`, `crm`, `customers`, `home`, `iam`, `integrations_control`, `inventory`, `market_intelligence`, `pricing`, `procurement`, `sales`, `shopping`, `suppliers`, `tenant_admin`

## Frontend Architecture

- Thin-client: all data via `@metalshopping/sdk-runtime` — no raw `fetch()`
- Feature logic in `packages/feature-*`; `apps/web` is the shell
- UI primitives: `packages/ui/src/index.ts` — check before creating components
- Design tokens only — no hardcoded hex values

## Monorepo Layout

| Path | Purpose |
|------|---------|
| `apps/server_core/` | Go modular monolith |
| `apps/web/` | React 18 thin-client |
| `apps/analytics_worker/` | Python compute/scoring |
| `contracts/api/openapi/` | OpenAPI 3.0 specs |
| `contracts/events/` | Versioned event schemas |
| `contracts/governance/` | Feature flags, policies, thresholds |
| `packages/ui/` | Shared React UI primitives |
| `packages/platform-sdk/` | SDK runtime |
| `packages/feature-analytics/` | Analytics surface |
| `packages/feature-products/` | Products surface |
| `packages/feature-auth-session/` | Auth/session |
| `docs/` | Architecture blueprint, ADRs |
| `ops/` | Docker, Kubernetes, Keycloak, observability |
