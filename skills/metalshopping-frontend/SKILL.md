---
name: metalshopping-frontend
description: Implement MetalShopping React pages and components. Enforces SDK-only data binding, design tokens, widget reuse from packages/ui, and required states. Use for any new page, component, or frontend feature.
---

# MetalShopping Frontend

## Before writing any component
1. Check `packages/ui/src/index.ts` — use existing widget if available:
   AppFrame, Button, Checkbox, FilterDropdown, MetricCard, MetricChip,
   SortHeaderButton, StatusBanner, StatusPill, SurfaceCard, SelectMenu
2. Check `packages/platform-sdk/src/index.ts` — use existing hook if available

## Widget ownership (where to create)
- 3+ pages use it → `packages/ui/src/`
- 1-2 pages → `packages/feature-<module>/src/`
- page-local → `apps/web/src/pages/<module>/`

## Data binding — only valid pattern
```tsx
import { useXxx } from '@metalshopping/platform-sdk'

function MyPage() {
  const { data, isLoading, error } = useXxx()
  if (isLoading) return <LoadingState />           // always required
  if (error) return <StatusBanner variant="error" message={error.message} />  // always required
  if (!data?.items?.length) return <EmptyState />  // always required
  return <MyContent data={data} />
}
```
No `fetch()`. No `useEffect + fetch`. No manual response parsing.

## Styling — always tokens, never hardcoded
```css
/* good */
color: var(--color-text-primary);
font-size: var(--font-size-sm);
padding: var(--spacing-md);

/* never */
color: #333;
font-size: 13px;
padding: 16px;
```
All styles in `.module.css`. No `style={{ }}` for layout or typography.

## Typography rules
| Element | Token |
|---|---|
| Page title | `--font-size-xl` + `--font-weight-semibold` |
| Section header | `--font-size-md` + `--font-weight-semibold` |
| Table header | `--font-size-sm` + `--color-text-secondary` |
| Table cell | `--font-size-sm` + `--font-weight-regular` |
| Metric value | `--font-size-lg` + `--font-weight-bold` |
| Label/caption | `--font-size-xs` + `--color-text-muted` |

## Reference HTML (when provided)
Extract: color intent, spacing rhythm, component boundaries
Ignore: px values, inline styles, legacy class names, HTML structure

## References
- `references/design-tokens.md` — full token list
- `references/table-pattern.md` — standard table structure
