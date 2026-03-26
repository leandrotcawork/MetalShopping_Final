---
id: cortex-frontend-index
title: Frontend Cortex Index
region: frontend
type: cortex-index
tags: [frontend, react, typescript, sdk, feature-packages]
updated_at: 2026-03-26
---

# Frontend Cortex

## Scope

`apps/web/` + `packages/feature-*` + `packages/ui/` + `packages/platform-sdk/`

## Architecture

**Thin-client pattern**: all data flows through the generated SDK. Feature logic lives in `packages/feature-*` packages; `apps/web` is the assembly shell.

## Package Map

| Package | Purpose |
|---------|---------|
| `apps/web` | Shell app — mounts feature packages, routing, layout |
| `packages/feature-analytics` | Analytics surfaces (11 read surfaces) |
| `packages/feature-products` | Products browsing and catalog |
| `packages/feature-auth-session` | Auth, session management |
| `packages/ui` | Shared UI primitives (CSS modules, design tokens) |
| `packages/platform-sdk` | SDK runtime (`@metalshopping/sdk-runtime`) |

## Absolute Rules

- **SDK boundary**: `sdk.*` only — no raw `fetch()`, no direct API calls
- **Design tokens**: `$metalshopping-design-system` tokens only — no hex values
- **Component check**: verify `packages/ui/src/index.ts` before creating new components
- **Async states**: every data-fetching component must have loading + error + empty states
- **Fetch pattern**: `useEffect + cancelled flag` (abort controller)

## Frontend Migration

Legacy UI modules are being migrated to `feature-*` packages. Migration charter: `docs/FRONTEND_MIGRATION_CHARTER.md`. Track parity via `docs/FRONTEND_MIGRATION_MATRIX.md`.

**Migration workflow**: literal copy → runnable mocks → parity validation → backend/SDK integration

## Build Commands

```bash
npm run web:typecheck          # TypeScript check
npm run web:build              # Vite build
npm run web:test               # Vitest tests
npm --workspace @metalshopping/web run dev   # Dev server :5173
```

## Sinapses in This Region

_Add links to `.brain/sinapses/<frontend-topic>.md` files as they are created._
