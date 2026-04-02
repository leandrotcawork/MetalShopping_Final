# Task Tracker

## Current State

- State: completed
- Active tranche: ERP integration review-routing remediation
- Source spec: user request after commit `72395c37f7834ebc51be5a24572d13c6073790a1`

## Current Task

- Task: Fix Task 6 ERP review conversion idempotency and zero-row handling
- State: completed
- Scope: `apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository.go`, `apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository_test.go`
- Decision log:
  - Plan mode: no, this is a targeted backend bug fix with an isolated repository change and regression tests.
  - Model: Codex for implementation and tests.
  - Claude vs Codex: Codex for code changes; verification remains mandatory before completion.
  - Parallel dispatch: no, the fix is localized and the main dependency is the repository SQL shape.

## Completed Tasks

- [x] Create `docs/MASTER_ORCHESTRATION_PLAN.md` as the live orchestration index
- [x] Register the orchestration document in `docs/PROJECT_SOT.md`
- [x] Align `docs/IMPLEMENTATION_PLAN.md` and `docs/PROGRESS.md` to the new orchestration state
- [x] Verify orchestration consistency and identify the recommended next front
- [x] Remediation Task 5 - replace placeholder ERP promotion with real product promotion in `server_core`
- [x] Fix Task 6 ERP review conversion idempotency and zero-row handling

## Notes

- The orchestration layer now sits between repository governance and front-specific spec work.
- The next step should be one detailed front spec selected from `docs/MASTER_ORCHESTRATION_PLAN.md`.
