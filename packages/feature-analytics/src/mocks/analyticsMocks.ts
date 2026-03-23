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
  const metric = (
    value: number,
    deltaMomPct: number | null,
    windowMonths = 6,
    flags?: { is_imputed?: boolean },
  ) => ({
    value,
    trend: {
      delta_mom_pct: deltaMomPct,
      is_available: deltaMomPct != null,
      window_months: windowMonths,
    },
    flags: flags || undefined,
  });

  const trend = (revenueDelta: number | null, marginDeltaPp: number | null, shareDeltaPp: number | null) => ({
    source: "SNAPSHOT",
    basis: "MoM",
    window_months: 1,
    target_month_ref: "2026-03",
    base_month_ref: "2026-02",
    revenue_delta_mom_pct: revenueDelta,
    margin_delta_mom_pp: marginDeltaPp,
    share_delta_mom_pp: shareDeltaPp,
    is_available: revenueDelta != null || marginDeltaPp != null || shareDeltaPp != null,
  });

  const topNodesByRevenue = [
    { node_id: 101, node_name: "Revestimentos", brand: "Eliane", revenue_brl: 245000, share_pct: 18.2, margin_pct: 26.4, trend: trend(3.2, 0.4, 0.3) },
    { node_id: 102, node_name: "Porcelanatos", brand: "Portobello", revenue_brl: 210000, share_pct: 15.6, margin_pct: 24.1, trend: trend(1.8, -0.3, 0.1) },
    { node_id: 103, node_name: "Metais", brand: "Deca", revenue_brl: 175000, share_pct: 13.0, margin_pct: 22.7, trend: trend(-0.9, 0.2, -0.2) },
    { node_id: 104, node_name: "Louças", brand: "Incepa", revenue_brl: 96000, share_pct: 7.1, margin_pct: 20.3, trend: trend(2.4, 0.1, 0.0) },
    { node_id: 105, node_name: "Ferragens", brand: "Vonder", revenue_brl: 71000, share_pct: 5.3, margin_pct: 18.8, trend: trend(4.9, -0.4, 0.2) },
    { node_id: 106, node_name: "Hidráulica", brand: "Tigre", revenue_brl: 88000, share_pct: 6.6, margin_pct: 21.2, trend: trend(0.7, 0.0, -0.1) },
    { node_id: 107, node_name: "Tintas", brand: "Suvinil", revenue_brl: 98000, share_pct: 7.3, margin_pct: 27.9, trend: trend(1.2, 0.6, 0.1) },
    { node_id: 108, node_name: "Elétrica", brand: "Tramontina", revenue_brl: 64000, share_pct: 4.8, margin_pct: 19.6, trend: trend(-1.4, -0.2, -0.1) },
    { node_id: 109, node_name: "Argamassas", brand: "Quartzolit", revenue_brl: 54000, share_pct: 4.0, margin_pct: 16.9, trend: trend(0.3, 0.1, 0.0) },
    { node_id: 110, node_name: "Coberturas", brand: "Brasilit", revenue_brl: 52000, share_pct: 3.9, margin_pct: 14.6, trend: trend(2.1, -0.6, 0.0) },
    { node_id: 111, node_name: "Pisos", brand: "Delta", revenue_brl: 48000, share_pct: 3.6, margin_pct: 17.4, trend: trend(0.9, 0.2, 0.0) },
    { node_id: 112, node_name: "Madeiras", brand: "Duratex", revenue_brl: 41000, share_pct: 3.0, margin_pct: 23.1, trend: trend(-0.2, 0.3, 0.0) },
  ];

  const capitalAllocationBase = [
    { node_id: 101, node_name: "Revestimentos", sku_count: 320, capital_brl: 120000, risk_level: "medium", risk_pct: 32 },
    { node_id: 102, node_name: "Porcelanatos", sku_count: 260, capital_brl: 98000, risk_level: "high", risk_pct: 54 },
    { node_id: 103, node_name: "Metais", sku_count: 210, capital_brl: 86000, risk_level: "medium", risk_pct: 18 },
    { node_id: 104, node_name: "Louças", sku_count: 140, capital_brl: 54000, risk_level: "low", risk_pct: 11 },
    { node_id: 105, node_name: "Ferragens", sku_count: 180, capital_brl: 42000, risk_level: "medium", risk_pct: 27 },
    { node_id: 106, node_name: "Hidráulica", sku_count: 190, capital_brl: 48000, risk_level: "low", risk_pct: 14 },
    { node_id: 107, node_name: "Tintas", sku_count: 175, capital_brl: 51000, risk_level: "medium", risk_pct: 21 },
    { node_id: 108, node_name: "Elétrica", sku_count: 155, capital_brl: 39000, risk_level: "low", risk_pct: 9 },
    { node_id: 109, node_name: "Argamassas", sku_count: 130, capital_brl: 28000, risk_level: "medium", risk_pct: 25 },
    { node_id: 110, node_name: "Coberturas", sku_count: 90, capital_brl: 22000, risk_level: "high", risk_pct: 48 },
  ];

  const capitalEfficiency = [
    { node_id: 101, node_name: "Revestimentos", capital_brl: 120000, risk_pct: 32, gmroi: 1.1, revenue_brl: 245000, gross_margin_brl: 60300 },
    { node_id: 102, node_name: "Porcelanatos", capital_brl: 98000, risk_pct: 54, gmroi: 0.9, revenue_brl: 210000, gross_margin_brl: 50600 },
    { node_id: 103, node_name: "Metais", capital_brl: 86000, risk_pct: 18, gmroi: 1.4, revenue_brl: 175000, gross_margin_brl: 43000 },
    { node_id: 104, node_name: "Louças", capital_brl: 54000, risk_pct: 11, gmroi: 1.3, revenue_brl: 96000, gross_margin_brl: 19500 },
    { node_id: 105, node_name: "Ferragens", capital_brl: 42000, risk_pct: 27, gmroi: 1.0, revenue_brl: 71000, gross_margin_brl: 13300 },
    { node_id: 106, node_name: "Hidráulica", capital_brl: 48000, risk_pct: 14, gmroi: 1.2, revenue_brl: 88000, gross_margin_brl: 18700 },
    { node_id: 107, node_name: "Tintas", capital_brl: 51000, risk_pct: 21, gmroi: 1.6, revenue_brl: 98000, gross_margin_brl: 27300 },
    { node_id: 108, node_name: "Elétrica", capital_brl: 39000, risk_pct: 9, gmroi: 1.05, revenue_brl: 64000, gross_margin_brl: 12500 },
    { node_id: 109, node_name: "Argamassas", capital_brl: 28000, risk_pct: 25, gmroi: 0.95, revenue_brl: 54000, gross_margin_brl: 9100 },
    { node_id: 110, node_name: "Coberturas", capital_brl: 22000, risk_pct: 48, gmroi: 0.8, revenue_brl: 52000, gross_margin_brl: 7600 },
  ];

  const topMargin = [
    { node_id: 107, node_name: "Tintas", revenue_brl: 98000, margin_pct: 27.9, gross_margin_brl: 27300, capital_brl: 51000, gmroi: 1.6, trend: trend(1.2, 0.6, 0.1) },
    { node_id: 101, node_name: "Revestimentos", revenue_brl: 245000, margin_pct: 26.4, gross_margin_brl: 60300, capital_brl: 120000, gmroi: 1.1, trend: trend(3.2, 0.4, 0.3) },
    { node_id: 112, node_name: "Madeiras", revenue_brl: 41000, margin_pct: 23.1, gross_margin_brl: 9400, capital_brl: 19000, gmroi: 1.2, trend: trend(-0.2, 0.3, 0.0) },
    { node_id: 103, node_name: "Metais", revenue_brl: 175000, margin_pct: 22.7, gross_margin_brl: 43000, capital_brl: 86000, gmroi: 1.4, trend: trend(-0.9, 0.2, -0.2) },
    { node_id: 104, node_name: "Louças", revenue_brl: 96000, margin_pct: 20.3, gross_margin_brl: 19500, capital_brl: 54000, gmroi: 1.3, trend: trend(2.4, 0.1, 0.0) },
    { node_id: 108, node_name: "Elétrica", revenue_brl: 64000, margin_pct: 19.6, gross_margin_brl: 12500, capital_brl: 39000, gmroi: 1.05, trend: trend(-1.4, -0.2, -0.1) },
    { node_id: 105, node_name: "Ferragens", revenue_brl: 71000, margin_pct: 18.8, gross_margin_brl: 13300, capital_brl: 42000, gmroi: 1.0, trend: trend(4.9, -0.4, 0.2) },
    { node_id: 111, node_name: "Pisos", revenue_brl: 48000, margin_pct: 17.4, gross_margin_brl: 8300, capital_brl: 21000, gmroi: 1.1, trend: trend(0.9, 0.2, 0.0) },
    { node_id: 109, node_name: "Argamassas", revenue_brl: 54000, margin_pct: 16.9, gross_margin_brl: 9100, capital_brl: 28000, gmroi: 0.95, trend: trend(0.3, 0.1, 0.0) },
    { node_id: 110, node_name: "Coberturas", revenue_brl: 52000, margin_pct: 14.6, gross_margin_brl: 7600, capital_brl: 22000, gmroi: 0.8, trend: trend(2.1, -0.6, 0.0) },
  ];

  const nodesAtRisk = [
    { node_id: 102, node_name: "Porcelanatos", risk_pct: 54, risk_level: "high", capital_at_risk_brl: 31000, capital_brl: 98000, revenue_brl: 210000, margin_pct: 24.1, gmroi: 0.9, financial_risk_priority_brl: 75000 },
    { node_id: 110, node_name: "Coberturas", risk_pct: 48, risk_level: "high", capital_at_risk_brl: 12000, capital_brl: 22000, revenue_brl: 52000, margin_pct: 14.6, gmroi: 0.8, financial_risk_priority_brl: 32000 },
    { node_id: 101, node_name: "Revestimentos", risk_pct: 32, risk_level: "medium", capital_at_risk_brl: 18000, capital_brl: 120000, revenue_brl: 245000, margin_pct: 26.4, gmroi: 1.1, financial_risk_priority_brl: 61000 },
    { node_id: 105, node_name: "Ferragens", risk_pct: 27, risk_level: "medium", capital_at_risk_brl: 9800, capital_brl: 42000, revenue_brl: 71000, margin_pct: 18.8, gmroi: 1.0, financial_risk_priority_brl: 23000 },
    { node_id: 109, node_name: "Argamassas", risk_pct: 25, risk_level: "medium", capital_at_risk_brl: 6400, capital_brl: 28000, revenue_brl: 54000, margin_pct: 16.9, gmroi: 0.95, financial_risk_priority_brl: 15500 },
    { node_id: 107, node_name: "Tintas", risk_pct: 21, risk_level: "medium", capital_at_risk_brl: 7800, capital_brl: 51000, revenue_brl: 98000, margin_pct: 27.9, gmroi: 1.6, financial_risk_priority_brl: 21000 },
    { node_id: 103, node_name: "Metais", risk_pct: 18, risk_level: "low", capital_at_risk_brl: 5400, capital_brl: 86000, revenue_brl: 175000, margin_pct: 22.7, gmroi: 1.4, financial_risk_priority_brl: 12000 },
    { node_id: 106, node_name: "Hidráulica", risk_pct: 14, risk_level: "low", capital_at_risk_brl: 4200, capital_brl: 48000, revenue_brl: 88000, margin_pct: 21.2, gmroi: 1.2, financial_risk_priority_brl: 9800 },
    { node_id: 104, node_name: "Louças", risk_pct: 11, risk_level: "low", capital_at_risk_brl: 2100, capital_brl: 54000, revenue_brl: 96000, margin_pct: 20.3, gmroi: 1.3, financial_risk_priority_brl: 7600 },
    { node_id: 108, node_name: "Elétrica", risk_pct: 9, risk_level: "low", capital_at_risk_brl: 1700, capital_brl: 39000, revenue_brl: 64000, margin_pct: 19.6, gmroi: 1.05, financial_risk_priority_brl: 5200 },
  ];

  return {
    message: "",
    kpis: {
      active_entities: metric(72, 1.4, 6),
      gross_revenue_6m_brl: metric(452000, 2.1, 6),
      margin_total_brl: metric(122300, 1.2, 6),
      margin_pct: metric(27.1, 0.4, 6),
      capital_total_brl: metric(144200, -0.9, 6, { is_imputed: false }),
      capital_at_risk_brl: metric(28600, 3.8, 6),
      potential_revenue_internal_brl: metric(158000, 0.7, 6),
      potential_revenue_market_brl: metric(512000, 1.1, 6),
    },
    scope: {
      level_label: "Grupo",
      trend_window_months: 6,
      window: { window_months: 6 },
      margin_policy: { low_pct: 12, good_pct: 25 },
      filter_options: {
        marcas: ["Portobello", "Deca", "Suvinil", "Tigre", "Tramontina", "Eliane", "Quartzolit"],
      },
      analysis_cards: {
        abc_mix: {
          a_max_cum_pct: 80,
          b_max_cum_pct: 95,
          a_count: 4,
          b_count: 12,
          c_count: 56,
          a_revenue_pct: 62.4,
          b_revenue_pct: 27.1,
          c_revenue_pct: 10.5,
        },
        gmroi_global: 1.18,
        capital_travado_brl: 80400,
        risco_global_pct: 22.7,
        margem_contrib_real_pct: 14.3,
      },
    },
    panels: {
      top_nodes_by_revenue: topNodesByRevenue,
      revenue_concentration: {
        top3_pct: 47.1,
        top5_pct: 62.2,
        top10_pct: 84.6,
        top3_mom_delta_pct: 0.8,
        top5_mom_delta_pct: 0.4,
        top10_mom_delta_pct: -0.1,
        risk_level: "medium",
        risk_reason: "Concentracao moderada em poucos grupos, mas com tendencia controlada.",
        is_trend_available: true,
        trend_source: "SNAPSHOT",
        trend_basis: "MoM",
        trend_window_months: 1,
        target_month_ref: "2026-03",
        base_month_ref: "2026-02",
      },
      capital_allocation_map: capitalAllocationBase.map((row) => ({
        ...row,
        size_weight: row.capital_brl,
      })),
      capital_allocation_map_spotlight: capitalAllocationBase.map((row) => ({
        ...row,
      })),
      capital_efficiency: capitalEfficiency,
      rankings: {
        top_margin: topMargin,
        nodes_at_risk: nodesAtRisk,
      },
      backlog: [
        { priority: "P0", title: "Reduzir imobilização", text: "Priorizar grupos com capital parado e risco alto.", cta: "Abrir spotlight", tone: "red" },
        { priority: "P0", title: "Ajustar preço", text: "Rever bandas de preço com perda de competitividade.", cta: "Ver ações", tone: "primary" },
        { priority: "P1", title: "Rever compras", text: "Ajustar cobertura para evitar excesso em baixa tração.", cta: "Abrir fila", tone: "neutral" },
        { priority: "P1", title: "Otimizar mix", text: "Melhorar pareto ABC e elevar GMROI por categoria.", cta: "Como ler", tone: "green" },
        { priority: "P2", title: "Higienizar dados", text: "Corrigir marca/classificação nos itens críticos.", cta: "Ver dados", tone: "neutral" },
      ],
    },
  };
}
