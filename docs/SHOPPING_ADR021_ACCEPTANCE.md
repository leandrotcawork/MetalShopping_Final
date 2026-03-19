# Shopping ADR-0021 Acceptance

## Status

Closed as implemented.

## Scope checked

- legacy workflow shape preserved in Shopping UI (`Upload -> Configurar -> Executar`)
- page consumes `@metalshopping/platform-sdk` only (no `fetch()` in page)
- run creation and run-request lifecycle status are contract-driven (`createRunRequest`, `getRunRequest`)
- supplier selection and execution parameters live in Step 2 and are backend-driven by bootstrap
- UI keeps stable primitives (`AppFrame`, `FilterDropdown`, `Checkbox`) without importing legacy transport shortcuts

## Evidence (2026-03-19)

- `npm.cmd --workspace @metalshopping/web run typecheck` -> pass
- `npm.cmd --workspace @metalshopping/web run build` -> pass
- no direct `fetch()` in `apps/web/src/pages/ShoppingPage.tsx`
- Shopping page uses SDK/runtime methods:
  - `shoppingApi.getBootstrap`
  - `shoppingApi.createRunRequest`
  - `shoppingApi.getRunRequest`
  - `shoppingApi.listRuns`
  - `shoppingApi.getRun`

## Files referenced

- `apps/web/src/pages/ShoppingPage.tsx`
- `apps/web/src/pages/ShoppingPage.module.css`
- `packages/platform-sdk/src/index.ts`
- `contracts/api/openapi/shopping_v1.openapi.yaml`
- `contracts/api/jsonschema/shopping_bootstrap_v1.schema.json`
- `contracts/api/jsonschema/shopping_create_run_request_v1.schema.json`
- `contracts/api/jsonschema/shopping_run_request_v1.schema.json`

