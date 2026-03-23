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
