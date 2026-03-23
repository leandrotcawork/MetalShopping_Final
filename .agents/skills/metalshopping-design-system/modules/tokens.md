# tokens.md — Design Tokens MetalShopping

Nunca use valores de cor, radius ou shadow hardcoded. Use as variáveis abaixo.
Definidas em `apps/web/src/app/global.css` como variáveis CSS `:root`.

## Paleta de cores

### Wine (primária — ações, destaques, brand)
```css
--ms-wine-900: #5f1227   /* hover escuro, estados pressionados */
--ms-wine-700: #8a1735   /* cor primária de ação e texto ativo */
--ms-wine-500: #c23b54   /* gradientes, botão primary hover */
```

### Ink (texto)
```css
--ms-ink-900: #251b22    /* títulos, texto principal */
--ms-ink-700: #4d3e47    /* texto secundário forte */
--ms-ink-500: #73606a    /* subtítulos, hints, labels */
```

### Line (bordas)
```css
--ms-line-200: #eadfe4   /* bordas padrão de cards */
--ms-line-100: #f3ebee   /* bordas sutis */
```

### Surface (fundos)
```css
--ms-surface-0:   #ffffff   /* fundo branco puro */
--ms-surface-50:  #fdfafc   /* fundo levemente rosado */
--ms-surface-100: #f8f2f5   /* fundo de seções, hover */
```

### Cores semânticas (use apenas em contexto específico)
```css
/* Sucesso */
color: #0b7c47;  background: #e6f6ee;

/* Aviso */
color: #a97900;  background: #fff8e8;

/* Erro */
color: #b91c1c;  background: #fef2f2;

/* Info */
color: #1b3a8a;  background: #eff6ff;
```

## Tipografia

### Família
```css
font-family: "Inter", -apple-system, system-ui, "Segoe UI", sans-serif;  /* UI */
font-family: "JetBrains Mono", ui-monospace, monospace;                  /* código */
```

### Escala de tamanhos (use sempre rem)
```
0.68rem   eyebrow / kicker uppercase
0.72rem   label de campo, metadado
0.75rem   subtitle de card
0.78rem   hint, legenda
0.82rem   texto de lista, célula de tabela
0.88rem   texto de parágrafo
0.94rem   botão
1.0rem    subtítulo de hero
1.12rem   título de SurfaceCard
1.25rem   valor de MetricCard
1.4rem    valor de MetricChip
2.0rem    título de AppFrame (h1)
```

### Pesos
```
600   texto de lista/célula leve
700   texto corrido, labels, botões secondary
800   títulos de card, labels uppercase, valores
900   títulos de hero, valores grandes, brand
```

## Border radius
```
6px    checkbox, badge pequeno
10px   botão small, input, pill de status
11px   avatar
12px   botão padrão, input, chip de filtro, MetricCard
14px   SurfaceCard small, MetricChip
16px   SurfaceCard padrão
20px   AppFrame hero, panel de wizard
999px  pill (status, badge arredondado)
```

## Sombras
```css
/* Card hero */
box-shadow: 0 12px 30px rgba(74, 39, 50, 0.06);

/* Dropdown menu */
box-shadow: 0 10px 38px rgba(2, 6, 23, 0.08);

/* Botão primary */
box-shadow: 0 4px 10px rgba(145, 19, 42, 0.22);

/* Focus ring (inputs, checkboxes) */
box-shadow: 0 0 0 4px rgba(199, 63, 103, 0.14);
```

## Espaçamento de gap (page layout)
```
4px    gap interno de chip/label
6px    gap interno de hero main
8px    gap entre itens de lista pequena
10px   gap de grid de filtros
12px   gap padrão de SurfaceCard interno
14px   gap de page stack (entre seções)
16px   gap de stack de página maior
20px   gap de container de shopping
24px   padding lateral de content area
```

## Transições padrão
```css
transition: all 0.2s ease;           /* genérico */
transition: all 180ms ease;          /* botões */
transition: all 0.16s ease;          /* itens de lista hover */
transition: width 0.5s ease;         /* progress bar */
```
