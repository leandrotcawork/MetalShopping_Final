# Worker Scaffold Checklist

## Before writing code
- [ ] confirmed the task cannot be done in a Go handler
- [ ] read `apps/integration_worker/shopping_price_worker.py` as reference
- [ ] output table DDL defined with tenant_id column on every table
- [ ] existing worker checked before creating a new file

## While implementing
- [ ] every write transaction sets `app.current_tenant_id` via `set_config`
- [ ] all inserts use `ON CONFLICT ... DO UPDATE` (idempotent)
- [ ] `log()` called at start, end, and on error with tenant_id and run_id
- [ ] `MS_DATABASE_URL` and other env vars fail-fast if missing
- [ ] no HTTP calls from worker to server_core
- [ ] no hardcoded DSNs, tenant IDs, or credentials

## Final review
- [ ] worker runs successfully and writes rows to Postgres
- [ ] re-running the worker does not corrupt or duplicate data
- [ ] Go reader in server_core reads from the same tables
- [ ] no business logic that belongs in Go domain layer
- [ ] table names follow `<domain>_<entity>_<noun>` convention
