# primitives.md — Inventário de packages/ui

Verifique aqui antes de criar qualquer componente.
Import: `import { X } from "@metalshopping/ui";`

---

## AppFrame
Hero de page com eyebrow, h1, subtitle e aside opcional.
```tsx
<AppFrame eyebrow="..." title="..." subtitle="..." aside={<div/>} fullWidth={false}>
  {children}
</AppFrame>
```
Props: `eyebrow` `title` `subtitle` `aside?` `fullWidth?` `children?`

---

## Button
Botão com variantes visuais.
```tsx
<Button variant="primary" disabled={false} onClick={fn}>Label</Button>
```
Props: `variant?: "primary" | "secondary" | "quiet"` + todos HTMLButtonAttributes
- `primary` — gradiente wine, branco, cta principal
- `secondary` — branco com borda, ação secundária
- `quiet` — borda suave, ação destrutiva ou de limpeza

---

## Checkbox
Checkbox estilizado com label opcional.
```tsx
<Checkbox
  checked={bool}
  onChange={(nextChecked) => fn()}
  label="Texto opcional"
  ariaLabel="Selecionar item X"
  disabled={false}
  id="optional-id"
/>
```
Props: `checked` `onChange` `label?` `ariaLabel?` `disabled?` `className?` `id?`

---

## FilterDropdown
Dropdown de seleção com busca interna quando > 10 opções.
```tsx
<FilterDropdown
  id="unique-id"
  value={currentValue}
  options={[{ value: "", label: "Todos" }, ...items]}
  onSelect={(value) => fn(value)}
  selectionMode="one"           // "one" | "duo" (multi)
  classNamesOverrides={{}}      // override de classes CSS
/>
```
Props: `id` `options` `onSelect` `value?` `values?` `selectionMode?` `disabled?` `searchThreshold?` `classNamesOverrides?`

Tipo: `SelectMenuOption = { label: string; value: string }`

---

## MetricCard
Card de métrica para grid no hero de page.
```tsx
<MetricCard
  label="Produtos"
  value="1.234 (98%)"
  hint="Total de produtos cadastrados"
/>
```
Props: `label: string` `value: ReactNode` `hint: ReactNode`

---

## MetricChip
Chip menor de métrica para aside do AppFrame hero.
```tsx
<MetricChip label="Na grade">1.234</MetricChip>
```
Props: `label: string` `children: ReactNode`

---

## SortHeaderButton
Botão de cabeçalho de coluna sortável.
```tsx
<SortHeaderButton indicator="↕" onClick={() => onSort("name")}>
  Nome
</SortHeaderButton>
```
Props: `indicator: string` `onClick: () => void` `children: ReactNode`
Indicadores: `"↕"` neutro, `"↑"` asc, `"↓"` desc

---

## StatusBanner
Banner de status (erro, sucesso) para uso no stack de page.
```tsx
<StatusBanner tone="error" className={styles.optionalClass}>
  Mensagem de erro aqui.
</StatusBanner>
```
Props: `tone?: "success" | "error"` `className?` `children`
- Use `tone="error"` para erros de fetch
- Use `tone="success"` para confirmações (raro)

---

## StatusPill
Pill de status para células de tabela ou listas.
```tsx
<StatusPill label="Ativo" tone="success" />
```
Props: `label: string` `tone?: "success" | "neutral" | "muted"`
- `success` — verde (ativo, concluído)
- `neutral` — azul (em andamento, padrão)
- `muted` — cinza (inativo, arquivado)

---

## SurfaceCard
Container de seção com header (título + subtitle + actions).
```tsx
<SurfaceCard
  title="Título"
  subtitle="Descrição"
  actions={<Button>...</Button>}
  tone="default"
  className={styles.optionalClass}
>
  {children}
</SurfaceCard>
```
Props: `title?` `subtitle?` `actions?` `tone?: "default" | "soft"` `className?` `children`

---

## O que NÃO está em packages/ui ainda

| Componente | Onde está | Status |
|------------|-----------|--------|
| ProductsPaginationBar | feature-products/components | Candidato a migrar |
| ProductsSelectionBar | feature-products/components | Candidato a migrar |
| StepPill + wizard | ShoppingPage (inline) | Feature-specific, não migrar |
| ProgressBar | ShoppingPage.module.css (inline) | Candidato futuro |
| LoadingState | não existe | Criar quando repetir 3x |
| ErrorState | não existe | Criar quando repetir 3x |
