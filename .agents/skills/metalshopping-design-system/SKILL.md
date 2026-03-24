---
name: metalshopping-design-system
description: Implement MetalShopping React UI — layout, components, tokens, tables, filters, pagination, selection, async states. Use for any frontend visual or component task. Enforces real tokens from global.css, packages/ui reuse, and useEffect+cancelled fetch pattern anchored to the actual codebase.
---

# MetalShopping Design System

## Before writing any UI
1. Check `references/primitives.md` — use existing widget if it covers the need
2. Check `references/tokens.md` — never hardcode hex, rem font-size, or spacing values
3. Read `tasks/lessons.md` — apply every lesson in this task

## Component ownership rule
- 3+ pages/features use it → `packages/ui/src/<Widget>.tsx`
- 1–2 uses → `packages/feature-<module>/src/components/<Component>.tsx`
- Page-local only → stays in page file

Promote test: *"Would this component work in an HR app without changes?"*
Yes → `packages/ui`. No → keep in feature.

---

## Data fetching — only valid pattern
```tsx
const [data, setData] = useState<T | null>(null);
const [loading, setLoading] = useState(true);
const [error, setError] = useState<string | null>(null);

useEffect(() => {
  let cancelled = false;
  async function load() {
    setLoading(true); setError(null);
    try {
      const result = await sdk.module.getX();
      if (!cancelled) setData(result);
    } catch (err) {
      if (!cancelled) setError(err instanceof Error ? err.message : "Falha ao carregar.");
    } finally {
      if (!cancelled) setLoading(false);
    }
  }
  void load();
  return () => { cancelled = true; };
}, [sdk, dependency]);
```
No `fetch()`. No hooks that don't exist in `@metalshopping/sdk-runtime`.
Every data-fetching component renders loading + error + empty state.

---

## Layout — every page

`AppFrame` wraps every route. Never write manual hero or h1.

```tsx
<AppFrame eyebrow="MetalShopping" title="Nome da Surface" subtitle="Descrição." aside={<div/>}>
  <div className={styles.stack}>
    <SurfaceCard title="Seção" subtitle="Desc" tone="default|soft">...</SurfaceCard>
  </div>
</AppFrame>
```
```css
.stack       { display: grid; gap: 14px; }
.twoCol      { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.metricsGrid { display: grid; grid-template-columns: repeat(4, minmax(0,1fr)); gap: 12px; }
@media (max-width: 1080px) {
  .metricsGrid { grid-template-columns: repeat(2,1fr); }
  .twoCol      { grid-template-columns: 1fr; }
}
```

`aside` pattern with MetricChip + actions:
```tsx
aside={
  <div className={styles.metricsGrid}>
    <MetricChip label="Na grade">{totalVisible}</MetricChip>
    <MetricChip label="Selecionados">{totalSelected}</MetricChip>
    <div style={{ gridColumn:"1/-1", display:"grid", gridTemplateColumns:"1fr 1fr", gap:8 }}>
      <Button variant="secondary">Ação secundária</Button>
      <Button variant="primary">Ação primária</Button>
    </div>
  </div>
}
```

---

## Loading, error, empty states

```tsx
{error   ? <StatusBanner tone="error">{error}</StatusBanner> : null}
{loading ? <p className={styles.loadingText}>Carregando...</p> : null}
```
```css
.loadingText { margin: 0; color: #73606a; font-size: .88rem; }
```

Empty state in table (single row with colSpan):
```tsx
<tr>
  <td colSpan={N} style={{ padding: 32, textAlign: "center", color: "#73606a" }}>
    {loading ? "Carregando..." : "Nenhum item encontrado para o filtro atual."}
  </td>
</tr>
```

---

## Table shell
Ref: `packages/feature-products/src/components/ProductsPortfolioTable.tsx`

```css
.tableWrap { overflow: auto; border-radius: 14px; border: 1px solid rgba(145,19,42,0.2); background: #fff; }
.table     { width: 100%; border-collapse: collapse; min-width: 860px; }
.table th  {
  padding: 10px 14px; text-align: left; font-size: .74rem; font-weight: 900;
  text-transform: uppercase; color: #8a1735;
  background: linear-gradient(180deg, #fff, #f8f3f5);
  position: sticky; top: 0; z-index: 1;
}
.table td  { padding: 10px 14px; border-top: 1px solid #f1e8ec; vertical-align: top; font-size: .82rem; color: #4d3e47; }
.table tbody tr:hover { background: rgba(145,19,42,0.04); }
.checkboxCol { width: 56px; text-align: center; }
.cellStrong  { display: block; color: #251b22; font-weight: 800; }
.cellMeta    { display: block; margin-top: 2px; color: #73606a; font-size: .82rem; }
.cellSmall   { display: block; margin-top: 4px; color: #73606a; font-size: .78rem; }
```

Sortable column header:
```tsx
<th><SortHeaderButton indicator={sortIndicator("col")} onClick={() => onSort("col")}>Label</SortHeaderButton></th>
```
`indicator`: `"↕"` neutral · `"↑"` asc · `"↓"` desc

Checkbox column:
```tsx
// thead
<th className={styles.checkboxCol}>
  <Checkbox checked={allVisible} onChange={togglePage} ariaLabel="Selecionar página" />
</th>
// tbody
<td className={styles.checkboxCol}>
  <Checkbox checked={selected} onChange={() => toggleRow(row.id)} ariaLabel={`Selecionar ${row.name}`} />
</td>
```

Do NOT extract the whole table to `packages/ui` — columns are always feature domain.
Extract only if the shell is identical in 3+ features and columns are configurable via props.

---

## Filters
Ref: `packages/feature-products/src/components/ProductsFiltersCard.tsx`

```css
.toolbar { display: grid; grid-template-columns: 1.6fr 1fr 1fr 1fr; gap: 10px; }
.fieldLabel { font-size: .72rem; font-weight: 900; letter-spacing: .08em; text-transform: uppercase; color: #73606a; }
.input {
  min-height: 38px; padding: 0 12px; border-radius: 12px;
  border: 1px solid rgba(145,19,42,0.22); background: #fff; color: #251b22;
}
.input:focus { outline: none; border-color: #c23b54; box-shadow: 0 0 0 3px rgba(184,56,79,.14); }
.filterFooter { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
.filterChips  { display: flex; flex-wrap: wrap; gap: 8px; }
.filterChip {
  display: inline-flex; align-items: center; gap: 8px; padding: 6px 10px;
  border: 1px solid rgba(145,19,42,0.18); border-radius: 999px;
  background: rgba(194,59,84,0.06); color: #7a1d33; font-size: .8rem; font-weight: 800; cursor: pointer;
}
.filterHint { color: #73606a; font-size: .82rem; font-weight: 700; }
@media (max-width: 1180px) { .toolbar { grid-template-columns: 1fr 1fr; } }
@media (max-width: 820px)  { .toolbar { grid-template-columns: 1fr; } }
```

```tsx
// FilterDropdown — always include "all" as first option
const opts: SelectMenuOption[] = [{ value: "", label: "Todos" }, ...items.map(i => ({ value: i, label: i }))];
<FilterDropdown id="unique-id" value={query.field} options={opts}
  onSelect={(v) => setQuery({ ...query, field: v, offset: 0 })} />

// Active chips
{activeFilters.length > 0
  ? activeFilters.map(f => (
      <button key={f.key} type="button" className={styles.filterChip}
        onClick={() => setQuery({ ...query, [f.key]: "", offset: 0 })}>
        {f.label} <span aria-hidden>×</span>
      </button>
    ))
  : <span className={styles.filterHint}>Nenhum filtro ativo.</span>}
<Button variant="quiet" disabled={activeFilters.length === 0} onClick={onClearAll}>Limpar filtros</Button>
```

Building `activeFilters`:
```ts
const activeFilters = [
  query.search.trim()  !== "" ? { key: "search",  label: `Busca: ${query.search.trim()}` }  : null,
  query.brand.trim()   !== "" ? { key: "brand",   label: `Marca: ${query.brand.trim()}` }   : null,
  query.status.trim()  !== "" ? { key: "status",  label: `Status: ${query.status.trim()}` } : null,
].filter((f): f is { key: string; label: string } => f !== null);
```

Status pills (simple filter bar, no dropdown):
```css
.filterBar { display: flex; flex-wrap: wrap; gap: 8px; }
.filterBtn { border: 1px solid #d1d5db; background: #fff; color: #374151; padding: 6px 10px; border-radius: 999px; font-size: .78rem; font-weight: 600; cursor: pointer; }
.filterBtnActive { border-color: #8a1735; color: #8a1735; background: #f7eff1; }
```

Always reset `offset: 0` when any filter changes.

---

## Pagination

Full bar (ref: `feature-products/src/components/ProductsPaginationBar.tsx`):
```tsx
<ProductsPaginationBar
  currentPage={currentPage} totalPages={totalPages} totalMatching={totalMatching}
  limit={query.limit} pageSizeOptions={[25, 50, 100]}
  canGoPrevious={canPrev} canGoNext={canNext}
  onChangeLimit={(l) => setQuery({ ...query, limit: l, offset: 0 })}
  onPrevious={() => setQuery({ ...query, offset: Math.max(0, query.offset - query.limit) })}
  onNext={() => setQuery({ ...query, offset: query.offset + query.limit })}
/>
```

Derived values (always compute these):
```ts
const totalPages  = Math.max(1, Math.ceil(totalMatching / query.limit));
const currentPage = Math.floor(query.offset / query.limit) + 1;
const canPrev     = query.offset > 0;
const canNext     = query.offset + query.limit < totalMatching;
```

Simple inline (internal wizard tables, no URL state):
```css
.pageRow { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.pageBtn { border: 0; background: transparent; color: #8a1735; font-size: 13px; font-weight: 800; cursor: pointer; }
.pageBtn:disabled { opacity: .45; cursor: not-allowed; }
.pageInfo { color: #73606a; font-size: 12px; font-weight: 700; }
```
```tsx
<div className={styles.pageRow}>
  <button className={styles.pageBtn} disabled={offset<=0||loading}
    onClick={() => setOffset(o => Math.max(0,o-limit))}>Página anterior</button>
  <span className={styles.pageInfo}>{returned} de {total} itens</span>
  <button className={styles.pageBtn} disabled={offset+returned>=total||loading}
    onClick={() => setOffset(o => o+limit)}>Próxima página</button>
</div>
```

---

## Row selection
Ref: `packages/feature-products/src/ProductsPortfolioPage.tsx` + `ProductsSelectionBar.tsx`

```ts
type SelectionMode = "explicit" | "filtered";
const [mode, setMode] = useState<SelectionMode>("explicit");
const [selectedIds, setSelectedIds] = useState<string[]>([]);

const totalSelected  = mode === "filtered" ? totalMatching : selectedIds.length;
const currentPageIds = useMemo(() => rows.map(r => r.id), [rows]);
const allVisible     = mode === "filtered" || (rows.length > 0 && rows.every(r => selectedIds.includes(r.id)));

function toggleRow(id: string) {
  setMode("explicit");
  setSelectedIds(cur => cur.includes(id) ? cur.filter(v => v !== id) : [...cur, id]);
}
function togglePage() {
  setMode("explicit");
  setSelectedIds(cur => {
    const shouldSelect = currentPageIds.some(id => !cur.includes(id));
    return shouldSelect ? Array.from(new Set([...cur, ...currentPageIds])) : cur.filter(id => !currentPageIds.includes(id));
  });
}
function selectFiltered() { setMode("filtered"); setSelectedIds([]); }
function clearSelection()  { setMode("explicit");  setSelectedIds([]); }
```

Selection row (summary bar):
```css
.selectionRow {
  display: inline-flex; align-items: center; flex-wrap: wrap; gap: 16px;
  padding: 7px 10px; border: 1px solid rgba(194,59,84,0.18); border-radius: 12px;
  background: rgba(194,59,84,0.05); color: #8a1735; font-size: .74rem; font-weight: 800;
}
```

Disable checkboxes when `mode === "filtered"`.
Use `"explicit"` mode for tables with row-level actions. Use `"filtered"` for bulk operations on full result sets.

---

## After task
1. `pnpm tsc --noEmit` passes
2. Real data visible in browser — no mocks
3. `git commit -m "feat(<scope>): implement <surface>"`

## References
- `references/tokens.md` — full token list: colors, typography, radius, shadows, spacing
- `references/primitives.md` — complete packages/ui inventory with props and promotion status
