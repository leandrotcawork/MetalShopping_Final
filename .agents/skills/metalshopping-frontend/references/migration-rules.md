# Migration Rules — Legacy to New Frontend

## What to preserve
- Visual hierarchy and composition patterns that work well
- Color direction and spacing feel
- Shell behavior and interaction patterns users are familiar with

## What to refactor
- Repeated cards/badges/filters/tables → extract to `packages/ui`
- Page-local API utilities → platform-sdk hooks
- Inconsistent CSS → CSS Modules + design tokens

## What to reject
- Manual DTO files when generated types exist
- Direct `fetch()` inside page/route components
- Generic `shared/` folders without explicit ownership
- Reusable widgets created inside `apps/web/src/pages/`

## Ownership map
| Location | Owns |
|---|---|
| `apps/web/src/pages/<m>/` | routing, page composition |
| `packages/generated/` | generated SDKs and types only |
| `packages/ui/` | reusable widgets (3+ uses) |
| `packages/feature-<m>/` | feature adapters, view models, local widgets |

## Flow when porting a legacy screen
1. OpenAPI contract finalized and SDK regenerated
2. Identify which packages/ui widgets cover the visual needs
3. Implement page consuming platform-sdk hook
4. Preserve visual feel — reimplement structure with tokens
