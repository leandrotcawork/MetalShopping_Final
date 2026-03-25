# tasks/lessons.md
# Read at the start of EVERY session before touching code.
# This file stores only structural/global lessons (architecture, contracts, process, platform).
# Do NOT add page-specific cosmetic fixes here.

---

## Lesson 1 — Tenant-safe DB access is mandatory
Date: 2026-03-23 | Trigger: baseline
Wrong:   Running adapter queries without tenant transaction/runtime tenant scope.
Correct: Every Postgres adapter query uses `pgdb.BeginTenantTx` and tenant-scoped predicates (`current_tenant_id()` where applicable).
Rule:    No tenant-scoped read/write may run outside tenant transaction and tenant filters.
Layer:   Go adapter

## Lesson 2 — Handlers must fail fast on auth and tenancy
Date: 2026-03-23 | Trigger: baseline
Wrong:   Calling service logic before validating principal/tenant context.
Correct: `PrincipalFromContext` (401) and `TenantFromContext` (403) execute before any handler operation.
Rule:    Auth and tenancy checks are always first in protected handlers.
Layer:   Go handler

## Lesson 3 — Outbox must be atomic with writes
Date: 2026-03-23 | Trigger: baseline
Wrong:   Appending events after `tx.Commit`.
Correct: Use `outbox.AppendInTx` in the same transaction before commit.
Rule:    Write + event publish intent must be atomic or not happen.
Layer:   Go adapter + events

## Lesson 4 — Worker writes require tenant context and idempotency
Date: 2026-03-23 | Trigger: baseline
Wrong:   Writing without `set_config('app.current_tenant_id', ...)` or without conflict-safe upsert semantics.
Correct: Start write tx with tenant `set_config` and use `ON CONFLICT ... DO UPDATE` where applicable.
Rule:    Worker write paths must be tenant-safe and retry-safe.
Layer:   Python worker

## Lesson 5 — Frontend data flow must use platform SDK contracts
Date: 2026-03-23 | Trigger: baseline
Wrong:   Fetching data directly via `fetch()` or bypassing contract-generated runtime types.
Correct: Use `@metalshopping/platform-sdk` hooks/runtime and keep loading/error/empty states explicit.
Rule:    Frontend data access is SDK-first, contract-aligned, and state-complete.
Layer:   Frontend

## Lesson 17 — Hover parity requires selector specificity on label elements
Date: 2026-03-23 | Trigger: correction
Wrong:   Relying on generic hover color rules that are overridden in composed table header styles.
Correct: Scope hover color to the exact interactive label selector in the spotlight header (`th .spotlightSkuSortBtn`) with sufficient specificity.
Rule:    For legacy visual parity, hover behavior must target the final rendered label node, not only parent containers.
Layer:   Frontend

## Lesson 18 — Table header hover must bind to explicit label node
Date: 2026-03-23 | Trigger: correction
Wrong:   Depending on inherited text color from sortable header button for hover state.
Correct: Wrap header text in a dedicated label element and style hover/focus directly on that label.
Rule:    Interactive table headers should expose a stable label selector for deterministic hover parity.
Layer:   Frontend

## Lesson 19 — Portaled UI must redeclare local CSS tokens
Date: 2026-03-23 | Trigger: correction
Wrong:   Using CSS variables scoped to page containers in components rendered via portal (`document.body`), causing unresolved hover colors.
Correct: Redeclare required tokens (`--wine`, `--muted`, `--surface`, etc.) on the portal root (`.drawer`) with dark-theme overrides.
Rule:    Any portaled surface must be self-sufficient for token resolution.
Layer:   Frontend

## Lesson 20 — Header hover must be bound to interactive target only
Date: 2026-03-23 | Trigger: correction
Wrong:   Applying hover color on `th:hover`, which changes label color even outside the clickable text area.
Correct: Bind hover/focus color only to the sortable button label (`.spotlightSkuSortBtn:hover .spotlightSkuHeadLabel`).
Rule:    Visual feedback must match the real interactive hit area.
Layer:   Frontend

## Lesson 21 — Feature CSS modules need local token baselines
Date: 2026-03-23 | Trigger: correction
Wrong:   Using `var(--surface)` / `var(--surface-border)` in feature styles without defining a local token baseline, causing transparent controls.
Correct: Define token defaults on the page root class and dark overrides in the same module.
Rule:    Any feature stylesheet that depends on design tokens must be self-sufficient for surface/border readability.
Layer:   Frontend

## Lesson 22 — Similar surfaces should share the same base fill
Date: 2026-03-23 | Trigger: correction
Wrong:   Using different background fills for peer blocks that must look like one surface family.
Correct: Keep peer containers (e.g., insights strip and table card) on the same base surface background and border.
Rule:    Visual parity improves when equivalent container roles use a single surface baseline.
Layer:   Frontend

## Lesson 23 — Feature code must import shared UI from package entrypoint
Date: 2026-03-23 | Trigger: correction
Wrong:   Importing duplicated/local UI wrappers instead of the registered shared component.
Correct: Import shared controls from `@metalshopping/ui` (`packages/ui/src/index.ts`) for consistent behavior and styling.
Rule:    If a component exists in `packages/ui`, feature modules consume it via package entrypoint.
Layer:   Frontend

## Lesson 24 — Remove local wrappers after migration to shared UI
Date: 2026-03-23 | Trigger: correction
Wrong:   Keeping dead wrapper files after switching to shared UI imports.
Correct: Delete redundant local wrappers to prevent accidental regressions to non-standard components.
Rule:    Migration to shared UI is complete only when obsolete wrapper paths are removed.
Layer:   Frontend

## Lesson 25 — Delete orphan facade files when usage hits zero
Date: 2026-03-23 | Trigger: correction
Wrong:   Leaving unused pass-through facade modules in the tree after import migration.
Correct: Remove zero-reference facades immediately once consumers are moved.
Rule:    Dead facade files create drift risk and must not remain after cleanup.
Layer:   Frontend

## Lesson 26 — Workspace root must provide token fallbacks
Date: 2026-03-23 | Trigger: correction
Wrong:   Relying on parent route tokens for workspace surfaces (`--surface`, `--muted`, `--wine`, `--success`), causing inconsistent rendering.
Correct: Define token fallbacks on workspace root `.page` with explicit dark-mode overrides.
Rule:    Self-contained route modules must declare required visual tokens at their own root.
Layer:   Frontend

## Lesson 27 — Second-pass parity must diff against legacy snapshot
Date: 2026-03-23 | Trigger: correction
Wrong:   Applying visual tweaks without checking exact delta versus legacy source-of-truth.
Correct: Run direct file diff against `legacy_snapshot` and keep only intentional structural deviations.
Rule:    In parity mode, every CSS delta from legacy must be explicit and justified.
Layer:   Frontend

## Lesson 28 — Do not downgrade legacy charts in parity phase
Date: 2026-03-23 | Trigger: correction
Wrong:   Replacing legacy chart components with simplified SVG placeholders during visual migration.
Correct: Preserve legacy chart implementation in parity phase and adapt only integration imports when required.
Rule:    Visual parity includes chart behavior and interaction fidelity, not only static layout.
Layer:   Frontend

## Lesson 6 — Reuse design system before adding UI primitives
Date: 2026-03-23 | Trigger: baseline
Wrong:   Creating local UI components already available in `packages/ui`.
Correct: Check `packages/ui/src/index.ts` first; extend shared primitives when needed.
Rule:    Prefer shared UI primitives over feature-local duplication.
Layer:   Frontend

## Lesson 7 — Generated artifacts are read-only outputs
Date: 2026-03-23 | Trigger: baseline
Wrong:   Editing files under `packages/generated/` manually.
Correct: Change contracts first, then regenerate SDK/artifacts.
Rule:    Source of truth is contracts; generated code is never hand-edited.
Layer:   Process

## Lesson 8 — Completion requires validation + commit
Date: 2026-03-23 | Trigger: baseline
Wrong:   Marking tasks complete without build/manual validation and commit.
Correct: Only mark done after required checks pass and commit is created.
Rule:    No completed task exists without evidence and commit.
Layer:   Process

## Lesson 9 — `tasks/todo.md` edits must be block-scoped
Date: 2026-03-23 | Trigger: baseline
Wrong:   Running global replacements that accidentally change unrelated feature blocks.
Correct: Edit only the intended feature section with scoped patches.
Rule:    Planning source-of-truth must not be mutated outside target scope.
Layer:   Process

## Lesson 10 — Legacy migration follows parity-first sequencing
Date: 2026-03-23 | Trigger: baseline
Wrong:   Mixing backend integration before visual parity or rewriting layout before baseline match.
Correct: Freeze baseline → literal copy → shims/mocks → parity pass → integration phases.
Rule:    For legacy migrations, visual parity is completed before backend adaptation.
Layer:   Frontend + Process

## Lesson 11 — Legacy CSS must define safe token fallbacks
Date: 2026-03-23 | Trigger: baseline
Wrong:   Using critical surface tokens without fallback, causing transparent/unstyled cards.
Correct: Define local fallback variables for surfaces/borders/radius in migrated CSS modules.
Rule:    Visual-critical styles in migrated pages must remain stable even if tokens are missing.
Layer:   Frontend

## Lesson 12 — Runtime behavior changes require operational verification
Date: 2026-03-23 | Trigger: baseline
Wrong:   Blaming code before checking migration/manifest/config/version/runtime restart state.
Correct: Verify DB migration, active manifest/config, restart status, and test data freshness first.
Rule:    Diagnose runtime/state drift before code-level fixes on worker/config-driven flows.
Layer:   Process

## Lesson 13 — Observability is part of the contract
Date: 2026-03-23 | Trigger: baseline
Wrong:   Returning generic errors/logs without actionable context.
Correct: Use structured error codes and request logs with `trace_id`, `action`, `result`, `duration_ms`.
Rule:    Debuggability requirements are mandatory, not optional.
Layer:   Backend + Process

## Lesson 14 — Keep this file high-signal
Date: 2026-03-23 | Trigger: baseline
Wrong:   Adding one-off UI tweaks (e.g., local border/padding adjustments) as global lessons.
Correct: Record only recurring, cross-cutting, architectural, or process-level rules.
Rule:    If a lesson is not reusable across modules, it belongs in feature notes, not here.
Layer:   Process

## Lesson 15 — Legacy migration must preserve interactive behavior
Date: 2026-03-23 | Trigger: correction
Wrong:   Replacing legacy interactive table logic with a simplified static version during visual migration.
Correct: Keep legacy interaction model (filters, sorting, pagination, state callbacks) and only adapt imports/wiring.
Rule:    In parity-first migration, functional UX behavior is part of visual fidelity and cannot be downgraded.
Layer:   Frontend

## Lesson 16 — Mock semantics must match UI contract keys
Date: 2026-03-23 | Trigger: correction
Wrong:   Feeding display labels like `Info` or `OK` into UI code that expects metric keys such as `giro_6m` and `margin_sales_pct`.
Correct: Mock payloads must preserve the exact semantic keys consumed by the migrated UI, even in visual-first phases.
Rule:    In legacy migration, mock data shape is part of the contract and cannot be approximated with human labels.
Layer:   Frontend

## Lesson 29 � Legacy copy must be normalized to local DTO shapes
Date: 2026-03-23 | Trigger: correction
Wrong:   Keeping legacy field names (`x`, `our`, `competitors`) and raw `Record<string, unknown>` values in typed code, causing runtime-safe but compile-broken paths.
Correct: Map legacy fields to current DTO contract (`date`, `our_price`, `suppliers`) and normalize unknown payload values with explicit converters before rendering.
Rule:    In parity migrations, literal copy is allowed, but every boundary to local contracts/types must be normalized explicitly.
Layer:   Frontend

## Lesson 30 � Preserve UTF-8 when patching legacy-copied frontend files
Date: 2026-03-23 | Trigger: correction
Wrong:   Editing UTF-8 files through a different code page (e.g., CP-1252) and writing back, which corrupts accents/emojis and changes UI text/icons.
Correct: Read/write migrated frontend files as UTF-8 and prefer patching tools that preserve encoding.
Rule:    Legacy parity files must keep original text encoding to avoid silent UI regressions.
Layer:   Frontend

## Lesson 31 � Remove `ts-nocheck` with minimal explicit callback typing
Date: 2026-03-23 | Trigger: correction
Wrong:   Keeping migrated files compile-green by disabling type checks globally (`// @ts-nocheck`).
Correct: Re-enable type checking and add explicit callback parameter typing at integration boundaries where DTOs are intentionally loose.
Rule:    Prefer localized type annotations over file-level typecheck suppression in migrated frontend modules.
Layer:   Frontend

## Lesson 32 � Simulator hero metrics need tolerant alias mapping
Date: 2026-03-23 | Trigger: correction
Wrong:   Depending on a single metric label in workspace payload causes business KPIs to disappear when label names vary.
Correct: Parse simulator baseline metrics using alias keys + deterministic fallback values (`preco real efetivo`, `gasto var`, gap baseline).
Rule:    UI-facing KPI extraction from loose payloads must be resilient to label variation.
Layer:   Frontend

## Lesson 33 � First-fold workspace KPIs must be rendered in ProductHero
Date: 2026-03-24 | Trigger: correction
Wrong:   Adding requested KPI fields only in tab-level content (e.g., Simulator) when the expected location is the workspace hero block.
Correct: Inject first-fold KPI metrics in `ProductHero`/`HeroMetrics` with resilient fallback mapping from available model fields.
Rule:    When a KPI is requested for the workspace header area, render it in the hero metrics source, not only inside tabs.
Layer:   Frontend

## Lesson 34 � Hero KPI order must be explicit, not payload-driven
Date: 2026-03-24 | Trigger: correction
Wrong:   Rendering workspace or simulator hero KPIs in the raw payload order, which lets backend/mock ordering drift break visual parity.
Correct: Build hero KPI lists from an explicit ordered view-model that matches the approved legacy surface.
Rule:    First-fold KPI strips must be rendered from fixed presentation order, never implicit source order.
Layer:   Frontend

## Lesson 35 � Workspace top bars that belong to the shell must be rendered as shell strips
Date: 2026-03-24 | Trigger: correction
Wrong:   Styling a workspace top bar as a self-contained route header/card with its own padded surface.
Correct: Render shell-level top bars in a full-width wrapper that owns background/border/sticky behavior, and keep the header component as inner constrained content only.
Rule:    If the bar is part of the shell, shell chrome lives in the outer wrapper, not in the page component itself.
Layer:   Frontend

## Lesson 36 � Shell bars inside padded app mains must break out at the route root
Date: 2026-03-24 | Trigger: correction
Wrong:   Recreating a shell top bar under a parent `main` with fixed padding but keeping it inset inside that padding box.
Correct: Match the legacy route shape (`section > header` + content container sibling) and let the outer header wrapper escape parent padding when needed.
Rule:    When app shell padding exists above the route, shell-level bars must break out at the route root or they will never visually match legacy.
Layer:   Frontend

## Lesson 37 � PowerShell entry scripts must declare `param` before executable statements
Date: 2026-03-25 | Trigger: correction
Wrong:   Setting `$ErrorActionPreference` before the `param(...)` block in a reusable PowerShell entry script.
Correct: Keep `param(...)` as the first executable construct, then apply runtime settings like `$ErrorActionPreference`.
Rule:    Any PowerShell script that accepts parameters must declare its parameter block before executable statements.
Layer:   Process
