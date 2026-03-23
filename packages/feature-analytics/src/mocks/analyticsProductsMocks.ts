import type { AnalyticsSkuRow, SkuStatus } from "../pages/analytics/contracts_products";
import { mockProducts } from "../pages/analytics/mocks/mock_products";
import { getWorkspaceModel } from "../pages/analytics/workspace/mock_workspace";

type ProductsIndexFilters = {
  search?: string;
  marca?: string;
  taxonomyLeafName?: string;
  status?: string;
  limit?: number;
  offset?: number;
};

function asList(value?: string): string[] {
  if (!value) return [];
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function unique(values: string[]): string[] {
  return Array.from(new Set(values));
}

function matchesSearch(row: AnalyticsSkuRow, search?: string): boolean {
  if (!search) return true;
  const needle = search.trim().toLowerCase();
  if (!needle) return true;
  return (
    row.pn.toLowerCase().includes(needle) ||
    row.description.toLowerCase().includes(needle) ||
    row.brand.toLowerCase().includes(needle) ||
    row.taxonomyLeafName.toLowerCase().includes(needle)
  );
}

function mapToIndexRow(row: AnalyticsSkuRow) {
  return {
    pn: row.pn,
    ean: row.ean,
    description: row.description,
    taxonomy_leaf_name: row.taxonomyLeafName,
    brand: row.brand,
    status: row.status,
    action: row.action,
    segment_id: row.classLabel,
    kpis: {
      gap_pct: row.metrics.gapPct,
      margem_contrib_pct: row.metrics.marginPct,
      slope: row.metrics.slope6,
      mean_units_6m: row.metrics.pme6,
      trend_pct_mom: row.trendPct,
      xyz: row.metrics.xyz6,
      estoque_un: row.stock,
      our_price: row.price,
      comp_price_mean: row.marketPrice,
      pme: row.metrics.pme6,
      giro_6m: row.metrics.giro6,
      dos: row.metrics.dos6,
      gmroi: row.metrics.gmroi6,
      cv: row.metrics.cv6,
      venda_1m_un: row.short.dem3,
    },
    alerts: [],
  };
}

function calcCapital(row: AnalyticsSkuRow): number {
  return Math.max(0, row.stock) * Math.max(0, row.price);
}

export function buildMockAnalyticsProductsIndexDto(filters: ProductsIndexFilters = {}) {
  const search = filters.search?.trim();
  const brandFilter = asList(filters.marca);
  const taxonomyFilter = asList(filters.taxonomyLeafName);
  const statusFilter = asList(filters.status) as SkuStatus[];
  const limit = Number.isFinite(filters.limit) ? Math.max(1, Number(filters.limit)) : 200;
  const offset = Number.isFinite(filters.offset) ? Math.max(0, Number(filters.offset)) : 0;

  let rows = mockProducts.filter((row) => matchesSearch(row, search));
  if (brandFilter.length) rows = rows.filter((row) => brandFilter.includes(row.brand));
  if (taxonomyFilter.length) rows = rows.filter((row) => taxonomyFilter.includes(row.taxonomyLeafName));
  if (statusFilter.length) rows = rows.filter((row) => statusFilter.includes(row.status));

  const total = rows.length;
  const paged = rows.slice(offset, offset + limit);

  return {
    snapshot: { as_of: "2026-03-22" },
    rows: paged.map(mapToIndexRow),
    filters: {
      marcas: unique(mockProducts.map((row) => row.brand)).sort(),
      taxonomy_leaf_names: unique(mockProducts.map((row) => row.taxonomyLeafName)).sort(),
    },
    paging: {
      offset,
      limit,
      returned: paged.length,
      total,
    },
  };
}

export function buildMockAnalyticsProductsOverviewDto() {
  const total = mockProducts.length;
  const capitalTotal = mockProducts.reduce((sum, row) => sum + calcCapital(row), 0);
  const capitalRisk = capitalTotal * 0.22;
  const potentialInternal = capitalTotal * 0.46;
  const potentialMarket = capitalTotal * 0.62;
  const marginAvg = mockProducts.reduce((sum, row) => sum + row.metrics.marginPct, 0) / Math.max(1, total);
  const trendPct = 2.4;

  const byStatus = mockProducts.reduce(
    (acc, row) => {
      acc[row.status] = (acc[row.status] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );

  const spotlightRows = mockProducts.slice(0, 6).map((row) => ({
    pn: row.pn,
    product: row.description,
    brand: row.brand,
    taxonomy_leaf_name: row.taxonomyLeafName,
    stock_value_brl: calcCapital(row),
    stock_qty: row.stock,
    financial_priority_score: calcCapital(row),
    financial_priority_tier: row.classLabel,
    margin_pct: row.metrics.marginPct,
    margin_sales_pct: row.metrics.marginPct,
    margin_unit_pct: row.metrics.marginPct,
    sales_6m_units: row.metrics.pme6,
    sales_1m_units: row.short.dem3,
    gap_vs_market_pct: row.metrics.gapPct,
    giro_6m: row.metrics.giro6,
    days_no_sales: Math.round(row.metrics.dos6 / 4),
    dos: row.metrics.dos6,
    capital_brl: calcCapital(row),
    contribution_brl: calcCapital(row) * 0.14,
  }));

  return {
    snapshot: { as_of: "2026-03-22" },
    scope: {
      filters_applied: {
        marca: "",
        taxonomy_leaf_name: "",
        status: "",
        search: "",
      },
      coverage: {
        is_partial: false,
        rows_considered: total,
        rows_total_available: total,
      },
    },
    quality: {
      null_total: 0,
      imputed_total: 0,
    },
    kpis: {
      trend_pct: { value: trendPct, trend: { is_available: true, delta_mom_pct: 1.1 }, flags: { is_imputed: false } },
      portfolio_active: {
        value: total,
        trend: { is_available: true, delta_mom_pct: 0.7 },
        flags: { is_imputed: false },
      },
      capital_total_brl: {
        value: capitalTotal,
        trend: { is_available: true, delta_mom_pct: 1.4 },
        flags: { is_imputed: false },
      },
      capital_at_risk_brl: {
        value: capitalRisk,
        trend: { is_available: true, delta_mom_pct: 2.1 },
        flags: { is_imputed: false },
      },
      potential_revenue_internal_brl: {
        value: potentialInternal,
        trend: { is_available: true, delta_mom_pct: 1.8 },
        flags: { is_imputed: false },
      },
      potential_revenue_market_brl: {
        value: potentialMarket,
        trend: { is_available: true, delta_mom_pct: 2.4 },
        flags: { is_imputed: false },
      },
      capital_brl: { value: capitalTotal, trend: { is_available: true, delta_mom_pct: 1.2 }, flags: { is_imputed: false } },
      weighted_margin_pct: {
        value: marginAvg,
        trend: { is_available: true, delta_mom_pct: 0.3 },
        flags: { is_imputed: false },
      },
    },
    matrix: {
      stars: { count: byStatus.info || 0 },
      potential: { count: byStatus.ok || 0 },
      attention: { count: byStatus.warn || 0 },
      critical: { count: byStatus.crit || 0 },
      spotlight: {
        display: {
          stars: ["Info"],
          potential: ["OK"],
          attention: ["Atencao"],
          critical: ["Critico"],
        },
        stars: spotlightRows.slice(0, 2),
        potential: spotlightRows.slice(2, 3),
        attention: spotlightRows.slice(3, 4),
        critical: spotlightRows.slice(4, 6),
      },
    },
    abc: {
      bands: [
        { class: "A", sku_count: Math.round(total * 0.2), margin_share_pct: 52, margin_brl: capitalTotal * 0.52 },
        { class: "B", sku_count: Math.round(total * 0.3), margin_share_pct: 33, margin_brl: capitalTotal * 0.33 },
        { class: "C", sku_count: Math.round(total * 0.5), margin_share_pct: 15, margin_brl: capitalTotal * 0.15 },
      ],
    },
    rankings: {
      top_margin: spotlightRows.slice(0, 3).map((row) => ({
        pn: row.pn,
        product: row.product,
        margin_pct: row.margin_pct,
        value_brl: row.contribution_brl,
      })),
      top_giro: spotlightRows.slice(1, 4).map((row) => ({
        pn: row.pn,
        product: row.product,
        giro: row.giro_6m,
        sales_units: row.sales_6m_units,
      })),
      worst_giro: spotlightRows.slice(2, 5).map((row) => ({
        pn: row.pn,
        product: row.product,
        days: row.days_no_sales,
        capital_brl: row.capital_brl,
      })),
    },
    alerts: {
      products_with_alerts: Math.round(total * 0.12),
    },
    trends_90d: {
      sales: { delta_pct: 3.1, points: [24, 36, 30, 42, 40, 46] },
      margin: { delta_pp: 1.2, points: [18, 19, 18.5, 20, 19.6, 21] },
      stock: { delta_pct: -2.4, points: [32, 30, 29, 27, 26, 28] },
    },
  };
}

export function buildMockAnalyticsProductWorkspaceDto(pn: string) {
  return {
    model: getWorkspaceModel(pn),
  };
}
