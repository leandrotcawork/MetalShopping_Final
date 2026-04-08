# System Pulse - MetalShopping Final
> Auto-updated: 2026-04-08 | Session: #5

## Project Identity
- **Name:** MetalShopping Final
- **Description:** Enterprise server-first platform for commercial strategy, pricing, procurement, CRM, and analytics.
- **Type:** Monorepo platform (backend + workers + web clients)
- **Stage:** Foundation implementation
- **Repository:** C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalShopping_Final\MetalShopping_Final

## Technology Stack
- **Language(s):** Go 1.23, TypeScript 5.7, React 18
- **Backend:** Go stdlib + pgx (server_core), modular monolith
- **Frontend:** React 18 + Vite
- **Database:** PostgreSQL (canonical state)
- **Testing:** go test, Vitest
- **Build/Deploy:** scripts/ (PowerShell), Docker required for contract generation, Vite build for web
- **Key libraries:** pgx, react, react-router-dom, vite, vitest

## Architecture Overview
Server-first modular monolith in apps/server_core, with specialized workers in
apps/integration_worker and others. Contracts live under contracts/ and drive
generated SDKs and types used by web clients. Postgres is the canonical store
with tenant isolation via tenant_id and RLS. Async integration uses outbox
events and workers for ingestion, normalization, and delivery. Packages/ hosts
shared UI, SDK runtime, and feature modules. Docs/ carries SoT, architecture,
ADRs, and implementation planning.

## Established Patterns and Conventions
- Contracts are the source of truth for APIs/events/governance.
- Generated SDKs and types are consumed by clients; do not edit generated code.
- Module structure follows domain/application/ports/adapters/transport.
- Tenant isolation is enforced in DB access paths.
- Outbox events are appended in the same transaction as writes.
- Frontend uses SDK runtime and design tokens from shared UI.

## Current Phase
- **Roadmap phase:** Phase 2 - Layer 1 Operational Intelligence
- **Phase goal:** Build the first operational intelligence surfaces on top of validated ERP data.
- **Last completed task:** T-001 0.1 ERP Integration (Gate A + Gate B acceptance complete).
- **Currently working on:** T-007 1.1 Analytics Home
- **Next up:** T-008 1.2 Pricing Intelligence
- **Blockers:** None

## Recent Changes (last 3-5 sessions)
- 2026-04-08: Gate B ERP Oracle run completed with canonical writes validated for products, prices, and inventory.
- 2026-04-08: Acceptance workbook updated and committed with Gate B evidence.
- 2026-04-08: ERP integration runtime tests and contract generation verified in the Oracle worktree.
- 2026-04-05: Added ERP run entity-step checkpoints and raw/staging batch ordinal persistence (`0038`, `0039`).
- 2026-04-05: Updated worker execution to structured connection config and dependency-aware entity flow.

## Active Architectural Decisions
- ADR-001: Use godror plus a typed query-runner API for Oracle ERP connectivity.

## Known Risks and Tech Debt
- Production identity integration not yet connected to a real issuer or JWKS.
- Broker delivery and worker consumption are not in place for outbox events.
- Operational governance surfaces still need admin mutation paths.

## Key File Locations
- Server core: apps/server_core/
- Integration worker: apps/integration_worker/
- Contracts: contracts/api/, contracts/events/, contracts/governance/
- Web client: apps/web/
- Shared UI: packages/ui/
- SDK runtime: packages/platform-sdk/
- Project SoT: docs/PROJECT_SOT.md
- Orchestration plan: docs/MASTER_ORCHESTRATION_PLAN.md
