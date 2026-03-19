# Shopping Level 1 Acceptance

## Status

Closed as `Pronto Nivel 1` in make-it-work-first mode.

## Scope checked

- contract-first Shopping read surface (`summary`, `runs`, `run detail`, `latest by product`)
- server_core module wired in composition
- sdk-runtime facade methods available
- React route `/shopping` bound to real API data (no placeholder)
- integration worker scaffold writes to Shopping Postgres tables

## Evidence (2026-03-19)

- `go build ./apps/server_core/...` -> pass
- `go test ./apps/server_core/...` -> pass
- `npm.cmd --workspace @metalshopping/web run typecheck` -> pass
- `npm.cmd --workspace @metalshopping/web run build` -> pass
- no direct `fetch()` in `apps/web/src/pages/ShoppingPage.tsx`
- structured request log in Shopping handler (`action`, `trace_id`, `result`, `duration_ms`)

## Files referenced

- `contracts/api/openapi/shopping_v1.openapi.yaml`
- `apps/server_core/internal/modules/shopping/*`
- `apps/server_core/migrations/0020_shopping_price_read_surfaces.sql`
- `packages/platform-sdk/src/index.ts`
- `apps/web/src/pages/ShoppingPage.tsx`
- `apps/integration_worker/shopping_price_worker.py`

## Notes

- Worker execution against live data source is operational follow-up and does not block Level 1 closure.
