# tasks/lessons.md
# Read at the start of EVERY session before touching any code.
# Add new lessons after every correction. Never delete existing lessons.

---

## Lesson A â€” pgdb.BeginTenantTx on every adapter query
Wrong:   db.QueryContext(ctx, "SELECT ... WHERE tenant_id=$1", id)
Correct: tx := pgdb.BeginTenantTx(...); tx.QueryContext(ctx, "... WHERE tenant_id=current_tenant_id()")
Rule:    Every Postgres adapter query uses pgdb.BeginTenantTx. No exceptions.
Layer:   Go adapter

## Lesson B â€” Handler checks principal then tenant before any operation
Wrong:   h.service.DoSomething(r.Context(), hardcodedID)
Correct: PrincipalFromContext â†’ 401; TenantFromContext â†’ 403; then service call
Rule:    Both checks mandatory before any service call in any handler.
Layer:   Go handler

## Lesson C â€” Outbox AppendInTx before tx.Commit, never after
Wrong:   tx.Commit(); outbox.Append(ctx, records)
Correct: outbox.AppendInTx(ctx, tx, records); tx.Commit()
Rule:    Outbox events inside same transaction as INSERT. Atomic or nothing.
Layer:   Go adapter + events

## Lesson D â€” Worker sets tenant context before every write transaction
Wrong:   cur.execute("INSERT INTO ...")
Correct: BEGIN; set_config('app.current_tenant_id',%s,true); INSERT; COMMIT
Rule:    set_config before every write tx. tenant_id from event payload, never hardcoded.
Layer:   Python worker

## Lesson E â€” No fetch() in React â€” platform-sdk only
Wrong:   useEffect(() => { fetch('/api/v1/x').then(...) }, [])
Correct: const { data, isLoading, error } = useXxx() from platform-sdk
Rule:    All frontend data via @metalshopping/platform-sdk hooks.
Layer:   Frontend

## Lesson F â€” Check packages/ui before creating any component
Wrong:   Creating StatusBadge when StatusPill already exists in packages/ui
Correct: Check packages/ui/src/index.ts first. Extend before creating new.
Rule:    Always check packages/ui exports before any new presentational component.
Layer:   Frontend

## Lesson 17 — Normalize trailing slashes in sub-resource routes
Date: 2026-03-22 | Trigger: correction
Wrong:   Parsing /runs/{id}/items/ without trimming trailing slash, causing run lookup to fail.
Correct: Trim trailing slash on run path before suffix matching.
Rule:    Normalize trailing slashes before route suffix parsing.
Layer:   Go handler

## Lesson 18 — Do not let inferred lookup_mode override lookup_policy
Date: 2026-03-22 | Trigger: correction
Wrong:   Using stored lookup_mode to force REFERENCE even when supplier policy is EAN_FIRST.
Correct: Honor lookup_mode only for manual/URL-backed signals; otherwise follow lookup_policy and infer mode from lookup_term.
Rule:    Inferred lookup_mode must not override lookup_policy unless manual override or URL exists.
Layer:   Python worker

## Lesson G â€” ADR done only when acceptance test passes and committed
Wrong:   Mark ADR Accepted after writing the document
Correct: Write â†’ implement â†’ run acceptance test â†’ commit "docs(adr): ADR-XXXX â€” verified and closed"
Rule:    ADR is DONE only when acceptance test passes and git commit is made.
Layer:   Process

## Lesson H â€” Every completed task needs a commit
Wrong:   Finishing tasks without committing, session ends with uncommitted work
Correct: git commit -m "feat(<scope>): <what>" after every task
Rule:    One commit per task. No uncommitted work at session end.
Layer:   Process

---
<!-- Project-specific lessons below this line -->

## Lesson 1 â€” ADR implementation must follow `$ms` plan + evidence
Date: 2026-03-21 | Trigger: correction
Wrong:   ADR checklist and acceptance evidence drift from `tasks/todo.md` and the `$ms` task workflow.
Correct: Keep `tasks/todo.md` as the execution source of truth (T1â€“T6), and mark ADR `accepted` only after real evidence + close-out commit.
Rule:    ADRs describe decisions; execution is tracked and verified via `$ms` tasks with evidence and commits.
Layer:   Process

## Lesson 2 â€” `tasks/todo.md` must match the `$ms` Phase 2 template
Date: 2026-03-21 | Trigger: correction
Wrong:   `tasks/todo.md` drifts (old formatting/encoding, missing Phase 2 framing, references to removed skills).
Correct: Rewrite `tasks/todo.md` as the `$ms` plan of record (Phase 1 summary + Phase 2 tasks + acceptance tests) and keep it aligned with the ADR.
Rule:    The plan that drives execution must be unambiguous, current, and skill-aligned.
Layer:   Process

## Lesson 3 â€” Cast Postgres placeholders used in string logic
Date: 2026-03-21 | Trigger: correction
Wrong:   Using `$3` (unknown type) in string checks like `$3 = ''` without an explicit cast.
Correct: Cast placeholders when mixing null/empty-string logic, e.g. `NULLIF($3::text, '')` and `CASE WHEN NULLIF($3::text, '') IS NULL THEN ...`.
Rule:    Ensure SQL placeholders have deterministic types when used in both value and predicate contexts.
Layer:   Go adapter

## Lesson 4 — Manual panel refresh must be isolated and supplier filter must not hide valid rows
Date: 2026-03-21 | Trigger: correction
Wrong:   Reusing a global reload tick for manual refresh and gating table render by "all suppliers" before checking loaded rows.
Correct: Use dedicated refresh state for manual panel updates and prioritize rendering rows whenever data exists, with supplier filter in true multi-select mode.
Rule:    Manual URL panel interactions must not trigger global screen reloads nor suppress valid SKU rows.
Layer:   Frontend

## Lesson 5 — psql set_config needs session or BEGIN
Date: 2026-03-21 | Trigger: correction
Wrong:   select set_config('app.current_tenant_id','tenant_default', true); then run SELECTs in later statements expecting tenant to persist.
Correct: Use select set_config('app.current_tenant_id','tenant_default', false) for session scope, or wrap queries in BEGIN; set_config(..., true); SELECT...; COMMIT.
Rule:    RLS depends on app.current_tenant_id; when validating via psql, ensure the setting persists across statements.
Layer:   Process

## Lesson 6 — SDK runtime must pass through optional contract fields
Date: 2026-03-21 | Trigger: correction
Wrong:   parseShoppingManualUrlCandidate ignores pnInterno/reference/ean so UI receives undefined.
Correct: SDK runtime validates and returns pnInterno/reference/ean when present in API payload.
Rule:    When contracts add optional fields, the runtime parser must surface them end-to-end.
Layer:   Frontend

## Lesson 7 — UI copy should match legacy table semantics
Date: 2026-03-21 | Trigger: correction
Wrong:   Showing "REF:" and "EAN:" labels when legacy UI uses values only.
Correct: Render reference on first line and EAN on second line without extra labels.
Rule:    Preserve legacy visual semantics unless explicitly changing UX copy.
Layer:   Frontend

## Lesson 8 — Manual product meta line breaks follow legacy layout
Date: 2026-03-21 | Trigger: correction
Wrong:   Showing fornecedor/marca/grupo on a single line with separators.
Correct: Render "Fornecedor: X Marca: Y" on the first line, and "Grupo: Z" on the second line.
Rule:    Match legacy line breaks for product meta in the manual URL table.
Layer:   Frontend

## Lesson 9 — "Mostrar URLs cadastradas" filters by productUrl
Date: 2026-03-21 | Trigger: correction
Wrong:   Showing rows without URL when toggle is active.
Correct: Filter manual URL candidates to rows with productUrl when the toggle is active.
Rule:    UI toggles must align to literal filter semantics.
Layer:   Frontend

## Lesson 10 — Data migration should not land as product code
Date: 2026-03-21 | Trigger: correction
Wrong:   Adding a one-off migration command under `apps/server_core/cmd/` for a data copy request.
Correct: Use ad-hoc DB scripts/commands (or a temporary local-only script) and keep the repo free of one-off migration executables unless explicitly requested.
Rule:    Keep product code focused; operational data moves should be reproducible without polluting app entrypoints.
Layer:   Process

## Lesson 11 — URL-only filter must be server-side with pagination
Date: 2026-03-21 | Trigger: correction
Wrong:   Filtering `productUrl` on the client after a paginated response.
Correct: Add a server-side `only_with_url` filter so pagination/total reflect rows with URL.
Rule:    Filters that change row eligibility must be applied before pagination.
Layer:   Frontend

## Lesson 12 — Show cumulative counts in paginated footer
Date: 2026-03-21 | Trigger: correction
Wrong:   "Mostrando" always shows page size (10) even on later pages.
Correct: Display `offset + returned` capped by total so page 2 shows 20, etc.
Rule:    Pagination footers should reflect cumulative items shown.
Layer:   Frontend

## Lesson 13 — Multi-select filters require array query params
Date: 2026-03-21 | Trigger: correction
Wrong:   Keeping single-value `brand_name`, `taxonomy_leaf0_name`, and `status` while asking for multi-select.
Correct: Accept arrays in the contract, parse repeated params server-side, and use `ANY($n)` in SQL.
Rule:    Multi-select UI must be backed by multi-value API filters.
Layer:   Backend + Frontend

## Lesson 14 — UI must not freeze at "queued" after async completion
Date: 2026-03-22 | Trigger: correction
Wrong:   Displaying the initial "queued" response only, causing confusion when the worker completes quickly.
Correct: Show live request status (polling result) next to the created request id.
Rule:    When we poll status, surface it in the same place the user watches.
Layer:   Frontend

## Lesson 15 — KPI cards must match the underlying metric
Date: 2026-03-22 | Trigger: correction
Wrong:   Labeling KPI cards as "OK / Not Found / Ambiguous / Error" while binding them to `ShoppingSummaryV1` (run counts).
Correct: Bind those KPI cards to per-run item counts grouped by `shopping_price_run_items.item_status` for the selected run.
Rule:    UI labels must reflect the metric source; run-level and item-level counters are not interchangeable.
Layer:   Frontend + Backend

## Lesson 16 — Playwright price parsing must use document HTML and price bounds
Date: 2026-03-22 | Trigger: correction
Wrong:   Using `page.content()` after `waitUntil=commit`, causing a minimal DOM snapshot and regex matching arbitrary digits, leading to wrong prices and DB numeric overflow.
Correct: Prefer `response.text()` (document HTML) and fallback to `page.content()` only when needed; reject out-of-range prices before writing to Postgres.
Rule:    Playwright PDP parsing must use the actual document HTML and enforce sane numeric bounds to keep writes safe.
Layer:   Python worker

## Lesson 19 — Run item log URL must not rely only on productUrl
Date: 2026-03-22 | Trigger: correction
Wrong:   Showing `URL: -` in the log when `productUrl` is empty, even though the worker wrote the attempted/final URL into `notes`.
Correct: Derive the displayed URL from `productUrl` OR parse `final_url/request_url` (or any http URL) from `notes` for debug visibility.
Rule:    Debug UX should surface attempted URLs without polluting durable URL fields.
Layer:   Frontend

## Lesson 20 — History panel must be bounded by detail panel height
Date: 2026-03-22 | Trigger: correction
Wrong:   Letting the recent history list define its own height, causing layout stretch and no internal scroll with many runs.
Correct: Measure the run detail card with `ResizeObserver` and set history scroll `maxHeight` from that value (mobile fallback fixed).
Rule:    Side-by-side panels must share bounded height to keep content readable and scrollable.
Layer:   Frontend

## Lesson 21 — Search URL for run logs should be computed at read-time
Date: 2026-03-22 | Trigger: correction
Wrong:   Depending on persisted `product_url` to display the executed lookup URL in run logs.
Correct: Compute the search URL from active manifest templates (`searchUrl`/`endpointTemplate`/`startUrl`) plus `lookup_term` while reading run items.
Rule:    Debug-only lookup URLs should be derived from runtime config, not persisted as durable data.
Layer:   Go adapter

## Lesson 22 — Search URL derivation must mirror runtime strategy keys
Date: 2026-03-22 | Trigger: correction
Wrong:   Deriving URLs from partial keys and ignoring `searchUrlTemplate` and VTEX (`baseUrl + persisted query params`), causing only Leroy URLs to appear.
Correct: Derive using the same strategy-facing config keys as the worker (`searchUrlTemplate`, template placeholders, and VTEX URL builder semantics).
Rule:    Debug projections must follow runtime contract keys, not ad-hoc subsets.
Layer:   Go adapter

## Lesson 23 — HTTP runtime should record search_url in notes
Date: 2026-03-22 | Trigger: correction
Wrong:   Returning only status notes for VTEX/HTML so the UI cannot show the executed search URL.
Correct: Prefix `notes` with `search_url=` for HTTP strategies when a lookup URL is built.
Rule:    Driver notes should include the actual lookup URL for debugging.
Layer:   Python worker

## Lesson 24 — VTEX needs separate storefront search template
Date: 2026-03-22 | Trigger: correction
Wrong:   Using the VTEX GraphQL persisted-query URL as `search_url`, which is not the human storefront search.
Correct: Add `debugSearchUrlTemplate` in manifest and expose it as `search_url`, while keeping GraphQL as `api_url` for deep debug.
Rule:    For VTEX suppliers, keep a dedicated storefront search URL template for logs.
Layer:   Worker + Config


## Lesson 25 — Skills must match harness behavior and stay compact
Date: 2026-03-22 | Trigger: correction
Wrong:   Letting `$ms` claim automatic plan mode and duplicating long implementation examples inside skills.
Correct: Use `update_plan` for complex work, ask the user to run `/plan` manually when needed, and keep skills short by pointing to repo references for concrete patterns.
Rule:    Skill workflows must reflect actual tool capabilities and minimize duplicated context.
Layer:   Process

## Lesson 26 — New feature package requires tsconfig path + include wiring
Date: 2026-03-22 | Trigger: correction
Wrong:   Importing `@metalshopping/feature-analytics` without adding path mapping/include in `apps/web/tsconfig.json`, breaking `tsc --noEmit`.
Correct: Register new workspace feature package in `compilerOptions.paths` and `include` for the web app before using it in routes/pages.
Rule:    Every new frontend feature package must be wired into web tsconfig resolution before consumption.
Layer:   Frontend

## Lesson 27 — New feature package also needs Vite alias wiring
Date: 2026-03-22 | Trigger: correction
Wrong:   Adding `@metalshopping/feature-analytics` only in `tsconfig.json`, so Vite build cannot resolve package entry.
Correct: Add the same feature package alias in `apps/web/vite.config.ts` (and test includes when applicable) before importing it in web routes/pages.
Rule:    In this workspace, every new feature package must be wired in both TypeScript and Vite resolution.
Layer:   Frontend

## Lesson 28 — Legacy-first visual parity needs snapshot copy plus runnable shell
Date: 2026-03-22 | Trigger: correction
Wrong:   Building a simplified Analytics Home that diverges from legacy layout before freezing visual parity.
Correct: Copy legacy TSX/CSS into a local snapshot and deliver a runnable analytics shell with tabs/cards matching legacy visual structure first, then adapt integration incrementally.
Rule:    For migration tasks marked "visual first", lock the UI parity baseline before deeper backend/SDK adaptation.
Layer:   Frontend

## Lesson 29 — Literal legacy copy requires compatibility shims before parity checks
Date: 2026-03-22 | Trigger: correction
Wrong:   Copying legacy Analytics TSX/CSS without restoring expected app/session/ui dependencies, causing compile/runtime break before visual validation.
Correct: Mirror legacy page tree and add local compatibility layer (`AppProviders`, ui wrappers, DTO adapters, registry resolver, mocks) first, then validate visual parity.
Rule:    In visual-first migrations, boot the copied legacy surface with shims and mocks before any backend integration.
Layer:   Frontend

## Lesson 30 — Legacy Analytics parity needs top navigation shell, not only home cards
Date: 2026-03-22 | Trigger: correction
Wrong:   Rendering only home cards without the legacy analytics top navigation context (tab rail/title shell).
Correct: Add a dedicated analytics top shell (brand + tab rail) in the legacy page container and keep tab routing on `/analytics/:tab?`.
Rule:    For visual parity on legacy analytics, replicate shell context (top rail + tab labels) before fine-tuning inner widgets.
Layer:   Frontend

## Lesson 31 — Visual parity depends on legacy payload shape, not only CSS
Date: 2026-03-22 | Trigger: correction
Wrong:   Keeping mock data in a simplified schema (`matrix`, generic action codes) so legacy widgets rendered empty blocks.
Correct: Shape mocks to the same legacy keys expected by the viewmodel (`actions_today.buckets`, `health_radar.cells`, `top_metal.best_*`, `kpis_products.capital_brl_total`).
Rule:    In legacy-first front migration, mirror the payload contract used by the legacy viewmodel before judging CSS parity.
Layer:   Frontend

## Lesson 32 — Legacy CSS modules need local token defaults to avoid flat/unstyled cards
Date: 2026-03-22 | Trigger: correction
Wrong:   Reusing legacy module classes that depend on global tokens (`--surface`, `--radius-lg`, `--grid-gap`) without defining them in the new scope.
Correct: Define fallback token variables at the analytics home page root so cards keep radius, padding, border and shadows.
Rule:    When migrating legacy CSS modules, provide local token defaults before visual parity tuning.
Layer:   Frontend

## Lesson 33 — Analytics Home must receive DTO and operational payload
Date: 2026-03-22 | Trigger: correction
Wrong:   Rendering `<AnalyticsHomePage />` without passing `dto`, while loading home data with `includeOperational: false`.
Correct: Render `<AnalyticsHomePage dto={dto} ... />` and request `api.home.workspace(..., { includeOperational: true })` so cards/lists receive data.
Rule:    Legacy-first visual pages only achieve parity when the copied component gets its expected payload shape.
Layer:   Frontend

## Lesson 34 — Animations must not hide content by default
Date: 2026-03-22 | Trigger: correction
Wrong:   Setting `opacity: 0` + translate on hero blocks, relying on CSS animation to reveal; when animations are disabled, the page looks blank.
Correct: Keep elements visible by default and define the hidden state in `@keyframes from { opacity: 0; ... }`.
Rule:    Never rely on animations to make core UI visible; animations are optional.
Layer:   Frontend

## Lesson 35 — Mock payloads must match viewmodel keys (or be defensively parsed)
Date: 2026-03-22 | Trigger: correction
Wrong:   Returning `kpis_series.data` with keys like `revenue_series/margin_series` while the viewmodel reads `sales_6m/margin_6m/runs_7d`, causing `undefined.length` crash.
Correct: Align mocks to the expected keys and default non-array fields to `[]` before reading `.length`.
Rule:    Visual-first migrations need strict payload-shape parity; otherwise UI fails before CSS parity.
Layer:   Frontend

## Lesson 36 — Top bar parity needs a right-side status + theme control and non-stretch tab rail
Date: 2026-03-22 | Trigger: correction
Wrong:   Leaving the tab rail in a `1fr` grid column without right controls, stretching the pill background across the full width and diverging from legacy.
Correct: Add right controls (Online + theme) and make the tab rail `fit-content` so it hugs tabs like legacy.
Rule:    For shell parity, replicate both layout slots (left/center/right) and constrain the center rail width.
Layer:   Frontend

## Lesson 37 — Legacy top bar container is a header strip, not a card
Date: 2026-03-22 | Trigger: correction
Wrong:   Styling the analytics top bar with card traits (container padding, rounded border, shadow, translucent panel).
Correct: Keep the top bar container flush and transparent, styling only inner controls (tab rail, status pill, theme button) to match legacy.
Rule:    In visual parity tasks, preserve the same visual hierarchy level as legacy containers.
Layer:   Frontend

## Lesson 38 — Match legacy shell by preserving header + inner wrapper split
Date: 2026-03-22 | Trigger: correction
Wrong:   Applying all top bar styles directly on header children without the `header > inner` structure used in legacy.
Correct: Keep a sticky header strip for backdrop/border and a dedicated inner wrapper for spacing/alignment (`padding: 14px 28px`).
Rule:    For pixel parity, copy both markup structure and CSS roles, not only visual tokens.
Layer:   Frontend

## Lesson 39 — Legacy migrations must neutralize missing deps and strict typing quickly
Date: 2026-03-22 | Trigger: correction
Wrong:   Copying legacy pages that import unavailable libs/types, leaving `tsc --noEmit` broken.
Correct: Add local shims (or replace deps with lightweight placeholders) and use `// @ts-nocheck` for large legacy files to keep migration runnable.
Rule:    Visual-first legacy migrations must compile before parity review; remove missing deps and strict type blockers early.
Layer:   Frontend

## Lesson 40 — Never leave invisible backdrops mounted across tab switches
Date: 2026-03-22 | Trigger: correction
Wrong:   Rendering fixed backdrops/drawers even when "closed" (relying on CSS `pointer-events: none`), allowing a style/regression to block all clicks on some routes.
Correct: Unmount backdrops/drawers when not open and force-close on tab switches.
Rule:    If a component can block interactions, it must not exist in the DOM when inactive.
Layer:   Frontend

## Lesson 41 — Package internals must not import their own public root
Date: 2026-03-22 | Trigger: correction
Wrong:   Importing DTO builders from `@metalshopping/feature-analytics` inside `packages/feature-analytics/src/pages/...`, creating a runtime cycle through `index.ts -> LegacyAnalyticsSurface -> AnalyticsPage -> AnalyticsProductsPage`.
Correct: Internal package files import sibling modules directly via relative paths; only external consumers use the package root.
Rule:    A package public barrel is for consumers, not for modules inside the same package.
Layer:   Frontend

## Lesson 42 — Analytics tab switches must reset shell-only UI state and scroll
Date: 2026-03-22 | Trigger: correction
Wrong:   Preserving the previous tab scroll position when switching analytics tabs, so `/analytics/products` could appear blank/frozen after navigating back from another surface.
Correct: On every analytics tab change, clear transient shell state (drawer/backdrop) and reset `window.scrollTo(0, 0)` before rendering the next surface.
Rule:    In SPA tab shells, unrelated surfaces must not inherit arbitrary scroll or blocking UI state from the previous tab.
Layer:   Frontend

## Lesson 43 — Context action callbacks consumed by effects must have stable identity
Date: 2026-03-22 | Trigger: correction
Wrong:   Recreating `setProductsOverviewSnapshot` on each provider render, while `AnalyticsProductsPage` used it in `useEffect` deps, causing repeated overview reloads and UI lock under tab navigation.
Correct: Expose snapshot callbacks with `useCallback` and keep lookup/read backed by a ref so effect dependencies stay stable.
Rule:    Any context callback used in hook dependency arrays must be referentially stable to prevent render-request loops.
Layer:   Frontend

## Lesson 44 — Auth/tenant failures must preserve HTTP status in logs
Date: 2026-03-22 | Trigger: correction
Wrong:   After writing a 403 response for missing tenant context, the handler still recorded `statusCode=401` in its deferred logger due to a hardcoded fallback.
Correct: Return the concrete auth status (401 vs 403) from the auth gate and propagate it to the request logger.
Rule:    Deferred request logging must reflect the actual response status, especially for auth/tenancy gates.
Layer:   Go handler

## Lesson 45 — Legacy UTF-8 BOM can break patching
Date: 2026-03-22 | Trigger: correction
Wrong:   Trying to `apply_patch` the first import lines of a TSX copied from legacy with a UTF-8 BOM, causing context mismatches.
Correct: Detect BOM (EF BB BF) and edit via BOM-safe rewrite (`Get-Content -Raw` + `Set-Content -Encoding utf8`) or normalize the file before patching.
Rule:    When migrating legacy files, handle UTF-8 BOM explicitly so automated patches match the expected lines.
Layer:   Process
