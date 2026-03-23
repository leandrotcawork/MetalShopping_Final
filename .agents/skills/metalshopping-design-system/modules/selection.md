# selection.md — Seleção de Linhas em Tabela

## Referência canônica
`packages/feature-products/src/components/ProductsSelectionBar.tsx`
`packages/feature-products/src/ProductsPortfolioPage.tsx` — lógica de estado

## Dois modos de seleção

### explicit — seleção manual de linhas
O usuário marca checkboxes individualmente ou clica "Selecionar página".
`selectedIds` é um array de IDs explícitos.

### filtered — selecionar todos os filtrados
O usuário clica "Selecionar filtrados". Semanticamente seleciona tudo que o filtro retorna,
não apenas a página visível. `selectedIds` fica vazio, o modo comunica a intenção.

```ts
type SelectionMode = "explicit" | "filtered";

const [selectionMode, setSelectionMode] = useState<SelectionMode>("explicit");
const [selectedIds, setSelectedIds] = useState<string[]>([]);

// Total selecionado depende do modo
const totalSelected = selectionMode === "filtered" ? totalMatching : selectedIds.length;
```

## Estado derivado obrigatório

```ts
// IDs da página atual
const currentPageIds = useMemo(() => rows.map((row) => row.id), [rows]);

// Todos da página atual estão selecionados?
const allVisibleSelected =
  selectionMode === "filtered" ||
  (rows.length > 0 && rows.every((row) => selectedIds.includes(row.id)));
```

## Funções de controle

```ts
function toggleRow(id: string) {
  setSelectionMode("explicit");
  setSelectedIds((current) =>
    current.includes(id) ? current.filter((v) => v !== id) : [...current, id],
  );
}

function toggleCurrentPage() {
  setSelectionMode("explicit");
  setSelectedIds((current) => {
    const shouldSelect = currentPageIds.some((id) => !current.includes(id));
    if (shouldSelect) return Array.from(new Set([...current, ...currentPageIds]));
    return current.filter((id) => !currentPageIds.includes(id));
  });
}

function selectFiltered() {
  setSelectionMode("filtered");
  setSelectedIds([]);
}

function clearSelection() {
  setSelectionMode("explicit");
  setSelectedIds([]);
}
```

## Checkbox no thead

```tsx
<th className={styles.checkboxColumn}>
  <Checkbox
    checked={allVisibleSelected}
    disabled={rows.length === 0 || selectionMode === "filtered"}
    label=""
    ariaLabel="Selecionar produtos da página"
    onChange={toggleCurrentPage}
  />
</th>
```

## Checkbox no tbody

```tsx
<td className={styles.checkboxColumn}>
  <Checkbox
    checked={selectionMode === "filtered" || selectedIds.includes(row.id)}
    disabled={selectionMode === "filtered"}
    label=""
    ariaLabel={`Selecionar ${row.name}`}
    onChange={() => toggleRow(row.id)}
  />
</td>
```

## SelectionBar — barra de ações e resumo

Use `ProductsSelectionBar` como referência. Tem dois modos:
- `mode="actions"` — botões de ação (exportar, selecionar página, etc.)
- `mode="summary"` — linha de resumo (modo, quantidade, fornecedores)

```tsx
// Barra de ações acima da tabela
<ProductsSelectionBar
  rowsCount={rows.length}
  allVisibleSelected={allVisibleSelected}
  totalSelected={totalSelected}
  selectionMode={selectionMode}
  mode="actions"
  onToggleCurrentPage={toggleCurrentPage}
  onSelectFiltered={selectFiltered}
  onClearSelection={clearSelection}
/>

// Resumo dentro da tabela (abaixo do header, acima dos dados)
<ProductsSelectionBar
  ...mesmas props...
  mode="summary"
/>
```

## CSS da selection row (resumo)

```css
.selectionRow {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  padding: 7px 10px;
  border: 1px solid rgba(194, 59, 84, 0.18);
  border-radius: 12px;
  background: rgba(194, 59, 84, 0.05);
  color: #8a1735;
  font-size: 0.74rem;
  font-weight: 800;
}
```

## Quando não usar seleção
Para tabelas de histórico/log sem ação em lote (ex: tabela de runs em ShoppingPage),
não implemente seleção. Tabela é somente leitura — clique na linha → abre detalhe.
