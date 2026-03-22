---
name: metalshopping-implement
description: Implement Go modules and Python workers for MetalShopping using the repo invariants and local reference patterns. Called by $ms for T2 and T3.
---

# MetalShopping Implement

Read `tasks/lessons.md` before coding. Use the existing module patterns in the repo; load reference files only when needed.

## 1) Pick the module shape

- `read-only` → reader + postgres reader + handler
- `write+events` → writer + outbox event + handler/service
- `CRUD+governance` → repository/service + governance adapter when required
- `scraping` → Python worker + Go read surface if the UI/server needs the result

## 2) Go invariants

Always enforce:
- `pgdb.BeginTenantTx(...)` on every Postgres adapter query
- `WHERE tenant_id = current_tenant_id()` on every tenant-scoped query
- handler auth order: `PrincipalFromContext` → `401`, then `TenantFromContext` → `403`
- structured error codes and request logging
- `outbox.AppendInTx(...)` before `Commit()` on write flows
- registration in `cmd/metalshopping-server/composition_modules.go`

## 3) Python worker invariants

Always enforce:
- `set_config('app.current_tenant_id', %s, true)` before every write transaction
- `ON CONFLICT ... DO UPDATE` for idempotent writes
- no calls to server_core HTTP endpoints
- fail fast on missing runtime config

## 4) Use the references, not duplicated examples

- Go patterns: `references/go-patterns.md`
- Worker patterns: `references/worker-patterns.md`

## 5) Closeout

- run relevant validation before marking the task done
- update `tasks/todo.md`
- write a lesson after any correction
- commit after the completed task
