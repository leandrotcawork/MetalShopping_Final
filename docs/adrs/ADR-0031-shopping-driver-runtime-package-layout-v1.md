# ADR-0031: Shopping Driver Runtime Package Layout v1 (integration_worker)

- Status: accepted
- Date: 2026-03-20

## Context

The Shopping Price worker is now functionally correct and governed (enabled/active/valid manifests, strategy validation, and a non-mock pilot).

However, the driver runtime is still concentrated in a single entrypoint script (`apps/integration_worker/shopping_price_worker.py`). This is acceptable for bootstrap, but it blocks the "big tech" goal:

- strategy implementations become hard to test in isolation
- adding new suppliers risks copy/paste growth
- shared building blocks (HTTP/VTEX, HTML search, Playwright PDP-first) are not packaged as reusable frameworks
- parallelization and rate limits cannot be reasoned about per family/strategy without refactoring the runtime boundary

We need to freeze the internal package layout so all future suppliers follow one predictable path.

## Decision

Extract the driver runtime into a dedicated Python package under `apps/integration_worker/src/`, while keeping `shopping_price_worker.py` as the orchestration entrypoint (claim -> load config -> execute -> persist).

Rules:

- `apps/integration_worker/shopping_price_worker.py` remains the executable entrypoint and owns:
  - queue/event claim loop and lifecycle updates
  - loading runtime eligibility (enabled/active/valid)
  - tenancy/session key setup for Postgres writes
  - persistence to Shopping read-surfaces
- A new package owns runtime execution and shared frameworks:
  - `apps/integration_worker/src/shopping_price_runtime/`
- Strategy executors must not query Postgres directly.
  - They receive inputs (product identifiers, persisted signal URL, manifest config) and return a normalized observation result.
- Strategy executors must not call `server_core` HTTP endpoints.
- Strategy executors must return a normalized `item_status` in:
  - `OK | NOT_FOUND | AMBIGUOUS | ERROR`
- Driver behavior is selected by `(family, strategy)` as frozen in ADR-0030.

### Proposed package layout (v1)

- `apps/integration_worker/src/shopping_price_runtime/models.py`
  - typed structs for `RuntimeConfig`, `LookupInputs`, `RuntimeObservation`
- `apps/integration_worker/src/shopping_price_runtime/dispatcher.py`
  - `execute(family, strategy, inputs, config) -> RuntimeObservation`
- `apps/integration_worker/src/shopping_price_runtime/http/`
  - `vtex_persisted_query.py`
  - `html_search.py`
  - `mock.py`
- `apps/integration_worker/src/shopping_price_runtime/playwright/`
  - `pdp_first.py`
  - `mock.py`
- `apps/integration_worker/src/shopping_price_runtime/shared/`
  - retry/backoff helpers, URL builders, parsing helpers, block detection helpers

## Contracts (touchpoints)

- No new external contracts are required.
- Existing driver manifest JSON Schemas remain authoritative:
  - `contracts/api/jsonschema/supplier_driver_manifest_http_v1.schema.json`
  - `contracts/api/jsonschema/supplier_driver_manifest_playwright_v1.schema.json`

## Implementation Checklist

1. Freeze the package layout and boundaries (this ADR)
   Skill: `metalshopping-adr-updates`
2. Extract strategy executors into `apps/integration_worker/src/shopping_price_runtime/*`
   Skill: `metalshopping-worker-scaffold` (review: `metalshopping-worker-patterns`)
3. Keep entrypoint as orchestrator and route via dispatcher
   Skill: `metalshopping-worker-scaffold`
4. Add unit tests per strategy module
   Skill: `metalshopping-worker-patterns`
5. Observability/security review
   Skill: `metalshopping-observability-security`

## Implementation Snapshot (2026-03-20)

- Extracted runtime package:
  - `apps/integration_worker/src/shopping_price_runtime/models.py`
  - `apps/integration_worker/src/shopping_price_runtime/lookup.py`
  - `apps/integration_worker/src/shopping_price_runtime/dispatcher.py`
  - `apps/integration_worker/src/shopping_price_runtime/http/strategies.py`
  - `apps/integration_worker/src/shopping_price_runtime/playwright/strategies.py`
  - `apps/integration_worker/src/shopping_price_runtime/shared/parsing.py`
  - `apps/integration_worker/src/shopping_price_runtime/headers.py`
- Entry point refactor:
  - `apps/integration_worker/shopping_price_worker.py` now delegates strategy execution to runtime dispatcher and keeps orchestration/persistence ownership.
- Build evidence (local):
  - `.venv\Scripts\python.exe -m py_compile apps/integration_worker/shopping_price_worker.py` -> pass
  - `.venv\Scripts\python.exe -m py_compile apps/integration_worker/src/shopping_price_runtime/dispatcher.py` -> pass
  - `.venv\Scripts\python.exe -m py_compile` on all package `.py` files -> pass

- Smoke evidence (outside sandbox):
  - `.env` loaded + `scripts/smoke_shopping_event_local.ps1` -> pass
  - worker executed in `event` mode and completed one claimed event without runtime/parser regressions.

## Acceptance Evidence (for Status: accepted)

- Build/test:
  - `python -m py_compile apps/integration_worker/shopping_price_worker.py` -> pass
  - `python -m py_compile apps/integration_worker/src/shopping_price_runtime/dispatcher.py` -> pass
- Smoke:
  - `scripts/smoke_shopping_event_local.ps1` (DEXCO VTEX non-mock) -> pass
- Code review:
  - entrypoint contains no strategy-specific parsing logic beyond dispatcher call

## Consequences

- Strategy work becomes modular and testable.
- Adding a supplier becomes "manifest config + (rare) bounded strategy extension" instead of editing a large monolith.
- Parallelization and rate limits can be implemented per family/strategy without tangling with DB orchestration.
