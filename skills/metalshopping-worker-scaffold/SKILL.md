---
name: metalshopping-worker-scaffold
description: Scaffold or review a MetalShopping Python worker inside `apps/integration_worker/` that ingests data into Postgres tables consumed by server_core. Use when implementing the Shopping Price scraper, analytics ingestion, or any async compute task that requires Playwright or Python-only libraries. Workers write to Postgres using tenant-aware upserts. Go server_core reads from those tables. Workers never call server_core HTTP endpoints.
---

# MetalShopping Worker Scaffold

## Overview

Workers in this repo are async data ingestion processes. They write
to Postgres tables that server_core reads via its normal query path.
The pattern is strictly: worker writes → Postgres → Go reads → API.

The canonical reference implementation is:
`apps/integration_worker/shopping_price_worker.py`

Read it before implementing any new worker.

## Decision: Go handler or Python worker?

| The task requires | Use |
|---|---|
| HTTP CRUD, business rules, Postgres query, auth | Go handler in server_core |
| Playwright scraping, Selenium, pandas, numpy, ML | Python worker in integration_worker |
| External API with Python-only SDK | Python worker |
| Scheduling, polling, batch ingestion | Python worker |

If Go can do it → Go. Workers exist only for what Go cannot.

## Workflow

1. Read the minimum repo context:
   `apps/integration_worker/README.md`
   `apps/integration_worker/shopping_price_worker.py` (canonical example)
   `docs/WORKER_OPERATING_MODEL.md`

2. Confirm the task cannot be done in a Go handler.
   If it can → stop, implement Go handler instead.

3. Define the Postgres output tables:
   - one table per entity type (runs, run_items, latest_snapshot, etc.)
   - include `tenant_id` column on every table
   - primary key strategy: uuid or composite (tenant_id, product_id)
   - write the CREATE TABLE DDL before writing the worker

4. Implement the worker following the canonical pattern:

   ### Tenancy — mandatory
   Every write transaction must set the tenant session context:
   ```python
   cur.execute("BEGIN")
  # RLS uses current_tenant_id() -> current_setting('app.tenant_id')
  cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
   # ... inserts use current_tenant_id() function
   cur.execute("COMMIT")
   ```
   Never write rows without setting tenant context first.

   ### Upsert pattern — mandatory
   All writes must be idempotent:
   ```python
   INSERT INTO table (...) VALUES (...)
   ON CONFLICT (pk_column) DO UPDATE SET
     field = EXCLUDED.field,
     updated_at = NOW()
   ```
   Re-running the worker must never corrupt or duplicate data.

   ### Logging — mandatory
   ```python
   log("worker_start", tenant_id=tenant_id, run_id=run_id)
   # ... work ...
   log("worker_end", tenant_id=tenant_id, rows_written=n)
   # on error:
   log("worker_error", tenant_id=tenant_id, error=str(exc))
   ```

   ### Environment config — mandatory
   ```python
   database_url = os.getenv("MS_DATABASE_URL", "").strip()
   # fail fast if missing
   if not database_url:
       sys.stderr.write("MS_DATABASE_URL required\n")
       sys.exit(2)
   ```
   No hardcoded DSNs, credentials, or tenant IDs.

5. Do NOT rewrite `shopping_price_worker.py` if it already handles
   the required tables. Extend it or add a new worker alongside it.

6. After the worker writes data, implement the Go reader in server_core:
   use `metalshopping-module-scaffold` for the handler.

7. Finish with the checklist in `references/worker-scaffold-checklist.md`.

## What workers must never do

- call server_core HTTP endpoints directly
- own canonical write semantics (no authoritative business state)
- expose HTTP endpoints of their own
- write without setting tenant session context
- hardcode tenant_id, DSN, or credentials in code
- contain business logic that belongs in Go domain/application layers

## Output table naming convention

Follow the pattern already established:
- `shopping_price_runs` → `<domain>_<entity>_<plural>`
- `shopping_price_run_items` → `<domain>_<entity>_<sub_entity>_<plural>`
- `shopping_price_latest_snapshot` → `<domain>_<entity>_latest_<noun>`

## References

- Canonical worker: `apps/integration_worker/shopping_price_worker.py`
- For the final review pass, read `references/worker-scaffold-checklist.md`.
