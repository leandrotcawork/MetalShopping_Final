# tasks/todo.md
# Active TODOs only. Completed history archived in `tasks/todo_archive_2026-03-22.md`.

## Active: ADR-0045 Manual URL candidates

## Tasks
- [x] T5b: frontend - manual URL panel UX fixes
      commit: "fix(web): stabilize manual URL panel interactions"
- [x] T6: ADR close-out - capture evidence and accept ADR-0045
      commit: "docs(adr): ADR-0045 manual URL candidates - verified and closed"

## Manual acceptance
- [x] In browser: Manual URL panel lists products even with empty signals; saving a URL creates signal and table reflects overlay
- [x] In browser: manual URL panel refreshes without layout jump, supplier=all behaves per spec, and toggle slider aligns

---

## Active: Shopping run progress (polling)

## Dev/acceptance
- [x] `go build ./apps/server_core/...` passes
- [x] `GET /api/v1/shopping/run-requests/{id}` returns progress fields
- [x] `GET /api/v1/shopping/runs/{run_id}/item-status-summary` returns grouped counts
- [x] In browser: progress bar updates over time with worker running
- [x] In browser: KPI cards reflect selected run item counts (OK/NOT_FOUND/AMBIGUOUS/ERROR)
- [x] In browser: "Historico recente" shows max N with "Ver tudo"
- [x] In dev: worker keeps running with empty queue when enabled (`MS_SHOPPING_KEEP_ALIVE=true`)

---

## Active: Obra Facil Playwright performance hardening

## Manual acceptance
- [x] Progress UI keeps updating during run (no regression)
