---
name: metalshopping-frontend
description: Implement MetalShopping React pages and components. Called by $ms for T5. Enforces SDK-only data binding, design tokens, widget reuse from packages/ui, and required loading/error/empty states. Prevents fetch(), hardcoded styles, duplicate components.
---

# MetalShopping Frontend

## Before writing any component
1. Check `packages/ui/src/index.ts` — use existing widget if it covers the need:
   `AppFrame` `Button` `Checkbox` `FilterDropdown` `MetricCard` `MetricChip`
   `SortHeaderButton` `StatusBanner` `StatusPill` `SurfaceCard` `SelectMenu`
2. Check `packages/platform-sdk/src/index.ts` — use existing hook if available

## Widget ownership decision
- Used in 3+ pages → `packages/ui/src/<Widget>.tsx`
- Used in 1-2 pages → `packages/feature-<module>/src/<Component>.tsx`
- Page-local only → stays in `apps/web/src/pages/<module>/`

## Data binding — only valid pattern
```tsx
import { useXxx } from '@metalshopping/platform-sdk'

export function MyPage() {
  const { data, isLoading, error } = useXxx()

  if (isLoading) return <LoadingState />
  if (error) return <StatusBanner variant="error" message={error.message} />
  if (!data) return <EmptyState message="No data yet" />

  return <MyContent data={data} />
}
```
No `fetch()`. No `useEffect + fetch`. No manual response parsing.
All 3 states (loading, error, empty) are required on every data-fetching component.

## Styling — tokens only, never hardcoded values
```css
/* correct */
color: var(--color-text-primary);
font-size: var(--font-size-sm);
padding: var(--spacing-md);
border-radius: var(--radius-md);

/* never */
color: #333;
font-size: 13px;
padding: 16px;
```
All styles in `.module.css`. No `style={{ }}` for layout or typography.

## Typography rules
| Element | Tokens |
|---|---|
| Page title | `--font-size-xl` + `--font-weight-semibold` |
| Section header | `--font-size-md` + `--font-weight-semibold` |
| Table header | `--font-size-sm` + `--color-text-secondary` |
| Table cell | `--font-size-sm` + `--font-weight-regular` |
| Metric value | `--font-size-lg` + `--font-weight-bold` |
| Label/caption | `--font-size-xs` + `--color-text-muted` |

## Reference HTML (when provided by Leandro)
Extract: color intent, spacing rhythm, component boundaries, typography hierarchy
Ignore: hardcoded px values, inline styles, legacy class names, raw HTML structure
Reimplement using design tokens and packages/ui widgets.

## After task
1. `pnpm tsc --noEmit` passes
2. No console errors in browser
3. Real data visible (no mocks)
4. `git commit -m "feat(<m>): implement React page"`

## References
- `references/design-tokens.md` — full token list
- `references/migration-rules.md` — porting from legacy frontend
