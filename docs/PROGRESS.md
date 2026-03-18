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

## Next

- keep ADR set complete and stable
- validate migrations `0008`, `0009`, and `0010` end-to-end in the running server with smoke coverage
- enforce contract validation and artifact generation in team workflow and CI
- connect the JWT/OIDC auth path to a real issuer configuration
- apply pricing migrations and database-backed governance defaults in runtime
- validate pricing write/list/current and outbox publication through smoke tests
- revise pricing semantics to align with the accepted canonical SKU data ownership model
- validate the revised pricing slice and migration path end-to-end after semantic alignment
- continue domain expansion from the proven tenant-aware foundation
- keep pricing semantically narrow while preparing `inventory` and procurement follow-on ownership
- apply inventory migration and validate the new write/list/current path in runtime
- prepare phase transition checklist from foundation hardening to domain expansion
- use the MetalDocs reuse matrix to decide the first extracted patterns without copying unsafe defaults

## Blockers

- none technical right now
- the only real blocker is changing core rules before freezing them

## Notes

- legacy backend code is intentionally not part of the active plan right now
- MetalDocs may inform selective reuse, but only through the migration matrix and frozen architecture rules
- future code work should start from the frozen documents and the current foundation baseline, not from ad hoc memory
