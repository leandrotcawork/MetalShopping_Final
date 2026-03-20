# Worker Patterns — MetalShopping

## When to use Python worker vs Go
Python worker: Playwright, numpy/pandas, Python-only lib, XLSX batch ingestion
Go handler: everything else — HTTP CRUD, Postgres query, business logic

## Worker operating model
Worker writes to Postgres → server_core reads from same tables → API exposes data
Worker NEVER calls server_core HTTP

## Event-driven mode
```bash
MS_SHOPPING_WORKER_MODE=event
MS_SHOPPING_MAX_QUEUE_CLAIMS=50
```
Worker claims events from outbox_events table, processes, marks done.
Reference: apps/integration_worker/shopping_price_worker.py (mode=event)

## Output table naming
<domain>_<entity>_<noun>  e.g.:
  shopping_price_runs
  shopping_price_run_items
  shopping_price_latest_snapshot

## Smoke tests
  scripts/smoke_shopping_driver_suite_local.ps1  — full supplier suite
  scripts/smoke_shopping_event_local.ps1         — single event flow

## Multi-strategy dispatch
See: apps/integration_worker/src/shopping_price_runtime/dispatcher.py
Strategies: http.vtex_persisted_query.v1, http.html_search.v1, playwright.pdp_first.v1
