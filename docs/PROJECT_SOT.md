# Project SoT

## Purpose

This document is the operational source of truth for the planning and foundation implementation phases of MetalShopping Final.

## Current state

- Phase: foundation implementation
- Architecture status: approved
- Code status: core foundation running with initial platform and business slices
- Legacy backend status: intentionally not in use
- MetalDocs reuse status: selective reuse only, guided by a transitional reuse matrix
- Next gate: keep execution aligned with frozen architecture while closing the remaining structural gaps

## Product identity

MetalShopping is not a traditional e-commerce product. It is an enterprise platform for:

- commercial strategy
- pricing
- market monitoring
- procurement
- CRM
- automations
- operational and strategic analytics

## Platform direction

- monorepo
- server-first
- modular monolith in `apps/server_core`
- specialized workers outside the core
- Postgres as canonical state
- Go in the core
- Python in workers during transition
- explicit contracts and governance outside app code
- thin clients for `web`, `desktop`, and `admin_console`

## Frozen platform rules

### 1. Tenant isolation

- Initial model: shared Postgres database, shared tables, `tenant_id`, and `RLS`
- Future exception: premium or regulated tenants may move to stronger isolation later
- Non-goal for now: one database per tenant at the start

### 2. Historical data

- There is no top-level `history` module
- Each domain owns its own history
- `platform/db/timeseries` is infrastructure support only
- Large temporal tables must be designed with partition and retention policies

### 3. Runtime governance

- `contracts/governance/*` defines schema
- `bootstrap/seeds/governance/*` defines initial defaults
- Effective state lives in the database
- Runtime resolution lives in `apps/server_core/internal/platform/governance/*`
- Core and workers must share the same semantics
- Hardcoded thresholds and policies are not allowed

### 4. Frontend model

- `web` and `desktop` consume the same `server_core`
- `admin_console` is the operational and governance surface
- Separate BFF is allowed only if client divergence becomes real
- Frontend consumes generated SDKs and generated types, not hand-maintained parallel types

### 5. Async integration model

- Relevant mutations publish versioned events
- Events live under `contracts/events/v1` first
- Workers consume through broker or queue semantics
- Workers must not become direct synchronous dependencies of the core
- Normal request serving in the core must not depend on worker round-trips

## Planning constraints

- Do not add product code just because a structure exists
- Prefer writing or updating SoT docs, ADRs, and phase plans first
- Avoid duplicate planning docs that restate the same rule in different wording

## Current implementation baseline

The repository now includes:

- executable `server_core` bootstrap
- Postgres platform foundation
- centralized auth flow with local static bootstrap and production-path JWT/OIDC adapter
- tenancy runtime and tenancy-aware Postgres session helpers
- runtime governance registry and resolvers
- database-backed governance for feature flags, thresholds, and policies
- platform outbox foundation and first real mutation-to-event path
- first structural module: `iam`
- first tenant-aware business module: `catalog`
- initial API, event, and governance contracts
- functional contract validation and generated artifact scripts

## Current structural gaps

The most important remaining gaps are:

- production identity integration is not yet connected to a real issuer or JWKS source
- outbox exists and catalog emits real events, but broker delivery and worker consumption are still not in place
- governance is operational in runtime, but broader operational surfaces still need administrative mutation paths
- catalog is now a strong canonical foundation, but pricing-readiness still needs one final review before domain expansion
- contract validation and generation are now functional locally and must next be enforced in team workflow and CI

## Planning deliverables

- official architecture doc
- ADR set for critical freezes
- contract conventions
- generated SDK strategy
- implementation plan by phase
- progress tracker
- AGENTS guidance for token-efficient work
- phase-by-phase execution discipline that keeps implementation aligned with the frozen architecture
- explicit canonical product model for `catalog` before pricing and inventory expansion
- explicit canonical pricing model before the first pricing implementation slice

## Key planning docs

- `ARCHITECTURE.md`
- `docs/SYSTEM_PRINCIPLES.md`
- `docs/ENGINEERING_PRINCIPLES.md`
- `docs/SERVER_CORE_OPERATING_MODEL.md`
- `docs/WORKER_OPERATING_MODEL.md`
- `docs/CONTRACT_EVOLUTION_RULES.md`
- `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
- `docs/MODULE_STANDARDS.md`
- `docs/PLATFORM_PACKAGE_STANDARDS.md`
- `docs/PLATFORM_BOUNDARIES.md`
- `docs/READMODEL_AND_EVENTS_RULES.md`
- `docs/MODULE_CREATION_CHECKLIST.md`
- `docs/CONTRACT_CONVENTIONS.md`
- `docs/SDK_GENERATION_STRATEGY.md`
- `docs/METALDOCS_REUSE_MATRIX.md` (transitional only, delete after migration execution)
- `docs/CATALOG_CANONICAL_MODEL.md`
- `docs/PRICING_READINESS_REVIEW.md`
- `docs/PRICING_CANONICAL_MODEL.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
