# AGENTS

## Scope

This file applies to everything under `apps/`.

## App rules

- `server_core` is the canonical sync core
- workers do not own canonical state
- clients remain thin
- planning comes before implementation in the current phase

## Boundaries

- `server_core`: auth, tenant, authz, governance, canonical writes, sync serving
- `analytics_worker`: compute and publication of analytical outputs
- `integration_worker`: external connectors, crawlers, imports, exports, normalization
- `automation_worker`: triggers, orchestration, async actions
- `notifications_worker`: delivery channels only
- `web`, `desktop`, `admin_console`: client surfaces only

## Token discipline

- Read `apps/server_core/AGENTS.md` before changing core planning artifacts
- Do not inspect every app if the task is scoped to one surface

