import { buildMockAnalyticsHomeDto, buildMockTaxonomyScopeOverview } from "./mocks/analyticsMocks";

export type AnalyticsHomeV2Dto = Record<string, unknown>;

export type AnalyticsTaxonomyScopeOverviewV1Dto = {
  scope: {
    filter_options?: {
      marcas?: string[];
    };
    analysis_cards?: {
      abc_mix?: {
        a_max_cum_pct?: number;
        b_max_cum_pct?: number;
      };
    };
  };
  panels: {
    capital_efficiency?: Array<Record<string, unknown>>;
    top_nodes_by_revenue?: Array<Record<string, unknown>>;
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
