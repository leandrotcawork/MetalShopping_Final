# tasks/lessons.md

Read this file at the start of EVERY session before touching any code.
Add new lessons after every correction. Never delete existing lessons.

---

## Lesson A — pgdb.BeginTenantTx on every adapter query
Wrong:  db.QueryContext(ctx, "SELECT ... WHERE tenant_id=$1", id)
Correct: tx := pgdb.BeginTenantTx(...); tx.QueryContext(ctx, "... WHERE tenant_id=current_tenant_id()")
Rule: Every Postgres adapter query goes through pgdb.BeginTenantTx. No exceptions.
Layer: Go adapter

## Lesson B — Handler checks principal then tenant before any operation
Wrong:  h.service.DoSomething(r.Context(), hardcodedID)
Correct: PrincipalFromContext → 401; TenantFromContext → 403; then service call
Rule: Both checks mandatory before any service call in any handler.
Layer: Go handler

## Lesson C — Outbox AppendInTx before tx.Commit, never after
Wrong:  tx.Commit(); outbox.Append(ctx, records)
Correct: outbox.AppendInTx(ctx, tx, records); tx.Commit()
Rule: Outbox events inside same transaction as INSERT. Atomic.
Layer: Go adapter + events

## Lesson D — Worker sets tenant context before every write tx
Wrong:  cur.execute("INSERT INTO ...")
Correct: cur.execute("BEGIN"); cur.execute("set_config('app.current_tenant_id',%s,true)",(tid,)); INSERT; COMMIT
Rule: set_config before every write transaction. tenant_id from event payload.
Layer: Python worker

## Lesson E — No fetch() in React — platform-sdk only
Wrong:  useEffect(() => { fetch('/api/v1/x').then(...) }, [])
Correct: const { data, isLoading, error } = useXxx() from platform-sdk
Rule: All frontend data via @metalshopping/platform-sdk hooks.
Layer: Frontend

## Lesson F — Check packages/ui before creating any component
Wrong:  Creating StatusBadge when StatusPill exists in packages/ui
Correct: Check packages/ui/src/index.ts first. Extend before creating.
Rule: Always check packages/ui exports before any new presentational component.
Layer: Frontend

## Lesson G — ADR done only when acceptance test passes and committed
Wrong:  Mark ADR Accepted after writing document
Correct: Write → implement → run test → commit "docs(adr): ADR-XXXX — verified and closed"
Rule: ADR is DONE only when acceptance test passes and commit is made.
Layer: Process

## Lesson H — Every completed task needs a commit
Wrong:  Finishing tasks without committing, session ends with uncommitted work
Correct: git commit -m "feat(<scope>): <what>" after every task
Rule: One commit per task. No uncommitted work at session end.
Layer: Process

---
<!-- Project-specific lessons below -->
