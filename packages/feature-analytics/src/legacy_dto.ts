import { buildMockAnalyticsHomeDto, buildMockTaxonomyScopeOverview } from "./mocks/analyticsMocks";

export type AnalyticsHomeV2Dto = Record<string, unknown>;

type AnalyticsMetricV1Dto = {
  value: number;
  trend?: {
    delta_mom_pct: number | null;
    is_available: boolean;
    window_months?: number;
  };
  flags?: {
    is_imputed?: boolean;
  };
};

export type AnalyticsTaxonomyScopeOverviewV1Dto = {
  message?: string;
  progress?: Record<string, unknown>;
  kpis: {
    active_entities: AnalyticsMetricV1Dto;
    gross_revenue_6m_brl: AnalyticsMetricV1Dto;
    margin_total_brl: AnalyticsMetricV1Dto;
    margin_pct: AnalyticsMetricV1Dto;
    capital_total_brl: AnalyticsMetricV1Dto;
    capital_at_risk_brl: AnalyticsMetricV1Dto;
    potential_revenue_internal_brl: AnalyticsMetricV1Dto;
    potential_revenue_market_brl: AnalyticsMetricV1Dto;
  };
  scope: {
    level_label?: string;
    trend_window_months?: number;
    window?: {
      window_months?: number;
    };
    margin_policy?: {
      low_pct?: number;
      good_pct?: number;
    };
    filter_options?: {
      marcas?: string[];
    };
    analysis_cards?: {
      abc_mix?: {
        a_max_cum_pct?: number;
        b_max_cum_pct?: number;
        a_count?: number;
        b_count?: number;
        c_count?: number;
        a_revenue_pct?: number;
        b_revenue_pct?: number;
        c_revenue_pct?: number;
      };
      gmroi_global?: number;
      capital_travado_brl?: number;
      risco_global_pct?: number;
      margem_contrib_real_pct?: number;
    };
  };
  panels: {
    top_nodes_by_revenue?: Array<Record<string, unknown>>;
    revenue_concentration?: Record<string, unknown>;
    capital_allocation_map?: Array<Record<string, unknown>>;
    capital_allocation_map_spotlight?: Array<Record<string, unknown>>;
    capital_efficiency?: Array<Record<string, unknown>>;
    rankings: {
      top_margin: Array<Record<string, unknown>>;
      nodes_at_risk: Array<Record<string, unknown>>;
    };
    backlog?: Array<Record<string, unknown>>;
  };
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object";
}

export function makeAnalyticsHomeV2Dto(
  envelope: { data?: Record<string, unknown> | null } | null | undefined,
  _snapshotScope: string,
): AnalyticsHomeV2Dto {
  if (isRecord(envelope?.data)) {
    return envelope.data as AnalyticsHomeV2Dto;
  }
  return buildMockAnalyticsHomeDto();
}

export function makeAnalyticsTaxonomyScopeOverviewV1Dto(
  data: unknown,
  _snapshotScope: string,
): AnalyticsTaxonomyScopeOverviewV1Dto {
  if (isRecord(data)) {
    return data as AnalyticsTaxonomyScopeOverviewV1Dto;
  }
  return buildMockTaxonomyScopeOverview();
}
