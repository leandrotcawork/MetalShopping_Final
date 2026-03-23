import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  makeAnalyticsTaxonomyScopeOverviewV1Dto,
  type AnalyticsHomeV2Dto,
  type AnalyticsTaxonomyScopeOverviewV1Dto,
} from "@metalshopping/feature-analytics";
import { useLocation, useNavigate } from "react-router-dom";

import { useAppSession } from "../../app/providers/AppProviders";
import { Chip } from "../../components/ui";
import { mapAnalyticsHomeViewModel } from "./analyticsHomeViewModel";
import { AnalyticsSpotlightDrawer } from "./components/AnalyticsSpotlightDrawer";
import { AbcMixSpotlight } from "./components/AbcMixSpotlight";
import { BentoGrid } from "./components/BentoGrid";
import { CapitalEfficiencySpotlight } from "./components/CapitalEfficiencySpotlight";
import { HeroCommandCenter } from "./components/HeroCommandCenter";
import { HeroTools } from "./components/HeroTools";
import { SpotlightCallout } from "./components/SpotlightCallout";
import { SpotlightNextSteps } from "./components/SpotlightNextSteps";
import { SpotlightSkuList } from "./components/SpotlightSkuList";
import { SpotlightSelectionWidget } from "./components/SpotlightSelectionWidget";
import styles from "./analytics_home.module.css";
import { resolveSpotlightContent, toSpotlightKey, type SpotlightExtra, type SpotlightState } from "./spotlightModel";

type AnalyticsHomePageProps = {
  updatedAtLabel?: string;
  dto?: AnalyticsHomeV2Dto | null;
  onRefresh?: () => Promise<void> | void;
  isRefreshing?: boolean;
};

type SpotlightTableState = {
  brandFilter: string[];
  taxonomyFilter: string[];
  stockTypeFilter: string[];
  pageSize: number;
  pageIndex: number;
  sortKey?: string;
  sortDir?: "asc" | "desc";
  stockOp: "gte" | "eq" | "lte";
  stockValue: number | null;
  scrollTop: number;
  savedAt: number;
};

const SPOTLIGHT_TABLE_STATE_PREFIX = "analytics:spotlight:table_state:v1:";

type WorkspaceReturnState = {
  from: string | null;
  fromScrollY: number | null;
  savedAt: number | null;
};

function readWorkspaceReturnState(): WorkspaceReturnState {
  try {
    const raw = window.sessionStorage.getItem("analytics:workspace:return");
    if (!raw) return { from: null, fromScrollY: null, savedAt: null };
    const parsed = JSON.parse(raw) as { from?: unknown; fromScrollY?: unknown; savedAt?: unknown };
    const from = String(parsed.from || "").trim() || null;
    const fromScrollYRaw = Number(parsed.fromScrollY);
    const savedAtRaw = Number(parsed.savedAt);
    return {
      from,
      fromScrollY: Number.isFinite(fromScrollYRaw) && fromScrollYRaw >= 0 ? fromScrollYRaw : null,
      savedAt: Number.isFinite(savedAtRaw) && savedAtRaw > 0 ? savedAtRaw : null,
    };
  } catch {
    return { from: null, fromScrollY: null, savedAt: null };
  }
}

function fromPathSpotlightKey(fromPath: string | null): string | null {
  const raw = String(fromPath || "").trim();
  if (!raw) return null;
  try {
    const parsed = new URL(raw, window.location.origin);
    const key = String(parsed.searchParams.get("spotlight") || "").trim();
    return key || null;
  } catch {
    return null;
  }
}

function normalizeFilterList(raw: unknown): string[] {
  if (Array.isArray(raw)) {
    return Array.from(
      new Set(
        raw
          .map((value) => String(value || "").trim())
          .filter((value) => value && value !== "all")
      )
    );
  }
  const single = String(raw || "").trim();
  if (!single || single === "all") return [];
  return [single];
}

function readFilterListFromQuery(query: URLSearchParams, key: string): string[] {
  return Array.from(
    new Set(
      query
        .getAll(key)
        .map((value) => String(value || "").trim())
        .filter((value) => value && value !== "all")
    )
  );
}

function writeFilterListToQuery(query: URLSearchParams, key: string, values: string[]) {
  query.delete(key);
  values.forEach((value) => {
    const token = String(value || "").trim();
    if (token) query.append(key, token);
  });
}

function spotlightStateStorageKey(spotlightKey: string): string {
  const token = String(spotlightKey || "").trim();
  return `${SPOTLIGHT_TABLE_STATE_PREFIX}${encodeURIComponent(token)}`;
}

function readSpotlightTableState(spotlightKey: string): SpotlightTableState | null {
  const token = String(spotlightKey || "").trim();
  if (!token) return null;
  try {
    const raw = window.sessionStorage.getItem(spotlightStateStorageKey(token));
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<SpotlightTableState>;
    const pageSizeRaw = Number(parsed.pageSize);
    const pageSize = [10, 20, 50].includes(pageSizeRaw) ? pageSizeRaw : 20;
    const pageIndexRaw = Number(parsed.pageIndex);
    const pageIndex = Number.isFinite(pageIndexRaw) && pageIndexRaw >= 0 ? Math.floor(pageIndexRaw) : 0;
    const stockOpRaw = String(parsed.stockOp || "").trim().toLowerCase();
    const stockOp: SpotlightTableState["stockOp"] = stockOpRaw === "eq" || stockOpRaw === "lte" ? (stockOpRaw as SpotlightTableState["stockOp"]) : "gte";
    const stockValueRaw = parsed.stockValue == null ? Number.NaN : Number(parsed.stockValue);
    const stockValue = Number.isFinite(stockValueRaw) ? stockValueRaw : null;
    const scrollTopRaw = Number(parsed.scrollTop);
    const scrollTop = Number.isFinite(scrollTopRaw) && scrollTopRaw >= 0 ? Math.floor(scrollTopRaw) : 0;
    const savedAtRaw = Number(parsed.savedAt);
    const savedAt = Number.isFinite(savedAtRaw) && savedAtRaw > 0 ? savedAtRaw : Date.now();
    return {
      brandFilter: normalizeFilterList(parsed.brandFilter),
      taxonomyFilter: normalizeFilterList(parsed.taxonomyFilter),
      stockTypeFilter: normalizeFilterList(parsed.stockTypeFilter),
      pageSize,
      pageIndex,
      sortKey: parsed.sortKey ? String(parsed.sortKey).trim() || undefined : undefined,
      sortDir: parsed.sortDir === "asc" || parsed.sortDir === "desc" ? parsed.sortDir : undefined,
      stockOp,
      stockValue,
      scrollTop,
      savedAt,
    };
  } catch {
    return null;
  }
}

function writeSpotlightTableState(spotlightKey: string, next: Omit<SpotlightTableState, "savedAt">) {
  const token = String(spotlightKey || "").trim();
  if (!token) return;
  try {
    window.sessionStorage.setItem(
      spotlightStateStorageKey(token),
      JSON.stringify({ ...next, savedAt: Date.now() } satisfies SpotlightTableState),
    );
  } catch {
    // Ignore storage failures; URL remains the primary state.
  }
}

function clearSpotlightTableState(spotlightKey: string) {
  const token = String(spotlightKey || "").trim();
  if (!token) return;
  try {
    window.sessionStorage.removeItem(spotlightStateStorageKey(token));
  } catch {
    // ignore
  }
}

function parseSpotlightStateFromSearch(search: string): SpotlightState {
  const query = new URLSearchParams(search);
  const spotlightKey = String(query.get("spotlight") || "").trim();
  if (!spotlightKey) {
    return { open: false, key: null };
  }
  const source = String(query.get("spotlightSource") || "").trim();
  return {
    open: true,
    key: toSpotlightKey(spotlightKey),
    extra: source ? { source } : undefined,
  };
}

export function AnalyticsHomePage({ updatedAtLabel, dto, onRefresh, isRefreshing = false }: AnalyticsHomePageProps) {
  const { api } = useAppSession();
  const navigate = useNavigate();
  const location = useLocation();
  const didRestoreScrollRef = useRef(false);
  const spotlightBodyRef = useRef<HTMLDivElement | null>(null);
  const didRestoreSpotlightScrollRef = useRef<string>("");
  const [spotlight, setSpotlight] = useState<SpotlightState>(() => parseSpotlightStateFromSearch(location.search));
  const [activeFilter, setActiveFilter] = useState<"all" | "critical" | "pricing" | "stock" | "data">("all");
  const [capitalScopeDto, setCapitalScopeDto] = useState<AnalyticsTaxonomyScopeOverviewV1Dto | null>(null);
  const [capitalScopeLoading, setCapitalScopeLoading] = useState(false);
  const [capitalScopeError, setCapitalScopeError] = useState("");
  const [capitalSearchQuery, setCapitalSearchQuery] = useState("");
  const [capitalBrandFilter, setCapitalBrandFilter] = useState<string[]>([]);
  const [abcScopeDto, setAbcScopeDto] = useState<AnalyticsTaxonomyScopeOverviewV1Dto | null>(null);
  const [abcScopeLoading, setAbcScopeLoading] = useState(false);
  const [abcScopeError, setAbcScopeError] = useState("");
  const [abcSearchQuery, setAbcSearchQuery] = useState("");
  const [abcBrandFilter, setAbcBrandFilter] = useState<string[]>([]);
  const model = useMemo(() => mapAnalyticsHomeViewModel(dto), [dto]);

  const openSpotlight = useCallback((key: string, extra?: SpotlightExtra) => {
    const nextKey = toSpotlightKey(key);
    clearSpotlightTableState(String(nextKey));
    setSpotlight({ open: true, key: nextKey, extra });

    const query = new URLSearchParams(location.search);
    query.set("spotlight", String(nextKey));
    if (extra?.source) query.set("spotlightSource", String(extra.source));
    else query.delete("spotlightSource");

    query.delete("spotlightBrand");
    query.delete("spotlightGroup");
    query.delete("spotlightStockType");
    query.delete("spotlightPageSize");
    query.delete("spotlightPageIndex");
    query.delete("spotlightSortKey");
    query.delete("spotlightSortDir");
    query.delete("spotlightStockValue");
    query.delete("spotlightStockOp");
    query.delete("spotlightScrollTop");

    const nextSearch = query.toString();
    const currentSearch = String(location.search || "").replace(/^\?/, "");
    if (nextSearch !== currentSearch) {
      navigate(`${location.pathname}${nextSearch ? `?${nextSearch}` : ""}${location.hash}`, { replace: true });
    }
  }, [location.hash, location.pathname, location.search, navigate]);

  const closeSpotlight = useCallback(() => {
    const currentKey = String(spotlight.key || "").trim();
    if (currentKey) clearSpotlightTableState(currentKey);
    setSpotlight({ open: false, key: null });
    if (location.pathname !== "/analytics/home" || location.search || location.hash) {
      navigate("/analytics/home", { replace: true });
    }
  }, [location.hash, location.pathname, location.search, navigate, spotlight.key]);

  const spotlightContent = useMemo(
    () => resolveSpotlightContent(spotlight.key, spotlight.extra, model),
    [spotlight.key, spotlight.extra, model]
  );

  useEffect(() => {
    if (!(spotlight.open && spotlight.key === "mini-capital")) return;
    let disposed = false;
    async function loadCapitalScope() {
      setCapitalScopeLoading(true);
      setCapitalScopeError("");
      try {
        const brandCsv = capitalBrandFilter.length ? capitalBrandFilter.join(",") : undefined;
        const env = await api.analytics.workspaceTaxonomyScope({
          level: 0,
          windowMonths: 6,
          search: capitalSearchQuery.trim() || undefined,
          marca: brandCsv,
          limit: 5000,
          offset: 0,
        });
        if (disposed) return;
        const mapped = makeAnalyticsTaxonomyScopeOverviewV1Dto(env.data, "current");
        setCapitalScopeDto(mapped);
      } catch (err) {
        if (disposed) return;
        setCapitalScopeError(err instanceof Error ? err.message : String(err));
      } finally {
        if (!disposed) setCapitalScopeLoading(false);
      }
    }
    void loadCapitalScope();
    return () => {
      disposed = true;
    };
  }, [api.analytics, capitalBrandFilter, capitalSearchQuery, spotlight.key, spotlight.open]);

  useEffect(() => {
    if (!(spotlight.open && spotlight.key === "mini-abc")) return;
    let disposed = false;
    async function loadAbcScope() {
      setAbcScopeLoading(true);
      setAbcScopeError("");
      try {
        const brandCsv = abcBrandFilter.length ? abcBrandFilter.join(",") : undefined;
        const env = await api.analytics.workspaceTaxonomyScope({
          level: 0,
          windowMonths: 6,
          search: abcSearchQuery.trim() || undefined,
          marca: brandCsv,
          limit: 5000,
          offset: 0,
        });
        if (disposed) return;
        const mapped = makeAnalyticsTaxonomyScopeOverviewV1Dto(env.data, "current");
        setAbcScopeDto(mapped);
      } catch (err) {
        if (disposed) return;
        setAbcScopeError(err instanceof Error ? err.message : String(err));
      } finally {
        if (!disposed) setAbcScopeLoading(false);
      }
    }
    void loadAbcScope();
    return () => {
      disposed = true;
    };
  }, [abcBrandFilter, abcSearchQuery, api.analytics, spotlight.key, spotlight.open]);

  const isActionSpotlight = spotlight.key === "mini-actions" || String(spotlight.key || "").startsWith("action-bucket:");
  const isAlertListSpotlight = spotlight.key === "mini-alerts" || spotlight.key === "stat-alerts";
  const isHeatOverviewSpotlight = spotlight.key === "mini-heat";
  const isPortfolioOverviewSpotlight = spotlight.key === "mini-portfolio";
  const isCapitalSpotlight = spotlight.key === "mini-capital";
  const isAbcSpotlight = spotlight.key === "mini-abc";
  const isPortfolioDetailSpotlight = String(spotlight.key || "").startsWith("pf-");
  const isAlertDetailSpotlight = String(spotlight.key || "").startsWith("alert:");
  const isHeatDetailSpotlight = String(spotlight.key || "").startsWith("heat-");
  const isListSpotlight =
    isActionSpotlight ||
    isAlertListSpotlight ||
    isCapitalSpotlight ||
    isAbcSpotlight ||
    isHeatOverviewSpotlight ||
    isPortfolioOverviewSpotlight ||
    isPortfolioDetailSpotlight ||
    isAlertDetailSpotlight ||
    isHeatDetailSpotlight;

  const selectedAction = useMemo(
    () => model.allActions.find((row) => row.key === spotlight.key) || null,
    [model.allActions, spotlight.key]
  );
  const selectedAlert = useMemo(
    () => model.allAlerts.find((row) => row.key === spotlight.key) || null,
    [model.allAlerts, spotlight.key]
  );
  const selectedHeatCell = useMemo(
    () => (spotlight.key ? model.heatCells[String(spotlight.key)] || null : null),
    [model.heatCells, spotlight.key]
  );
  const selectedPortfolio = useMemo(
    () => model.portfolio.find((row) => row.key === spotlight.key) || null,
    [model.portfolio, spotlight.key]
  );
  const capitalSpotlightRows = useMemo(() => {
    const efficiencyRows = capitalScopeDto?.panels.capital_efficiency || [];
    return efficiencyRows.map((row) => {
      const riskPct = Math.max(0, Number(row.risk_pct || 0));
      const riskLevel: "low" | "medium" | "high" =
        riskPct >= 30 ? "high" : riskPct >= 15 ? "medium" : "low";
      return {
        nodeId: Number(row.node_id || 0),
        nodeName: String(row.node_name || ""),
        capitalBrl: Math.max(0, Number(row.capital_brl || 0)),
        riskLevel,
        riskPct,
        gmroi: row.gmroi ?? null,
        revenueBrl: Math.max(0, Number(row.revenue_brl || 0)),
        marginBrl: Math.max(0, Number((row as { gross_margin_brl?: number }).gross_margin_brl ?? row.margin_brl ?? 0)),
      };
    });
  }, [capitalScopeDto]);
  const capitalBrandOptions = useMemo(
    () => capitalScopeDto?.scope.filter_options?.marcas || [],
    [capitalScopeDto?.scope.filter_options?.marcas],
  );
  const abcBrandOptions = useMemo(
    () => abcScopeDto?.scope.filter_options?.marcas || [],
    [abcScopeDto?.scope.filter_options?.marcas],
  );
  const abcSpotlightRows = useMemo(() => {
    const rows = abcScopeDto?.panels.top_nodes_by_revenue || [];
    const totalRevenue = rows.reduce((acc, row) => acc + Math.max(0, Number(row.revenue_brl || 0)), 0);
    const aMax = Number(abcScopeDto?.scope.analysis_cards?.abc_mix.a_max_cum_pct || 80);
    const bMax = Number(abcScopeDto?.scope.analysis_cards?.abc_mix.b_max_cum_pct || 95);
    let cum = 0;
    return rows
      .map((row) => {
        const revenue = Math.max(0, Number(row.revenue_brl || 0));
        const share = totalRevenue > 0 ? (revenue / totalRevenue) * 100 : 0;
        cum += share;
        const band: "A" | "B" | "C" = cum <= aMax ? "A" : cum <= bMax ? "B" : "C";
        return {
          nodeId: Number(row.node_id || 0),
          nodeName: String(row.node_name || ""),
          revenueBrl: revenue,
          sharePct: share,
          cumSharePct: cum,
          band,
          marginPct: row.margin_pct == null ? null : Number(row.margin_pct),
        };
      })
      .sort((a, b) => b.revenueBrl - a.revenueBrl);
  }, [abcScopeDto?.panels.top_nodes_by_revenue, abcScopeDto?.scope.analysis_cards?.abc_mix.a_max_cum_pct, abcScopeDto?.scope.analysis_cards?.abc_mix.b_max_cum_pct]);
  const spotlightTableStateFromQuery = useMemo(() => {
    const query = new URLSearchParams(location.search);
    const querySpotlightKey = String(query.get("spotlight") || "").trim();
    const currentSpotlightKey = String(spotlight.key || "").trim();
    const effectiveCurrentSpotlightKey = currentSpotlightKey || querySpotlightKey;
    const queryMatchesCurrent =
      !!querySpotlightKey && !!effectiveCurrentSpotlightKey && querySpotlightKey === effectiveCurrentSpotlightKey;
    if (!queryMatchesCurrent) {
      return {
        brandFilter: [] as string[],
        taxonomyFilter: [] as string[],
        stockTypeFilter: [] as string[],
        pageSize: 20,
        pageIndex: 0,
        sortKey: undefined,
        sortDir: undefined as "asc" | "desc" | undefined,
        stockOp: "gte" as const,
        stockValue: null as number | null,
        scrollTop: 0,
      };
    }

    const state = location.state as { restoreScrollY?: unknown } | null;
    const restoreScrollY = Number(state?.restoreScrollY);
    const workspaceReturn = readWorkspaceReturnState();
    const workspaceReturnSpotlight = fromPathSpotlightKey(workspaceReturn.from);
    const workspaceReturnIsRecent =
      workspaceReturn.savedAt != null && Date.now() - workspaceReturn.savedAt <= 5 * 60 * 1000;
    const isReturnNavigation =
      (Number.isFinite(restoreScrollY) && restoreScrollY >= 0) ||
      (!!workspaceReturnIsRecent &&
        !!workspaceReturnSpotlight &&
        workspaceReturnSpotlight === querySpotlightKey);
    const stored = isReturnNavigation ? readSpotlightTableState(querySpotlightKey) : null;

    if (stored && isReturnNavigation) {
      return {
        brandFilter: stored.brandFilter,
        taxonomyFilter: stored.taxonomyFilter,
        stockTypeFilter: stored.stockTypeFilter,
        pageSize: stored.pageSize,
        pageIndex: stored.pageIndex,
        sortKey: stored.sortKey,
        sortDir: stored.sortDir,
        stockOp: stored.stockOp,
        stockValue: stored.stockValue,
        scrollTop: stored.scrollTop,
      };
    }

    const brandFilter = readFilterListFromQuery(query, "spotlightBrand");
    const taxonomyFilter = readFilterListFromQuery(query, "spotlightGroup");
    const stockTypeFilter = readFilterListFromQuery(query, "spotlightStockType");
    const pageSizeRaw = Number(query.get("spotlightPageSize") || 20);
    const pageSize = [10, 20, 50].includes(pageSizeRaw) ? pageSizeRaw : 20;
    const pageIndexRaw = Number(query.get("spotlightPageIndex") || 0);
    const pageIndex = Number.isFinite(pageIndexRaw) && pageIndexRaw >= 0 ? Math.floor(pageIndexRaw) : 0;
    const sortKeyRaw = String(query.get("spotlightSortKey") || "").trim();
    const sortDirRaw = String(query.get("spotlightSortDir") || "").trim().toLowerCase();
    const sortKey = sortKeyRaw || undefined;
    const sortDir: "asc" | "desc" | undefined =
      sortDirRaw === "asc" || sortDirRaw === "desc" ? sortDirRaw : undefined;
    const stockOpRaw = String(query.get("spotlightStockOp") || "").trim().toLowerCase();
    const stockOp: "gte" | "eq" | "lte" =
      stockOpRaw === "eq" || stockOpRaw === "lte" ? stockOpRaw : "gte";
    const stockValueRaw = query.get("spotlightStockValue");
    const stockValueParsed = stockValueRaw == null ? Number.NaN : Number(stockValueRaw);
    const stockValue = Number.isFinite(stockValueParsed) ? stockValueParsed : null;
    const scrollTopRaw = Number(query.get("spotlightScrollTop") || 0);
    const scrollTop = Number.isFinite(scrollTopRaw) && scrollTopRaw >= 0 ? Math.floor(scrollTopRaw) : 0;

    return { brandFilter, taxonomyFilter, stockTypeFilter, pageSize, pageIndex, sortKey, sortDir, stockOp, stockValue, scrollTop };
  }, [location.search, spotlight.key, location.state]);
  const alertHeaderTone = selectedAlert
    ? selectedAlert.toneClass === "critical"
      ? "crit"
      : selectedAlert.toneClass === "warning"
        ? "warn"
        : "info"
    : "info";
  const spotlightHeaderChips = useMemo(() => {
    if (isCapitalSpotlight) {
      const totalCapital = capitalSpotlightRows.reduce((acc, row) => acc + Math.max(0, Number(row.capitalBrl || 0)), 0);
      const totalGroups = capitalSpotlightRows.length;
      const windowMonths = capitalScopeDto?.scope.window?.window_months || capitalScopeDto?.scope.trend_window_months || 6;
      const gmroiRows = capitalSpotlightRows
        .map((row) => (row.gmroi == null || !Number.isFinite(Number(row.gmroi)) ? null : Number(row.gmroi)))
        .filter((value): value is number => value != null);
      const gmroiAvg = gmroiRows.length ? gmroiRows.reduce((acc, cur) => acc + cur, 0) / gmroiRows.length : null;
      return (
        <>
          <Chip
            label={`Capital analisado: ${new Intl.NumberFormat("pt-BR", {
              style: "currency",
              currency: "BRL",
              notation: "compact",
              maximumFractionDigits: 1,
            }).format(totalCapital)}`}
            tone="info"
          />
          <Chip label={`${totalGroups} grupos`} tone="ok" />
          <Chip label={`GMROI medio: ${gmroiAvg == null ? "-" : gmroiAvg.toFixed(2)}`} tone="warn" />
          <Chip label={`Janela: ${windowMonths}M`} tone="info" />
        </>
      );
    }
    if (isAbcSpotlight) {
      const mix = abcScopeDto?.scope.analysis_cards?.abc_mix;
      const totalGroups = Number(mix?.a_count || 0) + Number(mix?.b_count || 0) + Number(mix?.c_count || 0);
      return (
        <>
          <Chip label={`A: ${Number(mix?.a_count || 0)} (${Number(mix?.a_revenue_pct || 0).toFixed(1)}%)`} tone="ok" />
          <Chip label={`B: ${Number(mix?.b_count || 0)} (${Number(mix?.b_revenue_pct || 0).toFixed(1)}%)`} tone="warn" />
          <Chip label={`C: ${Number(mix?.c_count || 0)} (${Number(mix?.c_revenue_pct || 0).toFixed(1)}%)`} tone="crit" />
          <Chip label={`${totalGroups} grupos`} tone="info" />
        </>
      );
    }
    if (!selectedAlert) return null;
    return (
      <>
        <Chip label="Alerta" tone={alertHeaderTone} />
        <Chip label={`Estoque total: ${selectedAlert.stockTotalLabel || "-"}`} tone={alertHeaderTone} />
      </>
    );
  }, [abcScopeDto?.scope.analysis_cards?.abc_mix, alertHeaderTone, capitalScopeDto?.scope.trend_window_months, capitalScopeDto?.scope.window?.window_months, capitalSpotlightRows, isAbcSpotlight, isCapitalSpotlight, selectedAlert]);

  const openWorkspace = useCallback(
    (
      pn: string,
      context?: {
        brandFilter: string[];
        taxonomyFilter: string[];
        stockTypeFilter: string[];
        pageSize: number;
        pageIndex?: number;
        sortKey?: string;
        sortDir?: "asc" | "desc";
        stockOp?: "gte" | "eq" | "lte";
        stockValue?: number;
      }
    ) => {
      const token = encodeURIComponent(String(pn || "").trim());
      if (!token) return;
      const query = new URLSearchParams(location.search);
      if (spotlight.key) query.set("spotlight", String(spotlight.key));
      if (spotlight.extra?.source) query.set("spotlightSource", String(spotlight.extra.source));
      if (context) {
        writeFilterListToQuery(query, "spotlightBrand", context.brandFilter || []);
        writeFilterListToQuery(query, "spotlightGroup", context.taxonomyFilter || []);
        writeFilterListToQuery(query, "spotlightStockType", context.stockTypeFilter || []);
        if (context.pageSize && context.pageSize !== 20) query.set("spotlightPageSize", String(context.pageSize));
        else query.delete("spotlightPageSize");
        if (context.pageIndex != null && Number.isFinite(Number(context.pageIndex)) && Number(context.pageIndex) >= 0) {
          query.set("spotlightPageIndex", String(Math.floor(Number(context.pageIndex))));
        } else {
          query.delete("spotlightPageIndex");
        }
        if (context.sortKey) query.set("spotlightSortKey", context.sortKey);
        else query.delete("spotlightSortKey");
        if (context.sortDir) query.set("spotlightSortDir", context.sortDir);
        else query.delete("spotlightSortDir");
        if (context.stockValue != null && Number.isFinite(Number(context.stockValue))) {
          query.set("spotlightStockValue", String(context.stockValue));
          query.set("spotlightStockOp", context.stockOp || "gte");
        } else {
          query.delete("spotlightStockValue");
          query.delete("spotlightStockOp");
        }
        const drawerScrollTop = Math.max(0, Math.round(spotlightBodyRef.current?.scrollTop || 0));
        if (spotlight.key) {
          writeSpotlightTableState(String(spotlight.key), {
            brandFilter: context.brandFilter || [],
            taxonomyFilter: context.taxonomyFilter || [],
            stockTypeFilter: context.stockTypeFilter || [],
            pageSize: context.pageSize || 20,
            pageIndex: context.pageIndex != null && Number.isFinite(Number(context.pageIndex)) && Number(context.pageIndex) >= 0
              ? Math.floor(Number(context.pageIndex))
              : 0,
            sortKey: context.sortKey || undefined,
            sortDir: context.sortDir,
            stockOp: context.stockOp || "gte",
            stockValue: context.stockValue != null && Number.isFinite(Number(context.stockValue))
              ? Number(context.stockValue)
              : null,
            scrollTop: drawerScrollTop,
          });
        }
      } else {
        query.delete("spotlightBrand");
        query.delete("spotlightGroup");
        query.delete("spotlightStockType");
        query.delete("spotlightPageSize");
        query.delete("spotlightPageIndex");
        query.delete("spotlightSortKey");
        query.delete("spotlightSortDir");
        query.delete("spotlightStockValue");
        query.delete("spotlightStockOp");
        query.delete("spotlightScrollTop");
      }
      const search = query.toString();
      const fromPath = `${location.pathname}${search ? `?${search}` : ""}${location.hash}`;
      const fromScrollY = Math.max(0, Math.round(window.scrollY || window.pageYOffset || 0));
      try {
        window.sessionStorage.setItem(
          "analytics:workspace:return",
          JSON.stringify({
            from: fromPath || "/analytics/home",
            fromScrollY,
            savedAt: Date.now(),
          }),
        );
      } catch {
        // Ignore storage failures; state-based return flow remains.
      }
      navigate(`/analytics/products/${token}/overview`, {
        state: { from: fromPath || "/analytics/home", fromScrollY },
      });
    },
    [location.hash, location.pathname, location.search, navigate, spotlight.extra?.source, spotlight.key],
  );

  const syncSpotlightTableStateToUrl = useCallback(
    (context: {
      brandFilter: string[];
      taxonomyFilter: string[];
      stockTypeFilter: string[];
      pageSize: number;
      pageIndex?: number;
      sortKey?: string;
      sortDir?: "asc" | "desc";
      stockOp?: "gte" | "eq" | "lte";
      stockValue?: number;
    }) => {
      if (!spotlight.open || !spotlight.key) return;
      const drawerScrollTop = Math.max(0, Math.round(spotlightBodyRef.current?.scrollTop || 0));
      writeSpotlightTableState(String(spotlight.key), {
        brandFilter: context.brandFilter || [],
        taxonomyFilter: context.taxonomyFilter || [],
        stockTypeFilter: context.stockTypeFilter || [],
        pageSize: context.pageSize || 20,
        pageIndex: context.pageIndex != null && Number.isFinite(Number(context.pageIndex)) && Number(context.pageIndex) >= 0
          ? Math.floor(Number(context.pageIndex))
          : 0,
        sortKey: context.sortKey || undefined,
        sortDir: context.sortDir,
        stockOp: context.stockOp || "gte",
        stockValue: context.stockValue != null && Number.isFinite(Number(context.stockValue))
          ? Number(context.stockValue)
          : null,
        scrollTop: drawerScrollTop,
      });

      const query = new URLSearchParams(location.search);
      query.set("spotlight", String(spotlight.key));
      if (spotlight.extra?.source) query.set("spotlightSource", String(spotlight.extra.source));
      else query.delete("spotlightSource");

      writeFilterListToQuery(query, "spotlightBrand", context.brandFilter || []);
      writeFilterListToQuery(query, "spotlightGroup", context.taxonomyFilter || []);
      writeFilterListToQuery(query, "spotlightStockType", context.stockTypeFilter || []);
      if (context.pageSize && context.pageSize !== 20) query.set("spotlightPageSize", String(context.pageSize));
      else query.delete("spotlightPageSize");
      if (context.pageIndex != null && Number.isFinite(Number(context.pageIndex)) && Number(context.pageIndex) >= 0) {
        query.set("spotlightPageIndex", String(Math.floor(Number(context.pageIndex))));
      } else {
        query.delete("spotlightPageIndex");
      }
      if (context.sortKey) query.set("spotlightSortKey", context.sortKey);
      else query.delete("spotlightSortKey");
      if (context.sortDir) query.set("spotlightSortDir", context.sortDir);
      else query.delete("spotlightSortDir");
      if (context.stockValue != null && Number.isFinite(Number(context.stockValue))) {
        query.set("spotlightStockValue", String(context.stockValue));
        query.set("spotlightStockOp", context.stockOp || "gte");
      } else {
        query.delete("spotlightStockValue");
        query.delete("spotlightStockOp");
      }

      const nextSearch = query.toString();
      const currentSearch = String(location.search || "").replace(/^\?/, "");
      if (nextSearch === currentSearch) return;
      navigate(`${location.pathname}${nextSearch ? `?${nextSearch}` : ""}${location.hash}`, { replace: true });
    },
    [location.hash, location.pathname, location.search, navigate, spotlight.extra?.source, spotlight.key, spotlight.open],
  );

  useEffect(() => {
    if (didRestoreScrollRef.current) return;
    const state = location.state as { restoreScrollY?: unknown } | null;
    const restoreScrollY = Number(state?.restoreScrollY);
    if (!Number.isFinite(restoreScrollY) || restoreScrollY < 0) return;
    didRestoreScrollRef.current = true;
    window.requestAnimationFrame(() => {
      window.scrollTo(0, restoreScrollY);
    });
  }, [location.state]);

  useEffect(() => {
    if (!spotlight.open) return;
    const targetTop = Math.max(0, Number(spotlightTableStateFromQuery.scrollTop || 0));
    const key = `${String(spotlight.key || "")}|${location.search}`;
    if (didRestoreSpotlightScrollRef.current === key) return;
    didRestoreSpotlightScrollRef.current = key;
    window.requestAnimationFrame(() => {
      if (spotlightBodyRef.current) {
        spotlightBodyRef.current.scrollTop = targetTop;
      }
    });
  }, [location.search, spotlight.key, spotlight.open, spotlightTableStateFromQuery.scrollTop]);

  useEffect(() => {
    const next = parseSpotlightStateFromSearch(location.search);
    if (!next.open || !next.key) return;
    setSpotlight((current) => {
      const nextSource = next.extra?.source;
      if (current.open && current.key === next.key && current.extra?.source === nextSource) return current;
      return next;
    });
  }, [location.search]);

  useEffect(() => {
    if (!onRefresh) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.repeat) return;
      if (event.ctrlKey || event.metaKey || event.altKey) return;
      if (event.key.toLowerCase() !== "r") return;
      event.preventDefault();
      void onRefresh();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onRefresh]);

  return (
    <div className={styles.page}>
      <section className={styles.hero}>
        <HeroCommandCenter onOpenSpotlight={openSpotlight} updatedAtLabel={updatedAtLabel} />
        <HeroTools onOpenSpotlight={openSpotlight} onFilterChange={setActiveFilter} miniStats={model.miniStats} />
      </section>
      <BentoGrid onOpenSpotlight={openSpotlight} activeFilter={activeFilter} model={model} />
      <AnalyticsSpotlightDrawer
        open={spotlight.open}
        title={spotlightContent.title}
        meta={spotlightContent.meta}
        headerChips={spotlightHeaderChips}
        bodyRef={spotlightBodyRef}
        onClose={closeSpotlight}
      >
        {spotlightContent.signals.length && !selectedAlert ? (
          <div className={styles.drawerSignals}>
            {spotlightContent.signals.map((signal) => (
              <Chip key={`${signal.label}-${signal.tone}`} label={signal.label} tone={signal.tone} />
            ))}
          </div>
        ) : null}
        {!isListSpotlight && spotlightContent.tags.length ? (
          <div className={styles.spotTagRow}>
            {spotlightContent.tags.map((tag) => (
              <span key={tag} className={styles.skuPill}>
                {tag}
              </span>
            ))}
          </div>
        ) : null}

        {spotlight.key === "mini-actions" ? (
          <SpotlightSelectionWidget
            listTitle="Acoes disponiveis"
            items={model.allActions.map((action) => ({
              key: action.key,
              title: action.name,
              meta: action.skuCount,
              className:
                styles[`spotlightActionCode_${action.actionCode}`] || styles[`spotlightActionTone_${action.signalClass}`],
            }))}
            onSelectItem={(key) => openSpotlight(key, { source: "mini-actions" })}
          />
        ) : null}

        {spotlight.key === "mini-alerts" || spotlight.key === "stat-alerts" ? (
          <SpotlightSelectionWidget
            listTitle="Alertas disponiveis"
            items={model.allAlerts.map((alert) => ({
              key: alert.key,
              title: alert.name,
              meta: `${alert.count} SKUs`,
              className:
                styles[`spotlightAlertCode_${String(alert.code || "").toUpperCase()}`] ||
                styles[`spotlightAlertTone_${alert.toneClass}`] ||
                "",
            }))}
            onSelectItem={(key) => openSpotlight(key, { source: "mini-alerts" })}
          />
        ) : null}
        {spotlight.key === "mini-heat" ? (
          <SpotlightSelectionWidget
            listTitle="Celulas do radar"
            items={Object.values(model.heatCells)
              .filter((cell) => Number(cell.count || 0) > 0)
              .sort((a, b) => Number(b.count || 0) - Number(a.count || 0))
              .map((cell) => ({
                key: cell.key,
                title: `${cell.impact} x ${cell.urgency}`,
                meta: `${cell.count} SKUs`,
              }))}
            onSelectItem={(key) => openSpotlight(key, { source: "mini-heat" })}
          />
        ) : null}
        {spotlight.key === "mini-portfolio" ? (
          <SpotlightSelectionWidget
            listTitle="Classes do portfolio"
            items={model.portfolio
              .slice()
              .sort((a, b) => Number(b.countSkus || 0) - Number(a.countSkus || 0))
              .map((row) => ({
                key: row.key,
                title: row.label,
                meta: `${row.countSkus} SKUs • ${row.pct}%`,
              }))}
            onSelectItem={(key) => openSpotlight(key, { source: "mini-portfolio" })}
          />
        ) : null}
        {isCapitalSpotlight ? (
          <CapitalEfficiencySpotlight
            rows={capitalSpotlightRows}
            searchQuery={capitalSearchQuery}
            onSearchQueryChange={setCapitalSearchQuery}
            brandOptions={capitalBrandOptions}
            brandFilter={capitalBrandFilter}
            onBrandFilterChange={setCapitalBrandFilter}
            loading={capitalScopeLoading}
            error={capitalScopeError}
          />
        ) : null}
        {isAbcSpotlight ? (
          <AbcMixSpotlight
            rows={abcSpotlightRows}
            aMaxCumPct={Number(abcScopeDto?.scope.analysis_cards?.abc_mix.a_max_cum_pct || 80)}
            bMaxCumPct={Number(abcScopeDto?.scope.analysis_cards?.abc_mix.b_max_cum_pct || 95)}
            searchQuery={abcSearchQuery}
            onSearchQueryChange={setAbcSearchQuery}
            brandOptions={abcBrandOptions}
            brandFilter={abcBrandFilter}
            onBrandFilterChange={setAbcBrandFilter}
            loading={abcScopeLoading}
            error={abcScopeError}
          />
        ) : null}

        {selectedAction ? (
          <SpotlightSelectionWidget
            tableTitle="SKUs da acao"
            tableThirdHeader="Detalhes"
            tableRows={(selectedAction.skuDetails || []).map((row) => ({
              pn: row.pn,
              description: row.description || "-",
              brand: row.brand || "-",
              taxonomyLeafName: row.taxonomyLeafName || "-",
              stockValue: row.stockValue || "-",
              stockValueNumeric: row.stockValueNumeric ?? null,
              stockQty: row.stockQty || "-",
              stockQtyNumeric: row.stockQtyNumeric ?? null,
              third: `Urgencia: ${row.urgencyLabel} (${row.urgencyScore})`,
              thirdNumeric: null,
              details: [
                { label: "Urgencia", value: `${row.urgencyLabel} (${row.urgencyScore})` },
                ...(row.decisionState ? [{ label: "Estado", value: row.decisionState }] : []),
                ...(row.valuePriorityTier ? [{ label: "Prioridade", value: row.valuePriorityTier }] : []),
              ],
            }))}
            tableEmptyText="Sem SKUs para esta acao no snapshot atual."
            tableDefaultSort={{ key: "stockValue", dir: "desc" }}
            tableMode="details_only"
            tableDetailsHeader="Detalhes da Acao"
            tableInitialFilters={spotlightTableStateFromQuery}
            onTableStateChange={syncSpotlightTableStateToUrl}
            onOpenSku={openWorkspace}
          />
        ) : null}

        {selectedAlert ? (
          <SpotlightSelectionWidget
            tableTitle="SKUs do alerta"
            tableThirdHeader="Detalhes"
            tableRows={(selectedAlert.skuDetails || []).map((row) => ({
              pn: row.pn,
              description: row.description || "-",
              brand: row.brand || "-",
              taxonomyLeafName: row.taxonomyLeafName || "-",
              stockType: row.stockType || "-",
              stockValue: row.stockValue || "-",
              stockValueNumeric: row.stockValueNumeric ?? null,
              stockQty: row.stockQty || "-",
              stockQtyNumeric: row.stockQtyNumeric ?? null,
              financialPriority: row.financialPriority || "-",
              financialPriorityScore: row.financialPriorityScore ?? null,
              third: row.details || "-",
              thirdNumeric: null,
              details: [
                { label: "", value: row.details || "-" },
              ],
            }))}
            tableEmptyText="Sem SKUs para este alerta no snapshot atual."
            tableDefaultSort={{ key: "stockValue", dir: "desc" }}
            tableMode="details_only"
            tableDetailsHeader="Detalhes do Alerta"
            tableInitialFilters={spotlightTableStateFromQuery}
            onTableStateChange={syncSpotlightTableStateToUrl}
            onOpenSku={openWorkspace}
          />
        ) : null}
        {selectedHeatCell ? (
          <SpotlightSelectionWidget
            tableTitle="SKUs da celula"
            tableThirdHeader="Detalhes"
            tableRows={(selectedHeatCell.skuDetails || []).map((row) => ({
              pn: row.pn,
              description: row.description || "-",
              brand: row.brand || "",
              taxonomyLeafName: row.taxonomyLeafName || "",
              stockValue: row.stockValue || "-",
              stockValueNumeric: row.stockValueNumeric ?? null,
              stockQty: row.stockQty || "-",
              stockQtyNumeric: row.stockQtyNumeric ?? null,
              third: row.details || "-",
              thirdNumeric: null,
              details: [
                { label: "", value: row.details || "-" },
              ],
            }))}
            tableEmptyText="Sem SKUs para esta celula no snapshot atual."
            tableDefaultSort={{ key: "stockValue", dir: "desc" }}
            tableMode="details_only"
            tableDetailsHeader="Detalhes do Radar"
            tableInitialFilters={spotlightTableStateFromQuery}
            onTableStateChange={syncSpotlightTableStateToUrl}
            onOpenSku={openWorkspace}
          />
        ) : null}
        {selectedPortfolio ? (
          <SpotlightSelectionWidget
            tableTitle="SKUs da classe"
            tableThirdHeader="Detalhes"
            tableRows={(selectedPortfolio.skuDetails || []).map((row) => ({
              pn: row.pn,
              description: row.description || "-",
              third: row.details || "-",
              thirdNumeric: null,
            }))}
            tableEmptyText="Sem SKUs para esta classe no snapshot atual."
            tableDefaultSort={{ key: "pn", dir: "asc" }}
            tableInitialFilters={spotlightTableStateFromQuery}
            onTableStateChange={syncSpotlightTableStateToUrl}
            onOpenSku={openWorkspace}
          />
        ) : null}

        {!isListSpotlight ? <SpotlightCallout title="Por que isso apareceu?" text={spotlightContent.why} /> : null}
        <SpotlightNextSteps items={spotlightContent.nextSteps} />
        {!isListSpotlight ? <SpotlightSkuList skus={spotlightContent.skus} /> : null}
      </AnalyticsSpotlightDrawer>
      <button
        type="button"
        className={`${styles.fab} ${isRefreshing ? styles.fabSpinning : ""}`}
        title="Atualizar dados (R)"
        aria-label="Atualizar dados"
        aria-busy={isRefreshing}
        onClick={() => void onRefresh?.()}
        disabled={isRefreshing}
      >
        {"\u27f3"}
      </button>
    </div>
  );
}
