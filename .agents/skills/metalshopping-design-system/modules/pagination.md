# pagination.md — Paginação

## Situação atual
`ProductsPaginationBar` existe em `packages/feature-products/src/components/ProductsPaginationBar.tsx`.
É o componente mais completo. Candidato a migrar para `packages/ui` como `PaginationBar`.

Até a migração: importe de `@metalshopping/feature-products` se já estiver na feature-products,
ou copie o padrão abaixo para outras features.

## Props de ProductsPaginationBar

```tsx
<ProductsPaginationBar
  currentPage={currentPage}       // número da página atual (1-based)
  totalPages={totalPages}         // total de páginas
  totalMatching={totalMatching}   // total de itens encontrados
  limit={query.limit}             // itens por página
  pageSizeOptions={[25, 50, 100]} // opções de tamanho de página
  canGoPrevious={canGoPrevious}   // boolean
  canGoNext={canGoNext}           // boolean
  onChangeLimit={(limit) => setQuery({ ...query, limit, offset: 0 })}
  onPrevious={() => setQuery({ ...query, offset: Math.max(0, query.offset - query.limit) })}
  onNext={() => setQuery({ ...query, offset: query.offset + query.limit })}
/>
```

## Cálculos auxiliares obrigatórios

```ts
const totalPages    = totalMatching === 0 ? 1 : Math.max(1, Math.ceil(totalMatching / query.limit));
const currentPage   = Math.floor(query.offset / query.limit) + 1;
const canGoPrevious = query.offset > 0;
const canGoNext     = query.offset + query.limit < totalMatching;
```

## Paginação simples (sem page size selector)

Para tabelas internas sem necessidade de alterar tamanho de página (ex: ShoppingPage):

```css
.selectRow {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin: 8px 0;
}

.selectRow button {
  border: 0;
  background: transparent;
  color: #8a1735;
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
}

.selectRow button:disabled { opacity: 0.45; cursor: not-allowed; }

.selectRow span {
  color: #73606a;
  font-size: 12px;
  font-weight: 700;
}
```

```tsx
<div className={styles.selectRow}>
  <button
    type="button"
    disabled={offset <= 0 || loading}
    onClick={() => setOffset((o) => Math.max(0, o - limit))}
  >
    Página anterior
  </button>
  <span>{returned} de {total} itens</span>
  <button
    type="button"
    disabled={offset + returned >= total || loading}
    onClick={() => setOffset((o) => o + limit)}
  >
    Próxima página
  </button>
</div>
```

## Quando usar qual padrão

| Caso | Padrão |
|------|--------|
| Feature com query persistida na URL, múltiplos tamanhos de página | `ProductsPaginationBar` |
| Tabela interna de wizard, sem URL, sem page size | Paginação simples acima |

## Migração futura para packages/ui

Quando `ProductsPaginationBar` for movido para `packages/ui`:
1. Mover arquivo para `packages/ui/src/PaginationBar.tsx`
2. Mover CSS module para `packages/ui/src/PaginationBar.module.css`
3. Exportar de `packages/ui/src/index.ts`
4. Atualizar import em `feature-products`
5. Atualizar este documento
