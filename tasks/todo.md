# Task Tracker

## Current State

- State: completed
- Active tranche: master orchestration planning
- Source spec: `docs/superpowers/specs/2026-04-01-master-orchestration-plan-design.md`

## Current Task

- Task: Remediation Task 5 - replace placeholder ERP promotion with real product promotion in `server_core`
- State: in-progress
- Scope: `apps/server_core/internal/modules/erp_integrations/{application,adapters,ports}`, `apps/server_core/cmd/metalshopping-server/composition_modules.go`
- Decision log:
  - Plan mode: yes, because this touches promotion/outbox semantics and multiple files.
  - Model: Codex for implementation and tests.
  - Claude vs Codex: Codex for code changes; review and verification remain mandatory before completion.
  - Parallel dispatch: yes, after discovery, to keep code wiring and tests moving together.

## Completed Tasks

- [x] Create `docs/MASTER_ORCHESTRATION_PLAN.md` as the live orchestration index
- [x] Register the orchestration document in `docs/PROJECT_SOT.md`
- [x] Align `docs/IMPLEMENTATION_PLAN.md` and `docs/PROGRESS.md` to the new orchestration state
- [x] Verify orchestration consistency and identify the recommended next front

## Notes

- The orchestration layer now sits between repository governance and front-specific spec work.
- The next step should be one detailed front spec selected from `docs/MASTER_ORCHESTRATION_PLAN.md`.
