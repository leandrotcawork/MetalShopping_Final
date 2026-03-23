// @ts-nocheck
import type { AnalyticsProductsOverviewV1Dto } from "../../../../../legacy_products_dto";
import copy from "../../../copy/products_overview_narratives.pt_BR.json";

type TrendTone = "positive" | "negative";

export type ProductsOverviewViewModel = {
  header: {
    title: string;
    subtitle: string;
    scopeLabel: string;
    qualityLabel: string;
  };
  kpis: Array<{
    label: string;
    value: string;
    change: string;
    tone: "positive" | "negative";
    helpItems: string[];
    showChangeArrow?: boolean;
  }>;
  matrix: {
    stars: number;
    potential: number;
    attention: number;
    critical: number;
  };
  matrixSpotlight: {
    display: Record<"stars" | "potential" | "attention" | "critical", string[]>;
    stars: MatrixSpotlightRowVm[];
    potential: MatrixSpotlightRowVm[];
    attention: MatrixSpotlightRowVm[];
    critical: MatrixSpotlightRowVm[];
  };
  abc: Array<{
    letter: "A" | "B" | "C";
    countLabel: string;
    shareLabel: string;
    fillPct: number;
    valueLabel: string;
  }>;
  rankings: {
    topMargin: Array<{ pn: string; product: string; metric: string; value: string }>;
    topGiro: Array<{ pn: string; product: string; metric: string; value: string }>;
    worstGiro: Array<{ pn: string; product: string; metric: string; value: string }>;
  };
  trends: Array<{
    label: string;
    change: string;
    tone: TrendTone;
    points: number[];
  }>;
};

type MatrixSpotlightRowVm = {
  pn: string;
  product: string;
  brand: string;
  taxonomyLeafName: string;
  stockValueBrl: number;
  stockQty: number;
  financialPriorityScore: number;
  financialPriorityTier: string;
  marginPct: number;
  marginSalesPct: number;
  marginUnitPct: number;
  giro6m: number;
  sales6mUnits: number;
  sales1mUnits: number;
  daysNoSales: number;
  dos: number;
  capitalBrl: number;
  contributionBrl: number;
  gapVsMarketPct: number;
};

function formatInteger(value: number): string {
  return new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 0 }).format(value);
}

function formatCompactBRL(value: number): string {
  const abs = Math.abs(value);
  if (abs >= 1_000_000) return `R$ ${(value / 1_000_000).toLocaleString("pt-BR", { maximumFractionDigits: 1 })}M`;
  if (abs >= 1_000) return `R$ ${(value / 1_000).toLocaleString("pt-BR", { maximumFractionDigits: 1 })}K`;
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL", maximumFractionDigits: 0 }).format(value);
}

function formatPercent(value: number, digits = 1): string {
  return `${value.toLocaleString("pt-BR", { minimumFractionDigits: digits, maximumFractionDigits: digits })}%`;
}

function mapSpotlightRows(
  rows: Array<{
    pn: string;
    product: string;
    brand?: string;
    taxonomy_leaf_name?: string;
    stock_value_brl?: number;
    stock_qty?: number;
    financial_priority_score?: number;
    financial_priority_tier?: string;
    margin_pct: number;
    margin_sales_pct?: number;
    margin_unit_pct?: number;
    sales_6m_units?: number;
    sales_1m_units?: number;
    gap_vs_market_pct?: number;
    giro_6m: number;
    days_no_sales: number;
    dos: number;
    capital_brl: number;
    contribution_brl: number;
  }> | undefined,
) {
  return (rows || []).map((row) => ({
    pn: row.pn || "-",
    product: row.product || "Sem descricao",
    brand: String(row.brand || "").trim() || "Sem marca",
    taxonomyLeafName: String(row.taxonomy_leaf_name || "").trim() || "Sem classificacao",
    stockValueBrl: Number.isFinite(row.stock_value_brl) ? Number(row.stock_value_brl) : (Number.isFinite(row.capital_brl) ? Number(row.capital_brl) : 0),
    stockQty: Number.isFinite(row.stock_qty) ? Number(row.stock_qty) : 0,
    financialPriorityScore: Number.isFinite(row.financial_priority_score) ? Number(row.financial_priority_score) : (Number.isFinite(row.stock_value_brl) ? Number(row.stock_value_brl) : (Number.isFinite(row.capital_brl) ? Number(row.capital_brl) : 0)),
    financialPriorityTier: String(row.financial_priority_tier || "").trim(),
    marginPct: Number.isFinite(row.margin_pct) ? row.margin_pct : 0,
    marginSalesPct: Number.isFinite(row.margin_sales_pct) ? Number(row.margin_sales_pct) : (Number.isFinite(row.margin_pct) ? row.margin_pct : 0),
    marginUnitPct: Number.isFinite(row.margin_unit_pct) ? Number(row.margin_unit_pct) : 0,
    sales6mUnits: Number.isFinite(row.sales_6m_units) ? Number(row.sales_6m_units) : 0,
    sales1mUnits: Number.isFinite(row.sales_1m_units) ? Number(row.sales_1m_units) : 0,
    gapVsMarketPct: Number.isFinite(row.gap_vs_market_pct) ? Number(row.gap_vs_market_pct) : 0,
    giro6m: Number.isFinite(row.giro_6m) ? row.giro_6m : 0,
    daysNoSales: Number.isFinite(row.days_no_sales) ? row.days_no_sales : 0,
    dos: Number.isFinite(row.dos) ? row.dos : 0,
    capitalBrl: Number.isFinite(row.capital_brl) ? row.capital_brl : 0,
    contributionBrl: Number.isFinite(row.contribution_brl) ? row.contribution_brl : 0,
  }));
}

function formatCoverage(dto: AnalyticsProductsOverviewV1Dto): string {
  const filters: string[] = [];
  if (dto.scope.filters_applied.marca) filters.push(`Marca: ${dto.scope.filters_applied.marca}`);
  if (dto.scope.filters_applied.taxonomy_leaf_name) {
    filters.push(`Classificacao: ${dto.scope.filters_applied.taxonomy_leaf_name}`);
  }
  if (dto.scope.filters_applied.status) filters.push(`Status: ${dto.scope.filters_applied.status}`);
  if (dto.scope.filters_applied.search) filters.push(`Busca: "${dto.scope.filters_applied.search}"`);
  const filterLabel = filters.length ? ` | ${filters.join(" | ")}` : "";
  const coverage = dto.scope.coverage.is_partial ? "amostra parcial" : "cobertura completa";
  return `${coverage} (${formatInteger(dto.scope.coverage.rows_considered)}/${formatInteger(dto.scope.coverage.rows_total_available)} SKUs)${filterLabel}`;
}

function toneByDelta(value: number | null): "positive" | "negative" {
  return (value ?? 0) >= 0 ? "positive" : "negative";
}

function resolveKpiHelpItems(
  key:
  | "portfolio_ativo"
  | "capital_imobilizado"
  | "capital_em_risco"
  | "receita_potencial_interna"
  | "receita_potencial_mercado"
  | "margem_ponderada",
): string[] {
  const source = (copy as {
    kpi_help?: Record<string, { items?: unknown }>;
  }).kpi_help;
  const bucket = source?.[key];
  const items = Array.isArray(bucket?.items) ? bucket.items : [];
  return items.map((item) => String(item || "").trim()).filter(Boolean);
}

export function mapProductsOverviewViewModel(dto: AnalyticsProductsOverviewV1Dto): ProductsOverviewViewModel {
  const trendDelta = dto.kpis.trend_pct.value;
  const stockDelta = dto.trends_90d.stock.delta_pct;
  const capitalMetric = dto.kpis.capital_total_brl;
  const capitalRiskMetric = dto.kpis.capital_at_risk_brl;
  const potentialInternalMetric = dto.kpis.potential_revenue_internal_brl;
  const potentialMarketMetric = dto.kpis.potential_revenue_market_brl;
  const capitalTotal = capitalMetric.value ?? 0;
  const capitalRisk = capitalRiskMetric.value ?? 0;
  const capitalRiskPct = capitalTotal > 0 ? (capitalRisk / capitalTotal) * 100 : 0;
  const capitalRiskTone: "positive" | "negative" = capitalRiskPct <= 15 ? "positive" : "negative";

  function trendChange(metric: { trend?: { delta_mom_pct: number | null; is_available: boolean } }, fallback: string): string {
    const trend = metric.trend;
    if (!trend?.is_available || trend.delta_mom_pct == null) return fallback;
    const value = trend.delta_mom_pct;
    return `${value >= 0 ? "+" : ""}${value.toFixed(1)}% vs mes anterior`;
  }

  function trendTone(metric: { trend?: { delta_mom_pct: number | null; is_available: boolean } }, _fallback: "positive" | "negative"): "positive" | "negative" {
    const trend = metric.trend;
    if (!trend?.is_available || trend.delta_mom_pct == null) return "negative";
    return trend.delta_mom_pct >= 0 ? "positive" : "negative";
  }

  function trendToneInverted(metric: { trend?: { delta_mom_pct: number | null; is_available: boolean } }, _fallback: "positive" | "negative"): "positive" | "negative" {
    const trend = metric.trend;
    if (!trend?.is_available || trend.delta_mom_pct == null) return "negative";
    return trend.delta_mom_pct >= 0 ? "negative" : "positive";
  }

  function hasTrend(metric: { trend?: { is_available: boolean } }): boolean {
    return Boolean(metric.trend?.is_available);
  }

  return {
    header: {
      title: "Analytics Overview",
      subtitle: "Visao geral estrategica do portfolio de produtos",
      scopeLabel: formatCoverage(dto),
      qualityLabel: `Qualidade: null ${dto.quality.null_total} | imputed ${dto.quality.imputed_total}`,
    },
    kpis: [
      {
        label: "Portfolio Ativo",
        value: formatInteger(dto.kpis.portfolio_active.value ?? 0),
        change: trendChange(dto.kpis.portfolio_active, "MoM indisponivel"),
        tone: trendTone(dto.kpis.portfolio_active, toneByDelta(trendDelta)),
        helpItems: resolveKpiHelpItems("portfolio_ativo"),
        showChangeArrow: hasTrend(dto.kpis.portfolio_active),
      },
      {
        label: "Capital Imobilizado",
        value: formatCompactBRL(capitalMetric.value ?? 0),
        change: trendChange(capitalMetric, "MoM indisponivel"),
        tone: trendTone(capitalMetric, capitalMetric.flags.is_imputed ? "negative" : "positive"),
        helpItems: resolveKpiHelpItems("capital_imobilizado"),
        showChangeArrow: hasTrend(capitalMetric),
      },
      {
        label: "Capital em Risco",
        value: formatCompactBRL(capitalRiskMetric.value ?? 0),
        change: trendChange(capitalRiskMetric, `MoM indisponivel | ${capitalRiskPct.toFixed(1)}% do imobilizado`),
        tone: trendToneInverted(capitalRiskMetric, capitalRiskTone),
        helpItems: resolveKpiHelpItems("capital_em_risco"),
        showChangeArrow: hasTrend(capitalRiskMetric),
      },
      {
        label: "Receita Potencial Interna",
        value: formatCompactBRL(potentialInternalMetric.value ?? 0),
        change: trendChange(potentialInternalMetric, "MoM indisponivel"),
        tone: trendTone(potentialInternalMetric, potentialInternalMetric.flags.is_imputed ? "negative" : "positive"),
        helpItems: resolveKpiHelpItems("receita_potencial_interna"),
        showChangeArrow: hasTrend(potentialInternalMetric),
      },
      {
        label: "Receita Potencial Mercado",
        value: formatCompactBRL(potentialMarketMetric.value ?? 0),
        change: trendChange(potentialMarketMetric, "MoM indisponivel"),
        tone: trendTone(potentialMarketMetric, potentialMarketMetric.flags.is_imputed ? "negative" : "positive"),
        helpItems: resolveKpiHelpItems("receita_potencial_mercado"),
        showChangeArrow: hasTrend(potentialMarketMetric),
      },
    ],
    matrix: {
      stars: dto.matrix.stars.count,
      potential: dto.matrix.potential.count,
      attention: dto.matrix.attention.count,
      critical: dto.matrix.critical.count,
    },
    matrixSpotlight: {
      display: {
        stars: Array.isArray(dto.matrix.spotlight?.display?.stars) ? dto.matrix.spotlight?.display?.stars ?? [] : [],
        potential: Array.isArray(dto.matrix.spotlight?.display?.potential) ? dto.matrix.spotlight?.display?.potential ?? [] : [],
        attention: Array.isArray(dto.matrix.spotlight?.display?.attention) ? dto.matrix.spotlight?.display?.attention ?? [] : [],
        critical: Array.isArray(dto.matrix.spotlight?.display?.critical) ? dto.matrix.spotlight?.display?.critical ?? [] : [],
      },
      stars: mapSpotlightRows(dto.matrix.spotlight?.stars),
      potential: mapSpotlightRows(dto.matrix.spotlight?.potential),
      attention: mapSpotlightRows(dto.matrix.spotlight?.attention),
      critical: mapSpotlightRows(dto.matrix.spotlight?.critical),
    },
    abc: dto.abc.bands.map((band) => ({
      letter: band.class,
      countLabel: `${formatInteger(band.sku_count)} produtos`,
      shareLabel: `${formatPercent(band.margin_share_pct)} da margem total`,
      fillPct: Math.max(5, Math.round(band.margin_share_pct)),
      valueLabel: formatCompactBRL(band.margin_brl),
    })),
    rankings: {
      topMargin: dto.rankings.top_margin.map((row) => ({
        pn: row.pn || "-",
        product: row.product || "Sem descricao",
        metric: formatPercent(row.margin_pct),
        value: formatCompactBRL(row.value_brl),
      })),
      topGiro: dto.rankings.top_giro.map((row) => ({
        pn: row.pn || "-",
        product: row.product || "Sem descricao",
        metric: `${row.giro.toFixed(1)}x`,
        value: `${formatInteger(row.sales_units)}un`,
      })),
      worstGiro: dto.rankings.worst_giro.map((row) => ({
        pn: row.pn || "-",
        product: row.product || "Sem descricao",
        metric: `${formatInteger(row.days)}d`,
        value: formatCompactBRL(row.capital_brl),
      })),
    },
    trends: [
      {
        label: "Vendas 90 dias",
        change: `${dto.trends_90d.sales.delta_pct >= 0 ? "+" : ""}${dto.trends_90d.sales.delta_pct.toFixed(1)}%`,
        tone: toneByDelta(dto.trends_90d.sales.delta_pct),
        points: dto.trends_90d.sales.points,
      },
      {
        label: "Margem 90 dias",
        change: `${dto.trends_90d.margin.delta_pp >= 0 ? "+" : ""}${dto.trends_90d.margin.delta_pp.toFixed(1)}pp`,
        tone: toneByDelta(dto.trends_90d.margin.delta_pp),
        points: dto.trends_90d.margin.points,
      },
      {
        label: "Estoque 90 dias",
        change: `${stockDelta >= 0 ? "+" : ""}${stockDelta.toFixed(1)}%`,
        tone: toneByDelta(stockDelta),
        points: dto.trends_90d.stock.points,
      },
    ],
  };
}
