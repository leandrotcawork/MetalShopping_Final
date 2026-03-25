---
id: sinapse-tenant-isolation-flow
title: Tenant Isolation Flow
region: sinapses
tags: [tenant-isolation, security, auth, multi-tenancy, cross-cutting]
links:
  - cortex/backend/index
  - cortex/database/index
  - hippocampus/conventions
weight: 0.95
updated_at: 2026-03-24T10:00:00Z
---

# Tenant Isolation Flow

How authentication → tenancy → database isolation works end-to-end.

## Request Entry Point

```
1. HTTP request arrives at Go handler
   GET /api/products
   Authorization: Bearer <JWT>
   X-Tenant-ID: tenant-abc

2. Handler starts:
   principal := platformauth.PrincipalFromContext(r)
   if principal == nil {
     return 401 Unauthorized
   }

3. Extract tenant:
   tenant := tenancy_runtime.TenantFromContext(r)
   if tenant == nil {
     return 403 Forbidden
   }

4. Verify tenant membership:
   if principal.TenantID != tenant.ID {
     return 403 Forbidden
   }

5. Proceed to adapter with tenant context in request.Context
```

## Adapter → Database Layer

```
6. Adapter receives context with tenant:
   func (a *ProductAdapter) List(ctx context.Context) ([]Product, error) {
     tx := pgdb.BeginTenantTx(ctx)
     // ← Sets PostgreSQL session variable: SET app.current_tenant_id TO 'tenant-abc'
   }

7. Every query includes tenant filter:
   SELECT * FROM products
   WHERE tenant_id = current_tenant_id()
   // current_tenant_id() returns 'tenant-abc' from session variable

8. Database Row-Level Security (RLS) enforces:
   - Constraint: tenant_id column equals current_tenant_id()
   - Effect: Query returns 0 rows if tenant mismatch
   - Fail-safe: Even if WHERE clause is buggy, RLS prevents cross-tenant leaks
```

## Why This Matters

| Layer | What Happens | Consequence |
|-------|------|------|
| **HTTP Handler** | Principal validation fails early | 401 before any logic runs |
| **Tenancy middleware** | Wrong tenant context detected | 403 prevents wrong-tenant access |
| **Adapter** | `BeginTenantTx` sets session variable | Current tenant implicit in all queries |
| **SQL WHERE clause** | `current_tenant_id()` filter | Data filtered at query time |
| **Database RLS** | Row-level security policy enforces | Fallback if app filter is forgotten |

**Defense in depth:** Every layer assumes the previous layer might fail. Each layer validates.

## Failure Scenarios

### Scenario 1: Developer forgets `BeginTenantTx`

```go
// ❌ WRONG
func (a *ProductAdapter) List(ctx context.Context) ([]Product, error) {
  rows, err := a.db.Query("SELECT * FROM products WHERE tenant_id = current_tenant_id()")
  // current_tenant_id() returns NULL because session variable was never set
  // Query returns 0 rows (correct by accident)
}
```

**Lesson:** Always use `BeginTenantTx`. Test with `SET app.current_tenant_id = NULL` to catch this.

### Scenario 2: Developer forgets `current_tenant_id()` in WHERE

```go
// ❌ WRONG
func (a *ProductAdapter) List(ctx context.Context) ([]Product, error) {
  tx := pgdb.BeginTenantTx(ctx)
  rows, err := tx.QueryContext(ctx, "SELECT * FROM products")
  // Session variable is set, but WHERE clause doesn't use it
  // Query returns ALL products for ALL tenants
}
```

**Catch:** Row-Level Security (RLS) policy on `products` table rejects the query or returns 0 rows.

### Scenario 3: Wrong tenant in request header

```
GET /api/products
X-Tenant-ID: wrong-tenant

1. tenancy_runtime.TenantFromContext reads 'wrong-tenant'
2. principal.TenantID != tenant.ID check fails
3. Return 403 Forbidden
4. Request never reaches adapter
```

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Fix |
|-------------|------|---|
| Caching `current_tenant_id()` between requests | Session variable changes per request; cache is stale | Call `BeginTenantTx` for every request |
| Passing tenant as function parameter instead of context | Decouples function from request context; easy to pass wrong tenant | Always derive tenant from context in handler; pass context to adapters |
| Admin endpoint with no tenant check | Admin can access all tenants' data | Every handler validates tenant, even admin endpoints |
| Query builder that constructs WHERE clauses dynamically | Easy to forget tenant filter in one branch | Use repository pattern that always appends tenant filter |

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.95
