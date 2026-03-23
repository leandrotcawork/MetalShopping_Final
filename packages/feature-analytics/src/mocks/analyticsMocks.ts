import type { AnalyticsHomeV2Dto, AnalyticsTaxonomyScopeOverviewV1Dto } from "../legacy_dto";

export function buildMockAnalyticsHomeDto(): AnalyticsHomeV2Dto {
  return {
    schemaVersion: "analytics_home_v2",
    snapshot: {
      resolved_id: "snap-legacy-home",
      as_of: "2026-03-15T21:00:00Z",
      servedAt: "2026-03-22T15:00:00Z",
    },
    blocks: {
      actions_today: {
        status: "OK",
        data: {
          buckets: [
            {
              action_code: "LIBERAR_CAPITAL",
              label: "Liberar capital",
              headline: "Liberar capital imobilizado",
              desc: "Objetivo: Alivio de capital",
              count: 4,
              top_drivers: ["Cobertura acima da meta", "Baixa tracao recente"],
              top_skus: [
                {
                  pn: "18578",
                  descricao: "Porcelanato 90x90 Delta",
                  marca: "Portobello",
                  taxonomy_leaf_name: "Porcelanato",
                  stock_value_brl: 48200,
                  stock_qty: 122,
                  urgency_score: 91,
                  urgency_label: "HIGH",
                },
              ],
            },
            {
              action_code: "REDUZIR_IMOBILIZACAO",
              label: "Reduzir imobilizacao",
              headline: "Reduzir capital parado",
              desc: "Objetivo: Alivio de capital",
              count: 15,
              top_drivers: ["Estoque sem giro", "Baixa cobertura de demanda"],
              top_skus: [
                {
                  pn: "90331",
                  descricao: "Torneira gourmet inox 70cm",
                  marca: "Deca",
                  taxonomy_leaf_name: "Metais",
                  stock_value_brl: 31500,
                  stock_qty: 87,
                  urgency_score: 82,
                  urgency_label: "HIGH",
                },
              ],
            },
            {
              action_code: "AJUSTAR_PRECO_BAIXAR",
              label: "Ajustar preco (baixar)",
              headline: "Ajustar preco para ganhar competitividade",
              desc: "Objetivo: Sem objetivo dominante",
              count: 5,
              top_drivers: ["Preco acima da banda", "Perda de sell-out"],
              top_skus: [
                {
                  pn: "11002",
                  descricao: "Argamassa ACIII premium 20kg",
                  marca: "Quartzolit",
                  taxonomy_leaf_name: "Argamassas",
                  stock_value_brl: 12200,
                  stock_qty: 410,
                  urgency_score: 74,
                  urgency_label: "MEDIUM",
                },
              ],
            },
            {
              action_code: "REVER_COMPRAS",
              label: "Rever compras",
              headline: "Revisar compras e cobertura",
              desc: "Objetivo: Alivio de capital",
              count: 6,
              top_drivers: ["Cobertura desalinhada", "Mix de reposicao"],
              top_skus: [
                {
                  pn: "77210",
                  descricao: "Louca sanitaria compacta",
                  marca: "Incepa",
                  taxonomy_leaf_name: "Loucas",
                  stock_value_brl: 9600,
                  stock_qty: 55,
                  urgency_score: 63,
                  urgency_label: "MEDIUM",
                },
              ],
            },
          ],
          items: [
            {
              action_code: "LIBERAR_CAPITAL",
              pn: "18578",
              descricao: "Porcelanato 90x90 Delta",
              marca: "Portobello",
              taxonomy_leaf_name: "Porcelanato",
              stock_value_brl: 48200,
              stock_qty: 122,
              urgency_score: 91,
              urgency_label: "HIGH",
              value_arbitration: {
                decision_state: "REVIEW",
                value_priority_tier: "P0",
                dominant_objective: "CAPITAL_RELIEF",
              },
              top_pns: ["18578", "90331"],
              top_taxonomy_leafs: ["Porcelanato"],
              top_brands: ["Portobello"],
            },
            {
              action_code: "REDUZIR_IMOBILIZACAO",
              pn: "90331",
              descricao: "Torneira gourmet inox 70cm",
              marca: "Deca",
              taxonomy_leaf_name: "Metais",
              stock_value_brl: 31500,
              stock_qty: 87,
              urgency_score: 82,
              urgency_label: "HIGH",
              value_arbitration: {
                decision_state: "REVIEW",
                value_priority_tier: "P0",
                dominant_objective: "CAPITAL_RELIEF",
              },
            },
            {
              action_code: "AJUSTAR_PRECO_BAIXAR",
              pn: "11002",
              descricao: "Argamassa ACIII premium 20kg",
              marca: "Quartzolit",
              taxonomy_leaf_name: "Argamassas",
              stock_value_brl: 12200,
              stock_qty: 410,
              urgency_score: 74,
              urgency_label: "MEDIUM",
              value_arbitration: {
                decision_state: "BLOCKED",
                value_priority_tier: "P0",
                dominant_objective: "NONE",
              },
            },
            {
              action_code: "REVER_COMPRAS",
              pn: "77210",
              descricao: "Louca sanitaria compacta",
              marca: "Incepa",
              taxonomy_leaf_name: "Loucas",
              stock_value_brl: 9600,
              stock_qty: 55,
              urgency_score: 63,
              urgency_label: "MEDIUM",
              value_arbitration: {
                decision_state: "REVIEW",
                value_priority_tier: "P1",
                dominant_objective: "CAPITAL_RELIEF",
              },
            },
          ],
        },
      },
      alerts_prioritarios: {
        status: "OK",
        data: {
          buckets: [
            {
              code: "price_above_market",
              label: "Preco acima do mercado",
              headline: "Sinais de perda de competitividade",
              count: 120,
              top_pns: ["18578", "90331", "11002"],
            },
            {
              code: "capital_imobilizado",
              label: "Capital imobilizado",
              headline: "Capital parado acima da banda",
              count: 42,
              top_pns: ["44109", "33110"],
            },
          ],
          top_skus: {
            price_above_market: [
              { pn: "18578", descricao: "Porcelanato 90x90 Delta", marca: "Portobello", taxonomy_leaf_name: "Porcelanato", stock_value_brl: 48200, stock_qty: 122 },
              { pn: "90331", descricao: "Torneira gourmet inox 70cm", marca: "Deca", taxonomy_leaf_name: "Metais", stock_value_brl: 31500, stock_qty: 87 },
            ],
            capital_imobilizado: [
              { pn: "44109", descricao: "Revestimento metro white", marca: "Eliane", taxonomy_leaf_name: "Revestimentos", stock_value_brl: 21800, stock_qty: 142 },
              { pn: "33110", descricao: "Assento sanitario soft close", marca: "Astra", taxonomy_leaf_name: "Loucas", stock_value_brl: 7600, stock_qty: 66 },
            ],
          },
        },
      },
      health_radar: {
        status: "OK",
        data: {
          impact_levels: ["I1", "I2", "I3", "I4", "I5"],
          urgency_levels: ["U1", "U2", "U3", "U4", "U5"],
          cells: [
            { impact: "I1", urgency: "U1", count: 12, top_drivers: ["Preco acima da banda"], pns: ["18578", "90331"] },
            { impact: "I1", urgency: "U2", count: 16, top_drivers: ["Margem apertada"], pns: ["11002", "44109"] },
            { impact: "I1", urgency: "U3", count: 10, top_drivers: ["Capital imobilizado"], pns: ["33110"] },
            { impact: "I2", urgency: "U1", count: 4, pns: ["77210"] },
            { impact: "I2", urgency: "U2", count: 5, pns: ["90331"] },
            { impact: "I3", urgency: "U2", count: 3, pns: ["18578"] },
          ],
        },
      },
      portfolio_distribution: {
        status: "OK",
        data: {
          buckets: [
            { key: "EXPAND", label: "Expandir", count_skus: 134, share_skus: 0.28, top_drivers: ["Demanda ascendente"], pns: ["18578", "90331"] },
            { key: "MONITOR", label: "Monitorar", count_skus: 219, share_skus: 0.46, top_drivers: ["Sazonalidade"], pns: ["11002", "77210"] },
            { key: "PRUNE", label: "Rever sortimento", count_skus: 121, share_skus: 0.26, top_drivers: ["Baixo giro"], pns: ["44109", "33110"] },
          ],
        },
      },
      top_metal: {
        status: "OK",
        data: {
          best_sku: { pn: "18578", descricao: "Porcelanato 90x90 Delta", lucro_mes: 48200, receita_liq_mes: 185000 },
          best_brand: { brand: "Portobello", lucro_mes: 121300, receita_liq_mes: 402000 },
          best_taxonomy: { taxonomy_leaf_name: "Revestimentos", lucro_mes: 176200, receita_liq_mes: 598000 },
        },
      },
      timeline: {
        status: "OK",
        data: {
          rows: [
            { date: "2026-03-20", title: "Snapshot atualizado", summary: "Reprocessamento concluido" },
            { date: "2026-03-21", title: "Regras recalculadas", summary: "Priorizacao revisada" },
            { date: "2026-03-22", title: "Painel publicado", summary: "Visual legacy ativo" },
          ],
        },
      },
      kpis_operational: {
        status: "OK",
        data: {
          priced_skus_count: 2380,
          with_seller_count: 1980,
          scraping_success_rate: 0.81,
          alerts_count: 120,
        },
      },
      kpis_products: {
        status: "OK",
        data: {
          products_active_count: 3838,
          capital_brl_total: 144200,
          potential_revenue_brl_total_market: 512000,
          weighted_margin_pct_total: 27.1,
        },
      },
      kpis_series: {
        status: "OK",
        data: {
          runs_7d: [
            { date: "2026-03-16", skus_processed: 120 },
            { date: "2026-03-17", skus_processed: 98 },
            { date: "2026-03-18", skus_processed: 156 },
            { date: "2026-03-19", skus_processed: 142 },
            { date: "2026-03-20", skus_processed: 131 },
            { date: "2026-03-21", skus_processed: 165 },
            { date: "2026-03-22", skus_processed: 150 },
          ],
          sales_6m: [
            { month: "2025-10", receita_liq: 360000 },
            { month: "2025-11", receita_liq: 372000 },
            { month: "2025-12", receita_liq: 401000 },
            { month: "2026-01", receita_liq: 415000 },
            { month: "2026-02", receita_liq: 438000 },
            { month: "2026-03", receita_liq: 452000 },
          ],
          margin_6m: [
            { month: "2025-10", margem_pct: 24.0 },
            { month: "2025-11", margem_pct: 25.0 },
            { month: "2025-12", margem_pct: 25.1 },
            { month: "2026-01", margem_pct: 25.8 },
            { month: "2026-02", margem_pct: 26.6 },
            { month: "2026-03", margem_pct: 27.1 },
          ],
          price_index_series: [
            { date: "2026-03-16", value: 0.88 },
            { date: "2026-03-17", value: 0.9 },
            { date: "2026-03-18", value: 0.92 },
            { date: "2026-03-19", value: 0.91 },
            { date: "2026-03-20", value: 0.93 },
            { date: "2026-03-21", value: 0.95 },
            { date: "2026-03-22", value: 0.96 },
          ],
          margin_series: [
            { month: "2025-10", margem_pct: 0.24 },
            { month: "2025-11", margem_pct: 0.25 },
            { month: "2025-12", margem_pct: 0.251 },
            { month: "2026-01", margem_pct: 0.258 },
            { month: "2026-02", margem_pct: 0.266 },
            { month: "2026-03", margem_pct: 0.271 },
          ],
          revenue_series: [
            { month: "2025-10", receita_brl: 360000 },
            { month: "2025-11", receita_brl: 372000 },
            { month: "2025-12", receita_brl: 401000 },
            { month: "2026-01", receita_brl: 415000 },
            { month: "2026-02", receita_brl: 438000 },
            { month: "2026-03", receita_brl: 452000 },
          ],
        },
      },
    },
  } as AnalyticsHomeV2Dto;
}

export function buildMockTaxonomyScopeOverview(): AnalyticsTaxonomyScopeOverviewV1Dto {
  return {
    scope: {
      filter_options: {
        marcas: ["Portobello", "Deca", "Suvinil", "Tigre"],
      },
      analysis_cards: {
        abc_mix: {
          a_max_cum_pct: 80,
          b_max_cum_pct: 95,
        },
      },
    },
    panels: {
      capital_efficiency: [
        { node_id: 1, node_name: "Revestimentos", capital_brl: 120000, risk_pct: 32, gmroi: 1.1, revenue_brl: 245000, gross_margin_brl: 60300 },
        { node_id: 2, node_name: "Metais", capital_brl: 86000, risk_pct: 18, gmroi: 1.4, revenue_brl: 175000, gross_margin_brl: 43000 },
      ],
      top_nodes_by_revenue: [
        { node_id: 10, node_name: "Porcelanato", revenue_brl: 210000, margin_pct: 24.1 },
        { node_id: 11, node_name: "Loucas", revenue_brl: 96000, margin_pct: 20.3 },
        { node_id: 12, node_name: "Ferragens", revenue_brl: 71000, margin_pct: 18.8 },
      ],
    },
  };
}
