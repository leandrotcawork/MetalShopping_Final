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

## Next

- keep ADR set complete and stable
- close the contract validation and generation gap behind the current authored contracts
- introduce real event publication and outbox discipline for at least one mutation flow
- evolve governance from first database-backed feature flags into broader thresholds and policies
- expand `catalog` from the current foundation slice to the canonical product model
- apply the new catalog taxonomy migration and validate the runtime slice end-to-end
- evolve bootstrap auth toward a production-grade identity integration path
- continue domain expansion from the proven tenant-aware foundation
- define real generators and validators behind the script entrypoints
- prepare phase transition checklist from planning to implementation
- use the MetalDocs reuse matrix to decide the first extracted patterns without copying unsafe defaults

## Blockers

- none technical right now
- the only real blocker is changing core rules before freezing them

## Notes

- legacy backend code is intentionally not part of the active plan right now
- MetalDocs may inform selective reuse, but only through the migration matrix and frozen architecture rules
- future code work should start from the frozen documents and the current foundation baseline, not from ad hoc memory
