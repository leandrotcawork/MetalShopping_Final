# tasks/todo.md
# Active TODOs only. Completed history archived in `tasks/todo_archive_2026-03-22.md`.

## Active: ADR-0045 Manual URL candidates

## Tasks
- [ ] T5b: frontend - manual URL panel UX fixes
      commit: "fix(web): stabilize manual URL panel interactions"
- [ ] T6: ADR close-out - capture evidence and accept ADR-0045
      commit: "docs(adr): ADR-0045 manual URL candidates - verified and closed"

## Manual acceptance
- [ ] In browser: Manual URL panel lists products even with empty signals; saving a URL creates signal and table reflects overlay
- [ ] In browser: manual URL panel refreshes without layout jump, supplier=all behaves per spec, and toggle slider aligns

---

## Active: Shopping run progress (polling)

## Dev/acceptance
- [ ] `go build ./apps/server_core/...` passes
- [ ] `GET /api/v1/shopping/run-requests/{id}` returns progress fields
- [ ] `GET /api/v1/shopping/runs/{run_id}/item-status-summary` returns grouped counts
- [ ] In browser: progress bar updates over time with worker running
- [ ] In browser: KPI cards reflect selected run item counts (OK/NOT_FOUND/AMBIGUOUS/ERROR)
- [ ] In browser: "Historico recente" shows max N with "Ver tudo"
- [ ] In dev: worker keeps running with empty queue when enabled (`MS_SHOPPING_KEEP_ALIVE=true`)

---

## Active: Obra Facil Playwright performance hardening

## Manual acceptance
- [ ] Progress UI keeps updating during run (no regression)
