# Worker Patterns

## Event-driven mode
```bash
MS_SHOPPING_WORKER_MODE=event
MS_SHOPPING_MAX_QUEUE_CLAIMS=50
MS_DATABASE_URL=postgres://...
```
Worker claims events from outbox_events, processes, marks done.
Smoke: `scripts/smoke_shopping_event_local.ps1`

## Multi-strategy dispatch (ref: src/shopping_price_runtime/dispatcher.py)
```python
family = config.family        # "http" or "playwright"
strategy = config.config_json.get("strategy")

if family == "http":
    return execute_http_runtime(config, strategy, ...)
if family == "playwright":
    return execute_playwright_runtime(config, strategy, ...)
```
Strategies: `http.vtex_persisted_query.v1`, `http.html_search.v1`, `playwright.pdp_first.v1`

## Output table naming
`<domain>_<entity>_<noun>` — e.g.:
- `shopping_price_runs`
- `shopping_price_run_items`
- `shopping_price_latest_snapshot`

## Smoke tests
- `scripts/smoke_shopping_driver_suite_local.ps1` — full supplier suite
- `scripts/smoke_shopping_event_local.ps1` — single event flow
- `scripts/smoke_shopping_queue_local.ps1` — queue mode

## Worker never does
- Call server_core HTTP endpoints
- Hardcode tenant_id, DSN, or credentials
- Write without setting `app.current_tenant_id`
- Use non-idempotent inserts
