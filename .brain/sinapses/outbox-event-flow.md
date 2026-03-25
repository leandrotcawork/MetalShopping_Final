---
id: sinapse-outbox-event-flow
title: Outbox Event Publishing Flow
region: sinapses
tags: [outbox, events, atomicity, transactional-consistency, cross-cutting]
links:
  - cortex/backend/index
  - cortex/database/index
  - hippocampus/conventions
weight: 0.94
updated_at: 2026-03-24T10:00:00Z
---

# Outbox Event Publishing Flow

How domain writes and event publishing are kept atomic using the transactional outbox pattern.

## The Problem

Without the outbox pattern:

```go
// ❌ WRONG: Data written, but event is lost if process crashes
err := tx.Commit()  // Domain write is persisted
if err != nil { return err }

// What if the process crashes here?
err = eventPublisher.Publish(event)  // Event might never be published
if err != nil { return err }
```

**Consequence:** Event-driven systems downstream miss updates. State becomes inconsistent.

## Outbox Pattern

All events are written to the database in the same transaction as the domain write:

```go
// ✅ CORRECT: Atomic write + event append
func (a *ProductAdapter) CreateProduct(ctx context.Context, p Product) error {
  tx := pgdb.BeginTenantTx(ctx)
  defer tx.Rollback()

  // 1. Insert the domain entity
  err := tx.ExecContext(ctx, `
    INSERT INTO products (id, tenant_id, name, price, created_at)
    VALUES ($1, $2, $3, $4, NOW())
  `, p.ID, p.TenantID, p.Name, p.Price)
  if err != nil { return err }

  // 2. Append the event to the outbox in the SAME transaction
  event := ProductCreatedEvent{
    ID: p.ID,
    Name: p.Name,
    Price: p.Price,
    Timestamp: time.Now(),
  }
  err = outbox.AppendInTx(tx, event)
  if err != nil { return err }

  // 3. Commit both together
  return tx.Commit()
}
```

## Database Outbox Table

```sql
CREATE TABLE outbox (
  id BIGSERIAL PRIMARY KEY,
  aggregate_id UUID NOT NULL,
  aggregate_type TEXT NOT NULL,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  tenant_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  published_at TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_outbox_unpublished ON outbox(published_at)
WHERE published_at IS NULL;
```

## Event Processing Flow

```
1. Background worker polls:
   SELECT * FROM outbox WHERE published_at IS NULL
   ORDER BY created_at ASC
   LIMIT 100

2. For each unpublished event:
   - Deserialize payload
   - Call appropriate event handler
   - Mark as published: UPDATE outbox SET published_at = NOW() WHERE id = ...

3. Event handlers are idempotent:
   - Safe to retry on failure
   - No side effects on duplicate processing
```

## Why Idempotency Matters

Without idempotency:

```go
// ❌ WRONG: Processing event twice causes double-work
func OnProductCreated(event ProductCreatedEvent) {
  // Publishes to analytics queue
  // Notifies customers
  // If this handler runs twice, customers get 2 notifications
}
```

With idempotency:

```go
// ✅ CORRECT: Safe to run multiple times with same result
func OnProductCreated(event ProductCreatedEvent) {
  // Use event.ID as deduplication key
  exists, err := analyticsDB.EventExists(event.ID)
  if exists { return nil } // Already processed

  // Insert only once
  return analyticsDB.RecordEvent(event)
}
```

## Failure Scenarios

### Scenario 1: Process crashes during event processing

```
1. Worker picks event from outbox
2. Calls OnProductCreated(event)
3. Process crashes before updating published_at
4. Next worker picks same event again
5. Event handler must be idempotent (Scenario above)
```

### Scenario 2: Database connection lost during append

```go
err := outbox.AppendInTx(tx, event)
if err != nil {
  tx.Rollback()  // Both product and event are rolled back
  return err     // Client retries → product and event both succeed
}
```

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Fix |
|-------------|------|---|
| Append event after `Commit()` | Transaction is no longer atomic; event might be lost | Always use `AppendInTx()` before `Commit()` |
| Publish event directly without outbox | Synchronous processing couples services; slow, unreliable | All events go to outbox; worker processes asynchronously |
| Event handlers without idempotency checks | Duplicate processing causes side effects | Use event ID as deduplication key in handlers |
| No index on `published_at` | Worker does full table scan; slow with large outbox | Index on `WHERE published_at IS NULL` |

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.94
