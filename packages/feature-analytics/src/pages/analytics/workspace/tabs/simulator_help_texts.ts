import type { TooltipHelpCopy } from "../components/InfoTooltipLabel";

export const SIMULATOR_HELP: Record<string, TooltipHelpCopy> = {
  "preco-rs": {
    title: "Preco (R$)",
    items: [
      "Preco final sugerido para o produto no cenario.",
      "Afeta desconto, margem e gap de mercado.",
    ],
  },
  "desconto-pct": {
    title: "Desconto (%)",
    items: [
      "Percentual reduzido sobre o preco atual.",
      "Descontos altos aumentam risco de margem baixa.",
    ],
  },
  "margem-alvo-pct": {
    title: "Margem Alvo (%)",
    items: [
      "Margem de contribuicao desejada no cenario.",
      "Ao ajustar, o simulador recalcula o preco necessario.",
    ],
  },
  "frete-encargos-pct": {
    title: "Fretes/Encargos (%)",
    items: [
      "Percentual aplicado em cima do Custo medio base.",
      "Exemplo: custo medio R$100 e frete/encargos 12% => novo custo R$112.",
      "Impacta diretamente break-even, margem e markup simulados.",
    ],
  },
  "margem-simulada-pct": {
    title: "Margem Simulada (%)",
    items: [
      "Percentual de contribuicao apos custos no cenario.",
      "Valores baixos sinalizam pressao de rentabilidade.",
    ],
  },
  "markup-simulado-pct": {
    title: "Markup Simulado (%)",
    items: [
      "Relacao percentual entre preco e custo ajustado.",
      "Ajuda a validar consistencia da formacao de preco.",
    ],
  },
  "contribuicao-unidade": {
    title: "Contribuicao por Unidade",
    items: [
      "Valor em reais que sobra por unidade no cenario.",
      "Base para sustentar despesas e resultado.",
    ],
  },
  "novo-custo": {
    title: "Novo custo",
    items: [
      "Custo medio ajustado por Fretes/Encargos no cenario.",
      "Formula: novo custo = custo medio x (1 + fretes/encargos %).",
    ],
  },
  "gap-mercado-pct": {
    title: "Gap vs Mercado (%)",
    items: [
      "Diferenca percentual entre preco simulado e mercado medio.",
      "Negativo indica preco abaixo do mercado.",
    ],
  },
  "novo-preco": {
    title: "Novo Preco",
    items: [
      "Preco final resultante do cenario atual.",
      "Consolida efeitos de desconto, margem e custos.",
    ],
  },
  "posicao-no-range": {
    title: "Posicao no Range",
    items: [
      "Mostra onde o preco esta entre min/med/max de mercado.",
      "Atual e Simulado permitem comparar deslocamento de posicao.",
    ],
  },
};
