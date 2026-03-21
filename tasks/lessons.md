# tasks/lessons.md
# Read at the start of EVERY session before touching any code.
# Add new lessons after every correction. Never delete existing lessons.

---

## Lesson A — pgdb.BeginTenantTx on every adapter query
Wrong:   db.QueryContext(ctx, "SELECT ... WHERE tenant_id=$1", id)
Correct: tx := pgdb.BeginTenantTx(...); tx.QueryContext(ctx, "... WHERE tenant_id=current_tenant_id()")
Rule:    Every Postgres adapter query uses pgdb.BeginTenantTx. No exceptions.
Layer:   Go adapter

## Lesson B — Handler checks principal then tenant before any operation
Wrong:   h.service.DoSomething(r.Context(), hardcodedID)
Correct: PrincipalFromContext → 401; TenantFromContext → 403; then service call
Rule:    Both checks mandatory before any service call in any handler.
Layer:   Go handler

## Lesson C — Outbox AppendInTx before tx.Commit, never after
Wrong:   tx.Commit(); outbox.Append(ctx, records)
Correct: outbox.AppendInTx(ctx, tx, records); tx.Commit()
Rule:    Outbox events inside same transaction as INSERT. Atomic or nothing.
Layer:   Go adapter + events

## Lesson D — Worker sets tenant context before every write transaction
Wrong:   cur.execute("INSERT INTO ...")
Correct: BEGIN; set_config('app.current_tenant_id',%s,true); INSERT; COMMIT
Rule:    set_config before every write tx. tenant_id from event payload, never hardcoded.
Layer:   Python worker

## Lesson E — No fetch() in React — platform-sdk only
Wrong:   useEffect(() => { fetch('/api/v1/x').then(...) }, [])
Correct: const { data, isLoading, error } = useXxx() from platform-sdk
Rule:    All frontend data via @metalshopping/platform-sdk hooks.
Layer:   Frontend

## Lesson F — Check packages/ui before creating any component
Wrong:   Creating StatusBadge when StatusPill already exists in packages/ui
Correct: Check packages/ui/src/index.ts first. Extend before creating new.
Rule:    Always check packages/ui exports before any new presentational component.
Layer:   Frontend

## Lesson G — ADR done only when acceptance test passes and committed
Wrong:   Mark ADR Accepted after writing the document
Correct: Write → implement → run acceptance test → commit "docs(adr): ADR-XXXX — verified and closed"
Rule:    ADR is DONE only when acceptance test passes and git commit is made.
Layer:   Process

## Lesson H — Every completed task needs a commit
Wrong:   Finishing tasks without committing, session ends with uncommitted work
Correct: git commit -m "feat(<scope>): <what>" after every task
Rule:    One commit per task. No uncommitted work at session end.
Layer:   Process

---
<!-- Project-specific lessons below this line -->

## Lesson 1 — ADR implementation must follow `$ms` plan + evidence
Date: 2026-03-21 | Trigger: correction
Wrong:   ADR checklist and acceptance evidence drift from `tasks/todo.md` and the `$ms` task workflow.
Correct: Keep `tasks/todo.md` as the execution source of truth (T1–T6), and mark ADR `accepted` only after real evidence + close-out commit.
Rule:    ADRs describe decisions; execution is tracked and verified via `$ms` tasks with evidence and commits.
Layer:   Process

## Lesson 2 — `tasks/todo.md` must match the `$ms` Phase 2 template
Date: 2026-03-21 | Trigger: correction
Wrong:   `tasks/todo.md` drifts (old formatting/encoding, missing Phase 2 framing, references to removed skills).
Correct: Rewrite `tasks/todo.md` as the `$ms` plan of record (Phase 1 summary + Phase 2 tasks + acceptance tests) and keep it aligned with the ADR.
Rule:    The plan that drives execution must be unambiguous, current, and skill-aligned.
Layer:   Process

## Lesson 3 — Cast Postgres placeholders used in string logic
Date: 2026-03-21 | Trigger: correction
Wrong:   Using `$3` (unknown type) in string checks like `$3 = ''` without an explicit cast.
Correct: Cast placeholders when mixing null/empty-string logic, e.g. `NULLIF($3::text, '')` and `CASE WHEN NULLIF($3::text, '') IS NULL THEN ...`.
Rule:    Ensure SQL placeholders have deterministic types when used in both value and predicate contexts.
Layer:   Go adapter
