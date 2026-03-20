# Project SoT

## Purpose

This document is the operational source of truth for the planning and foundation implementation phases of MetalShopping Final.

## Current state

- Phase: foundation implementation
- Delivery mode: make it work first
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
- follow `make it work -> make it clean -> make it fast` for new module delivery
- Prefer writing or updating SoT docs, ADRs, and phase plans first
- Avoid duplicate planning docs that restate the same rule in different wording
- Structural changes must be paired with `docs/PROJECT_SOT.md` and `docs/PROGRESS.md` updates in the same tranche

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
- explicit canonical SKU data ownership model aligned to legacy `products` and `product_erp`
- first tenant-aware `pricing` module slice implemented and semantically realigned against legacy replacement-cost and average-cost language
- first tenant-aware `inventory` module slice implemented with canonical stock-position ownership
- first thin-client `Products` surface implemented with generated frontend transport and backend-owned sorting
- backend-owned `auth/session` foundation implemented with OIDC cookie-session runtime and generated transport support
- login hardening now includes browser-safe CSRF defense for cookie-backed mutations, centralized generated browser HTTP runtime, and a thinner auth composition in `apps/web`
- login closure governance is now frozen with explicit scope, DoD, SDK boundary references, and tranche execution plan
- `sdk_ts` hardening now targets OpenAPI Generator via Docker-backed orchestration instead of handwritten TypeScript emission
- `server_core` startup composition root is now decomposed by capability for clearer ownership and maintenance
- frontend auth ownership now keeps `feature-auth-session` runtime-agnostic, with session login URL composition owned by the generated SDK facade
- login visual baseline now has shared token files and an explicit sync/check workflow across React fallback and Keycloak theme
- auth/session bootstrap mode is now explicit (`required`, `optional`, `disabled`) with fail-fast default to avoid silent runtime degradation
- `server_core` now uses signal-aware graceful shutdown and shared context cancellation for HTTP server and outbox dispatcher lifecycle
- frontend authored runtime/facade moved to `packages/platform-sdk`, while `scripts/generate_contract_artifacts.ps1` now orchestrates OpenAPI generation without hand-emitting transport code
- SoT drift guard now supports branch diff mode (`-BaseRef`) for CI-grade checks and covers workspace manifests plus web runtime config files in addition to local working-tree validation
- first `Home` Level 1 slice is implemented with contract-first endpoint, server_core handler, generated SDK binding, and real KPI rendering in `apps/web`

## Current structural gaps

The most important remaining gaps are:

- production identity integration is not yet connected to a real issuer or JWKS source
- generated artifact drift check is now wired in pull request CI together with contract validation
- login closure governance remains documented, and operational module work now follows make-it-work-first sequencing
- outbox exists and catalog emits real events, but broker delivery and worker consumption are still not in place
- governance is operational in runtime, but broader operational surfaces still need administrative mutation paths
- Shopping frontend parity still needs to converge on the legacy workflow visuals and operational density without reintroducing legacy shortcuts (ADR-0036..ADR-0039)
- catalog is now a strong canonical foundation, pricing semantics have been realigned against the accepted SKU ownership model, inventory owns live stock position, and the next gate is freezing procurement so supplier-side replenishment semantics do not leak into existing modules
- contract validation, generated artifact drift checks, backend tests, web typecheck/build, and SoT documentation consistency are now enforced on `pull_request` CI workflow

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
- explicit canonical SKU data ownership model spanning `catalog`, `pricing`, `inventory`, `procurement`, and analytics
- explicit canonical pricing model before the first pricing implementation slice
- explicit canonical procurement model before procurement contracts and runtime code
- explicit operational surface recovery plan before frontend implementation accelerates
- explicit frontend quality gates and products read-surface ownership before frontend implementation spreads
- explicit frontend migration charter freezing legacy visual reuse with modern package and API boundaries
- explicit frontend migration playbook freezing legacy-study-first execution before surface ports continue
- explicit decision record for the next implementation area
- explicit web auth session model before login UI implementation
- explicit login and identity architecture before local issuer bootstrap and login UI
- explicit login MVP closure scope, DoD, and SDK boundary governance before opening new authenticated surfaces
- explicit frontend migration matrix for preserve/refactor/reject decisions by legacy artifact
- explicit data contract map for Home, Shopping, and Analytics before each module implementation

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
- `docs/SKU_CANONICAL_DATA_MODEL.md`
- `docs/PRICING_READINESS_REVIEW.md`
- `docs/PRICING_CANONICAL_MODEL.md`
- `docs/PRICING_LEGACY_SIGNAL_BOUNDARIES.md`
- `docs/PRICING_IMPLEMENTATION_PLAN.md`
- `docs/INVENTORY_CANONICAL_MODEL.md`
- `docs/PROCUREMENT_CANONICAL_MODEL.md`
- `docs/PROCUREMENT_IMPLEMENTATION_PLAN.md`
- `docs/OPERATIONAL_SURFACES_PLAN.md`
- `docs/PRODUCTS_SURFACE_IMPLEMENTATION_PLAN.md`
- `docs/PRODUCTS_READMODEL_OWNERSHIP.md`
- `docs/FRONTEND_QUALITY_GATES.md`
- `docs/FRONTEND_MIGRATION_CHARTER.md`
- `docs/FRONTEND_MIGRATION_PLAYBOOK.md`
- `docs/NEXT_EXECUTION_DECISION.md`
- `docs/WEB_AUTH_SESSION_IMPLEMENTATION_PLAN.md`
- `docs/LOGIN_AND_IDENTITY_ARCHITECTURE.md`
- `docs/LOGIN_MVP_SCOPE.md`
- `docs/LOGIN_DOD.md`
- `docs/LOGIN_MVP_EXECUTION_PLAN.md`
- `docs/SDK_BOUNDARY.md`
- `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`
- `docs/FRONTEND_MIGRATION_MATRIX.md`
- `docs/DATA_CONTRACT_MAP.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
