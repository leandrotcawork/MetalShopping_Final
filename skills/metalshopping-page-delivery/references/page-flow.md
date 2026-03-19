# Page Flow

## Read order

1. `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`
2. `docs/FRONTEND_MIGRATION_CHARTER.md`
3. Legacy page file if it exists (inspect before porting)

## Ownership map

| Location | Owns |
|---|---|
| `apps/web/src/pages/<module>/` | route and page composition |
| `packages/feature-<module>/` | feature adapter, view model builder |
| `packages/ui/` | widgets used in 3+ places |
| `packages/generated/` | SDK types — generated, never edit manually |

## Files this skill normally touches

- `apps/web/src/pages/<module>/index.tsx`
- `packages/feature-<module>/src/` (if adapter needed)
- `packages/ui/src/` (only for extracted widgets)

## Rules

- SDK is regenerated before the page starts
- legacy visual language is preserved where it already works
- page files do not contain transport or fetch logic
- no new `shared/` or ambiguous ownership folders
- widget extraction only when 3+ occurrences — not earlier
