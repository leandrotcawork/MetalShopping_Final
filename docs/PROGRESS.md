# Progress

## Snapshot

- Phase: foundation implementation
- Status: active
- Architecture: frozen at direction level
- Critical freezes: documented
- Product code: implementation started with first platform package

## Done

- monorepo scaffold created
- architecture document created
- legacy-to-target mapping documented
- planning-first SoT created
- implementation phases documented
- ADR track started
- AGENTS guidance being added
- contract conventions documented
- SDK generation strategy documented
- contract templates created
- script workflow entrypoints defined
- engineering standards documented
- global system principles documented
- engineering principles documented
- core and worker operating models documented
- contract evolution rules documented
- observability and security baseline documented
- module standards documented
- platform boundaries documented
- readmodel and event rules documented
- platform package standards documented
- module creation checklist documented
- module and platform templates created
- MetalDocs reuse review completed and migration matrix documented
- first real platform package implemented in `server_core/internal/platform/db/postgres`
- foundational auth middleware and principal context implemented in `server_core/internal/platform/auth`
- first real business module implemented in `server_core/internal/modules/iam`
- first real IAM contracts implemented in `contracts/api` and `contracts/events`
- server_core bootstrap wired with dotenv, Postgres readiness, auth, and IAM route registration
- tenancy runtime context and middleware implemented in `server_core/internal/platform/tenancy_runtime`
- tenancy-aware Postgres session helpers and RLS runtime function added to the platform foundation
- first tenant-aware business module implemented in `server_core/internal/modules/catalog`
- first tenant-aware business contracts implemented in `contracts/api` and `contracts/events` for catalog
- governance runtime registry and resolvers implemented in `server_core/internal/platform/governance`
- first real governance contracts added for feature flags, thresholds, and policies
- governance bootstrap registry added and first runtime-controlled module path wired into `catalog` product creation
- first database-backed governance slice implemented for feature flags with runtime loading from Postgres
- first server_core unit tests added and validated
- canonical catalog product model analyzed from legacy signals and frozen as the next gate before pricing
- canonical catalog taxonomy slice implemented with tenant-aware tables, API reads, and richer product master fields
- canonical catalog product identifiers slice implemented with tenant-aware persistence and API reads
- catalog product description added to the canonical product shape, contracts, and create/read flow
- platform outbox implemented with transactional event append for `catalog.product_created`
- production-path JWT/OIDC authentication adapter added while preserving static bootstrap auth for local development
- database-backed governance thresholds and policies implemented and wired into real runtime behavior
- contract validation and generation scripts implemented and generating minimal artifacts in `packages/generated`
- `catalog` reviewed as ready for the first `pricing` slice
- initial canonical pricing model frozen before implementation
- next implementation area explicitly frozen as `pricing`
- full first-slice pricing implementation plan documented before code
- pricing phase 1 contracts completed across governance, API, JSON Schemas, and events
- generated artifacts refreshed after pricing contract authoring
- first tenant-aware `pricing` module slice implemented in `server_core`
- pricing runtime guards, outbox publication path, HTTP transport, and bootstrap wiring implemented
- pricing unit tests added and passing
- canonical SKU data ownership model frozen from legacy `products` and `product_erp` semantics
- pricing revision started to replace generic cost-basis and margin-floor persistence with legacy-aligned cost semantics
- pricing write path now deduplicates no-op reruns so history is change-based instead of execution-based
- legacy `product_erp` signal boundaries are now frozen so price, stock, procurement, tax, and advisory semantics do not collapse back into a single module
- first tenant-aware `inventory` module slice implemented with canonical stock position ownership, contracts, outbox event, and HTTP transport
- canonical inventory model frozen with `on_hand_quantity`, `last_purchase_at`, and `last_sale_at` as first owned semantics
- procurement birth constraints frozen so supplier-side replenishment semantics do not leak into `pricing`, `inventory`, or direct ERP reads
- operational surface recovery order frozen as `Products -> Shopping -> Home`
- first `Products` surface implementation plan frozen as the next real UI slice
- `Products` read-surface ownership and frontend quality gates frozen before web code expansion
- frontend migration charter frozen so legacy MetalShopping visuals are preserved without reusing weak DTO, API, CSS, and package patterns
- dedicated frontend migration skill added to keep future UI work aligned with the charter
- first `Products` thin-client surface hardened with generated frontend transport and backend-owned sorting
- next auth/session boundary for `web` frozen on OIDC plus `HttpOnly` cookie sessions
- backend-owned `auth/session` runtime implemented in `server_core` with cookie session storage, OIDC callback flow, governance-aware timeouts, and generated web transport support
- login and identity architecture frozen with Keycloak as the initial IdP, tenant claim mapping, and a cross-channel identity model for future app surfaces
- reproducible local Keycloak bootstrap assets added for realm import, tenant claim mapping, local test users, and repeatable IAM role seeding
- login hardening completed with backend-owned CSRF protection for cookie-backed mutations, centralized generated browser HTTP runtime, thinner auth routing composition, and a frozen login visual baseline shared with the Keycloak theme
- official `sdk_ts` migration target frozen on OpenAPI Generator `typescript-fetch` with Docker-backed orchestration and a repo-specific integration skill

## Next

- keep ADR set complete and stable
- freeze and implement the backend-owned `auth/session` surface before login UI work
- bootstrap Keycloak locally as the first real issuer
- start the local Keycloak issuer from the committed bootstrap assets
- configure realm, client, redirect URI, and `tenant_id` mapper
- seed internal IAM roles for the imported Keycloak subject ids before switching the backend to JWT mode
- connect `.env` to the Keycloak local issuer and validate `auth/session` end to end
- validate migrations `0008`, `0009`, and `0010` end-to-end in the running server with smoke coverage
- enforce contract validation and artifact generation in team workflow and CI
- connect the JWT/OIDC auth path to a real issuer configuration
- bootstrap authenticated session state in `apps/web`
- validate the hardened login flow end-to-end against the running Keycloak issuer after every local restart
- replace the remaining handwritten `sdk_ts` emission path with the official generator-backed flow
- create and adopt the repo skill that freezes front+back integration rules around contracts, generation, runtime, and docs sync
- apply pricing migrations and database-backed governance defaults in runtime
- validate pricing write/list/current and outbox publication through smoke tests
- revise pricing semantics to align with the accepted canonical SKU data ownership model
- validate the revised pricing slice and migration path end-to-end after semantic alignment
- continue domain expansion from the proven tenant-aware foundation
- keep pricing and inventory semantically narrow while preparing procurement follow-on ownership
- freeze procurement canonical model and upstream integration gate before procurement contracts or runtime code
- decide the first canonical procurement inputs and contracts that integration must publish
- scaffold the first thin-client operational surface around `Products`
- define the first `Products` read surface contract and scaffold `apps/web`
- validate contract generation after adding the `Products` read surface contract
- execute the first `Products` UI slice using the frozen frontend migration charter and new frontend skill
- prepare phase transition checklist from foundation hardening to domain expansion
- use the MetalDocs reuse matrix to decide the first extracted patterns without copying unsafe defaults

## Blockers

- none technical right now
- the only real blocker is changing core rules before freezing them

## Notes

- legacy backend code is intentionally not part of the active plan right now
- MetalDocs may inform selective reuse, but only through the migration matrix and frozen architecture rules
- future code work should start from the frozen documents and the current foundation baseline, not from ad hoc memory
