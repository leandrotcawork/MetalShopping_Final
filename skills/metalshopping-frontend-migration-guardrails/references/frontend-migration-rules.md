# Frontend Migration Rules

## Keep

- visual hierarchy and composition patterns from strong legacy pages
- established MetalShopping color direction and CSS feel
- package-oriented frontend decomposition
- operational workflow emphasis in `Products`, `Shopping`, and `Home`

## Refactor

- repeated cards, badges, filters, and tables into `packages/ui`
- page-local normalization into feature adapters and view-model builders
- local API utilities into generated or feature-owned clients
- inconsistent styling into CSS Modules plus shared UI primitives

## Reject

- manual DTO files under frontend packages when generated artifacts exist
- direct `fetch` inside route or page components
- generic `shared` buckets without explicit ownership
- app-level component folders acting as long-term homes for reusable widgets
- legacy sidecar assumptions as the source of truth for target runtime contracts

## Ownership map

### `apps/web`

- routing
- providers
- page composition
- app shell

### `packages/generated`

- generated SDKs
- generated DTOs and schema-derived types

### `packages/ui`

- reusable widgets
- design primitives
- shared view building blocks

### `packages/feature-*`

- feature adapters
- feature view models
- feature widgets and composition helpers

## Review questions

1. Does the migrated surface still look and feel like legacy MetalShopping in a good way?
2. Is any page doing transport parsing or request assembly directly?
3. Did a repeated visual pattern remain buried inside one page instead of becoming a widget?
4. Is any frontend package inventing parallel contract types?
5. Is the folder ownership explicit and future-proof?
