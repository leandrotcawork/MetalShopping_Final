---
id: cortex-database-index
title: Database Domain
region: cortex/database
tags: [database, postgres, multi-tenant, schema]
links:
  - hippocampus/architecture
  - hippocampus/conventions
  - hippocampus/decisions_log
weight: 0.88
updated_at: 2026-03-24T10:00:00Z
---

# Database Domain

MetalShopping uses **PostgreSQL with shared-database multi-tenancy**. Every table includes tenant isolation at the database and application level.

## Multi-Tenant Architecture

### Shared Database, Multiple Tenants

All tenants' data lives in the same PostgreSQL instance. Tenant isolation is enforced through:

1. **Row-Level Security (RLS)** — `current_tenant_id()` PostgreSQL function
2. **Application-level filters** — Every query includes `WHERE tenant_id = current_tenant_id()`
3. **Transaction context** — Each request sets the tenant context before querying

### Tenant Context Setup

Every Go adapter query starts like this:

```go
// Postgres adapter
func (p *ProductAdapter) ListByTenant(ctx context.Context) ([]Product, error) {
  tx := pgdb.BeginTenantTx(ctx)  // ← Sets up context with current_tenant_id()
  defer tx.Rollback()

  var products []Product
  err := tx.QueryContext(ctx, `
    SELECT id, name, price FROM products
    WHERE tenant_id = current_tenant_id()  // ← RLS + app-level filter
  `).Scan(&products)

  return products, err
}
```

**Critical rule:** Never query without filtering by `current_tenant_id()` on tenant-scoped tables.

## Schema Patterns

### Every Tenant-Scoped Table

```sql
CREATE TABLE products (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,                    -- ← Tenancy column
  name TEXT NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Tenant isolation RLS policy
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
CREATE POLICY products_tenant_isolation ON products
  USING (tenant_id = current_tenant_id());
```

### Platform Tables (No Tenant Scope)

Some tables are global (not tenant-scoped):

```sql
CREATE TABLE tenants (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  email TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Migrations

All schema changes go through SQL migrations in `migrations/`:

```
migrations/
  001_initial_schema.sql
  002_add_products_table.sql
  003_add_products_tenant_isolation.sql
  ...
```

**Rule:** Migrations must be:
- Idempotent (safe to run multiple times with `BEGIN; ... ON CONFLICT; COMMIT;`)
- Reversible (every UP has a corresponding DOWN)
- Tested (migration passes on test database before prod)

## Queries & Read Models

### Complex Queries → Read Models

For analytics or complex joins, create denormalized **read models**:

```sql
-- products_analytics read model (materialized view or scheduled refresh)
CREATE MATERIALIZED VIEW products_analytics AS
SELECT
  p.id, p.name, p.tenant_id,
  COUNT(o.id) as order_count,
  SUM(o.total) as total_revenue,
  AVG(o.total) as avg_order_value
FROM products p
LEFT JOIN orders o ON o.product_id = p.id
WHERE p.tenant_id = current_tenant_id()
GROUP BY p.id, p.name, p.tenant_id;
```

Read models are always queried with tenant context, same as live tables.

## Indexes & Performance

Key indexing patterns:

```sql
-- Tenant + ID (most common filter pair)
CREATE INDEX idx_products_tenant_id ON products(tenant_id, id);

-- Tenant + timestamp (for time-range queries)
CREATE INDEX idx_orders_tenant_created ON orders(tenant_id, created_at);

-- Foreign key lookups
CREATE INDEX idx_orders_product_id ON orders(product_id);
```

**Rule:** Never create an index without a performance baseline (EXPLAIN ANALYZE).

## Outbox Table (Event Publishing)

Events are stored durably before being published:

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

CREATE INDEX idx_outbox_unpublished ON outbox(published_at) WHERE published_at IS NULL;
```

All domain writes append to outbox in the same transaction. A background worker processes unpublished events.

## Known Pitfalls

These lessons capture repeated mistakes in database design and usage:

- [[../backend/lessons/lesson-0001]] — Tenant-safe DB access is mandatory
- [[../backend/lessons/lesson-0003]] — Outbox must be atomic with writes

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.88
