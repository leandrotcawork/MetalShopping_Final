# ADR-0025: Shopping Run Requested Event And Broker Consumption (Phase 2)

- Status: draft
- Date: 2026-03-19

## Context

ADR-0018 defines Phase 1 run orchestration via a Postgres DB queue:

- `server_core` writes a run request row (tenant-scoped)
- worker claims via `SELECT ... FOR UPDATE SKIP LOCKED`

This is sufficient for make-it-work-first. For a production-grade platform, we still need broker-driven delivery so workers do not depend on polling a DB table and so we can scale execution and routing. The broker upgrade must preserve the same semantics and keep the DB queue ledger as the source of truth for idempotency and audit.

## Decision

Phase 2 introduces an explicit domain event for run delivery:

- `server_core` appends `shopping_run_requested.v1` to the outbox when a run request is created.
- workers consume the event through the broker and start execution by transitioning the same DB ledger row from `queued` to `claimed`/`running`.

Rules:

- the run request table remains the authoritative ledger (idempotency + audit)
- the event is delivery, not the ledger
- workers must still be safe under duplicates and re-delivery
- no synchronous worker dependency is introduced in HTTP request paths

## Contracts (touchpoints)

- Event contract (new):
  - `contracts/events/v1/shopping_run_requested.v1.json`
  - payload contains: `run_request_id`, `tenant_id`, and minimal execution hints required for routing (avoid embedding large scopes)
- Platform outbox (existing): `apps/server_core/internal/platform/outbox/*`
- Shopping OpenAPI: unchanged for Phase 2 (still `POST /api/v1/shopping/runs`)
- Governance: optional later (feature flag for broker delivery), but not required for the initial Phase 2

## Implementation checklist

- Add the event contract: `metalshopping-event-contracts`
- Publish to outbox from Shopping run request creation path: `metalshopping-module-scaffold` + `metalshopping-platform-packages`
- Worker consume path:
  - implement broker consumer + idempotent processing
  - claim/update the DB ledger row as the first durable action
  Skills: `metalshopping-worker-patterns`, `metalshopping-worker-scaffold`
- SDK generation is unaffected unless OpenAPI changes: `metalshopping-sdk-generation`
- Observability/security review (correlation ids, PII, retries): `metalshopping-observability-security`

## Consequences

- Execution can scale without DB polling as the primary trigger.
- Auditing and idempotency remain anchored in the run request ledger.
- Phase 1 and Phase 2 can coexist behind a runtime switch if needed.

