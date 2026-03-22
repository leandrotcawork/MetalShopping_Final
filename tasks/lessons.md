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

