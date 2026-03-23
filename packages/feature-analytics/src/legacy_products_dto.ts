import {
  buildMockAnalyticsProductWorkspaceDto,
  buildMockAnalyticsProductsIndexDto,
  buildMockAnalyticsProductsOverviewDto,
} from "./mocks/analyticsProductsMocks";

export type ProductsAnalyticsIndexRowV1 = any;
export type AnalyticsProductsIndexV1Dto = any;
export type AnalyticsProductsOverviewV1Dto = any;
export type AnalyticsProductWorkspaceV1Dto = { model: any };
export type WorkspaceHeroMetricV1 = { label: string; value: string };
export type WorkspaceHistoryIndicatorV1 = { label: string; value: string; fill_pct: number };
export type WorkspaceHistoryPricePointV1 = {
  date: string;
  our_price?: number | null;
  market_mean?: number | null;
  suppliers?: Record<string, number | null>;
};
export type WorkspaceHistorySupplierLinkV1 = { label: string; url?: string | null };
export type WorkspaceHistorySalesPointV1 = {
  date?: string;
  month?: string;
  units?: number | null;
  revenue?: number | null;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object";
}

export function makeAnalyticsProductsIndexV1Dto(
  data: unknown,
  _snapshotScope: string,
): AnalyticsProductsIndexV1Dto {
  if (isRecord(data) && Array.isArray((data as { rows?: unknown }).rows)) {
    return data as AnalyticsProductsIndexV1Dto;
  }
  return buildMockAnalyticsProductsIndexDto();
}

export function makeAnalyticsProductsOverviewV1Dto(
  data: unknown,
  _snapshotScope: string,
): AnalyticsProductsOverviewV1Dto {
  if (isRecord(data) && isRecord((data as { kpis?: unknown }).kpis)) {
    return data as AnalyticsProductsOverviewV1Dto;
  }
  return buildMockAnalyticsProductsOverviewDto();
}

export function makeAnalyticsProductWorkspaceV1Dto(
  data: unknown,
  _snapshotScope: string,
): AnalyticsProductWorkspaceV1Dto {
  if (isRecord(data) && isRecord((data as { model?: unknown }).model)) {
    return data as AnalyticsProductWorkspaceV1Dto;
  }
  const pn = isRecord(data) ? String((data as { pn?: unknown }).pn || "") : "";
  return buildMockAnalyticsProductWorkspaceDto(pn || "33584");
}
