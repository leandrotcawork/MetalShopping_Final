# Review Lenses

## L1 — Contract
- Handler response matches OpenAPI contract exactly?
- Endpoints declared in contracts/api/openapi/?
- SDK regenerated after last contract change?
- Breaking change without version bump?

## L2 — Tenancy
- pgdb.BeginTenantTx on every adapter query?
- current_tenant_id() in every WHERE on tenant tables?
- Any query can return cross-tenant rows?
- Worker sets app.current_tenant_id before every write tx?
- Hardcoded tenant_id anywhere?

## L3 — Auth
- PrincipalFromContext checked → 401 before operation?
- TenantFromContext checked → 403 before operation?
- Any endpoint callable without authentication?
- Authorization before data access (not after)?

## L4 — Boundaries
- Business logic in domain (not transport, not adapter)?
- Handler: parse → auth → service → writeJSON only?
- Service: no HTTP, no DB, no framework deps?
- Adapter: no business logic?
- Worker: no canonical business logic?

## L5 — Outbox / Events
- AppendInTx called before tx.Commit (never after)?
- Idempotency key: "event_name:aggregate_id"?
- Event payload has all fields the worker needs?
- Event version explicit?

## L6 — Frontend
- No fetch() in pages or components?
- All data via platform-sdk hooks?
- No hardcoded colors/sizes/spacing?
- Loading + error + empty states present?
- No widget duplicating one in packages/ui?

## L7 — Observability
- Handler logs: trace_id, action, result, duration_ms?
- Worker logs: start, end, rows_written, error with tenant_id?
- Silent failure paths? (errors caught without log)
- Can you debug in prod without local reproduction?

## L8 — Idempotency
- Writes safe to retry?
- Worker inserts use ON CONFLICT DO UPDATE?
- Re-deploy won't cause duplicate processing?

## L9 — Scalability
- N+1 queries in list endpoints?
- Missing index for known query pattern?
- Query breaks at 100k rows what works at 100?
- analytics/CRM modules can read this data in future?

## L10 — SDK / Governance
- packages/generated/ manually edited?
- Manual DTO type duplicating generated type?
- Business rule hardcoded that should be governed?
- Governance key registered in bootstrap?

## Severity mapping
Critical: L2, L3, L5 (outbox), L4 (business logic in transport)
High: L6 fetch, L1 contract drift, L7 silent failures
Medium: L9 N+1, L10 missing key, L8 non-idempotent
Low: naming, log format, test coverage
