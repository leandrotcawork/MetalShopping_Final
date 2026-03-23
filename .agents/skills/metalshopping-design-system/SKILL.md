---
name: metalshopping-design-system
description: Skill modular de design system para o frontend MetalShopping.. 
---

Leia este índice primeiro, depois carregue apenas o módulo relevante para a tarefa.

## Regra de ouro
Nunca invente valores de cor, tipografia ou espaçamento. Sempre consulte `tokens.md`.
Nunca crie um componente sem verificar `primitives.md` — ele pode já existir em `packages/ui`.

## Quando carregar cada módulo

| Tarefa | Módulo |
|--------|--------|
| Qualquer valor CSS (cor, radius, shadow, font) | `tokens.md` |
| Montar layout de page, hero, grid de seções | `layout.md` |
| Estado de carregamento ou erro | `async-state.md` |
| Qualquer tabela de dados | `table.md` |
| Busca, dropdowns de filtro, chips ativos | `filters.md` |
| Navegação entre páginas de lista | `pagination.md` |
| Seleção de linhas em tabela | `selection.md` |
| Verificar se componente já existe antes de criar | `primitives.md` |

## Estrutura dos módulos
```
modules/
  tokens.md        cores, tipografia, border-radius, sombras
  layout.md        AppFrame, SurfaceCard, page stack, grid padrão
  async-state.md   loading/error — o que criar + como usar
  table.md         shell visual, empty state, padrão de colunas
  filters.md       toolbar + FilterDropdown + chips ativos + limpar
  pagination.md    PaginationBar — quando usar e como montar
  selection.md     SelectionBar, modo explicit vs filtered
  primitives.md    inventário completo de packages/ui com props
```

## Critério de promoção para packages/ui
Pergunta: "este componente funcionaria num app de RH sem mudança?"
- Sim → candidato a `packages/ui`
- Não → fica na feature ou documentado aqui como padrão

## Pacotes do design system
- Primitivos: `@metalshopping/ui`
- Tokens CSS: definidos em `apps/web/src/app/global.css` como variáveis `:root`
- Fontes: Inter (UI) e JetBrains Mono (código/mono)
