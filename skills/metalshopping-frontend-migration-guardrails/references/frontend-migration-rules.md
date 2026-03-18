# Frontend Migration Rules

## Keep

- visual hierarchy and composition patterns from strong legacy pages
- established MetalShopping color direction and CSS feel
- package-oriented frontend decomposition
- operational workflow emphasis in `Products`, `Shopping`, and `Home`
- shell behavior and interaction patterns that already worked well in the legacy app

## Refactor

- repeated cards, badges, filters, and tables into `packages/ui`
- page-local normalization into feature adapters and view-model builders
- local API utilities into generated or feature-owned clients
- inconsistent styling into CSS Modules plus shared UI primitives
- shell, typography, and table density into a reusable baseline before surface proliferation

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
2. Was the shell, typography, and widget baseline extracted before the page was ported?
3. Is any page doing transport parsing or request assembly directly?
4. Did a repeated visual pattern remain buried inside one page instead of becoming a widget?
5. Is any frontend package inventing parallel contract types?
6. Is the folder ownership explicit and future-proof?
