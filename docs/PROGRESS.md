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
- operational surface recovery order now frozen as `Home -> Shopping -> Analytics -> CRM` under make-it-work-first
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
- `server_core` composition root refactored into capability-focused composition files to reduce bootstrap concentration in `main.go`
- frontend auth ownership hardened so `feature-auth-session` no longer depends on app runtime `apiBaseUrl`, while SDK composition is centralized in the web runtime provider
- login visual baseline hardened with shared token source-of-truth and cross-surface sync/check automation for React fallback and Keycloak theme
- auth/session bootstrap now enforces explicit mode semantics with fail-fast default (`required`) to avoid silent OIDC/session fallback
- `server_core` runtime now shuts down gracefully with signal-aware context cancellation and outbox dispatcher lifecycle wiring
- frontend runtime/facade extraction completed in `packages/platform-sdk`, removing transport code emission from `generate_contract_artifacts.ps1`
- SoT drift guard hardened with branch-diff mode (`-BaseRef`) and expanded structural scope (workspace manifests plus web runtime config) so CI can detect structural changes beyond local working tree
- pull request CI workflow now enforces contract validation, generated artifact drift check, backend tests, web typecheck/build, and SoT drift guard sequencing
- login closure governance was frozen with explicit scope (`docs/LOGIN_MVP_SCOPE.md`), DoD (`docs/LOGIN_DOD.md`), SDK boundary rules (`docs/SDK_BOUNDARY.md`), and tranche plan (`docs/LOGIN_MVP_EXECUTION_PLAN.md`)
- ADR-0015 and ADR-0016 were added to bind login closure governance and generated-vs-runtime SDK boundary semantics
- login T2 SDK/runtime hardening implemented with stable workspace alias imports in `packages/platform-sdk`, explicit CI guards for deep-relative-import and `as unknown as`, and environment-aware CI behavior for Docker availability
- local developer guidance was updated so Docker-based artifact generation is optional outside CI and Windows `esbuild EPERM` fallback is explicit (including WSL script path)
- login T3 local smoke automation implemented in `scripts/smoke_auth_session_local.ps1` and validated end-to-end against local Keycloak (`login -> me -> refresh/logout with CSRF`)
- OpenAPI generation check was hardened to run against a sanitized local contract mirror so JSON Schema canonical `$id` values do not force remote resolution during SDK generation drift checks
- frontend auth route behavior was hardened with explicit route policy functions plus unit tests for redirect/manual/authenticated modes and auto-redirect no-loop guard semantics
- first Home Level 1 slice delivered with real backend KPI summary endpoint and generated SDK consumption in the web surface
- Home backend ownership tightened from generic `internal/handlers` into explicit module structure `internal/modules/home` (`adapters/postgres`, `application`, `ports`, `transport/http`) without changing API contract behavior
- make-it-work-first development guideline documented as active delivery mode
- Home Level 1 acceptance was formally closed with objective evidence (`go build`, `go test`, `web:typecheck`, `web:build`, runtime boundary grep) in `docs/HOME_LEVEL1_ACCEPTANCE.md`
- Shopping API contract surface was frozen in draft (`contracts/api/openapi/shopping_v1.openapi.yaml`) with summary, runs list/detail, and latest-by-product schemas
- Shopping Level 1 advanced with server_core read endpoints wired, sdk-runtime facade methods available, and `apps/web` route `/shopping` now bound to real API data instead of placeholder
- integration worker scaffold for Shopping Price added in `apps/integration_worker/shopping_price_worker.py` (worker writes Postgres, Go reads Postgres)
- root `SKILLS.md` maps specialist skills by ordered step without a dedicated orchestrator skill
- Shopping Level 1 acceptance was formally closed with evidence in `docs/SHOPPING_LEVEL1_ACCEPTANCE.md`
- ADR-0021 frontend migration closure recorded with legacy workflow preserved and thin-client boundaries enforced (`docs/SHOPPING_ADR021_ACCEPTANCE.md`)
- Shopping Price Phase 2 ADRs (ADR-0025..ADR-0028) accepted with objective evidence (`docs/SHOPPING_ADR025_ACCEPTANCE.md` -> `docs/SHOPPING_ADR028_ACCEPTANCE.md`)
- Shopping Driver Runtime v1 (ADR-0029) accepted with pilot supplier (`OBRA_FACIL`) executed end-to-end in event smoke (`docs/SHOPPING_ADR029_ACCEPTANCE.md`)
- Driver strategy framework ADR (ADR-0030) accepted to freeze `family + strategy` scaling model before implementing real non-mock suppliers at scale
- Backend completion ADR tranche drafted for driver framework parity (ADR-0031..ADR-0034)
- ADR-0031 accepted: runtime extracted from `shopping_price_worker.py` into `apps/integration_worker/src/shopping_price_runtime/*`, compile checks passed, and smoke (`scripts/smoke_shopping_event_local.ps1`) passed outside sandbox
- ADR-0032 implemented: bounded runtime parallelism + HTTP rate limiting + retry status policy with schema and server-side validation updates; acceptance now depends on ADR-0033 multi-supplier smoke evidence
- ADR-0033 implemented: multi-supplier smoke suite and DB evidence report generator added (`scripts/smoke_shopping_driver_suite_local.ps1`, `scripts/smoke_shopping_driver_suite_report.py`) with generated report in `docs/SHOPPING_DRIVER_SUITE_ACCEPTANCE.md`
- ADR-0034 implemented: `playwright.pdp_first.v1` non-mock runtime delivered (real browser navigation, selector extraction, anti-block notes, bounded retries) with Playwright toolchain installed and direct runtime smoke returning `OK` + `PLAYWRIGHT` + non-zero price
- ADR-0042 accepted: `http.html_dom_first_card.v1` validated with `ABC` smoke `OK` (`run_request_id=aba7655c-d422-4960-8765-8627638fad47`)
- ADR-0040 accepted: legacy suppliers pack validated (`TELHA_NORTE`, `LEROY`, `ABC`) with non-mock smoke evidence and enabled suppliers directory list
- ADR-0041 accepted: `http.leroy_search_sellers.v1` validated with `LEROY` smoke `OK` and `http_status=200` (`run_request_id=fc383948-a124-4db1-b428-6c775eba8b8c`)
- ADR-0043/ADR-0044 accepted: legacy catalog import pipeline executed for `tenant_default` with reports and idempotent rerun evidence
- governance audit design spec added to consolidate documentation precedence, agent entrypoints, and SoT ownership before the next planning wave

## Next

- consolidate `PROJECT_SOT`, `AGENTS`, `CLAUDE`, `CODEX`, `ARCHITECTURE`, `IMPLEMENTATION_PLAN`, and `PROGRESS` under the accepted governance hierarchy
- create the master orchestration plan only after the governance tranche is fully aligned
- keep ADR set complete and stable
- implement manual URL candidates listing from catalog (ADR-0045) so the panel works when signals are empty
- execute Shopping frontend parity ADR tranche (ADR-0036..ADR-0039) before expanding Shopping UX/features
- execute Shopping legacy suppliers driver pack ADR tranche (ADR-0040..ADR-0042) to add `TELHA_NORTE`, `LEROY`, `ABC` under governed manifests (complete)
- freeze and follow the Shopping Price Level 2 ADR set (ADR-0017 .. ADR-0024) before implementation work starts
- keep CI workflow scope aligned with future structural package boundaries and new quality gates
- apply pricing migrations and database-backed governance defaults in runtime
- validate pricing write/list/current and outbox publication through smoke tests
- revise pricing semantics to align with the accepted canonical SKU data ownership model
- validate the revised pricing slice and migration path end-to-end after semantic alignment
- continue domain expansion from the proven tenant-aware foundation
- keep pricing and inventory semantically narrow while preparing procurement follow-on ownership
- freeze procurement canonical model and upstream integration gate before procurement contracts or runtime code
- decide the first canonical procurement inputs and contracts that integration must publish
- keep frontend migration execution aligned with the new migration matrix and data contract map
- prepare phase transition checklist from foundation hardening to domain expansion
- use the MetalDocs reuse matrix to decide the first extracted patterns without copying unsafe defaults

## Blockers

- none technical right now
- the only real blocker is changing core rules before freezing them

## Notes

- legacy backend code is intentionally not part of the active plan right now
- MetalDocs may inform selective reuse, but only through the migration matrix and frozen architecture rules
- future code work should start from the frozen documents and the current foundation baseline, not from ad hoc memory
