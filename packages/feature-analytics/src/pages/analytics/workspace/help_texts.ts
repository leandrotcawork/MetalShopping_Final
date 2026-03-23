export type HelpCopy = {
  title: string;
  items: string[];
};

export const ANALYTICS_HELP: Record<string, HelpCopy> = {
  "Margem contrib. (%)": {
    title: "Margem de contribuicao (%)",
    items: [
      "Percentual de contribuicao unitaria apos custos variaveis.",
      "Quanto maior, maior espaco para desconto e promocao com seguranca.",
    ],
  },
  Markup: {
    title: "Markup",
    items: [
      "Relacao entre preco e custo do item.",
      "Ajuda a medir folga de precificacao e protecao de margem.",
    ],
  },
  "GMROI (6M)": {
    title: "GMROI (6M)",
    items: [
      "Formula: GMROI = margem_total_6m / investimento_medio_estoque.",
      "Investimento medio = estoque_medio_un * custo unitario medio (janela).",
      "Exemplo: GMROI 1.4 => cada R$1 no estoque gerou R$1,40 de margem bruta.",
      "Regra pratica: acima de 1 tende a indicar retorno saudavel.",
    ],
  },
  "COGS (6M)": {
    title: "COGS (6M)",
    items: [
      "Custo total das mercadorias vendidas nos ultimos 6 meses.",
      "Base para leitura de rentabilidade e eficiencia operacional.",
    ],
  },
  "PME (dias)": {
    title: "PME (dias)",
    items: [
      "Formula: PME = estoque_medio_un / demanda_diaria.",
      "Mostra por quantos dias, em media, o item fica em estoque.",
      "Exemplo: PME 30d => o item permanece cerca de 30 dias estocado.",
      "Menor PME costuma indicar operacao mais eficiente.",
    ],
  },
  "Giro (6M)": {
    title: "Giro (6M)",
    items: [
      "Formula: Giro_6m = unidades_vendidas_6m / estoque_medio_un.",
      "Indica quantas vezes o estoque medio girou no periodo.",
      "Exemplo: Giro 2.0 => vendeu o equivalente a 2 estoques medios em 6 meses.",
      "Maior giro tende a indicar melhor tracao de vendas.",
    ],
  },
  "DOS (dias)": {
    title: "DOS (dias)",
    items: [
      "Formula: DOS = estoque_atual_un / demanda_diaria.",
      "Representa por quantos dias o estoque atual cobre a venda no ritmo atual.",
      "Exemplo: DOS 84d => mantendo o ritmo, o estoque cobre cerca de 84 dias.",
      "Quanto maior, maior risco de sobreestoque.",
    ],
  },
  "Pressao de reposicao": {
    title: "Pressao de reposicao",
    items: [
      "Indicador de pressao para reposicao do item no contexto atual.",
      "Valores mais altos sinalizam reposicao mais pressionada.",
    ],
  },
};
