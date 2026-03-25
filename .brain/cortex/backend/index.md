---
id: cortex-backend-index
title: Backend Domain
region: cortex/backend
tags: [backend, go, modular-monolith, architecture]
links:
  - hippocampus/architecture
  - hippocampus/conventions
  - hippocampus/decisions_log
weight: 0.9
updated_at: 2026-03-24T10:00:00Z
---

# Backend Domain

MetalShopping's Go backend implements a **modular monolith** with strict layer separation and tenant-aware data access.

## Architecture

### Module Structure (Strict Layers)

Every module under `internal/modules/<name>/` follows this hierarchy:

```
domain/           Business entities, invariants, value objects
application/      Use-case handlers, command/query orchestration
ports/            Input/output interfaces (repository, event, external service)
adapters/         Postgres persistence, external API clients
transport/        HTTP handlers, request/response serialization
events/           Domain events, publisher
readmodel/        Denormalized query views, projections
```

**Rule:** No cross-layer shortcuts. Adapters never import transport. Transport never imports domain directly.

### Module Registration

Every new module must be registered in `composition_modules.go`. This is the dependency injection entry point — without registration, the module won't initialize.

## Platform Infrastructure

All modules depend on shared infrastructure in `internal/platform/`:

| Component | Purpose |
|-----------|---------|
| `db/postgres/` | Tenant-aware transaction and query helpers (`BeginTenantTx`, `current_tenant_id()`) |
| `auth/` | JWT validation, principal extraction, middleware |
| `tenancy_runtime/` | Tenant context extraction from request headers |
| `governance/` | Feature flags, policies, thresholds runtime resolution |
| `outbox/` | Transactional event publishing (append events within transactions) |
| `logger/` | Structured logging with trace ID, action, duration, result |

## Tenant Safety (Non-Negotiable)

Every database query must:
1. **Start with tenant tx**: `pgdb.BeginTenantTx(ctx)` at adapter entry
2. **Filter by tenant**: `current_tenant_id()` in every WHERE clause on tenant-scoped tables
3. **Validate principal**: `platformauth.PrincipalFromContext(r)` → 401 if missing
4. **Validate tenant**: `tenancy_runtime.TenantFromContext(r)` → 403 if wrong

**Anti-pattern:** Running a SELECT without `current_tenant_id()` in WHERE. This is a data leak.

## Event Publishing (Outbox Pattern)

All state changes that trigger events must use the transactional outbox:

```
1. Start transaction: tx := pgdb.BeginTenantTx(ctx)
2. Perform write (INSERT/UPDATE)
3. Append event: outbox.AppendInTx(tx, event)
4. Commit: tx.Commit()
```

Events are processed asynchronously. All event handlers must be idempotent (safe to retry).

## API Contracts

All HTTP endpoints are defined in `contracts/api/openapi/`. These are the source of truth for:
- Request/response shapes
- Authentication requirements
- Status codes
- Error formats

**Rule:** Contracts are hand-authored. Generate SDK and types from contracts, never the reverse.

## Logging & Observability

Every handler must log:

```go
logger.Info("action_result",
  zap.String("trace_id", traceID),
  zap.String("action", "create_product"),
  zap.String("result", "success"),
  zap.Int64("duration_ms", elapsed),
)
```

All errors carry structured error codes: `MODULE_ENTITY_REASON` (e.g., `PRODUCTS_SKU_DUPLICATE`).

## Testing

- Unit tests: Pure domain logic, no database
- Integration tests: Real Postgres with tenant isolation (use test fixtures)
- Contract tests: Verify HTTP handlers match OpenAPI spec

## Known Pitfalls

These lessons capture repeated mistakes in backend development:

- [[lessons/lesson-0001]] — Tenant-safe DB access is mandatory
- [[lessons/lesson-0002]] — Handlers must fail fast on auth and tenancy
- [[lessons/lesson-0003]] — Outbox must be atomic with writes
- [[lessons/lesson-0004]] — Worker writes require tenant context and idempotency
- [[lessons/lesson-0013]] — Observability is part of the contract

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.9
