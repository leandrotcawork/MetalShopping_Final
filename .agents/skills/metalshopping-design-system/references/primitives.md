# packages/ui — Primitives Inventory

Import from: `@metalshopping/ui`
Source: `packages/ui/src/index.ts`

Check here before creating any component. If it exists — use it.

---

## AppFrame
Hero wrapper. Required on every route. Never write manual hero or `<h1>`.
```tsx
<AppFrame
  eyebrow="MetalShopping"      // uppercase kicker — always "MetalShopping" or module name
  title="Surface Name"
  subtitle="Short description up to 64ch."
  aside={<ReactNode />}        // optional right column: MetricChip grid, Button row
  fullWidth={false}            // true = removes max-width and padding (use in sub-components)
>
  {children}                   // rendered below hero
</AppFrame>
```

---

## Button
```tsx
<Button variant="primary|secondary|quiet" disabled={false} onClick={fn} className="">
  Label
</Button>
```
- `primary` — wine gradient, white text, main CTA
- `secondary` — white bg, border, secondary action
- `quiet` — soft border, destructive or clear action
Accepts all `ButtonHTMLAttributes`.

---

## Checkbox
```tsx
<Checkbox
  checked={bool}
  onChange={(nextChecked?) => fn()}
  label="Optional text label"
  ariaLabel="Describe the selection"   // required when no label
  disabled={false}
  id="optional-html-id"
  className=""
/>
```

---

## FilterDropdown
Dropdown with internal search when options > `searchThreshold`.
```tsx
<FilterDropdown
  id="unique-id-per-page"
  value={currentValue}                 // selected string value
  options={SelectMenuOption[]}         // { value: string; label: string }[]
  onSelect={(value: string) => fn()}
  selectionMode="one"                  // "one" | "duo" (multi-select)
  values={[]}                          // used when selectionMode="duo"
  disabled={false}
  searchThreshold={10}                 // show search input when options > N
  classNamesOverrides={{               // override wrap/trigger CSS class
    wrap: styles.myWrap,
    trigger: styles.myTrigger,
  }}
/>
```
Always include `{ value: "", label: "Todos" }` as the first option.

---

## MetricCard
For 4-column metric grids in page stack.
```tsx
<MetricCard
  label="Label uppercase"
  value="1.234 (98%)"         // ReactNode — formatted string or JSX
  hint="Short explanation"    // ReactNode
/>
```

---

## MetricChip
For AppFrame `aside` hero column.
```tsx
<MetricChip label="Na grade">
  {totalVisible}
</MetricChip>
```

---

## SortHeaderButton
For sortable `<th>` columns.
```tsx
<SortHeaderButton
  indicator="↕"             // "↕" neutral | "↑" asc | "↓" desc
  onClick={() => onSort("column_key")}
>
  Column Label
</SortHeaderButton>
```

---

## StatusBanner
For fetch errors rendered in the page stack.
```tsx
<StatusBanner tone="error" className={styles.optional}>
  Error message string
</StatusBanner>
```
- `tone="error"` — pink bg, dark red text (use for fetch errors)
- `tone="success"` — white bg, muted text (use for confirmations — rare)

---

## StatusPill
For status values in table cells.
```tsx
<StatusPill label="Ativo" tone="success" />
<StatusPill label="Inativo" tone="muted" />
<StatusPill label="Em andamento" tone="neutral" />
```
- `success` — green (active, completed)
- `neutral` — blue (in progress, default)
- `muted` — grey (inactive, archived)

---

## SurfaceCard
Section container within page stack.
```tsx
<SurfaceCard
  title="Section Title"
  subtitle="Optional short description"
  actions={<Button>...</Button>}   // optional — right-aligned in header
  tone="default"                   // "default" (white) | "soft" (warm gradient)
  className={styles.optional}
>
  {children}
</SurfaceCard>
```
- `tone="soft"` — informational/read-only cards
- `tone="default"` — interactive cards

---

## What does NOT exist in packages/ui yet

| Component | Current location | Promote when |
|-----------|-----------------|--------------|
| `ProductsPaginationBar` | `feature-products/src/components/` | Used in 3+ features |
| `ProductsSelectionBar` | `feature-products/src/components/` | Used in 3+ features |
| `PaginationSimple` | inline in ShoppingPage | Extract when repeated |
| `LoadingState` | doesn't exist | Create when pattern repeats 3+ times |
| `ErrorState` | doesn't exist — use StatusBanner | Create when pattern repeats 3+ times |
| `StepWizard / StepPill` | inline in ShoppingPage | Feature-specific — do not promote |
| `ProgressBar` | inline CSS in ShoppingPage | Promote when used elsewhere |

---

## Promotion checklist (when adding to packages/ui)
1. Component has zero domain knowledge — no DTO types, no API imports
2. Works without changes in a different app (HR, CRM, etc.)
3. CSS is in `.module.css` alongside the component
4. Exported from `packages/ui/src/index.ts`
5. Props are typed — no `any`
