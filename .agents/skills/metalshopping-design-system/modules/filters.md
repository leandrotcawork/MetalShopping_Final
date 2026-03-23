# filters.md — Padrões de Filtros

## Componentes disponíveis
- `FilterDropdown` de `@metalshopping/ui` — dropdown de seleção única ou múltipla
- `Checkbox` de `@metalshopping/ui` — para filtros booleanos
- `Button` de `@metalshopping/ui` — para "Limpar filtros"

## Toolbar de filtros (padrão completo)

Usado em `ProductsFiltersCard`. Use como referência para qualquer toolbar com dropdowns.

```css
.toolbar {
  display: grid;
  grid-template-columns: 1.6fr 1fr 1fr 1fr;  /* busca mais larga + 3 dropdowns */
  gap: 10px;
}

.field {
  display: grid;
  gap: 8px;
}

.label {
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #73606a;
}

.input {
  min-height: 38px;
  padding: 0 12px;
  border-radius: 12px;
  border: 1px solid rgba(145, 19, 42, 0.22);
  background: #fff;
  color: #251b22;
}

.input:focus {
  outline: none;
  border-color: #c23b54;
  box-shadow: 0 0 0 3px rgba(184, 56, 79, 0.14);
}

@media (max-width: 1180px) { .toolbar { grid-template-columns: 1fr 1fr; } }
@media (max-width: 820px)  { .toolbar { grid-template-columns: 1fr; } }
```

```tsx
<div className={styles.toolbar}>
  <label className={styles.field}>
    <span className={styles.label}>Busca</span>
    <input
      className={styles.input}
      value={searchDraft}
      placeholder="Texto de busca..."
      onChange={(e) => onSearchChange(e.target.value)}
    />
  </label>

  <div className={styles.field}>
    <span className={styles.label}>Marca</span>
    <FilterDropdown
      id="filter-brand"
      value={query.brand}
      options={brandOptions}
      onSelect={(v) => onChangeQuery({ ...query, brand: v, offset: 0 })}
    />
  </div>

  {/* mais FilterDropdowns... */}
</div>
```

## FilterDropdown — uso

```tsx
import { FilterDropdown, type SelectMenuOption } from "@metalshopping/ui";

// Sempre inclua opção "todos" como primeiro item
const options: SelectMenuOption[] = [
  { value: "", label: "Todas as marcas" },
  ...brands.map((b) => ({ value: b, label: b })),
];

<FilterDropdown
  id="unique-id"              // único na página
  value={currentValue}        // string — valor selecionado
  options={options}
  onSelect={(value) => handleSelect(value)}
  classNamesOverrides={{      // opcional — para ajuste de tamanho
    wrap: styles.dropdownWrap,
    trigger: styles.dropdownTrigger,
  }}
/>
```

## Chips de filtros ativos

Exibidos abaixo do toolbar. Cada filtro ativo vira um chip clicável que remove o filtro.

```css
.filterFooter {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.filterChips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filterChip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 6px 10px;
  border: 1px solid rgba(145, 19, 42, 0.18);
  border-radius: 999px;
  background: rgba(194, 59, 84, 0.06);
  color: #7a1d33;
  font-size: 0.8rem;
  font-weight: 800;
  cursor: pointer;
}

.filterHint {
  color: #73606a;
  font-size: 0.82rem;
  font-weight: 700;
}
```

```tsx
<div className={styles.filterFooter}>
  <div className={styles.filterChips}>
    {activeFilters.length > 0 ? (
      activeFilters.map((filter) => (
        <button
          key={filter.key}
          type="button"
          className={styles.filterChip}
          onClick={() => onRemoveFilter(filter.key)}
        >
          <span>{filter.label}</span>
          <span aria-hidden>×</span>
        </button>
      ))
    ) : (
      <span className={styles.filterHint}>Nenhum filtro ativo.</span>
    )}
  </div>

  <Button variant="quiet" disabled={activeFilters.length === 0} onClick={onClearAll}>
    Limpar filtros
  </Button>
</div>
```

## Pills de status simples (sem dropdown)

Para filtrar por status com pills visuais (ex: Todos / Queued / Running):

```css
.filterBar { display: flex; flex-wrap: wrap; gap: 8px; }

.filterButton {
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.filterActive {
  border-color: #8a1735;
  color: #8a1735;
  background: #f7eff1;
}
```

```tsx
{statusOptions.map((opt) => (
  <button
    key={opt.value}
    type="button"
    className={`${styles.filterButton} ${selected === opt.value ? styles.filterActive : ""}`.trim()}
    onClick={() => setSelected(opt.value)}
  >
    {opt.label}
  </button>
))}
```

## Construindo activeFilters a partir de query

```ts
const activeFilters = [
  query.search.trim()  !== "" ? { key: "search",  label: `Busca: ${query.search.trim()}` }  : null,
  query.brand.trim()   !== "" ? { key: "brand",   label: `Marca: ${query.brand.trim()}` }   : null,
  query.status.trim()  !== "" ? { key: "status",  label: `Status: ${query.status.trim()}` } : null,
].filter((f): f is { key: string; label: string } => f !== null);
```

## Reset de paginação ao filtrar

Sempre que qualquer filtro mudar, resete o offset para 0:

```ts
onChangeQuery({ ...query, brand: value, offset: 0 })
```
