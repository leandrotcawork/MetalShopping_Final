# ADR-0017: Shopping Price Operational Workflow Write Surface

- Status: accepted
- Date: 2026-03-19

## Context

Shopping Price Level 1 is already implemented as a read surface:

- Go `server_core` reads from Postgres tables
- a Python worker is expected to write snapshots to those tables
- the web UI consumes the surface via generated SDK + platform runtime

However, the legacy Shopping flow is an operational workflow (not just a page):

- bootstrap
- input preparation (XLSX or catalog selection)
- run submission
- progress
- result
- history

Level 1 closed only the read surface. To continue the frontend migration and implement the real shopping core, we must freeze a backend-owned write surface for run submission and workflow bootstrap.

## Decision

We will extend Shopping v1 to include a backend-owned workflow write surface.

The Shopping bounded-context owns:

- run submission (create/queue)
- run lifecycle visibility (status/progress/history)
- result serving through read models written by the worker

The workflow contract must include, at minimum:

- `GET /api/v1/shopping/bootstrap`
  - returns workflow bootstrap data used by the UI (defaults, allowed statuses)
  - returns supplier selection options by consuming the Suppliers directory (see ADR-0019)
- `POST /api/v1/shopping/runs`
  - creates a run request with explicit input scope (catalog selection and/or XLSX upload reference)
  - never triggers the worker synchronously

Worker execution remains strictly:

`worker writes -> Postgres -> server_core reads -> API -> thin clients`

## Contracts (touchpoints)

- OpenAPI: `contracts/api/openapi/shopping_v1.openapi.yaml`
  - `GET /api/v1/shopping/bootstrap`
  - `GET /api/v1/shopping/summary`
  - `POST /api/v1/shopping/runs` (run request creation, queued)
  - `GET /api/v1/shopping/run-requests/{run_request_id}` (status)
- JSON Schemas (v1): `contracts/api/jsonschema/shopping_bootstrap_v1.schema.json`, `shopping_summary_v1.schema.json`, `shopping_create_run_request_v1.schema.json`, `shopping_create_run_response_v1.schema.json`, `shopping_run_request_v1.schema.json`
- Events: none in Phase 1; Phase 2 may add `shopping.run_requested.v1` per ADR-0018
- Governance: none required unless a feature-flag gate is introduced later

## Implementation checklist (Level 2)

- Contract-first changes: `metalshopping-openapi-contracts`
- Go write/read surfaces: `metalshopping-module-scaffold` (principal + tenant checks, tenant-scoped DB access)
- Worker execution: `metalshopping-worker-patterns` + `metalshopping-worker-scaffold`
- SDK refresh after contract change: `metalshopping-sdk-generation`
- UI binding with legacy workflow preserved: `metalshopping-frontend-migration-guardrails` + `metalshopping-page-delivery`

## Consequences

- The UI can preserve the legacy workflow patterns without reintroducing legacy coupling.
- Run submission becomes explicit, contract-driven, and tenant-safe.
- The core does not depend on a synchronous worker round-trip.
- Subsequent improvements (cancel run, manual URLs, per-supplier configuration) become incremental contract expansions, not ad hoc UI logic.

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Contract authoring: `metalshopping-openapi-contracts`
- Module implementation (Go): `metalshopping-module-scaffold`
- Worker execution model: `metalshopping-worker-patterns` and `metalshopping-worker-scaffold`
- SDK refresh: `metalshopping-sdk-generation`
- Page binding (legacy visual, thin-client): `metalshopping-frontend-migration-guardrails` and `metalshopping-page-delivery`
