# ADR-0018: Shopping Run Orchestration via DB Queue (Phase 1)

- Status: accepted
- Date: 2026-03-19

## Context

The platform architecture freezes:

- `server_core` must not depend on workers synchronously in the normal request path
- async work must be decoupled (queue semantics, outbox/inbox)
- contracts are explicit and versioned

Today, the repo has outbox foundations, but broker delivery and worker consumption are not yet in place. Shopping Price needs a run submission mechanism that is production-grade in semantics (idempotent, tenant-safe, observable) without requiring a full broker integration on day one.

## Decision

Phase 1 (make-it-work-first) will use Postgres as the orchestration queue for Shopping runs:

- `server_core` writes a run request row to a tenant-scoped table
- the integration worker claims queued work using safe DB-queue semantics
- the worker updates run lifecycle fields and writes read-model outputs

Queue semantics must be:

- claim via `SELECT ... FOR UPDATE SKIP LOCKED`
- idempotent (a re-run must not corrupt data)
- observable (structured logs + status fields)
- tenant-safe (RLS enforced through `current_tenant_id()`)

Phase 2 (future) upgrades the same semantics to outbox/broker delivery:

- `server_core` publishes `shopping.run_requested.v1` (outbox)
- workers consume and process

The Phase 1 data model must not block Phase 2. The DB queue table is treated as the authoritative request ledger (and can later be fed by outbox delivery).

## Contracts (touchpoints)

- OpenAPI: `contracts/api/openapi/shopping_v1.openapi.yaml`
  - `POST /api/v1/shopping/runs` writes a run request (queued)
  - `GET /api/v1/shopping/run-requests/{run_request_id}` exposes lifecycle fields
  - `GET /api/v1/shopping/runs` / `GET /api/v1/shopping/runs/{run_id}` expose run history and progress
- JSON Schemas: `contracts/api/jsonschema/shopping_create_run_request_v1.schema.json`, `shopping_run_request_v1.schema.json`, `shopping_run_list_v1.schema.json`, `shopping_run_v1.schema.json`
- Events (Phase 2 only): add `contracts/events/v1/shopping_run_requested.v1.json` (name frozen when Phase 2 starts)
- Governance: none required in Phase 1

## Implementation checklist

- DB queue table + RLS: `apps/server_core/migrations/0021_shopping_price_run_requests.sql`
- Worker claim loop (Phase 1):
  - `SELECT ... FOR UPDATE SKIP LOCKED` to claim queued requests
  - update lifecycle fields (`queued` -> `claimed` -> `running` -> `completed`/`failed`)
  - idempotent writes for outputs (see ADR-0020)
  Skills: `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`
- Observability baseline: structured logs with `trace_id`, action, result, and durations
  Skill: `metalshopping-observability-security`

## Consequences

- We can ship end-to-end workflow execution before broker delivery is implemented.
- Semantics remain aligned with the frozen async integration direction.
- Migration to broker is an incremental integration change, not a rewrite of Shopping.

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Contract authoring (run submission): `metalshopping-openapi-contracts`
- Go module implementation (write surface): `metalshopping-module-scaffold`
- Worker polling/claim pattern: `metalshopping-worker-patterns` and `metalshopping-worker-scaffold`
- Observability baseline review: `metalshopping-observability-security`
- Event contract (Phase 2): `metalshopping-event-contracts`
