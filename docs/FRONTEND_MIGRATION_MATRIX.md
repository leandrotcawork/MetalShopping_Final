# Frontend Migration Matrix

## Purpose

Freeze legacy reuse decisions for frontend migration using one explicit matrix:

- preserve visually
- refactor structurally
- reject

This avoids ad hoc decisions while moving Home, Shopping, and Analytics.

## Source legacy reference

- `MetalDocs/frontend/apps/web/src/styles.css`
- `MetalDocs/frontend/apps/web/src/components/DocumentWorkspaceShell.tsx`
- `MetalDocs/frontend/apps/web/src/components/OperationsCenter.tsx`

## Matrix

| Artifact | Decision | Why | Target ownership |
| --- | --- | --- | --- |
| Color palette and typography rhythm | preserve visually | Visual identity is already trusted by users | `apps/web/src/app/global.css` + `packages/ui` |
| Sidebar and topbar interaction feel | preserve visually | Legacy shell flow is strong | `apps/web` shell composition |
| Card, chip, table density language | preserve visually | Preserves product familiarity | `packages/ui` primitives |
| Repeated page widgets | refactor structurally | Must become reusable and testable | `packages/ui` |
| Feature-specific data formatting | refactor structurally | Keep pages thin and deterministic | `packages/feature-*` view-model/adapters |
| API request assembly inside pages | reject | Breaks thin-client boundary | use `@metalshopping/sdk-runtime` only |
| Manual frontend DTOs | reject | Creates contract drift | generated contracts only |
| Generic shared dump folders | reject | Blurs ownership and scales poorly | explicit package boundaries only |

## Enforced migration flow

For each new surface:

1. classify legacy artifacts with this matrix
2. freeze OpenAPI contract for the surface
3. implement backend endpoint with real data
4. regenerate SDK
5. bind React page through sdk-runtime

## Current status

- Home: in progress on this matrix
- Shopping: pending classification and contract freeze
- Analytics: pending classification and contract freeze
