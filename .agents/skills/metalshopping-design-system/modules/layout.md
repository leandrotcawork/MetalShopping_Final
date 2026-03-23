# layout.md — Padrões de Layout de Page

## AppFrame — wrapper obrigatório de toda page

Todo conteúdo de rota começa com `AppFrame`. Nunca escreva hero/título manual.

```tsx
import { AppFrame } from "@metalshopping/ui";

<AppFrame
  eyebrow="MetalShopping"          // kicker uppercase, sempre "MetalShopping" ou módulo
  title="Nome da Surface"
  subtitle="Descrição curta do que o usuário vê aqui."
  aside={<div>...</div>}           // opcional — chips, KPIs, ações do hero
>
  {/* conteúdo da page */}
</AppFrame>
```

Props:
- `eyebrow` string — kicker acima do título
- `title` string — h1 da page
- `subtitle` ReactNode — descrição, máx 64ch
- `aside?` ReactNode — coluna direita do hero (KPIs, MetricChip, botões)
- `fullWidth?` boolean — remove max-width e padding (use em sub-componentes dentro de features)
- `children?` ReactNode — conteúdo abaixo do hero

### Padrão de aside com MetricChip
```tsx
aside={
  <div style={{ display: "grid", gridTemplateColumns: "repeat(4, minmax(0,1fr))", gap: 8 }}>
    <MetricChip label="Na grade">{totalVisible}</MetricChip>
    <MetricChip label="Selecionados">{totalSelected}</MetricChip>
    <div style={{ gridColumn: "1 / -1", display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
      <Button variant="secondary">Ação secundária</Button>
      <Button variant="primary">Ação primária</Button>
    </div>
  </div>
}
```

## Page stack — container de seções

Todo o conteúdo abaixo do hero usa um stack vertical com gap fixo.
Nunca use margin individual entre seções — use o stack.

```css
/* Em FeatureName.module.css */
.stack {
  display: grid;
  gap: 14px;   /* gap padrão entre seções de page */
}
```

```tsx
<AppFrame ...>
  <div className={styles.stack}>
    <SurfaceCard ...>...</SurfaceCard>
    <SurfaceCard ...>...</SurfaceCard>
  </div>
</AppFrame>
```

Gap por contexto:
- `14px` — gap padrão entre SurfaceCards em page
- `16px` — gap levemente maior (HomePage usa 16px)
- `20px` — gap de container de wizard/shopping

## SurfaceCard — container de seção

Use para agrupar conteúdo relacionado dentro do stack.

```tsx
import { SurfaceCard } from "@metalshopping/ui";

<SurfaceCard
  title="Título da seção"
  subtitle="Descrição curta"    // opcional
  actions={<Button>...</Button>} // opcional — alinhado à direita do header
  tone="default"                 // "default" | "soft"
>
  {/* conteúdo */}
</SurfaceCard>
```

- `tone="soft"` → fundo levemente gradiente (use para cards de leitura/informativo)
- `tone="default"` → fundo branco puro (use para cards com interação)

## Grid de métricas no hero

```css
.metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

@media (max-width: 1080px) {
  .metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 700px) {
  .metrics { grid-template-columns: 1fr; }
}
```

```tsx
<div className={styles.metrics}>
  <MetricCard label="Total" value="1.234" hint="Descrição do dado" />
  <MetricCard label="Ativos" value="980 (79%)" hint="Itens ativos" />
</div>
```

## Grid operacional de dois painéis

Para duas colunas iguais dentro de um stack (ex: cobertura + snapshot):

```css
.operationalGrid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 12px;
}

@media (max-width: 1080px) {
  .operationalGrid { grid-template-columns: 1fr; }
}
```

## Regras gerais
- Nunca use `margin-top` / `margin-bottom` entre seções — use gap no container
- Nunca crie h1 ou hero manual — use `AppFrame`
- Nunca crie card container manual — use `SurfaceCard`
- Padding lateral do content area: `24px` (definido no AppShell, não repita)
