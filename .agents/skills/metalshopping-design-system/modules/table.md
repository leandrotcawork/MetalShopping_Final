# table.md — Padrão de Tabela

## Shell visual obrigatório

Toda tabela usa este shell. Não invente wrapper, bordas ou cores diferentes.

```css
/* Em Feature.module.css */
.tableWrap {
  overflow: auto;
  border-radius: 14px;
  border: 1px solid rgba(145, 19, 42, 0.2);
  background: #fff;
}

.table {
  width: 100%;
  border-collapse: collapse;
  min-width: 860px;  /* ajuste conforme número de colunas */
}

.table th {
  padding: 10px 14px;
  text-align: left;
  font-size: 0.74rem;
  font-weight: 900;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #8a1735;                              /* --ms-wine-700 */
  background: linear-gradient(180deg, #fff, #f8f3f5);
  position: sticky;
  top: 0;
  z-index: 1;
}

.table td {
  padding: 10px 14px;
  border-top: 1px solid #f1e8ec;              /* --ms-line-100 derivado */
  vertical-align: top;
  color: #4d3e47;                              /* --ms-ink-700 */
  font-size: 0.82rem;
}

.table tbody tr:hover {
  background: rgba(145, 19, 42, 0.04);
}
```

```tsx
<div className={styles.tableWrap}>
  <table className={styles.table}>
    <thead>
      <tr>
        {/* colunas com SortHeaderButton se sortável, texto simples se não */}
      </tr>
    </thead>
    <tbody>
      {/* rows */}
    </tbody>
  </table>
</div>
```

## Coluna de checkbox (para tabelas com seleção)

```css
.checkboxColumn {
  width: 56px;
  text-align: center;
}
```

```tsx
// thead
<th className={styles.checkboxColumn}>
  <Checkbox checked={allSelected} onChange={onTogglePage} ariaLabel="Selecionar página" />
</th>

// tbody
<td className={styles.checkboxColumn}>
  <Checkbox checked={selected} onChange={() => onToggleRow(row.id)} ariaLabel={`Selecionar ${row.name}`} />
</td>
```

## Coluna sortável

Use `SortHeaderButton` de `@metalshopping/ui` para colunas com ordenação.

```tsx
import { SortHeaderButton } from "@metalshopping/ui";

<th>
  <SortHeaderButton indicator={sortIndicator("name")} onClick={() => onSort("name")}>
    Nome
  </SortHeaderButton>
</th>
```

`indicator` recebe: `"↕"` (neutro), `"↑"` (asc), `"↓"` (desc).

## Célula com dado principal + metadado

```css
.cellStrong { display: block; color: #251b22; font-weight: 800; }
.cellMeta   { display: block; margin-top: 2px; color: #73606a; font-size: 0.82rem; }
.cellSmall  { display: block; margin-top: 4px; color: #73606a; font-size: 0.78rem; }
```

```tsx
<td>
  <span className={styles.cellStrong}>{row.name}</span>
  <span className={styles.cellSmall}>Ref: {row.ref} · EAN: {row.ean}</span>
</td>
```

## Empty state e loading state

```tsx
// Sempre como linha única com colSpan
{rows.length === 0 || loading ? (
  <tr>
    <td colSpan={totalCols} style={{ padding: 32, textAlign: "center", color: "#73606a" }}>
      {loading ? "Carregando..." : "Nenhum item encontrado para o filtro atual."}
    </td>
  </tr>
) : rows.map(...)}
```

## Quando extrair para componente

NÃO extraia a tabela inteira para `packages/ui`. As colunas são sempre domínio da feature.

Extraia para `packages/ui` apenas se:
- O shell visual (wrapper + thead style + hover) for idêntico em 3+ features
- As colunas forem configuráveis via prop (column definitions)

Até lá: copie o shell CSS, mantenha colunas na feature.

## Referência canônica
`packages/feature-products/src/components/ProductsPortfolioTable.tsx` — tabela mais completa do sistema. Use como referência para seleção, sort e slots de actions/footer.
