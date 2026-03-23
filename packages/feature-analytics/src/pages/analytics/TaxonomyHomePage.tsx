// @ts-nocheck
import { ApiClientError } from "../../app/apiClient";
import {
  makeAnalyticsTaxonomyScopeOverviewV1Dto,
  type AnalyticsTaxonomyScopeOverviewV1Dto,
} from "../../legacy_dto";
import { FilterDropdown } from "../../components/ui/FilterDropdown";
import { SelectMenu, type SelectMenuOption } from "../../components/ui/SelectMenu";
import { createSpotlightSelectClassNames } from "../../components/ui/spotlightSelect";
import {
  Chart as ChartJSCore,
  Tooltip,
  Legend,
  type ChartData,
  type ChartOptions,
} from "chart.js";
import { useEffect, useMemo, useRef, useState } from "react";
import { Chart } from "react-chartjs-2";
import { TreemapController, TreemapElement } from "chartjs-chart-treemap";
import {
  ResponsiveContainer,
  ComposedChart,
  ScatterChart,
  Scatter,
  CartesianGrid,
  XAxis,
  YAxis,
  ZAxis,
  Tooltip as ReTooltip,
  Legend as ReLegend,
  Bar as ReBar,
  Line as ReLine,
} from "recharts";

import { useAppSession } from "../../app/providers/AppProviders";
import copy from "./copy/products_overview_narratives.pt_BR.json";
import { AnalyticsConcentrationSection } from "./components/AnalyticsConcentrationSection";
import { AnalyticsPerformanceSection, type AnalyticsPerformanceItem } from "./components/AnalyticsPerformanceSection";
import { SpotlightDataTable, type SpotlightDataTableColumn } from "./components/SpotlightDataTable";
import { KpiCard } from "./components/products/overview/KpiCard";
import { AnalyticsSpotlightDrawer } from "../analytics_home/components/AnalyticsSpotlightDrawer";
import styles from "./taxonomy_home.module.css";
import { taxonomyEmojiForNodeName } from "./taxonomy_visuals";

ChartJSCore.register(
  TreemapController,
  TreemapElement,
  Tooltip,
  Legend,
);

type KpiCard = {
  label: string;
  value: string;
  change: string;
  tone: "positive" | "negative";
  helpItems: string[];
  showChangeArrow?: boolean;
};

type TopNode = {
  nodeId: number;
  name: string;
  revenueBrl: number;
  revenue: string;
  sharePct: number;
  marginPct: number | null;
  trend: {
    source: "DB_FALLBACK" | "SNAPSHOT";
    basis: string;
    window_months?: number;
    target_month_ref: string | null;
    base_month_ref: string | null;
    revenue_delta_mom_pct: number | null;
    margin_delta_mom_pp: number | null;
    share_delta_mom_pp: number | null;
    is_available: boolean;
  };
  tone: "a" | "b" | "c" | "d" | "e" | "f";
};

type ActionCard = {
  priority: string;
  title: string;
  text: string;
  cta: string;
  tone: "primary" | "green" | "red" | "neutral";
};

type LevelOption = {
  level: number;
  label: string;
};

type EfficiencyChartRow = {
  node_short: string;
  node_name: string;
  margem_bruta_brl: number;
  capital_brl: number;
  gmroi: number | null;
};

type EfficiencySpotlightRow = {
  node_name: string;
  revenue_brl: number;
  margem_bruta_brl: number;
  capital_brl: number;
  gmroi: number | null;
  risk_pct: number;
  risk_level: "low" | "medium" | "high";
};

type EfficiencyBarShapeProps = {
  x?: number;
  y?: number;
  width?: number;
  height?: number;
  fill?: string;
};

type EfficiencyTooltipEntry = {
  name?: string;
  value?: number | null;
  color?: string;
  payload?: EfficiencyChartRow;
};

type AllocationSpotlightRow = {
  node_id: number;
  node_name: string;
  sku_count: number;
  capital_brl: number;
  capital_share_pct: number;
  risk_level: "low" | "medium" | "high";
  risk_pct: number;
  revenue_brl: number;
  margin_pct: number | null;
  share_pct: number;
  priority_score: number;
};

type AbcSpotlightRow = {
  node_id: number;
  node_name: string;
  revenue_brl: number;
  gross_margin_brl: number;
  share_pct: number;
  margin_pct: number | null;
  band: "A" | "B" | "C";
  cum_share_pct: number;
};

type AbcSpotlightCurveType = "revenue" | "gross_margin";
type AbcSpotlightBandFilter = "all" | "A" | "B" | "C";
type TopMarginSpotlightCurveType = "gross_margin" | "margin_pct" | "gmroi";

type TopMarginSpotlightRow = {
  node_id: number;
  node_name: string;
  gross_margin_brl: number;
  revenue_brl: number;
  margin_pct: number | null;
  capital_brl: number;
  gmroi: number | null;
  share_margin_pct: number;
  trend?: {
    source: "SNAPSHOT" | "DB_FALLBACK";
    basis: "MoM" | "WINDOW_EDGE" | "SNAPSHOT_MOM";
    window_months: number;
    target_month_ref: string | null;
    base_month_ref: string | null;
    revenue_delta_mom_pct: number | null;
    margin_delta_mom_pp: number | null;
    share_delta_mom_pp: number | null;
    is_available: boolean;
  };
};

type RiskSpotlightMetric = "capital_at_risk_brl" | "risk_pct";

type RiskSpotlightRow = {
  node_id: number;
  node_name: string;
  risk_level: "low" | "medium" | "high";
  risk_pct: number;
  capital_at_risk_brl: number;
  capital_brl: number;
  revenue_brl: number;
  margin_pct: number | null;
  gmroi: number | null;
  financial_risk_priority_brl: number;
};

type SpotlightNumericOp = "gte" | "eq" | "lte";
type SpotlightSortDir = "asc" | "desc";

const PERFORMANCE_LEVEL_MAX_ITEMS = 6;
const TREND_WINDOW_OPTIONS = [1, 3, 6, 12] as const;
const BOTTOM_TABLE_MAX_ROWS = 5;
const EFFICIENCY_BAR_PAIR_GAP_PX = 3;
const EFFICIENCY_SPOTLIGHT_CHART_MAX_ITEMS = 16;
const ABC_PARETO_MARGIN = { top: 14, right: 12, left: 0, bottom: 8 } as const;
const ABC_PARETO_X_AXIS_HEIGHT = 56;
const ABC_PARETO_Y_AXIS_WIDTH = 52;
const SPOTLIGHT_SELECT_CLASSNAMES = createSpotlightSelectClassNames({
  wrap: styles.spotlightFilterSelectWrap,
});
const HEADER_SELECT_CLASSNAMES = createSpotlightSelectClassNames({
  wrap: styles.headerSelectWrap,
});
const SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES = createSpotlightSelectClassNames({
  wrap: styles.spotlightOperatorSelectWrap,
});

function asBrlCompact(value: number): string {
  const abs = Math.abs(value);
  if (abs >= 1_000_000) return `R$ ${(value / 1_000_000).toLocaleString("pt-BR", { maximumFractionDigits: 1 })}M`;
  if (abs >= 1_000) return `R$ ${(value / 1_000).toLocaleString("pt-BR", { maximumFractionDigits: 1 })}K`;
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL", maximumFractionDigits: 0 }).format(value);
}

function asPct(value: number): string {
  return `${value.toLocaleString("pt-BR", { maximumFractionDigits: 1, minimumFractionDigits: 1 })}%`;
}

function kpiTitleByLevel(levelLabel: string): string {
  const token = String(levelLabel || "").toLowerCase();
  if (token.includes("categoria")) return "Categorias ativas";
  if (token.includes("sub")) return "Subcategorias ativas";
  return "Grupos ativos";
}

function trendPeriodLabel(windowMonths: number): string {
  return windowMonths === 1 ? "mes anterior" : `${windowMonths} meses`;
}

function toTrendText(
  metric: { trend?: { delta_mom_pct: number | null; is_available: boolean; window_months?: number } },
  fallbackWindowMonths: number,
): string {
  const trend = metric.trend;
  if (!trend?.is_available || trend.delta_mom_pct == null) return "";
  const windowMonths = Number(trend.window_months || fallbackWindowMonths || 1);
  const val = trend.delta_mom_pct;
  return `${val >= 0 ? "+" : ""}${val.toFixed(1)}% vs ${trendPeriodLabel(windowMonths)}`;
}

function trendTone(
  metric: { trend?: { delta_mom_pct: number | null; is_available: boolean } },
  mode: "positive_when_up" | "negative_when_up",
  fallback: "positive" | "negative" = "positive",
): "positive" | "negative" {
  const trend = metric.trend;
  if (!trend?.is_available || trend.delta_mom_pct == null) return "negative";
  if (mode === "negative_when_up") return trend.delta_mom_pct > 0 ? "negative" : "positive";
  return trend.delta_mom_pct >= 0 ? "positive" : "negative";
}

function hasTrend(metric: { trend?: { is_available: boolean } }): boolean {
  return Boolean(metric.trend?.is_available);
}

function resolveKpiHelpItems(
  key:
    | "capital_imobilizado"
    | "capital_em_risco"
    | "receita_potencial_interna"
    | "receita_potencial_mercado"
    | "receita_bruta_6m",
): string[] {
  const source = (copy as { kpi_help?: Record<string, { items?: unknown }> }).kpi_help;
  const bucket = source?.[key];
  const items = Array.isArray(bucket?.items) ? bucket.items : [];
  return items.map((item) => String(item || "").trim()).filter(Boolean);
}

function grossRevenueHelpItems(windowMonths: number): string[] {
  return [
    `Receita total agregada da janela selecionada (${windowMonths}M) no escopo atual.`,
    "Consolidada por nivel das classificacoes com os mesmos filtros aplicados na tela.",
    "Use para comparar escala de receita entre os niveis (grupo/categoria/subcategoria).",
  ];
}

function activeEntitiesHelpItems(levelLabel: string): string[] {
  const noun = kpiTitleByLevel(levelLabel).toLowerCase();
  return [
    `Total de ${noun} no recorte atual (com filtros).`,
    "Conta entidades distintas presentes no nivel selecionado das classificacoes.",
    "Use para medir cobertura real da analise por nivel hierarquico.",
  ];
}

function buildTaxonomyScopeCacheKey(level: number, windowMonths: number): string {
  return JSON.stringify({ level, windowMonths });
}

function trendClassBySign(value: number | null): string {
  if (value == null) return styles.metricTrendDown;
  return value >= 0 ? styles.metricTrendUp : styles.metricTrendDown;
}

function trendPctText(value: number | null): string | null {
  if (value == null) return null;
  const arrow = value >= 0 ? "↑" : "↓";
  const magnitude = Math.abs(value).toFixed(1).replace(".", ",");
  return `${arrow} ${magnitude}%`;
}

function trendPpText(value: number | null): string | null {
  if (value == null) return null;
  const arrow = value >= 0 ? "↑" : "↓";
  const magnitude = Math.abs(value).toFixed(1).replace(".", ",");
  return `${arrow} ${magnitude}%`;
}

function tableTrendPctCompact(value: number | null): string | null {
  if (value == null) return null;
  const arrow = value >= 0 ? "↑" : "↓";
  const magnitude = Math.abs(value).toFixed(1).replace(".", ",");
  return `${arrow}${magnitude}%`;
}

function riskFillColor(riskLevel: string): string {
  if (riskLevel === "high") return "#b91c1c";
  if (riskLevel === "medium") return "#eab308";
  return "#0f766e";
}

function riskLabelColor(riskLevel: string): string {
  if (riskLevel === "high") return "#fff7ed";
  if (riskLevel === "medium") return "#422006";
  return "#ecfeff";
}

function riskLevelFromPct(riskPct: number): "low" | "medium" | "high" {
  if (riskPct >= 30) return "high";
  if (riskPct >= 15) return "medium";
  return "low";
}

function shortNodeLabel(label: string, max = 14): string {
  const raw = String(label || "").trim();
  if (raw.length <= max) return raw;
  return `${raw.slice(0, Math.max(0, max - 1))}…`;
}

function percentile(values: number[], q: number): number | null {
  if (!values.length) return null;
  const vv = [...values].filter((v) => Number.isFinite(v)).sort((a, b) => a - b);
  if (!vv.length) return null;
  if (vv.length === 1) return vv[0];
  const qq = Math.min(1, Math.max(0, q));
  const pos = qq * (vv.length - 1);
  const lo = Math.floor(pos);
  const hi = Math.min(vv.length - 1, lo + 1);
  const frac = pos - lo;
  return vv[lo] * (1 - frac) + vv[hi] * frac;
}

function normalizeSpotlightNumericOp(value: string): SpotlightNumericOp {
  if (value === "eq" || value === "lte") return value;
  return "gte";
}

function normalizeSpotlightSortDir(value: string): SpotlightSortDir {
  return value === "asc" ? "asc" : "desc";
}

function clampSharePctInput(raw: string): string {
  const normalized = String(raw || "").trim().replace(",", ".");
  if (!normalized) return "0";
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed)) return "0";
  if (parsed < 0) return "0";
  if (parsed > 100) return "100";
  return String(parsed);
}

function clampNonNegativeDecimalInput(raw: string): string {
  const normalized = String(raw || "").trim().replace(",", ".");
  if (!normalized) return "";
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed)) return "";
  if (parsed < 0) return "0";
  if (parsed > 9999) return "9999";
  return normalized;
}

function clampTopProductsInput(raw: string): string {
  const normalized = String(raw || "").trim().replace(",", ".");
  if (!normalized) return "0";
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed)) return "0";
  if (parsed < 0) return "0";
  if (parsed > 999) return "999";
  return String(Math.floor(parsed));
}

function toggleMultiValue(current: string[], value: string): string[] {
  if (!value || value === "all") return [];
  if (current.includes(value)) return current.filter((item) => item !== value);
  return [...current, value];
}

function mapEfficiencySpotlightRow(item: Record<string, unknown>): EfficiencySpotlightRow {
  const riskPct = Number(item.risk_pct || 0);
  return {
    node_name: String(item.node_name || ""),
    revenue_brl: Number(item.revenue_brl || 0),
    margem_bruta_brl: Number((item.gross_margin_brl as number | undefined) ?? item.margin_brl ?? 0),
    capital_brl: Number(item.capital_brl || 0),
    gmroi: item.gmroi == null ? null : Number(item.gmroi),
    risk_pct: riskPct,
    risk_level: riskLevelFromPct(riskPct),
  };
}

type RiskScatterShapeProps = {
  cx?: number;
  cy?: number;
  payload?: RiskSpotlightRow;
};

type EfficiencyXAxisTickProps = {
  x?: number;
  y?: number;
  payload?: {
    value?: string;
    payload?: EfficiencyChartRow;
  };
};

export function TaxonomyHomePage() {
  const { api, getTaxonomyScopeSnapshot, setTaxonomyScopeSnapshot } = useAppSession();
  const [levels, setLevels] = useState<LevelOption[]>([]);
  const [selectedLevel, setSelectedLevel] = useState<number>(0);
  const [selectedTrendWindowMonths, setSelectedTrendWindowMonths] = useState<number>(6);
  const [dto, setDto] = useState<AnalyticsTaxonomyScopeOverviewV1Dto | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [refreshingMessage, setRefreshingMessage] = useState("");
  const [isPerformanceSpotlightOpen, setIsPerformanceSpotlightOpen] = useState(false);
  const [isAllocationSpotlightOpen, setIsAllocationSpotlightOpen] = useState(false);
  const [isEfficiencySpotlightOpen, setIsEfficiencySpotlightOpen] = useState(false);
  const [isAbcSpotlightOpen, setIsAbcSpotlightOpen] = useState(false);
  const [isTopMarginSpotlightOpen, setIsTopMarginSpotlightOpen] = useState(false);
  const [isRiskSpotlightOpen, setIsRiskSpotlightOpen] = useState(false);
  const [abcSpotlightSearch, setAbcSpotlightSearch] = useState("");
  const [abcSpotlightGroupFilters, setAbcSpotlightGroupFilters] = useState<string[]>([]);
  const [abcSpotlightCurveType, setAbcSpotlightCurveType] = useState<AbcSpotlightCurveType>("revenue");
  const [abcSpotlightBandFilter, setAbcSpotlightBandFilter] = useState<AbcSpotlightBandFilter>("all");
  const [topMarginSpotlightSearch, setTopMarginSpotlightSearch] = useState("");
  const [topMarginSpotlightGroupFilters, setTopMarginSpotlightGroupFilters] = useState<string[]>([]);
  const [topMarginSpotlightCurveType, setTopMarginSpotlightCurveType] = useState<TopMarginSpotlightCurveType>("gross_margin");
  const [riskSpotlightSearch, setRiskSpotlightSearch] = useState("");
  const [riskSpotlightGroupFilters, setRiskSpotlightGroupFilters] = useState<string[]>([]);
  const [riskSpotlightRiskFilter, setRiskSpotlightRiskFilter] = useState<"all" | "high" | "medium" | "low">("all");
  const [riskSpotlightMetric, setRiskSpotlightMetric] = useState<RiskSpotlightMetric>("capital_at_risk_brl");
  const [riskSpotlightMetricOp, setRiskSpotlightMetricOp] = useState<SpotlightNumericOp>("gte");
  const [riskSpotlightMetricInput, setRiskSpotlightMetricInput] = useState("");
  const [efficiencySpotlightSearch, setEfficiencySpotlightSearch] = useState("");
  const [efficiencySpotlightGroupFilters, setEfficiencySpotlightGroupFilters] = useState<string[]>([]);
  const [efficiencySpotlightGmroiOp, setEfficiencySpotlightGmroiOp] = useState<SpotlightNumericOp>("gte");
  const [efficiencySpotlightGmroiInput, setEfficiencySpotlightGmroiInput] = useState("");
  const [spotlightRevenueSortDir, setSpotlightRevenueSortDir] = useState<SpotlightSortDir>("desc");
  const [spotlightMarginSortDir, setSpotlightMarginSortDir] = useState<SpotlightSortDir>("desc");
  const [spotlightShareOp, setSpotlightShareOp] = useState<SpotlightNumericOp>("gte");
  const [spotlightShareTargetPctInput, setSpotlightShareTargetPctInput] = useState("100");
  const [spotlightTopProductsOp, setSpotlightTopProductsOp] = useState<SpotlightNumericOp>("gte");
  const [spotlightTopProductsInput, setSpotlightTopProductsInput] = useState("0");
  const [spotlightGroupFilter, setSpotlightGroupFilter] = useState("all");
  const [spotlightBrandFilter, setSpotlightBrandFilter] = useState("all");
  const [spotlightSearch, setSpotlightSearch] = useState("");
  const [allocationSpotlightSearch, setAllocationSpotlightSearch] = useState("");
  const [allocationSpotlightRiskFilter, setAllocationSpotlightRiskFilter] = useState<"all" | "high" | "medium" | "low">("all");
  const [allocationSpotlightShareOp, setAllocationSpotlightShareOp] = useState<SpotlightNumericOp>("gte");
  const [allocationSpotlightShareTargetInput, setAllocationSpotlightShareTargetInput] = useState("100");
  const [allocationSpotlightTopOp, setAllocationSpotlightTopOp] = useState<SpotlightNumericOp>("gte");
  const [allocationSpotlightTopInput, setAllocationSpotlightTopInput] = useState("0");
  const [allocationSpotlightSelectedNodeId, setAllocationSpotlightSelectedNodeId] = useState<number | null>(null);
  const allocationPriorityRef = useRef<HTMLElement | null>(null);
  const [allocationTopHeight, setAllocationTopHeight] = useState(320);
  const abcTopBandsRef = useRef<HTMLElement | null>(null);
  const [abcChartHeight, setAbcChartHeight] = useState(300);
  const efficiencySpotlightBodyRef = useRef<HTMLDivElement | null>(null);
  const scopeCacheKey = useMemo(
    () => buildTaxonomyScopeCacheKey(selectedLevel, selectedTrendWindowMonths),
    [selectedLevel, selectedTrendWindowMonths],
  );
  const levelSelectOptions = useMemo<SelectMenuOption[]>(
    () =>
      levels.map((opt) => ({
        value: String(opt.level),
        label: opt.label,
      })),
    [levels],
  );
  const trendWindowSelectOptions = useMemo<SelectMenuOption[]>(
    () => TREND_WINDOW_OPTIONS.map((months) => ({ value: String(months), label: `${months}M` })),
    [],
  );

  useEffect(() => {
    let disposed = false;
    async function loadLevels() {
      try {
        const env = await api.taxonomy.levels({ enabled_only: true });
        if (disposed) return;
        const rows = Array.isArray(env.data?.rows) ? env.data.rows : [];
        const options = rows
          .map((item) => ({
            level: Number(item?.level ?? 0),
            label: String(item?.label || "").trim() || `Hierarquia ${Number(item?.level ?? 0) + 1}`,
          }))
          .filter((item) => Number.isFinite(item.level))
          .sort((a, b) => a.level - b.level);
        const nextLevels = options.length ? options : [{ level: 0, label: "Grupo" }, { level: 1, label: "Categoria" }, { level: 2, label: "Subcategoria" }];
        setLevels(nextLevels);
        setSelectedLevel((curr) => (nextLevels.some((it) => it.level === curr) ? curr : nextLevels[0].level));
      } catch {
        if (disposed) return;
        const fallback = [{ level: 0, label: "Grupo" }, { level: 1, label: "Categoria" }, { level: 2, label: "Subcategoria" }];
        setLevels(fallback);
        setSelectedLevel((curr) => (fallback.some((it) => it.level === curr) ? curr : 0));
      }
    }
    void loadLevels();
    return () => {
      disposed = true;
    };
  }, [api.taxonomy]);

  useEffect(() => {
    const cached = getTaxonomyScopeSnapshot(scopeCacheKey);
    if (!cached) return;
    setDto(cached.data);
    if (cached.data.status === "refreshing") {
      setRefreshingMessage(cached.data.message || "Atualizacao de snapshot em andamento. Aguarde e atualize a pagina.");
    }
  }, [getTaxonomyScopeSnapshot, scopeCacheKey]);

  useEffect(() => {
    let disposed = false;
    async function loadScope() {
      setLoading(true);
      setError("");
      setRefreshingMessage("");
      try {
        const env = await api.analytics.workspaceTaxonomyScope({
          level: selectedLevel,
          windowMonths: selectedTrendWindowMonths,
          limit: 50,
          offset: 0,
        });
        if (disposed) return;
        const mapped = makeAnalyticsTaxonomyScopeOverviewV1Dto(env.data, "current");
        setDto(mapped);
        setTaxonomyScopeSnapshot(scopeCacheKey, mapped);
        if (mapped.status === "refreshing") {
          setRefreshingMessage(mapped.message || "Atualizacao de snapshot em andamento. Aguarde e atualize a pagina.");
        }
      } catch (err) {
        if (disposed) return;
        const apiErr = err instanceof ApiClientError ? err : null;
        setError(apiErr?.message || (err instanceof Error ? err.message : String(err)));
        setDto(null);
      } finally {
        if (!disposed) setLoading(false);
      }
    }
    void loadScope();
    return () => {
      disposed = true;
    };
  }, [api.analytics, scopeCacheKey, selectedLevel, selectedTrendWindowMonths, setTaxonomyScopeSnapshot]);

  const levelLabel = dto?.scope.level_label || levels.find((it) => it.level === selectedLevel)?.label || "Grupo";
  const leaf0Label = levels.find((it) => it.level === 0)?.label || "Grupo";
  const leaf0LabelLower = leaf0Label.toLocaleLowerCase("pt-BR");
  const leaf0LabelPlural = leaf0LabelLower.endsWith("s") ? leaf0Label : `${leaf0Label}s`;
  const leaf0LabelPluralLower = leaf0LabelPlural.toLocaleLowerCase("pt-BR");
  const kpis: KpiCard[] = useMemo(() => {
    if (!dto) return [];
    const trendWindowMonths = dto.scope.window?.window_months || dto.scope.trend_window_months || selectedTrendWindowMonths;
    const active = dto.kpis.active_entities;
    const grossRevenue = dto.kpis.gross_revenue_6m_brl;
    const marginTotal = dto.kpis.margin_total_brl;
    const marginPct = dto.kpis.margin_pct;
    const capTotal = dto.kpis.capital_total_brl;
    const capRisk = dto.kpis.capital_at_risk_brl;
    const potInternal = dto.kpis.potential_revenue_internal_brl;
    const potMarket = dto.kpis.potential_revenue_market_brl;
    const capRiskText = toTrendText(capRisk, trendWindowMonths);
    const marginText = `Margem: ${asBrlCompact(marginTotal.value || 0)} (${asPct(marginPct.value || 0)})`;
    return [
      {
        label: kpiTitleByLevel(levelLabel),
        value: `${Math.round(active.value || 0).toLocaleString("pt-BR")}`,
        change: toTrendText(active, trendWindowMonths),
        tone: trendTone(active, "positive_when_up"),
        helpItems: activeEntitiesHelpItems(levelLabel),
        showChangeArrow: hasTrend(active),
      },
      {
        label: "Receita Bruta",
        value: asBrlCompact(grossRevenue.value || 0),
        change: marginText,
        tone: "positive",
        helpItems: grossRevenueHelpItems(trendWindowMonths),
        showChangeArrow: false,
      },
      {
        label: "Capital imobilizado",
        value: asBrlCompact(capTotal.value || 0),
        change: toTrendText(capTotal, trendWindowMonths),
        tone: trendTone(capTotal, "positive_when_up", capTotal.flags?.is_imputed ? "negative" : "positive"),
        helpItems: resolveKpiHelpItems("capital_imobilizado"),
        showChangeArrow: hasTrend(capTotal),
      },
      {
        label: "Capital em risco",
        value: asBrlCompact(capRisk.value || 0),
        change: capRiskText,
        tone: trendTone(capRisk, "negative_when_up"),
        helpItems: resolveKpiHelpItems("capital_em_risco"),
        showChangeArrow: hasTrend(capRisk),
      },
      {
        label: "Receita potencial interna",
        value: asBrlCompact(potInternal.value || 0),
        change: toTrendText(potInternal, trendWindowMonths),
        tone: trendTone(potInternal, "positive_when_up", potInternal.flags?.is_imputed ? "negative" : "positive"),
        helpItems: resolveKpiHelpItems("receita_potencial_interna"),
        showChangeArrow: hasTrend(potInternal),
      },
      {
        label: "Receita potencial mercado",
        value: asBrlCompact(potMarket.value || 0),
        change: toTrendText(potMarket, trendWindowMonths),
        tone: trendTone(potMarket, "positive_when_up", potMarket.flags?.is_imputed ? "negative" : "positive"),
        helpItems: resolveKpiHelpItems("receita_potencial_mercado"),
        showChangeArrow: hasTrend(potMarket),
      },
    ];
  }, [dto, levelLabel, selectedTrendWindowMonths]);

  const topNodes: TopNode[] = useMemo(() => {
    const rows = dto?.panels.top_nodes_by_revenue || [];
    const tones: Array<TopNode["tone"]> = ["a", "b", "c", "d", "e", "f"];
    return rows.slice(0, PERFORMANCE_LEVEL_MAX_ITEMS).map((row, idx) => ({
      nodeId: row.node_id,
      name: row.node_name,
      revenueBrl: row.revenue_brl,
      revenue: asBrlCompact(row.revenue_brl),
      sharePct: row.share_pct,
      marginPct: row.margin_pct,
      trend: row.trend,
      tone: tones[idx] || "f",
    }));
  }, [dto]);
  const performanceSectionItems = useMemo<AnalyticsPerformanceItem[]>(() => (
    topNodes.map((row) => ({
      id: String(row.nodeId),
      name: row.name,
      icon: taxonomyEmojiForNodeName(row.name),
      metrics: [
        {
          label: "Receita",
          value: row.revenue,
          delta: trendPctText(row.trend.revenue_delta_mom_pct),
          deltaTone: row.trend.revenue_delta_mom_pct != null && row.trend.revenue_delta_mom_pct >= 0 ? "positive" : "negative",
        },
        {
          label: "Margem (%)",
          value: row.marginPct == null ? "-" : asPct(row.marginPct),
          delta: trendPpText(row.trend.margin_delta_mom_pp),
          deltaTone: row.trend.margin_delta_mom_pp != null && row.trend.margin_delta_mom_pp >= 0 ? "positive" : "negative",
        },
        {
          label: "Participacao",
          value: asPct(row.sharePct),
          delta: trendPpText(row.trend.share_delta_mom_pp),
          deltaTone: row.trend.share_delta_mom_pp != null && row.trend.share_delta_mom_pp >= 0 ? "positive" : "negative",
        },
      ],
    }))
  ), [topNodes]);
  const performanceSpotlightRows = useMemo(() => {
    const parseNumberOrZero = (raw: string): number => {
      const normalized = String(raw || "").trim().replace(",", ".");
      if (!normalized) return 0;
      const parsed = Number(normalized);
      return Number.isFinite(parsed) ? parsed : 0;
    };
    const shareTarget = Math.min(100, Math.max(0, parseNumberOrZero(spotlightShareTargetPctInput)));
    const topProductsTarget = Math.max(0, Math.floor(parseNumberOrZero(spotlightTopProductsInput)));
    const searchToken = String(spotlightSearch || "").trim().toLocaleLowerCase("pt-BR");
    const byRevenueDesc = [...(dto?.panels.top_nodes_by_revenue || [])]
      .map((row) => {
        const brand =
          String(
            (row as { brand?: string; marca?: string }).brand ??
            (row as { brand?: string; marca?: string }).marca ??
            "",
          ).trim() || "Sem marca";
        return { ...row, _brand: brand };
      })
      .sort((a, b) => b.revenue_brl - a.revenue_brl);
    const filtered = byRevenueDesc.filter((row) => {
      const groupOk = spotlightGroupFilter === "all" ? true : row.node_name === spotlightGroupFilter;
      const brandOk = spotlightBrandFilter === "all" ? true : row._brand === spotlightBrandFilter;
      const searchOk = searchToken
        ? `${String(row.node_name || "").toLocaleLowerCase("pt-BR")} ${String(row._brand || "").toLocaleLowerCase("pt-BR")}`.includes(searchToken)
        : true;
      return groupOk && brandOk && searchOk;
    });
    const byRevenueAsc = [...filtered].sort((a, b) => Number(a.revenue_brl || 0) - Number(b.revenue_brl || 0));

    const selectByCumulativeShare = (): typeof filtered => {
      if (!filtered.length) return [];
      if (shareTarget >= 100) return filtered;
      if (spotlightShareOp === "gte") {
        if (shareTarget <= 0) return filtered;
        const selected: typeof filtered = [];
        let cumulativeShare = 0;
        for (const row of filtered) {
          selected.push(row);
          cumulativeShare += Math.max(0, Number(row.share_pct || 0));
          if (cumulativeShare >= shareTarget) break;
        }
        return selected;
      }
      if (spotlightShareOp === "lte") {
        const selected: typeof filtered = [];
        let cumulativeShare = 0;
        for (const row of byRevenueAsc) {
          const next = cumulativeShare + Math.max(0, Number(row.share_pct || 0));
          if (next > shareTarget) break;
          selected.push(row);
          cumulativeShare = next;
        }
        return selected;
      }
      let cumulativeShare = 0;
      let bestCount = 0;
      let bestDiff = Number.POSITIVE_INFINITY;
      for (let index = 0; index < filtered.length; index += 1) {
        cumulativeShare += Math.max(0, Number(filtered[index].share_pct || 0));
        const diff = Math.abs(cumulativeShare - shareTarget);
        if (diff < bestDiff) {
          bestDiff = diff;
          bestCount = index + 1;
        }
      }
      return filtered.slice(0, Math.max(0, bestCount));
    };

    const selectedRows = selectByCumulativeShare();
    const selectByTopProducts = (): typeof selectedRows => {
      if (!selectedRows.length) return [];
      if (topProductsTarget <= 0) return selectedRows;
      const byRevenueDescSelected = [...selectedRows].sort((a, b) => Number(b.revenue_brl || 0) - Number(a.revenue_brl || 0));
      if (spotlightTopProductsOp === "gte") {
        return byRevenueDescSelected.slice(0, topProductsTarget);
      }
      if (spotlightTopProductsOp === "lte") {
        const byRevenueAscSelected = [...selectedRows].sort((a, b) => Number(a.revenue_brl || 0) - Number(b.revenue_brl || 0));
        return byRevenueAscSelected.slice(0, topProductsTarget);
      }
      return byRevenueDescSelected.slice(0, topProductsTarget);
    };
    const topProductsRows = selectByTopProducts();
    const marginSortValue = (value: number | null): number => {
      if (value == null) return spotlightMarginSortDir === "asc" ? Number.POSITIVE_INFINITY : Number.NEGATIVE_INFINITY;
      return Number(value);
    };
    return [...topProductsRows].sort((a, b) => {
      const revenueCmp = spotlightRevenueSortDir === "asc"
        ? Number(a.revenue_brl || 0) - Number(b.revenue_brl || 0)
        : Number(b.revenue_brl || 0) - Number(a.revenue_brl || 0);
      if (revenueCmp !== 0) return revenueCmp;
      const marginCmp = spotlightMarginSortDir === "asc"
        ? marginSortValue(a.margin_pct) - marginSortValue(b.margin_pct)
        : marginSortValue(b.margin_pct) - marginSortValue(a.margin_pct);
      if (marginCmp !== 0) return marginCmp;
      return String(a.node_name || "").localeCompare(String(b.node_name || ""), "pt-BR");
    });
  }, [
    dto,
    spotlightRevenueSortDir,
    spotlightMarginSortDir,
    spotlightShareOp,
    spotlightShareTargetPctInput,
    spotlightTopProductsOp,
    spotlightTopProductsInput,
    spotlightGroupFilter,
    spotlightBrandFilter,
    spotlightSearch,
  ]);
  const performanceSpotlightGroupOptions = useMemo(
    () => ["all", ...Array.from(new Set((dto?.panels.top_nodes_by_revenue || []).map((row) => row.node_name).filter(Boolean)))] as string[],
    [dto],
  );
  const performanceSpotlightGroupSelectOptions = useMemo<SelectMenuOption[]>(
    () =>
      performanceSpotlightGroupOptions.map((option) => ({
        value: option,
        label: option === "all" ? `Todos os ${leaf0LabelPluralLower}` : option,
      })),
    [leaf0LabelPluralLower, performanceSpotlightGroupOptions],
  );
  const performanceSpotlightBrandOptions = useMemo(() => {
    const set = new Set<string>();
    for (const row of dto?.panels.top_nodes_by_revenue || []) {
      const brand =
        String(
          (row as { brand?: string; marca?: string }).brand ??
          (row as { brand?: string; marca?: string }).marca ??
          "",
        ).trim() || "Sem marca";
      set.add(brand);
    }
    return ["all", ...Array.from(set).sort((a, b) => a.localeCompare(b, "pt-BR"))];
  }, [dto]);
  const performanceSpotlightBrandSelectOptions = useMemo<SelectMenuOption[]>(
    () =>
      performanceSpotlightBrandOptions.map((option) => ({
        value: option,
        label: option === "all" ? "Todas as marcas" : option,
      })),
    [performanceSpotlightBrandOptions],
  );
  const performanceSpotlightShareOperatorOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: ">=", value: "gte" },
      { label: "=", value: "eq" },
      { label: "<=", value: "lte" },
    ],
    [],
  );
  const performanceSpotlightSortDirOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Maior -> Menor", value: "desc" },
      { label: "Menor -> Maior", value: "asc" },
    ],
    [],
  );
  const performanceSpotlightCumShare = useMemo(
    () => performanceSpotlightRows.reduce((acc, row) => acc + Math.max(0, Number(row.share_pct || 0)), 0),
    [performanceSpotlightRows],
  );
  const performanceSpotlightMeta = `${performanceSpotlightRows.length} ${leaf0LabelPluralLower} | participacao acumulada ${asPct(performanceSpotlightCumShare)}`;

  const concentration = dto?.panels.revenue_concentration || {
    top3_pct: 0,
    top5_pct: 0,
    top10_pct: 0,
    top3_mom_delta_pct: null,
    top5_mom_delta_pct: null,
    risk_level: "medium" as const,
    risk_reason: "",
    is_trend_available: false,
    trend_source: "DB_FALLBACK" as const,
    trend_basis: "MoM",
    trend_window_months: selectedTrendWindowMonths,
    target_month_ref: null,
    base_month_ref: null,
  };
  const concentrationTrendWindow = concentration.trend_window_months || dto?.scope.window?.window_months || dto?.scope.trend_window_months || selectedTrendWindowMonths;
  const concentrationTrendLabel = concentrationTrendWindow === 1 ? "1M" : `${concentrationTrendWindow}M`;
  const concentrationRiskLabel =
    concentration.risk_level === "high"
      ? "Risco alto"
      : concentration.risk_level === "medium"
        ? "Risco medio"
        : "Risco controlado";
  const concentrationRiskTone =
    concentration.risk_level === "high"
      ? styles.riskHigh
      : concentration.risk_level === "medium"
        ? styles.riskMedium
        : styles.riskLow;
  const allocation = dto?.panels.capital_allocation_map || [];
  const allocationSpotlightSource = (dto?.panels.capital_allocation_map_spotlight && dto.panels.capital_allocation_map_spotlight.length > 0)
    ? dto.panels.capital_allocation_map_spotlight
    : allocation;
  const allocationByNodeId = useMemo(() => {
    const index = new Map<number, { revenue_brl: number; margin_pct: number | null; share_pct: number }>();
    for (const row of dto?.panels.top_nodes_by_revenue || []) {
      index.set(Number(row.node_id || 0), {
        revenue_brl: Number(row.revenue_brl || 0),
        margin_pct: row.margin_pct == null ? null : Number(row.margin_pct),
        share_pct: Number(row.share_pct || 0),
      });
    }
    return index;
  }, [dto]);
  const allocationTotalCapital = useMemo(
    () => allocationSpotlightSource.reduce((acc, item) => acc + Math.max(0, Number(item.capital_brl || 0)), 0),
    [allocationSpotlightSource],
  );
  const allocationSpotlightRows = useMemo<AllocationSpotlightRow[]>(() => {
    const parseNumberOrZero = (raw: string): number => {
      const normalized = String(raw || "").trim().replace(",", ".");
      if (!normalized) return 0;
      const parsed = Number(normalized);
      return Number.isFinite(parsed) ? parsed : 0;
    };
    const shareTarget = Math.min(100, Math.max(0, parseNumberOrZero(allocationSpotlightShareTargetInput)));
    const topTarget = Math.max(0, Math.floor(parseNumberOrZero(allocationSpotlightTopInput)));
    const searchToken = String(allocationSpotlightSearch || "").trim().toLocaleLowerCase("pt-BR");
    const riskWeight = (risk: "low" | "medium" | "high"): number => {
      if (risk === "high") return 1.4;
      if (risk === "medium") return 1.15;
      return 1;
    };

    const rows: AllocationSpotlightRow[] = allocationSpotlightSource.map((item) => {
      const nodeId = Number(item.node_id || 0);
      const enriched = allocationByNodeId.get(nodeId);
      const capital = Math.max(0, Number(item.capital_brl || 0));
      const riskPct = Math.max(0, Number(item.risk_pct || 0));
      return {
        node_id: nodeId,
        node_name: String(item.node_name || ""),
        sku_count: Math.max(0, Number(item.sku_count || 0)),
        capital_brl: capital,
        capital_share_pct: allocationTotalCapital > 0 ? (capital / allocationTotalCapital) * 100 : 0,
        risk_level: item.risk_level,
        risk_pct: riskPct,
        revenue_brl: Number(enriched?.revenue_brl || 0),
        margin_pct: enriched?.margin_pct ?? null,
        share_pct: Number(enriched?.share_pct || 0),
        priority_score: capital * (riskWeight(item.risk_level) + Math.max(0, riskPct / 100)),
      };
    });

    const filtered = rows.filter((row) => {
      const riskOk = allocationSpotlightRiskFilter === "all" ? true : row.risk_level === allocationSpotlightRiskFilter;
      const searchOk = searchToken ? row.node_name.toLocaleLowerCase("pt-BR").includes(searchToken) : true;
      return riskOk && searchOk;
    });

    const byCapitalDesc = [...filtered].sort((a, b) => b.capital_brl - a.capital_brl);
    const byCapitalAsc = [...filtered].sort((a, b) => a.capital_brl - b.capital_brl);

    const selectByShare = (): AllocationSpotlightRow[] => {
      if (!filtered.length) return [];
      if (shareTarget >= 100) return filtered;
      if (allocationSpotlightShareOp === "gte") {
        if (shareTarget <= 0) return filtered;
        const selected: AllocationSpotlightRow[] = [];
        let cumulative = 0;
        for (const row of byCapitalDesc) {
          selected.push(row);
          cumulative += Math.max(0, Number(row.capital_share_pct || 0));
          if (cumulative >= shareTarget) break;
        }
        return selected;
      }
      if (allocationSpotlightShareOp === "lte") {
        const selected: AllocationSpotlightRow[] = [];
        let cumulative = 0;
        for (const row of byCapitalAsc) {
          const next = cumulative + Math.max(0, Number(row.capital_share_pct || 0));
          if (next > shareTarget) break;
          selected.push(row);
          cumulative = next;
        }
        return selected;
      }
      let cumulative = 0;
      let bestCount = 0;
      let bestDiff = Number.POSITIVE_INFINITY;
      for (let i = 0; i < byCapitalDesc.length; i += 1) {
        cumulative += Math.max(0, Number(byCapitalDesc[i].capital_share_pct || 0));
        const diff = Math.abs(cumulative - shareTarget);
        if (diff < bestDiff) {
          bestDiff = diff;
          bestCount = i + 1;
        }
      }
      return byCapitalDesc.slice(0, Math.max(0, bestCount));
    };

    const rowsByShare = selectByShare();
    const selectByTop = (): AllocationSpotlightRow[] => {
      if (!rowsByShare.length) return [];
      if (topTarget <= 0) return rowsByShare;
      const desc = [...rowsByShare].sort((a, b) => b.capital_brl - a.capital_brl);
      if (allocationSpotlightTopOp === "gte") return desc.slice(0, topTarget);
      if (allocationSpotlightTopOp === "lte") return [...rowsByShare].sort((a, b) => a.capital_brl - b.capital_brl).slice(0, topTarget);
      return desc.slice(0, topTarget);
    };

    const rowsAfterTop = selectByTop();
    return [...rowsAfterTop].sort((a, b) => {
      const capitalCmp = b.capital_brl - a.capital_brl;
      if (capitalCmp !== 0) return capitalCmp;
      return a.node_name.localeCompare(b.node_name, "pt-BR");
    });
  }, [
    allocationSpotlightSource,
    allocationByNodeId,
    allocationSpotlightRiskFilter,
    allocationSpotlightSearch,
    allocationSpotlightShareOp,
    allocationSpotlightShareTargetInput,
    allocationSpotlightTopInput,
    allocationSpotlightTopOp,
    allocationTotalCapital,
  ]);
  const allocationSpotlightRowsTotalShare = useMemo(
    () => allocationSpotlightRows.reduce((acc, row) => acc + Math.max(0, Number(row.capital_share_pct || 0)), 0),
    [allocationSpotlightRows],
  );
  const allocationSpotlightMeta = useMemo(
    () => `${allocationSpotlightRows.length} ${leaf0LabelPluralLower} | capital coberto ${asPct(allocationSpotlightRowsTotalShare)}`,
    [allocationSpotlightRows.length, allocationSpotlightRowsTotalShare, leaf0LabelPluralLower],
  );
  const allocationSpotlightRiskOptions = useMemo<SelectMenuOption[]>(
    () => [
      { value: "all", label: "Todos riscos" },
      { value: "high", label: "Alto risco" },
      { value: "medium", label: "Medio risco" },
      { value: "low", label: "Baixo risco" },
    ],
    [],
  );
  const allocationSpotlightPriorityRows = useMemo(
    () => [...allocationSpotlightRows].sort((a, b) => b.priority_score - a.priority_score).slice(0, 5),
    [allocationSpotlightRows],
  );
  const allocationSpotlightSelectedRows = useMemo(() => {
    if (allocationSpotlightSelectedNodeId == null) return allocationSpotlightRows;
    return allocationSpotlightRows.filter((row) => row.node_id === allocationSpotlightSelectedNodeId);
  }, [allocationSpotlightRows, allocationSpotlightSelectedNodeId]);

  useEffect(() => {
    if (!isAllocationSpotlightOpen) return;
    const timer = window.setTimeout(() => {
      window.dispatchEvent(new Event("resize"));
    }, 180);
    return () => window.clearTimeout(timer);
  }, [isAllocationSpotlightOpen, allocationSpotlightSelectedNodeId, allocationSpotlightRows.length]);

  useEffect(() => {
    if (!isAllocationSpotlightOpen || !allocationPriorityRef.current) return undefined;

    const node = allocationPriorityRef.current;
    const syncHeight = () => {
      const next = Math.max(320, Math.ceil(node.getBoundingClientRect().height));
      setAllocationTopHeight((current) => (Math.abs(current - next) <= 1 ? current : next));
    };

    syncHeight();

    const observer = new ResizeObserver(() => syncHeight());
    observer.observe(node);

    return () => observer.disconnect();
  }, [allocationSpotlightPriorityRows, allocationSpotlightSelectedNodeId, isAllocationSpotlightOpen]);
  const isDarkTheme =
    typeof document !== "undefined" &&
    document.documentElement.getAttribute("data-theme") === "dark";
  const efficiency = dto?.panels.capital_efficiency || [];
  const topMargin = dto?.panels.rankings.top_margin || [];
  const nodesAtRisk = dto?.panels.rankings.nodes_at_risk || [];
  const topMarginRows = topMargin.slice(0, BOTTOM_TABLE_MAX_ROWS);
  const nodesAtRiskRows = nodesAtRisk.slice(0, BOTTOM_TABLE_MAX_ROWS);
  const topMarginSpotlightCurveOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Margem Bruta", value: "gross_margin" },
      { label: "Margem %", value: "margin_pct" },
      { label: "GMROI", value: "gmroi" },
    ],
    [],
  );
  const topMarginSpotlightGroupOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: `Todos os ${leaf0LabelPluralLower}`, value: "all" },
      ...Array.from(new Set(topMargin.map((row) => String(row.node_name || "").trim()).filter(Boolean)))
        .sort((a, b) => a.localeCompare(b, "pt-BR"))
        .map((group) => ({ label: group, value: group })),
    ],
    [leaf0LabelPluralLower, topMargin],
  );
  const topMarginSpotlightRows = useMemo<TopMarginSpotlightRow[]>(() => {
    const token = String(topMarginSpotlightSearch || "").trim().toLocaleLowerCase("pt-BR");
    const baseRows = topMargin
      .filter((row) => {
        const nodeName = String(row.node_name || "");
        const searchOk = token ? nodeName.toLocaleLowerCase("pt-BR").includes(token) : true;
        const groupOk = topMarginSpotlightGroupFilters.length ? topMarginSpotlightGroupFilters.includes(nodeName) : true;
        return searchOk && groupOk;
      })
      .map((row) => ({
        node_id: Number(row.node_id || 0),
        node_name: String(row.node_name || ""),
        gross_margin_brl: Number((row as { gross_margin_brl?: number }).gross_margin_brl || 0),
        revenue_brl: Number(row.revenue_brl || 0),
        margin_pct: row.margin_pct == null ? null : Number(row.margin_pct),
        capital_brl: Number((row as { capital_brl?: number }).capital_brl || 0),
        gmroi: (row as { gmroi?: number | null }).gmroi == null ? null : Number((row as { gmroi?: number | null }).gmroi),
        share_margin_pct: 0,
        trend: row.trend,
      }));
    const totalGrossMargin = baseRows.reduce((acc, row) => acc + Math.max(0, row.gross_margin_brl), 0);
    return baseRows.map((row) => ({
      ...row,
      share_margin_pct: totalGrossMargin > 0 ? (Math.max(0, row.gross_margin_brl) / totalGrossMargin) * 100 : 0,
    }));
  }, [topMargin, topMarginSpotlightGroupFilters, topMarginSpotlightSearch]);
  const topMarginMetricValue = (row: TopMarginSpotlightRow): number => {
    if (topMarginSpotlightCurveType === "margin_pct") return row.margin_pct == null ? Number.NEGATIVE_INFINITY : row.margin_pct;
    if (topMarginSpotlightCurveType === "gmroi") return row.gmroi == null ? Number.NEGATIVE_INFINITY : row.gmroi;
    return row.gross_margin_brl;
  };
  const topMarginSpotlightSortedRows = useMemo(
    () => [...topMarginSpotlightRows].sort((a, b) => topMarginMetricValue(b) - topMarginMetricValue(a)),
    [topMarginSpotlightRows, topMarginSpotlightCurveType],
  );
  const topMarginSpotlightChartRows = useMemo(
    () =>
      topMarginSpotlightSortedRows
        .map((row) => ({
          node_short: shortNodeLabel(row.node_name, 20),
          node_name: row.node_name,
          metric_value: topMarginMetricValue(row),
        }))
        .filter((row) => Number.isFinite(row.metric_value))
        .slice(0, 12),
    [topMarginSpotlightSortedRows, topMarginSpotlightCurveType],
  );
  const topMarginSpotlightTotalGrossMargin = useMemo(
    () => topMarginSpotlightRows.reduce((acc, row) => acc + Math.max(0, row.gross_margin_brl), 0),
    [topMarginSpotlightRows],
  );
  const topMarginSpotlightTop5Share = useMemo(
    () => topMarginSpotlightSortedRows.slice(0, 5).reduce((acc, row) => acc + row.share_margin_pct, 0),
    [topMarginSpotlightSortedRows],
  );
  const topMarginSpotlightTop10Share = useMemo(
    () => topMarginSpotlightSortedRows.slice(0, 10).reduce((acc, row) => acc + row.share_margin_pct, 0),
    [topMarginSpotlightSortedRows],
  );
  const topMarginSpotlightMetricLabel =
    topMarginSpotlightCurveType === "gmroi"
      ? "GMROI"
      : topMarginSpotlightCurveType === "margin_pct"
        ? "Margem %"
        : "Margem Bruta";
  const topMarginSpotlightColumns = useMemo<SpotlightDataTableColumn<TopMarginSpotlightRow>[]>(
    () => [
      { id: "node_name", header: leaf0Label, accessor: (row) => row.node_name },
      { id: "gross_margin_brl", header: "Margem Bruta", accessor: (row) => row.gross_margin_brl, cell: (row) => asBrlCompact(row.gross_margin_brl) },
      { id: "share_margin_pct", header: "Share Margem", accessor: (row) => row.share_margin_pct, cell: (row) => asPct(row.share_margin_pct) },
      { id: "revenue_brl", header: "Receita", accessor: (row) => row.revenue_brl, cell: (row) => asBrlCompact(row.revenue_brl) },
      {
        id: "margin_pct",
        header: "Margem %",
        accessor: (row) => row.margin_pct,
        cell: (row) => <span className={marginToneClass(row.margin_pct)}>{row.margin_pct == null ? "-" : asPct(row.margin_pct)}</span>,
        sortValue: (row) => (row.margin_pct == null ? Number.NEGATIVE_INFINITY : row.margin_pct),
      },
      { id: "capital_brl", header: "Capital", accessor: (row) => row.capital_brl, cell: (row) => asBrlCompact(row.capital_brl) },
      {
        id: "gmroi",
        header: "GMROI",
        accessor: (row) => row.gmroi,
        cell: (row) => (row.gmroi == null ? "-" : `${Number(row.gmroi).toLocaleString("pt-BR", { maximumFractionDigits: 2 })}x`),
        sortValue: (row) => (row.gmroi == null ? Number.NEGATIVE_INFINITY : row.gmroi),
      },
    ],
    [leaf0Label],
  );
  const riskSpotlightMetricOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Capital em Risco", value: "capital_at_risk_brl" },
      { label: "Risco %", value: "risk_pct" },
    ],
    [],
  );
  const riskSpotlightRiskOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Todos os Riscos", value: "all" },
      { label: "Alto risco", value: "high" },
      { label: "Medio risco", value: "medium" },
      { label: "Baixo risco", value: "low" },
    ],
    [],
  );
  const riskSpotlightGroupOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: `Todos os ${leaf0LabelPluralLower}`, value: "all" },
      ...Array.from(new Set(nodesAtRisk.map((row) => String(row.node_name || "").trim()).filter(Boolean)))
        .sort((a, b) => a.localeCompare(b, "pt-BR"))
        .map((group) => ({ label: group, value: group })),
    ],
    [leaf0LabelPluralLower, nodesAtRisk],
  );
  const riskSpotlightRows = useMemo<RiskSpotlightRow[]>(() => {
    const token = String(riskSpotlightSearch || "").trim().toLocaleLowerCase("pt-BR");
    const metricFilterValue =
      riskSpotlightMetricInput.trim() === ""
        ? null
        : Number(riskSpotlightMetricInput.replace(",", "."));
    return nodesAtRisk
      .map((row) => ({
        node_id: Number(row.node_id || 0),
        node_name: String(row.node_name || ""),
        risk_level: row.risk_level || riskLevelFromPct(Number(row.risk_pct || 0)),
        risk_pct: Number(row.risk_pct || 0),
        capital_at_risk_brl: Number(row.capital_at_risk_brl || 0),
        capital_brl: Number(row.capital_brl || 0),
        revenue_brl: Number(row.revenue_brl || 0),
        margin_pct: row.margin_pct == null ? null : Number(row.margin_pct),
        gmroi: row.gmroi == null ? null : Number(row.gmroi),
        financial_risk_priority_brl: Number(row.financial_risk_priority_brl || 0),
      }))
      .filter((row) => {
        const searchOk = token ? row.node_name.toLocaleLowerCase("pt-BR").includes(token) : true;
        const groupOk = riskSpotlightGroupFilters.length ? riskSpotlightGroupFilters.includes(row.node_name) : true;
        const riskOk = riskSpotlightRiskFilter === "all" ? true : row.risk_level === riskSpotlightRiskFilter;
        const metricValue = riskSpotlightMetric === "risk_pct" ? row.risk_pct : row.capital_at_risk_brl;
        const metricOk =
          metricFilterValue == null || !Number.isFinite(metricFilterValue)
            ? true
            : riskSpotlightMetricOp === "eq"
              ? metricValue === metricFilterValue
              : riskSpotlightMetricOp === "lte"
                ? metricValue <= metricFilterValue
                : metricValue >= metricFilterValue;
        return searchOk && groupOk && riskOk && metricOk;
      });
  }, [
    nodesAtRisk,
    riskSpotlightGroupFilters,
    riskSpotlightMetric,
    riskSpotlightMetricInput,
    riskSpotlightMetricOp,
    riskSpotlightRiskFilter,
    riskSpotlightSearch,
  ]);
  const riskSpotlightSortedRows = useMemo(
    () =>
      [...riskSpotlightRows].sort((a, b) => {
        const priorityCmp = b.financial_risk_priority_brl - a.financial_risk_priority_brl;
        if (priorityCmp !== 0) return priorityCmp;
        const riskCmp = b.risk_pct - a.risk_pct;
        if (riskCmp !== 0) return riskCmp;
        return b.capital_at_risk_brl - a.capital_at_risk_brl;
      }),
    [riskSpotlightRows],
  );
  const riskSpotlightChartRows = useMemo(
    () => riskSpotlightSortedRows.slice(0, 24),
    [riskSpotlightSortedRows],
  );
  const riskSpotlightTotalCapitalAtRisk = useMemo(
    () => riskSpotlightRows.reduce((acc, row) => acc + Math.max(0, row.capital_at_risk_brl), 0),
    [riskSpotlightRows],
  );
  const riskSpotlightHighRiskCount = useMemo(
    () => riskSpotlightRows.filter((row) => row.risk_level === "high").length,
    [riskSpotlightRows],
  );
  const riskSpotlightHighRiskShare = useMemo(() => {
    if (!riskSpotlightRows.length) return 0;
    return (riskSpotlightHighRiskCount / riskSpotlightRows.length) * 100;
  }, [riskSpotlightHighRiskCount, riskSpotlightRows.length]);
  const riskSpotlightColumns = useMemo<SpotlightDataTableColumn<RiskSpotlightRow>[]>(
    () => [
      { id: "node_name", header: leaf0Label, accessor: (row) => row.node_name },
      {
        id: "risk_pct",
        header: "Risco",
        accessor: (row) => row.risk_pct,
        cell: (row) => (
          <span className={row.risk_level === "high" ? styles.badgeRed : row.risk_level === "medium" ? styles.badgeOrange : styles.badgeGray}>
            {asPct(row.risk_pct)}
          </span>
        ),
      },
      { id: "capital_at_risk_brl", header: "Capital em Risco", accessor: (row) => row.capital_at_risk_brl, cell: (row) => asBrlCompact(row.capital_at_risk_brl) },
      { id: "capital_brl", header: "Capital", accessor: (row) => row.capital_brl, cell: (row) => asBrlCompact(row.capital_brl) },
      { id: "revenue_brl", header: "Receita", accessor: (row) => row.revenue_brl, cell: (row) => asBrlCompact(row.revenue_brl) },
      {
        id: "margin_pct",
        header: "Margem %",
        accessor: (row) => row.margin_pct,
        cell: (row) => <span className={marginToneClass(row.margin_pct)}>{row.margin_pct == null ? "-" : asPct(row.margin_pct)}</span>,
        sortValue: (row) => (row.margin_pct == null ? Number.NEGATIVE_INFINITY : row.margin_pct),
      },
      {
        id: "gmroi",
        header: "GMROI",
        accessor: (row) => row.gmroi,
        cell: (row) => (row.gmroi == null ? "-" : `${Number(row.gmroi).toLocaleString("pt-BR", { maximumFractionDigits: 2 })}x`),
        sortValue: (row) => (row.gmroi == null ? Number.NEGATIVE_INFINITY : row.gmroi),
      },
    ],
    [leaf0Label],
  );
  const backlog: ActionCard[] = (dto?.panels.backlog || []).slice(0, 4);
  const allocationMax = Math.max(1, ...allocation.map((item) => item.capital_brl || 0));
  const allocationTreeData = useMemo<ChartData<"treemap">>(() => {
    // chartjs-chart-treemap gera `data` internamente a partir de `tree`; manter sem `data`
    // evita inconsistência de hitbox/hover após refresh do payload.
    return {
      datasets: [
        {
          key: "v",
          tree: allocation.map((item) => ({
            n: item.node_name,
            sc: item.sku_count,
            cap: Math.max(0, item.capital_brl || 0),
            v: Math.max(0, item.size_weight ?? item.capital_brl ?? 0),
            rl: item.risk_level,
            rp: item.risk_pct,
          })),
          borderColor: isDarkTheme ? "rgba(148, 163, 184, 0.34)" : "#d6dbe5",
          borderWidth: 1,
          spacing: 1,
          labels: {
            display: true,
            formatter: (ctx: unknown) => {
              const raw = (ctx as { raw?: { _data?: { n?: string; cap?: number }; w?: number; h?: number } }).raw;
              const row = raw?._data;
              const area = Number(raw?.w || 0) * Number(raw?.h || 0);
              const name = String(row?.n || "");
              if (!name || area < 1800) return "";
              if (area < 3200) return name;
              return [name, asBrlCompact(Number(row?.cap || 0))];
            },
            color: (ctx: unknown) => {
              const row = (ctx as { raw?: { _data?: { rl?: string } } }).raw?._data;
              return riskLabelColor(String(row?.rl || "low"));
            },
            font: {
              size: 11,
              weight: 700,
              family: "Inter, sans-serif",
              style: "normal",
              lineHeight: 1.2,
            },
            overflow: "fit",
          },
          backgroundColor: (ctx: unknown) => {
            const row = (ctx as { raw?: { _data?: { rl?: string; cap?: number } } }).raw?._data;
            const base = riskFillColor(String(row?.rl || "low"));
            const alpha = 0.55 + (Math.min(1, Number(row?.cap || 0) / allocationMax) * 0.35);
            return `${base}${Math.round(alpha * 255).toString(16).padStart(2, "0")}`;
          },
          hoverBorderWidth: 1.2,
          hoverBorderColor: isDarkTheme ? "rgba(226, 232, 240, 0.62)" : "#94a3b8",
        },
      ],
    } as unknown as ChartData<"treemap">;
  }, [allocation, allocationMax, isDarkTheme]);
  const allocationTreeKey = useMemo(
    () => allocation.map((item) => `${item.node_id}:${item.capital_brl}:${item.risk_pct}:${item.sku_count}`).join("|"),
    [allocation],
  );
  const allocationTreeOptions = useMemo<ChartOptions<"treemap">>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    events: ["mousemove", "mouseout", "click", "touchstart", "touchmove"],
    interaction: {
      mode: "nearest",
      intersect: true,
    },
    hover: {
      mode: "nearest",
      intersect: true,
    },
    plugins: {
      legend: { display: false },
      tooltip: {
        enabled: true,
        position: "nearest",
        displayColors: false,
        callbacks: {
          title: (items) => {
            const row = (items[0]?.raw as { _data?: { n?: string } } | undefined)?._data;
            return String(row?.n || leaf0Label);
          },
          label: (item): string[] => {
            const row = (item.raw as { _data?: { cap?: number; rp?: number; sc?: number } } | undefined)?._data;
            const capital = asBrlCompact(Number(row?.cap || 0));
            const risk = Number.isFinite(Number(row?.rp)) ? `${Number(row?.rp).toFixed(1).replace(".", ",")}%` : "n/d";
            const skus = Math.max(0, Number(row?.sc || 0)).toLocaleString("pt-BR");
            return [`Capital alocado: ${capital}`, `Risco: ${risk}`, `SKUs: ${skus}`];
          },
        },
      },
    },
  }), [leaf0Label]);
  const allocationSpotlightTreeMax = Math.max(1, ...allocationSpotlightRows.map((row) => row.capital_brl || 0));
  const allocationSpotlightTreeData = useMemo<ChartData<"treemap">>(() => {
    const maxCapital = Number.isFinite(allocationSpotlightTreeMax) && allocationSpotlightTreeMax > 0 ? allocationSpotlightTreeMax : 1;
    const resolveFill = (row?: { rl?: string; cap?: number }): string => {
      const riskLevel = String(row?.rl || "low");
      const cap = Number(row?.cap);
      const safeCap = Number.isFinite(cap) ? cap : 0;
      const base = riskFillColor(riskLevel);
      const alpha = 0.55 + (Math.min(1, safeCap / maxCapital) * 0.35);
      const alphaHex = Math.round(alpha * 255).toString(16).padStart(2, "0");
      return `${base}${alphaHex}`;
    };
    return {
      datasets: [
        {
          key: "v",
          tree: allocationSpotlightRows.map((row) => ({
            node_id: row.node_id,
            n: row.node_name,
            cap: row.capital_brl,
            cp: row.capital_share_pct,
            rp: row.risk_pct,
            rl: row.risk_level,
            sku: row.sku_count,
            rev: row.revenue_brl,
            m: row.margin_pct,
            v: Math.max(0, row.capital_brl),
          })),
          borderColor: isDarkTheme ? "rgba(148, 163, 184, 0.38)" : "#d6dbe5",
          borderWidth: 1,
          spacing: 1,
          labels: {
            display: true,
            formatter: (ctx: unknown) => {
              const raw = (ctx as { raw?: { _data?: { n?: string; cap?: number; cp?: number }; w?: number; h?: number } }).raw;
              const row = raw?._data;
              const area = Number(raw?.w || 0) * Number(raw?.h || 0);
              const name = String(row?.n || "");
              if (!name || area < 1800) return "";
              if (area < 3200) return name;
              return [name, `${asBrlCompact(Number(row?.cap || 0))} (${asPct(Number(row?.cp || 0))})`];
            },
            color: (ctx: unknown) => {
              const row = (ctx as { raw?: { _data?: { rl?: string } } }).raw?._data;
              return riskLabelColor(String(row?.rl || "low"));
            },
            font: {
              size: 11,
              weight: 700,
              family: "Inter, sans-serif",
              style: "normal",
              lineHeight: 1.2,
            },
            overflow: "fit",
          },
          backgroundColor: (ctx: unknown) => {
            const row = (ctx as { raw?: { _data?: { rl?: string; cap?: number } } }).raw?._data;
            return resolveFill(row);
          },
          hoverBackgroundColor: (ctx: unknown) => {
            const row = (ctx as { raw?: { _data?: { rl?: string; cap?: number } } }).raw?._data;
            return resolveFill(row);
          },
          hoverBorderWidth: 1.3,
          hoverBorderColor: isDarkTheme ? "rgba(226, 232, 240, 0.7)" : "#94a3b8",
        },
      ],
    } as unknown as ChartData<"treemap">;
  }, [allocationSpotlightRows, allocationSpotlightTreeMax, isDarkTheme]);
  const allocationSpotlightTreeKey = useMemo(
    () => `${allocationSpotlightSelectedNodeId ?? "all"}|${allocationSpotlightRows.map((row) => `${row.node_id}:${row.capital_brl}:${row.risk_pct}:${row.priority_score}`).join("|")}`,
    [allocationSpotlightRows, allocationSpotlightSelectedNodeId],
  );
  const allocationSpotlightTreeOptions = useMemo<ChartOptions<"treemap">>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    onClick: (_event, elements) => {
      const first = elements[0];
      if (!first) return;
      const nextNodeId = Number(allocationSpotlightRows[first.index]?.node_id || 0);
      if (!nextNodeId) return;
      setAllocationSpotlightSelectedNodeId((curr) => (curr === nextNodeId ? null : nextNodeId));
    },
    plugins: {
      legend: { display: false },
      tooltip: {
        enabled: true,
        displayColors: false,
        callbacks: {
          title: (items) => {
            const row = (items[0]?.raw as { _data?: { n?: string } } | undefined)?._data;
            return String(row?.n || leaf0Label);
          },
          label: (item): string[] => {
            const row = (item.raw as { _data?: { cap?: number; cp?: number; rp?: number; sku?: number; rev?: number; m?: number | null } } | undefined)?._data;
            const cap = asBrlCompact(Number(row?.cap || 0));
            const capShare = asPct(Number(row?.cp || 0));
            const risk = Number.isFinite(Number(row?.rp)) ? `${Number(row?.rp).toFixed(1).replace(".", ",")}%` : "n/d";
            const sku = Math.max(0, Number(row?.sku || 0)).toLocaleString("pt-BR");
            const rev = asBrlCompact(Number(row?.rev || 0));
            const margin = row?.m == null ? "-" : asPct(Number(row.m));
            return [`Capital: ${cap}`, `% do total: ${capShare}`, `Receita: ${rev}`, `Margem: ${margin}`, `Risco: ${risk}`, `SKUs: ${sku}`];
          },
        },
      },
    },
  }), [allocationSpotlightRows, leaf0Label]);
  const efficiencyRows = useMemo(() => efficiency.slice(0, 6), [efficiency]);
  const efficiencyAvgGmroi = useMemo(() => {
    const valid = efficiencyRows.map((item) => item.gmroi).filter((value): value is number => value != null && Number.isFinite(value));
    if (!valid.length) return null;
    return valid.reduce((acc, cur) => acc + cur, 0) / valid.length;
  }, [efficiencyRows]);
  const efficiencyGoodCount = useMemo(() => {
    if (efficiencyAvgGmroi == null) return 0;
    return efficiencyRows.filter((item) => {
      const gmroi = Number(item.gmroi ?? Number.NaN);
      return Number.isFinite(gmroi) && gmroi >= efficiencyAvgGmroi;
    }).length;
  }, [efficiencyAvgGmroi, efficiencyRows]);
  const efficiencyWorstNode = useMemo(() => {
    const candidates = efficiencyRows.filter((item) => Number(item.capital_brl || 0) > 0);
    if (!candidates.length) return null;
    return [...candidates].sort((a, b) => {
      const grossA = Number((a as { gross_margin_brl?: number }).gross_margin_brl ?? a.margin_brl ?? 0);
      const grossB = Number((b as { gross_margin_brl?: number }).gross_margin_brl ?? b.margin_brl ?? 0);
      const ratioA = grossA / Math.max(1, Number(a.capital_brl || 0));
      const ratioB = grossB / Math.max(1, Number(b.capital_brl || 0));
      return ratioA - ratioB;
    })[0];
  }, [efficiencyRows]);
  const abcTotalGroups = useMemo(() => {
    const mix = dto?.scope.analysis_cards?.abc_mix;
    return Number(mix?.a_count || 0) + Number(mix?.b_count || 0) + Number(mix?.c_count || 0);
  }, [dto?.scope.analysis_cards?.abc_mix]);
  const abcSpotlightCurveOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Receita", value: "revenue" },
      { label: "Margem Bruta", value: "gross_margin" },
    ],
    [],
  );
  const abcSpotlightBandOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Todas", value: "all" },
      { label: "A", value: "A" },
      { label: "B", value: "B" },
      { label: "C", value: "C" },
    ],
    [],
  );
  const abcCardGroupSharePct = (count: number | null | undefined): number => {
    if (!abcTotalGroups) return 0;
    return (Number(count || 0) / abcTotalGroups) * 100;
  };
  const abcSpotlightGroupOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: `Todos os ${leaf0LabelPluralLower}`, value: "all" },
      ...Array.from(new Set((dto?.panels.top_nodes_by_revenue || []).map((row) => String(row.node_name || "").trim()).filter(Boolean)))
        .sort((a, b) => a.localeCompare(b, "pt-BR"))
        .map((group) => ({ label: group, value: group })),
    ],
    [dto?.panels.top_nodes_by_revenue, leaf0LabelPluralLower],
  );
  const abcSpotlightRows = useMemo<AbcSpotlightRow[]>(() => {
    const source = dto?.panels.top_nodes_by_revenue || [];
    const aMax = Number(dto?.scope.analysis_cards?.abc_mix.a_max_cum_pct || 80);
    const bMax = Number(dto?.scope.analysis_cards?.abc_mix.b_max_cum_pct || 95);
    const token = String(abcSpotlightSearch || "").trim().toLocaleLowerCase("pt-BR");
    const filtered = source.filter((row) => {
      const nodeName = String(row.node_name || "");
      const searchOk = token ? nodeName.toLocaleLowerCase("pt-BR").includes(token) : true;
      const groupOk = abcSpotlightGroupFilters.length ? abcSpotlightGroupFilters.includes(nodeName) : true;
      return searchOk && groupOk;
    });
    const metricAccessor = (row: typeof filtered[number]): number =>
      abcSpotlightCurveType === "gross_margin"
        ? Number((row as { gross_margin_brl?: number; margin_brl?: number }).gross_margin_brl ?? (row as { margin_brl?: number }).margin_brl ?? 0)
        : Number(row.revenue_brl || 0);
    const sorted = [...filtered].sort((a, b) => metricAccessor(b) - metricAccessor(a));
    const totalMetric = sorted.reduce((acc, row) => acc + Math.max(0, metricAccessor(row)), 0);
    let cumShare = 0;
    return sorted
      .map((row) => {
        const share = totalMetric > 0 ? (Math.max(0, metricAccessor(row)) / totalMetric) * 100 : 0;
        cumShare += share;
        const band: "A" | "B" | "C" = cumShare <= aMax ? "A" : cumShare <= bMax ? "B" : "C";
        return {
          ...row,
          gross_margin_brl: Number(
            (row as { gross_margin_brl?: number; margin_brl?: number }).gross_margin_brl ??
            (row as { margin_brl?: number }).margin_brl ??
            0
          ),
          share_pct: share,
          band,
          cum_share_pct: cumShare,
        };
      })
      .filter((row) => (abcSpotlightBandFilter === "all" ? true : row.band === abcSpotlightBandFilter));
  }, [
    abcSpotlightBandFilter,
    abcSpotlightCurveType,
    abcSpotlightGroupFilters,
    abcSpotlightSearch,
    dto?.panels.top_nodes_by_revenue,
    dto?.scope.analysis_cards?.abc_mix.a_max_cum_pct,
    dto?.scope.analysis_cards?.abc_mix.b_max_cum_pct,
  ]);
  const abcParetoChartRows = useMemo(
    () =>
      abcSpotlightRows.slice(0, 30).map((row) => ({
        node_short: shortNodeLabel(String(row.node_name || ""), 14),
        node_name: String(row.node_name || ""),
        band: row.band,
        share_pct: Number(row.share_pct || 0),
        cum_share_pct: Number((row as { cum_share_pct?: number }).cum_share_pct || 0),
      })),
    [abcSpotlightRows],
  );
  const abcParetoBandSpans = useMemo(() => {
    if (!abcParetoChartRows.length) return [] as Array<{ band: "A" | "B" | "C"; leftPct: number; widthPct: number }>;
    const total = abcParetoChartRows.length;
    return (["A", "B", "C"] as const)
      .map((band) => {
        const firstIndex = abcParetoChartRows.findIndex((row) => row.band === band);
        const lastIndex = abcParetoChartRows.length - 1 - [...abcParetoChartRows].reverse().findIndex((row) => row.band === band);
        if (firstIndex < 0 || lastIndex < 0) return null;
        const leftPct = Math.max(0, ((firstIndex - 0.5) / total) * 100);
        const rightPct = Math.min(100, ((lastIndex + 0.5) / total) * 100);
        return {
          band,
          leftPct,
          widthPct: Math.max(0, rightPct - leftPct),
        };
      })
      .filter((item): item is { band: "A" | "B" | "C"; leftPct: number; widthPct: number } => item != null);
  }, [abcParetoChartRows]);
  const abcTopBands = useMemo(
    () => ({
      A: abcSpotlightRows.filter((row) => row.band === "A").slice(0, 5),
      B: abcSpotlightRows.filter((row) => row.band === "B").slice(0, 5),
      C: abcSpotlightRows.filter((row) => row.band === "C").slice(0, 5),
    }),
    [abcSpotlightRows],
  );
  const abcSpotlightBandSummary = useMemo(() => {
    const totalRows = abcSpotlightRows.length;
    const totalMetric = abcSpotlightRows.reduce(
      (acc, row) => acc + Math.max(0, abcSpotlightCurveType === "gross_margin" ? row.gross_margin_brl : row.revenue_brl),
      0,
    );
    const summarizeBand = (band: "A" | "B" | "C") => {
      const rows = abcSpotlightRows.filter((row) => row.band === band);
      const metricValue = rows.reduce(
        (acc, row) => acc + Math.max(0, abcSpotlightCurveType === "gross_margin" ? row.gross_margin_brl : row.revenue_brl),
        0,
      );
      return {
        count: rows.length,
        entityPct: totalRows > 0 ? (rows.length / totalRows) * 100 : 0,
        metricPct: totalMetric > 0 ? (metricValue / totalMetric) * 100 : 0,
      };
    };
    return {
      A: summarizeBand("A"),
      B: summarizeBand("B"),
      C: summarizeBand("C"),
    };
  }, [abcSpotlightCurveType, abcSpotlightRows]);
  const allocationSpotlightColumns = useMemo<SpotlightDataTableColumn<AllocationSpotlightRow>[]>(
    () => [
      {
        id: "node_name",
        header: leaf0Label,
        accessor: (row) => row.node_name,
      },
      {
        id: "capital_brl",
        header: "Capital",
        accessor: (row) => row.capital_brl,
        cell: (row) => asBrlCompact(row.capital_brl),
      },
      {
        id: "capital_share_pct",
        header: "% Total",
        accessor: (row) => row.capital_share_pct,
        cell: (row) => asPct(row.capital_share_pct),
      },
      {
        id: "risk_pct",
        header: "Risco",
        accessor: (row) => row.risk_pct,
        cell: (row) => (
          <span
            className={
              row.risk_level === "high" ? styles.badgeRed : row.risk_level === "medium" ? styles.badgeOrange : styles.badgeGray
            }
          >
            {asPct(row.risk_pct)}
          </span>
        ),
      },
      {
        id: "revenue_brl",
        header: "Receita",
        accessor: (row) => row.revenue_brl,
        cell: (row) => asBrlCompact(row.revenue_brl),
      },
      {
        id: "margin_pct",
        header: "Margem",
        accessor: (row) => row.margin_pct,
        cell: (row) => (
          <span className={marginToneClass(row.margin_pct)}>
            {row.margin_pct == null ? "-" : asPct(row.margin_pct)}
          </span>
        ),
        sortValue: (row) => (row.margin_pct == null ? Number.NEGATIVE_INFINITY : row.margin_pct),
      },
      {
        id: "sku_count",
        header: "SKUs",
        accessor: (row) => row.sku_count,
        cell: (row) => row.sku_count.toLocaleString("pt-BR"),
      },
    ],
    [leaf0Label],
  );
  const abcSpotlightColumns = useMemo<SpotlightDataTableColumn<AbcSpotlightRow>[]>(
    () => [
      {
        id: "node_name",
        header: leaf0Label,
        accessor: (row) => row.node_name,
      },
      {
        id: "band",
        header: "Faixa",
        accessor: (row) => row.band,
      },
      {
        id: "revenue_brl",
        header: "Receita",
        accessor: (row) => row.revenue_brl,
        cell: (row) => asBrlCompact(row.revenue_brl),
      },
      {
        id: "gross_margin_brl",
        header: "Margem Bruta",
        accessor: (row) => row.gross_margin_brl,
        cell: (row) => asBrlCompact(row.gross_margin_brl),
      },
      {
        id: "share_pct",
        header: "Share",
        accessor: (row) => row.share_pct,
        cell: (row) => asPct(row.share_pct),
      },
      {
        id: "cum_share_pct",
        header: "Acumulado",
        accessor: (row) => row.cum_share_pct,
        cell: (row) => asPct(row.cum_share_pct),
      },
      {
        id: "margin_pct",
        header: "Margem",
        accessor: (row) => row.margin_pct,
        cell: (row) => (
          <span className={marginToneClass(row.margin_pct)}>
            {row.margin_pct == null ? "-" : asPct(row.margin_pct)}
          </span>
        ),
        sortValue: (row) => (row.margin_pct == null ? Number.NEGATIVE_INFINITY : row.margin_pct),
      },
    ],
    [leaf0Label],
  );

  useEffect(() => {
    if (!isAbcSpotlightOpen || !abcTopBandsRef.current) return undefined;

    const node = abcTopBandsRef.current;
    const syncHeight = () => {
      const next = Math.max(220, Math.ceil(node.getBoundingClientRect().height));
      setAbcChartHeight((current) => (Math.abs(current - next) <= 1 ? current : next));
    };

    syncHeight();

    const observer = new ResizeObserver(() => syncHeight());
    observer.observe(node);

    return () => observer.disconnect();
  }, [abcTopBands, isAbcSpotlightOpen]);
  const marginPolicyLowPct = Number(dto?.scope.margin_policy?.low_pct ?? 12);
  const marginPolicyGoodPct = Number(dto?.scope.margin_policy?.good_pct ?? 25);
  const marginToneClass = (value: number | null): string => {
    if (value == null) return "";
    if (value >= marginPolicyGoodPct) return styles.good;
    if (value >= marginPolicyLowPct) return styles.warn;
    return styles.bad;
  };
  const efficiencyChartRows = useMemo<EfficiencyChartRow[]>(
    () =>
      efficiencyRows.map((item) => ({
        node_short: shortNodeLabel(item.node_name),
        node_name: String(item.node_name || ""),
        margem_bruta_brl: Number((item as { gross_margin_brl?: number }).gross_margin_brl ?? item.margin_brl ?? 0),
        capital_brl: Number(item.capital_brl || 0),
        gmroi: item.gmroi == null ? null : Number(item.gmroi),
      })),
    [efficiencyRows],
  );
  const efficiencySpotlightGroupOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: `Todos os ${leaf0LabelPluralLower}`, value: "all" },
      ...Array.from(new Set(efficiency.map((item) => String(item.node_name || "").trim()).filter(Boolean)))
        .sort((a, b) => a.localeCompare(b, "pt-BR"))
        .map((group) => ({ label: group, value: group })),
    ],
    [efficiency, leaf0LabelPluralLower],
  );
  const efficiencySpotlightRows = useMemo<EfficiencySpotlightRow[]>(() => {
    const token = String(efficiencySpotlightSearch || "").trim().toLocaleLowerCase("pt-BR");
    const gmroiFilterValue =
      efficiencySpotlightGmroiInput.trim() === ""
        ? null
        : Number(efficiencySpotlightGmroiInput.replace(",", "."));
    return efficiency
      .filter((item) => {
        const name = String(item.node_name || "").toLocaleLowerCase("pt-BR");
        const searchOk = token ? name.includes(token) : true;
        const groupOk = efficiencySpotlightGroupFilters.length
          ? efficiencySpotlightGroupFilters.includes(String(item.node_name || ""))
          : true;
        const gmroiValue = item.gmroi == null ? null : Number(item.gmroi);
        const gmroiOk =
          gmroiFilterValue == null || !Number.isFinite(gmroiFilterValue)
            ? true
            : gmroiValue == null
              ? false
              : efficiencySpotlightGmroiOp === "eq"
                ? gmroiValue === gmroiFilterValue
                : efficiencySpotlightGmroiOp === "lte"
                  ? gmroiValue <= gmroiFilterValue
                  : gmroiValue >= gmroiFilterValue;
        return searchOk && groupOk && gmroiOk;
      })
      .map((item) => ({
        node_name: String(item.node_name || ""),
        revenue_brl: Number(item.revenue_brl || 0),
        margem_bruta_brl: Number((item as { gross_margin_brl?: number }).gross_margin_brl ?? item.margin_brl ?? 0),
        capital_brl: Number(item.capital_brl || 0),
        gmroi: item.gmroi == null ? null : Number(item.gmroi),
        risk_pct: Number(item.risk_pct || 0),
        risk_level: riskLevelFromPct(Number(item.risk_pct || 0)),
      }));
  }, [
    efficiency,
    efficiencySpotlightGmroiInput,
    efficiencySpotlightGmroiOp,
    efficiencySpotlightGroupFilters,
    efficiencySpotlightSearch,
  ]);
  const efficiencySpotlightChartRows = useMemo<EfficiencyChartRow[]>(
    () =>
      [...efficiencySpotlightRows]
        .sort((a, b) => Number(b.capital_brl || 0) - Number(a.capital_brl || 0))
        .slice(0, EFFICIENCY_SPOTLIGHT_CHART_MAX_ITEMS)
        .map((item) => ({
        node_short: shortNodeLabel(item.node_name),
        node_name: item.node_name,
        margem_bruta_brl: item.margem_bruta_brl,
        capital_brl: item.capital_brl,
        gmroi: item.gmroi,
      })),
    [efficiencySpotlightRows],
  );
  const efficiencySpotlightGmroiAxisMax = useMemo(() => {
    const gmroiValues = efficiencySpotlightChartRows
      .map((row) => Number(row.gmroi))
      .filter((value) => Number.isFinite(value) && value >= 0) as number[];
    const p95 = percentile(gmroiValues, 0.95);
    if (p95 == null) return 0.1;
    return Math.max(0.1, p95 * 1.15);
  }, [efficiencySpotlightChartRows]);
  const efficiencySpotlightColumns = useMemo<SpotlightDataTableColumn<EfficiencySpotlightRow>[]>(
    () => [
      {
        id: "node_name",
        header: leaf0Label,
        accessor: (row) => row.node_name,
      },
      {
        id: "capital_brl",
        header: "Capital",
        accessor: (row) => row.capital_brl,
        cell: (row) => asBrlCompact(row.capital_brl),
      },
      {
        id: "margem_bruta_brl",
        header: "Margem Bruta",
        accessor: (row) => row.margem_bruta_brl,
        cell: (row) => asBrlCompact(row.margem_bruta_brl),
      },
      {
        id: "revenue_brl",
        header: "Receita",
        accessor: (row) => row.revenue_brl,
        cell: (row) => asBrlCompact(row.revenue_brl),
      },
      {
        id: "gmroi",
        header: "GMROI",
        accessor: (row) => row.gmroi,
        cell: (row) =>
          row.gmroi == null
            ? "-"
            : Number(row.gmroi).toLocaleString("pt-BR", { maximumFractionDigits: 2 }),
        sortValue: (row) => (row.gmroi == null ? Number.POSITIVE_INFINITY : row.gmroi),
      },
    ],
    [leaf0Label]
  );
  const renderEfficiencyTooltip = (props: any) => {
    const active = Boolean(props?.active);
    const payload = (Array.isArray(props?.payload) ? props.payload : []) as EfficiencyTooltipEntry[];
    if (!active || !payload || !payload.length) return null;
    const row = payload[0]?.payload as EfficiencyChartRow | undefined;
    if (!row) return null;
    return (
      <div
        style={{
          background: "color-mix(in srgb, var(--surface) 96%, transparent)",
          border: "1px solid color-mix(in srgb, var(--muted) 24%, transparent)",
          borderRadius: 8,
          padding: "8px 10px",
          boxShadow: "0 6px 18px rgba(15, 23, 42, 0.08)",
          minWidth: 180,
        }}
      >
        <div style={{ fontSize: 12, fontWeight: 800, color: "var(--ink)", marginBottom: 6 }}>{row.node_name}</div>
        {payload.map((entry: EfficiencyTooltipEntry, idx: number) => {
          const metricName = String(entry?.name || "");
          const metricValue =
            metricName === "GMROI"
              ? `${Number(entry?.value || 0).toFixed(2).replace(".", ",")}x`
              : asBrlCompact(Number(entry?.value || 0));
          return (
            <div key={`${metricName}:${idx}`} style={{ display: "flex", justifyContent: "space-between", gap: 8, fontSize: 12 }}>
              <span style={{ color: entry?.color || "#64748b", fontWeight: 700 }}>{metricName}</span>
              <span style={{ color: "var(--ink)", fontWeight: 800 }}>{metricValue}</span>
            </div>
          );
        })}
      </div>
    );
  };

  const renderReceitaBar = ({ x = 0, y = 0, width = 0, height = 0, fill = "#cbd5e1" }: EfficiencyBarShapeProps) => {
    const rectWidth = Math.max(1, width - EFFICIENCY_BAR_PAIR_GAP_PX / 2);
    const rectHeight = Math.abs(height);
    const rectY = height >= 0 ? y : y + height;
    return <rect x={x} y={rectY} width={rectWidth} height={rectHeight} fill={fill} />;
  };

  const renderCapitalBar = ({ x = 0, y = 0, width = 0, height = 0, fill = "#8b1538" }: EfficiencyBarShapeProps) => {
    const shift = EFFICIENCY_BAR_PAIR_GAP_PX / 2;
    const rectWidth = Math.max(1, width - shift);
    const rectHeight = Math.abs(height);
    const rectY = height >= 0 ? y : y + height;
    return <rect x={x + shift} y={rectY} width={rectWidth} height={rectHeight} fill={fill} />;
  };

  const renderEfficiencyXAxisTick = ({ x = 0, y = 0, payload }: EfficiencyXAxisTickProps) => {
    const shortLabel = String(payload?.payload?.node_short || payload?.value || "");
    const fullLabel = String(payload?.payload?.node_name || payload?.value || shortLabel);
    return (
      <g transform={`translate(${x},${y})`}>
        <text
          x={0}
          y={0}
          dy={14}
          textAnchor="middle"
          fill="#64748b"
          fontSize={11}
          fontWeight={700}
          style={{ cursor: "default" }}
        >
          <title>{fullLabel}</title>
          {shortLabel}
        </text>
      </g>
    );
  };
  const renderRiskScatterPoint = (rawProps: unknown) => {
    const { cx = 0, cy = 0, payload } = (rawProps as RiskScatterShapeProps | null) || {};
    const row = payload;
    if (!row) return <g />;
    const radiusBase = Math.max(5, Math.min(14, Math.sqrt(Math.max(0, row.revenue_brl || 0)) / 26));
    const fill =
      row.risk_level === "high" ? "#b91c1c" : row.risk_level === "medium" ? "#f59e0b" : "#10b981";
    return (
      <g>
        <circle cx={cx} cy={cy} r={radiusBase + 2} fill="rgba(15, 23, 42, 0.18)" />
        <circle cx={cx} cy={cy} r={radiusBase} fill={fill} fillOpacity={0.82} stroke="#ffffff" strokeOpacity={0.5} strokeWidth={1.5} />
      </g>
    );
  };

  useEffect(() => {
    if (!isEfficiencySpotlightOpen) return;
    const node = efficiencySpotlightBodyRef.current;
    if (!node) return;
    node.scrollTo({ top: 0, behavior: "auto" });
  }, [isEfficiencySpotlightOpen]);

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <h1 className={styles.title}>Classificacoes</h1>
          <p className={styles.subtitle}>Analytics por nivel de hierarquia</p>
        </div>
        <div className={styles.headerActions}>
          <label className={styles.selectLabel} htmlFor="taxonomy-level">
            Nivel
          </label>
          <SelectMenu
            id="taxonomy-level"
            value={String(selectedLevel)}
            options={levelSelectOptions}
            onSelect={(value) => setSelectedLevel(Number(value))}
            classNames={HEADER_SELECT_CLASSNAMES}
          />
          <label className={styles.selectLabel} htmlFor="taxonomy-trend-window">
            Periodo
          </label>
          <SelectMenu
            id="taxonomy-trend-window"
            value={String(selectedTrendWindowMonths)}
            options={trendWindowSelectOptions}
            onSelect={(value) => setSelectedTrendWindowMonths(Number(value))}
            classNames={HEADER_SELECT_CLASSNAMES}
          />
        </div>
      </header>

      {refreshingMessage ? <div className={styles.notice}>{refreshingMessage}</div> : null}
      {error ? <div className={styles.error}>{error}</div> : null}
      {loading && !dto ? <div className={styles.notice}>Carregando analytics de classificacoes...</div> : null}

      <section className={styles.kpiGrid6}>
        {kpis.map((kpi) => (
          <KpiCard
            key={kpi.label}
            label={kpi.label}
            value={kpi.value}
            change={kpi.change}
            tone={kpi.tone}
            helpItems={kpi.helpItems}
            showChangeArrow={kpi.showChangeArrow}
          />
        ))}
      </section>

      <section className={styles.twoCol}>
        <AnalyticsPerformanceSection
          title="Performance por nivel"
          hint={`Receita e participacao dos principais ${leaf0LabelPluralLower} do escopo.`}
          spotlight={{ label: "Ver Detalhes", onClick: () => setIsPerformanceSpotlightOpen(true) }}
          items={performanceSectionItems}
          solid
        />

        <AnalyticsConcentrationSection
          title="Concentracao de receita"
          hint={`Monitorar dependencia em poucos ${leaf0LabelPluralLower}.`}
          riskBadge={concentrationRiskLabel}
          riskTone={concentration.risk_level === "high" ? "negative" : concentration.risk_level === "low" ? "positive" : "neutral"}
          stats={[
            {
              label: `Top 3 (${concentrationTrendLabel})`,
              value: trendPctText(concentration.top3_mom_delta_pct) || "-",
              valueTone: concentration.top3_mom_delta_pct == null ? "neutral" : concentration.top3_mom_delta_pct >= 0 ? "positive" : "negative",
            },
            {
              label: `Top 5 (${concentrationTrendLabel})`,
              value: trendPctText(concentration.top5_mom_delta_pct) || "-",
              valueTone: concentration.top5_mom_delta_pct == null ? "neutral" : concentration.top5_mom_delta_pct >= 0 ? "positive" : "negative",
            },
            {
              label: "Base trend",
              value: concentration.trend_source === "SNAPSHOT" ? "SNAPSHOT" : "DB",
            },
          ]}
          bars={[
            { label: `Top 3 ${leaf0LabelPluralLower}`, valueText: asPct(concentration.top3_pct), percent: concentration.top3_pct, tone: "top3" },
            { label: `Top 5 ${leaf0LabelPluralLower}`, valueText: asPct(concentration.top5_pct), percent: concentration.top5_pct, tone: "top5" },
            { label: `Top 10 ${leaf0LabelPluralLower}`, valueText: asPct(concentration.top10_pct), percent: concentration.top10_pct, tone: "top10" },
          ]}
          quote={concentration.risk_reason || "Receita concentrada exige monitoramento continuo de risco e cobertura."}
        />
      </section>

      <section className={styles.twoCol}>
        <article
          className={`${styles.panel} ${styles.panelClickable}`}
          role="button"
          tabIndex={0}
          onClick={() => setIsAllocationSpotlightOpen(true)}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              setIsAllocationSpotlightOpen(true);
            }
          }}
        >
          <div className={styles.panelHeadRow}>
            <div className={styles.panelHead}>
              <h2 className={styles.panelTitle}>Mapa de alocacao de capital</h2>
              <p className={styles.panelSub}>{`Tamanho por capital investido e risco por ${leaf0LabelLower}`}</p>
            </div>
            <span className={styles.panelActionText}>Ver Detalhes</span>
          </div>
          <div className={styles.allocLegend}>
            <span><i className={styles.allocLegendLow} /> Baixo risco</span>
            <span><i className={styles.allocLegendMid} /> Medio risco</span>
            <span><i className={styles.allocLegendHigh} /> Alto risco</span>
          </div>
          <div className={styles.allocTreemapWrap}>
            <Chart
              key={allocationTreeKey}
              type="treemap"
              data={allocationTreeData}
              options={allocationTreeOptions}
              redraw
            />
          </div>
        </article>

        <article
          className={`${styles.panel} ${styles.panelClickable}`}
          role="button"
          tabIndex={0}
          onClick={() => setIsEfficiencySpotlightOpen(true)}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              setIsEfficiencySpotlightOpen(true);
            }
          }}
        >
          <div className={styles.panelHeadRow}>
            <div className={styles.panelHead}>
              <h2 className={styles.panelTitle}>Eficiencia do capital</h2>
              <p className={styles.panelSub}>Margem Bruta vs Capital (barras) e GMROI (linha).</p>
            </div>
            <span className={styles.panelActionText}>Ver Detalhes</span>
          </div>
          <p className={styles.efficiencyInsightGood}>
            {efficiencyAvgGmroi == null
              ? "GMROI ainda sem base para insight."
              : `${efficiencyGoodCount} ${leaf0LabelPluralLower} com eficiencia acima da media de GMROI.`}
          </p>
          <div className={styles.chart}>
            <ResponsiveContainer width="100%" height={250}>
              <ComposedChart
                data={efficiencyChartRows}
                margin={{ top: 14, right: 10, left: 0, bottom: 0 }}
                barCategoryGap="28%"
                barGap={2}
              >
                <CartesianGrid stroke="rgba(148, 163, 184, 0.14)" vertical={false} />
                <XAxis
                  dataKey="node_name"
                  tick={renderEfficiencyXAxisTick}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  yAxisId="left"
                  tick={{ fill: "#64748b", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  tickCount={4}
                  tickFormatter={(value) => asBrlCompact(Number(value || 0))}
                />
                <YAxis
                  yAxisId="right"
                  orientation="right"
                  tick={{ fill: "#64748b", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  tickCount={4}
                  domain={[0, (dataMax: number) => Math.max(0.1, Number(dataMax || 0) * 1.1)]}
                  tickFormatter={(value) => `${Number(value || 0).toFixed(1).replace(".", ",")}x`}
                />
                <ReTooltip
                  cursor={{ fill: "rgba(15, 23, 42, 0.04)" }}
                  content={renderEfficiencyTooltip}
                />
                <ReLegend
                  align="right"
                  verticalAlign="top"
                  iconType="square"
                  wrapperStyle={{ top: 0, right: 20, fontSize: 12, fontWeight: 700, color: "#475569" }}
                />
                <ReBar
                  yAxisId="left"
                  dataKey="margem_bruta_brl"
                  name="Margem Bruta (R$)"
                  fill="#cbd5e1"
                  barSize={10}
                  radius={[0, 0, 0, 0]}
                  shape={renderReceitaBar}
                />
                <ReBar
                  yAxisId="left"
                  dataKey="capital_brl"
                  name="Capital (R$)"
                  fill="#8b1538"
                  barSize={10}
                  radius={[0, 0, 0, 0]}
                  shape={renderCapitalBar}
                />
                <ReLine
                  yAxisId="right"
                  dataKey="gmroi"
                  name="GMROI"
                  type="monotone"
                  stroke="#059669"
                  strokeWidth={2}
                  dot={{ r: 3, fill: "#059669" }}
                  activeDot={{ r: 4, fill: "#059669" }}
                  connectNulls
                />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
          <p className={styles.efficiencyInsightBad}>
            {efficiencyWorstNode
              ? `Alerta: ${efficiencyWorstNode.node_name} mostra ineficiencia de capital.`
              : "Sem alerta critico de ineficiencia no recorte atual."}
          </p>
        </article>
      </section>

      <AnalyticsSpotlightDrawer
        open={isPerformanceSpotlightOpen}
        title={`Performance por nivel - ${leaf0LabelPlural}`}
        meta={performanceSpotlightMeta}
        cardClassName={styles.performanceSpotlightCard}
        bodyClassName={styles.performanceSpotlightBody}
        onClose={() => setIsPerformanceSpotlightOpen(false)}
      >
        <section className={styles.spotlightFilters}>
          <label className={`${styles.spotlightField} ${styles.spotlightFieldSearch} ${styles.spotlightFieldSearchOrder}`}>
            <span>Buscar</span>
            <input
              type="search"
              placeholder={`Buscar ${leaf0LabelLower} ou marca`}
              value={spotlightSearch}
              onChange={(event) => setSpotlightSearch(event.target.value)}
            />
          </label>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldBrandOrder}`}>
            <span>Marca</span>
            <SelectMenu
              id="taxonomy-spotlight-brand-filter"
              value={spotlightBrandFilter}
              options={performanceSpotlightBrandSelectOptions}
              onSelect={(value) => setSpotlightBrandFilter(value)}
              classNames={SPOTLIGHT_SELECT_CLASSNAMES}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldGroupOrder}`}>
            <span>{leaf0Label}</span>
            <SelectMenu
              id="taxonomy-spotlight-group-filter"
              value={spotlightGroupFilter}
              options={performanceSpotlightGroupSelectOptions}
              onSelect={(value) => setSpotlightGroupFilter(value)}
              classNames={SPOTLIGHT_SELECT_CLASSNAMES}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber} ${styles.spotlightFieldMarginOrder}`}>
            <span>Margem (%)</span>
            <SelectMenu
              id="taxonomy-spotlight-margin-sort-filter"
              value={spotlightMarginSortDir}
              options={performanceSpotlightSortDirOptions}
              onSelect={(value) => setSpotlightMarginSortDir(normalizeSpotlightSortDir(value))}
              classNames={SPOTLIGHT_SELECT_CLASSNAMES}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber} ${styles.spotlightFieldRevenueOrder}`}>
            <span>Receita</span>
            <SelectMenu
              id="taxonomy-spotlight-revenue-sort-filter"
              value={spotlightRevenueSortDir}
              options={performanceSpotlightSortDirOptions}
              onSelect={(value) => setSpotlightRevenueSortDir(normalizeSpotlightSortDir(value))}
              classNames={SPOTLIGHT_SELECT_CLASSNAMES}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber} ${styles.spotlightFieldTopProductsOrder}`}>
            <span>Top produtos</span>
            <div className={styles.spotlightValueWithOperator}>
              <div className={styles.spotlightOperatorSelect}>
                <SelectMenu
                  id="taxonomy-spotlight-top-products-op-filter"
                  value={spotlightTopProductsOp}
                  options={performanceSpotlightShareOperatorOptions}
                  onSelect={(value) => setSpotlightTopProductsOp(normalizeSpotlightNumericOp(value))}
                  classNames={SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES}
                />
              </div>
              <div className={styles.spotlightInputWithSuffix}>
                <input
                  className={styles.spotlightInputCompact}
                  type="number"
                  min={0}
                  max={999}
                  step={1}
                  value={spotlightTopProductsInput}
                  onChange={(event) => setSpotlightTopProductsInput(clampTopProductsInput(event.target.value))}
                  onBlur={(event) => setSpotlightTopProductsInput(clampTopProductsInput(event.target.value))}
                />
              </div>
            </div>
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber} ${styles.spotlightFieldShareOrder}`}>
            <span>Participacao acumulada (%)</span>
            <div className={styles.spotlightValueWithOperator}>
              <div className={styles.spotlightOperatorSelect}>
                <SelectMenu
                  id="taxonomy-spotlight-share-op-filter"
                  value={spotlightShareOp}
                  options={performanceSpotlightShareOperatorOptions}
                  onSelect={(value) => setSpotlightShareOp(normalizeSpotlightNumericOp(value))}
                  classNames={SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES}
                />
              </div>
              <div className={styles.spotlightInputWithSuffix}>
                <input
                  className={styles.spotlightInputCompact}
                  type="number"
                  min={0}
                  max={100}
                  step={1}
                  value={spotlightShareTargetPctInput}
                  onChange={(event) => setSpotlightShareTargetPctInput(clampSharePctInput(event.target.value))}
                  onBlur={(event) => setSpotlightShareTargetPctInput(clampSharePctInput(event.target.value))}
                />
                <span className={styles.spotlightInputSuffixPct}> %</span>
              </div>
            </div>
          </div>
        </section>
        <div className={`${styles.spotlightGrid} ${styles.performanceSpotlightGrid}`}>
          {performanceSpotlightRows.map((row) => (
            <article key={`${row.node_id}:${row.node_name}`} className={styles.categoryCard}>
              <div className={styles.categoryHeader}>
                <div className={styles.categoryIcon}>{taxonomyEmojiForNodeName(row.node_name)}</div>
                <div className={styles.categoryName}>{row.node_name}</div>
              </div>
              <div className={styles.categoryMetrics}>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>Receita</div>
                  <div className={styles.metricValue}>{asBrlCompact(row.revenue_brl)}</div>
                  {trendPctText(row.trend.revenue_delta_mom_pct) ? (
                    <div className={trendClassBySign(row.trend.revenue_delta_mom_pct)}>{trendPctText(row.trend.revenue_delta_mom_pct)}</div>
                  ) : null}
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>Margem (%)</div>
                  <div className={styles.metricValue}>{row.margin_pct == null ? "-" : asPct(row.margin_pct)}</div>
                  {trendPpText(row.trend.margin_delta_mom_pp) ? (
                    <div className={trendClassBySign(row.trend.margin_delta_mom_pp)}>{trendPpText(row.trend.margin_delta_mom_pp)}</div>
                  ) : null}
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>Participacao</div>
                  <div className={styles.metricValue}>{asPct(row.share_pct)}</div>
                  {trendPpText(row.trend.share_delta_mom_pp) ? (
                    <div className={trendClassBySign(row.trend.share_delta_mom_pp)}>{trendPpText(row.trend.share_delta_mom_pp)}</div>
                  ) : null}
                </div>
              </div>
            </article>
          ))}
        </div>
        {!performanceSpotlightRows.length ? (
          <p className={styles.spotlightEmpty}>Nenhum {leaf0LabelLower} encontrado com os filtros atuais.</p>
        ) : null}
      </AnalyticsSpotlightDrawer>

      <AnalyticsSpotlightDrawer
        open={isEfficiencySpotlightOpen}
        title={`Eficiencia do capital - ${leaf0LabelPlural}`}
        meta={`${efficiencySpotlightRows.length} ${leaf0LabelPluralLower} no recorte`}
        cardClassName={styles.efficiencySpotlightCard}
        bodyClassName={styles.performanceSpotlightBody}
        bodyRef={efficiencySpotlightBodyRef}
        onClose={() => setIsEfficiencySpotlightOpen(false)}
      >
        <section className={styles.spotlightFilters}>
          <label className={`${styles.spotlightField} ${styles.spotlightFieldSearch}`}>
            <span>Buscar</span>
            <input
              type="search"
              placeholder={`Buscar ${leaf0LabelLower}`}
              value={efficiencySpotlightSearch}
              onChange={(event) => setEfficiencySpotlightSearch(event.target.value)}
            />
          </label>
          <div className={styles.spotlightField}>
            <span>{leaf0Label}</span>
            <FilterDropdown
              id="taxonomy-efficiency-spotlight-group-filter"
              values={efficiencySpotlightGroupFilters}
              options={efficiencySpotlightGroupOptions}
              selectionMode="duo"
              onSelect={(value) => setEfficiencySpotlightGroupFilters((current) => toggleMultiValue(current, value))}
              classNamesOverrides={{
                wrap: styles.spotlightFilterSelectWrap,
              }}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber}`}>
            <span>GMROI</span>
            <div className={styles.spotlightValueWithOperator}>
              <div className={styles.spotlightOperatorSelect}>
                <SelectMenu
                  id="taxonomy-efficiency-spotlight-gmroi-op"
                  value={efficiencySpotlightGmroiOp}
                  options={performanceSpotlightShareOperatorOptions}
                  onSelect={(value) => setEfficiencySpotlightGmroiOp(normalizeSpotlightNumericOp(value))}
                  classNames={SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES}
                />
              </div>
              <div className={styles.spotlightInputWithSuffix}>
                <input
                  className={styles.spotlightInputCompact}
                  type="number"
                  min={0}
                  step={0.01}
                  placeholder="Todos"
                  value={efficiencySpotlightGmroiInput}
                  onChange={(event) => setEfficiencySpotlightGmroiInput(clampNonNegativeDecimalInput(event.target.value))}
                  onBlur={(event) => setEfficiencySpotlightGmroiInput(clampNonNegativeDecimalInput(event.target.value))}
                />
              </div>
            </div>
          </div>
        </section>

        <article className={`${styles.panel} ${styles.efficiencySpotlightChartPanel}`}>
          <div className={`${styles.chart} ${styles.efficiencySpotlightChart}`}>
            {isEfficiencySpotlightOpen ? (
              <ResponsiveContainer key={`efficiency-open-${efficiencySpotlightChartRows.length}`} width="100%" height="100%">
                <ComposedChart
                  data={efficiencySpotlightChartRows}
                  margin={{ top: 14, right: 10, left: 0, bottom: 0 }}
                  barCategoryGap="24%"
                  barGap={2}
                >
                  <CartesianGrid stroke="rgba(148, 163, 184, 0.14)" vertical={false} />
                  <XAxis
                    dataKey="node_name"
                    tick={renderEfficiencyXAxisTick}
                    axisLine={false}
                    tickLine={false}
                    interval="preserveStartEnd"
                    minTickGap={24}
                    tickMargin={6}
                  />
                  <YAxis
                    yAxisId="left"
                    tick={{ fill: "#64748b", fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickCount={4}
                    tickFormatter={(value) => asBrlCompact(Number(value || 0))}
                  />
                  <YAxis
                    yAxisId="right"
                    orientation="right"
                    tick={{ fill: "#64748b", fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickCount={4}
                    domain={[0, efficiencySpotlightGmroiAxisMax]}
                    tickFormatter={(value) => `${Number(value || 0).toFixed(2).replace(".", ",")}x`}
                  />
                  <ReTooltip content={renderEfficiencyTooltip} />
                  <ReLegend wrapperStyle={{ fontSize: 11 }} />
                  <ReBar
                    yAxisId="left"
                    dataKey="margem_bruta_brl"
                    name="Margem Bruta (R$)"
                    fill="#cbd5e1"
                    barSize={10}
                    radius={[0, 0, 0, 0]}
                    shape={renderReceitaBar}
                  />
                  <ReBar
                    yAxisId="left"
                    dataKey="capital_brl"
                    name="Capital (R$)"
                    fill="#8b1538"
                    barSize={10}
                    radius={[0, 0, 0, 0]}
                    shape={renderCapitalBar}
                  />
                  <ReLine
                    yAxisId="right"
                    dataKey="gmroi"
                    name="GMROI"
                    type="monotone"
                    stroke="#059669"
                    strokeWidth={2}
                    dot={{ r: 3, fill: "#059669" }}
                    activeDot={{ r: 4, fill: "#059669" }}
                    connectNulls
                  />
                </ComposedChart>
              </ResponsiveContainer>
            ) : null}
          </div>
        </article>

        <section className={styles.allocationSpotlightTableWrap}>
          <div className={styles.panelHeadInline}>
            <h3 className={styles.panelTitleMini}>Detalhe de Eficiencia</h3>
          </div>
          <SpotlightDataTable
            rows={efficiencySpotlightRows}
            columns={efficiencySpotlightColumns}
            emptyText="Sem itens no recorte atual."
            rowKey={(row) => `eff-row-${row.node_name}`}
            defaultSort={{ id: "gmroi", desc: false }}
          />
        </section>
      </AnalyticsSpotlightDrawer>

      <AnalyticsSpotlightDrawer
        open={isAllocationSpotlightOpen}
        title={`Mapa de alocacao de capital - ${leaf0LabelPlural}`}
        meta={allocationSpotlightMeta}
        cardClassName={styles.allocationSpotlightCard}
        bodyClassName={styles.allocationSpotlightBody}
        onClose={() => {
          setIsAllocationSpotlightOpen(false);
          setAllocationSpotlightSelectedNodeId(null);
        }}
      >
        <section className={`${styles.spotlightFilters} ${styles.allocationSpotlightFilters}`}>
          <label className={`${styles.spotlightField} ${styles.spotlightFieldSearch}`}>
            <span>Buscar</span>
            <input
              type="search"
              placeholder={`Buscar ${leaf0LabelLower}`}
              value={allocationSpotlightSearch}
              onChange={(event) => setAllocationSpotlightSearch(event.target.value)}
            />
          </label>
          <div className={styles.spotlightField}>
            <span>Risco</span>
            <SelectMenu
              id="taxonomy-allocation-spotlight-risk-filter"
              value={allocationSpotlightRiskFilter}
              options={allocationSpotlightRiskOptions}
              onSelect={(value) => setAllocationSpotlightRiskFilter(
                value === "high" || value === "medium" || value === "low" ? value : "all",
              )}
              classNames={SPOTLIGHT_SELECT_CLASSNAMES}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber}`}>
            <span>Top produtos</span>
            <div className={styles.spotlightValueWithOperator}>
              <div className={styles.spotlightOperatorSelect}>
                <SelectMenu
                  id="taxonomy-allocation-spotlight-top-op"
                  value={allocationSpotlightTopOp}
                  options={performanceSpotlightShareOperatorOptions}
                  onSelect={(value) => setAllocationSpotlightTopOp(normalizeSpotlightNumericOp(value))}
                  classNames={SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES}
                />
              </div>
              <div className={styles.spotlightInputWithSuffix}>
                <input
                  className={styles.spotlightInputCompact}
                  type="number"
                  min={0}
                  max={999}
                  step={1}
                  value={allocationSpotlightTopInput}
                  onChange={(event) => setAllocationSpotlightTopInput(clampTopProductsInput(event.target.value))}
                  onBlur={(event) => setAllocationSpotlightTopInput(clampTopProductsInput(event.target.value))}
                />
              </div>
            </div>
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber}`}>
            <span>Participacao acumulada (%)</span>
            <div className={styles.spotlightValueWithOperator}>
              <div className={styles.spotlightOperatorSelect}>
                <SelectMenu
                  id="taxonomy-allocation-spotlight-share-op"
                  value={allocationSpotlightShareOp}
                  options={performanceSpotlightShareOperatorOptions}
                  onSelect={(value) => setAllocationSpotlightShareOp(normalizeSpotlightNumericOp(value))}
                  classNames={SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES}
                />
              </div>
              <div className={styles.spotlightInputWithSuffix}>
                <input
                  className={styles.spotlightInputCompact}
                  type="number"
                  min={0}
                  max={100}
                  step={1}
                  value={allocationSpotlightShareTargetInput}
                  onChange={(event) => setAllocationSpotlightShareTargetInput(clampSharePctInput(event.target.value))}
                  onBlur={(event) => setAllocationSpotlightShareTargetInput(clampSharePctInput(event.target.value))}
                />
                <span className={styles.spotlightInputSuffixPct}> %</span>
              </div>
            </div>
          </div>
        </section>

        <section className={styles.allocationSpotlightTop}>
          <div className={styles.allocationSpotlightTreemap} style={{ height: `${allocationTopHeight}px` }}>
            <Chart
              key={allocationSpotlightTreeKey}
              type="treemap"
              data={allocationSpotlightTreeData}
              options={allocationSpotlightTreeOptions}
              redraw
            />
          </div>
          <aside ref={allocationPriorityRef} className={styles.allocationSpotlightPriority}>
            <div className={styles.panelHeadInline}>
              <h3 className={styles.panelTitleMini}>Top Prioridades</h3>
              {allocationSpotlightSelectedNodeId != null ? (
                <button
                  type="button"
                  className={styles.linkBtn}
                  onClick={() => setAllocationSpotlightSelectedNodeId(null)}
                >
                  Limpar seleção
                </button>
              ) : null}
            </div>
            <div className={styles.allocationPriorityList}>
              {allocationSpotlightPriorityRows.map((row, idx) => (
                <button
                  type="button"
                  key={`alloc-priority-${row.node_id}`}
                  className={`${styles.allocationPriorityItem} ${allocationSpotlightSelectedNodeId === row.node_id ? styles.allocationPriorityItemActive : ""}`}
                  onClick={() => setAllocationSpotlightSelectedNodeId((curr) => (curr === row.node_id ? null : row.node_id))}
                >
                  <span className={styles.allocationPriorityRank}>{idx + 1}</span>
                  <span className={styles.allocationPriorityName}>{row.node_name}</span>
                  <span className={styles.allocationPriorityMeta}>
                    {asBrlCompact(row.capital_brl)} | {asPct(row.risk_pct)}
                  </span>
                </button>
              ))}
              {!allocationSpotlightPriorityRows.length ? <p className={styles.spotlightEmpty}>Sem itens para priorizar no recorte.</p> : null}
            </div>
          </aside>
        </section>

        <section className={styles.allocationSpotlightTableWrap}>
          <div className={styles.panelHeadInline}>
            <h3 className={styles.panelTitleMini}>Detalhe do Capital</h3>
          </div>
          <SpotlightDataTable
            rows={allocationSpotlightSelectedRows}
            columns={allocationSpotlightColumns}
            emptyText="Sem itens no recorte atual."
            rowKey={(row) => `alloc-row-${row.node_id}`}
            defaultSort={{ id: "capital_brl", desc: true }}
          />
        </section>
      </AnalyticsSpotlightDrawer>

      <AnalyticsSpotlightDrawer
        open={isTopMarginSpotlightOpen}
        title={`Top ${leaf0LabelPluralLower} por margem`}
        meta={`${topMarginSpotlightRows.length} ${leaf0LabelPluralLower} | margem total ${asBrlCompact(topMarginSpotlightTotalGrossMargin)}`}
        cardClassName={styles.efficiencySpotlightCard}
        bodyClassName={styles.performanceSpotlightBody}
        onClose={() => setIsTopMarginSpotlightOpen(false)}
      >
        <section className={styles.spotlightFilters}>
          <label className={`${styles.spotlightField} ${styles.spotlightFieldSearch}`}>
            <span>Buscar</span>
            <input
              type="search"
              placeholder={`Buscar ${leaf0LabelLower}`}
              value={topMarginSpotlightSearch}
              onChange={(event) => setTopMarginSpotlightSearch(event.target.value)}
            />
          </label>
          <div className={styles.spotlightField}>
            <span>{leaf0Label}</span>
            <FilterDropdown
              id="taxonomy-top-margin-spotlight-group-filter"
              values={topMarginSpotlightGroupFilters}
              options={topMarginSpotlightGroupOptions}
              selectionMode="duo"
              onSelect={(value) => setTopMarginSpotlightGroupFilters((current) => toggleMultiValue(current, value))}
              classNamesOverrides={{ wrap: styles.spotlightFilterSelectWrap }}
            />
          </div>
          <div className={styles.spotlightField}>
            <span>Tipo de Curva</span>
            <FilterDropdown
              id="taxonomy-top-margin-spotlight-curve-type"
              value={topMarginSpotlightCurveType}
              options={topMarginSpotlightCurveOptions}
              onSelect={(value) =>
                setTopMarginSpotlightCurveType(
                  value === "margin_pct" || value === "gmroi" ? value : "gross_margin"
                )
              }
              classNamesOverrides={{ wrap: styles.spotlightFilterSelectWrap }}
            />
          </div>
        </section>

        <section className={styles.concStats}>
          <article className={styles.concStatCard}>
            <span>Margem Bruta Total</span>
            <strong>{asBrlCompact(topMarginSpotlightTotalGrossMargin)}</strong>
          </article>
          <article className={styles.concStatCard}>
            <span>Top 5 share margem</span>
            <strong>{asPct(topMarginSpotlightTop5Share)}</strong>
          </article>
          <article className={styles.concStatCard}>
            <span>Top 10 share margem</span>
            <strong>{asPct(topMarginSpotlightTop10Share)}</strong>
          </article>
        </section>

        <article className={`${styles.panel} ${styles.efficiencySpotlightChartPanel}`}>
          <div className={`${styles.chart} ${styles.efficiencySpotlightChart}`}>
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart
                data={topMarginSpotlightChartRows}
                layout="vertical"
                margin={{ top: 10, right: 18, left: 8, bottom: 6 }}
              >
                <CartesianGrid stroke="rgba(148, 163, 184, 0.14)" horizontal={false} />
                <XAxis
                  type="number"
                  tick={{ fill: "#64748b", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  tickFormatter={(value) =>
                    topMarginSpotlightCurveType === "gmroi"
                      ? `${Number(value || 0).toFixed(1).replace(".", ",")}x`
                      : topMarginSpotlightCurveType === "margin_pct"
                        ? `${Number(value || 0).toFixed(0)}%`
                        : asBrlCompact(Number(value || 0))
                  }
                />
                <YAxis
                  type="category"
                  dataKey="node_short"
                  width={180}
                  tick={{ fill: "#64748b", fontSize: 11, fontWeight: 700 }}
                  axisLine={false}
                  tickLine={false}
                />
                <ReTooltip
                  formatter={(value: number) => [
                    topMarginSpotlightCurveType === "gmroi"
                      ? `${Number(value || 0).toFixed(2).replace(".", ",")}x`
                      : topMarginSpotlightCurveType === "margin_pct"
                        ? `${Number(value || 0).toFixed(1)}%`
                        : asBrlCompact(Number(value || 0)),
                    topMarginSpotlightMetricLabel,
                  ]}
                  labelFormatter={(_, payload) => String((payload?.[0]?.payload as { node_name?: string } | undefined)?.node_name || "")}
                />
                <ReBar
                  dataKey="metric_value"
                  name={topMarginSpotlightMetricLabel}
                  fill="#8b1538"
                  radius={[0, 6, 6, 0]}
                />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        </article>

        <section className={styles.allocationSpotlightTableWrap}>
          <div className={styles.panelHeadInline}>
            <h3 className={styles.panelTitleMini}>Detalhe da Margem</h3>
          </div>
          <SpotlightDataTable
            rows={topMarginSpotlightRows}
            columns={topMarginSpotlightColumns}
            emptyText="Sem grupos no recorte atual."
            rowKey={(row) => `top-margin-row-${row.node_id}`}
            defaultSort={{ id: "gross_margin_brl", desc: true }}
          />
        </section>
      </AnalyticsSpotlightDrawer>

      <AnalyticsSpotlightDrawer
        open={isRiskSpotlightOpen}
        title={`${leaf0LabelPlural} em risco`}
        meta={`${riskSpotlightRows.length} ${leaf0LabelPluralLower} | capital em risco ${asBrlCompact(riskSpotlightTotalCapitalAtRisk)}`}
        cardClassName={styles.efficiencySpotlightCard}
        bodyClassName={styles.performanceSpotlightBody}
        onClose={() => setIsRiskSpotlightOpen(false)}
      >
        <section className={styles.spotlightFilters}>
          <label className={`${styles.spotlightField} ${styles.spotlightFieldSearch}`}>
            <span>Buscar</span>
            <input
              type="search"
              placeholder={`Buscar ${leaf0LabelLower}`}
              value={riskSpotlightSearch}
              onChange={(event) => setRiskSpotlightSearch(event.target.value)}
            />
          </label>
          <div className={styles.spotlightField}>
            <span>{leaf0Label}</span>
            <FilterDropdown
              id="taxonomy-risk-spotlight-group-filter"
              values={riskSpotlightGroupFilters}
              options={riskSpotlightGroupOptions}
              selectionMode="duo"
              onSelect={(value) => setRiskSpotlightGroupFilters((current) => toggleMultiValue(current, value))}
              classNamesOverrides={{ wrap: styles.spotlightFilterSelectWrap }}
            />
          </div>
          <div className={styles.spotlightField}>
            <span>Risco</span>
            <FilterDropdown
              id="taxonomy-risk-spotlight-risk-filter"
              value={riskSpotlightRiskFilter}
              options={riskSpotlightRiskOptions}
              onSelect={(value) =>
                setRiskSpotlightRiskFilter(
                  value === "high" || value === "medium" || value === "low" ? value : "all"
                )
              }
              classNamesOverrides={{ wrap: styles.spotlightFilterSelectWrap }}
            />
          </div>
          <div className={styles.spotlightField}>
            <span>Métrica</span>
            <FilterDropdown
              id="taxonomy-risk-spotlight-metric-filter"
              value={riskSpotlightMetric}
              options={riskSpotlightMetricOptions}
              onSelect={(value) =>
                setRiskSpotlightMetric(value === "risk_pct" ? "risk_pct" : "capital_at_risk_brl")
              }
              classNamesOverrides={{ wrap: styles.spotlightFilterSelectWrap }}
            />
          </div>
          <div className={`${styles.spotlightField} ${styles.spotlightFieldNumber}`}>
            <span>Valor</span>
            <div className={styles.spotlightValueWithOperator}>
              <div className={styles.spotlightOperatorSelect}>
                <SelectMenu
                  id="taxonomy-risk-spotlight-metric-op"
                  value={riskSpotlightMetricOp}
                  options={performanceSpotlightShareOperatorOptions}
                  onSelect={(value) => setRiskSpotlightMetricOp(normalizeSpotlightNumericOp(value))}
                  classNames={SPOTLIGHT_OPERATOR_SELECT_CLASSNAMES}
                />
              </div>
              <div className={styles.spotlightInputWithSuffix}>
                <input
                  className={styles.spotlightInputCompact}
                  type="number"
                  min={0}
                  step={riskSpotlightMetric === "risk_pct" ? 0.1 : 1}
                  placeholder="Todos"
                  value={riskSpotlightMetricInput}
                  onChange={(event) => setRiskSpotlightMetricInput(clampNonNegativeDecimalInput(event.target.value))}
                  onBlur={(event) => setRiskSpotlightMetricInput(clampNonNegativeDecimalInput(event.target.value))}
                />
                {riskSpotlightMetric === "risk_pct" ? <span className={styles.spotlightInputSuffixPct}> %</span> : null}
              </div>
            </div>
          </div>
        </section>

        <section className={styles.concStats}>
          <article className={styles.concStatCard}>
            <span>Capital em Risco</span>
            <strong>{asBrlCompact(riskSpotlightTotalCapitalAtRisk)}</strong>
          </article>
          <article className={styles.concStatCard}>
            <span>Alto risco</span>
            <strong>{riskSpotlightHighRiskCount.toLocaleString("pt-BR")}</strong>
          </article>
          <article className={styles.concStatCard}>
            <span>Share alto risco</span>
            <strong>{asPct(riskSpotlightHighRiskShare)}</strong>
          </article>
        </section>

        <article className={`${styles.panel} ${styles.efficiencySpotlightChartPanel}`}>
          <div className={`${styles.chart} ${styles.efficiencySpotlightChart}`}>
            <ResponsiveContainer width="100%" height="100%">
              <ScatterChart margin={{ top: 16, right: 20, left: 20, bottom: 28 }}>
                <CartesianGrid stroke="rgba(148, 163, 184, 0.14)" />
                <XAxis
                  type="number"
                  dataKey="capital_brl"
                  name="Capital"
                  label={{
                    value: "Capital (R$)",
                    position: "insideBottom",
                    offset: -10,
                    fill: "#94a3b8",
                    fontSize: 12,
                  }}
                  tick={{ fill: "#64748b", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  tickFormatter={(value) => asBrlCompact(Number(value || 0))}
                />
                <YAxis
                  type="number"
                  dataKey="risk_pct"
                  name="Risco"
                  width={48}
                  label={{
                    value: "Risco (%)",
                    angle: -90,
                    position: "insideLeft",
                    fill: "#94a3b8",
                    fontSize: 12,
                  }}
                  tick={{ fill: "#64748b", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  tickFormatter={(value) => `${Number(value || 0).toFixed(0)}%`}
                />
                <ZAxis type="number" dataKey="revenue_brl" range={[70, 500]} />
                <ReTooltip
                  formatter={(value: number, name: string) => [
                    name === "Risco"
                      ? `${Number(value || 0).toFixed(1)}%`
                      : asBrlCompact(Number(value || 0)),
                    name,
                  ]}
                  labelFormatter={(_, payload) => String((payload?.[0]?.payload as { node_name?: string } | undefined)?.node_name || "")}
                />
                <Scatter
                  name="Grupos em risco"
                  data={riskSpotlightChartRows}
                  shape={renderRiskScatterPoint}
                />
              </ScatterChart>
            </ResponsiveContainer>
          </div>
        </article>

        <section className={styles.allocationSpotlightTableWrap}>
          <div className={styles.panelHeadInline}>
            <h3 className={styles.panelTitleMini}>Detalhe do Risco</h3>
          </div>
          <SpotlightDataTable
            rows={riskSpotlightRows}
            columns={riskSpotlightColumns}
            emptyText="Sem grupos no recorte atual."
            rowKey={(row) => `risk-row-${row.node_id}`}
            defaultSort={{ id: "capital_at_risk_brl", desc: true }}
          />
        </section>
      </AnalyticsSpotlightDrawer>

      <AnalyticsSpotlightDrawer
        open={isAbcSpotlightOpen}
        title={`Mix ABC - ${leaf0LabelPlural}`}
        meta="Curva de Pareto por receita e classificacao A/B/C"
        cardClassName={styles.abcSpotlightCard}
        bodyClassName={styles.abcSpotlightBody}
        onClose={() => setIsAbcSpotlightOpen(false)}
      >
        <section className={styles.spotlightFilters}>
          <label className={`${styles.spotlightField} ${styles.spotlightFieldSearch}`}>
            <span>Buscar</span>
            <input
              type="search"
              placeholder={`Buscar ${leaf0LabelLower}`}
              value={abcSpotlightSearch}
              onChange={(event) => setAbcSpotlightSearch(event.target.value)}
            />
          </label>
          <div className={styles.spotlightField}>
            <span>Tipo de Curva</span>
            <FilterDropdown
              id="taxonomy-abc-spotlight-curve-type"
              value={abcSpotlightCurveType}
              options={abcSpotlightCurveOptions}
              onSelect={(value) => setAbcSpotlightCurveType(value === "gross_margin" ? "gross_margin" : "revenue")}
              classNamesOverrides={{
                wrap: styles.spotlightFilterSelectWrap,
              }}
            />
          </div>
          <div className={styles.spotlightField}>
            <span>{leaf0Label}</span>
            <FilterDropdown
              id="taxonomy-abc-spotlight-group-filter"
              values={abcSpotlightGroupFilters}
              options={abcSpotlightGroupOptions}
              selectionMode="duo"
              onSelect={(value) => setAbcSpotlightGroupFilters((current) => toggleMultiValue(current, value))}
              classNamesOverrides={{
                wrap: styles.spotlightFilterSelectWrap,
              }}
            />
          </div>
          <div className={styles.spotlightField}>
            <span>Faixa ABC</span>
            <FilterDropdown
              id="taxonomy-abc-spotlight-band-filter"
              value={abcSpotlightBandFilter}
              options={abcSpotlightBandOptions}
              onSelect={(value) =>
                setAbcSpotlightBandFilter(
                  value === "A" || value === "B" || value === "C" ? value : "all"
                )
              }
              classNamesOverrides={{
                wrap: styles.spotlightFilterSelectWrap,
              }}
            />
          </div>
        </section>
        <section className={`${styles.panel} ${styles.abcSpotlightBandsPanel}`}>
          <div className={styles.panelHeadInline}>
            <h3 className={styles.panelTitleMini}>Faixas ABC</h3>
          </div>
          <div className={`${styles.abcColumns} ${styles.abcSpotlightColumns}`}>
            <div className={`${styles.abcColumn} ${styles.abcColumnSep} ${styles.abcToneA}`}>
              <div className={styles.abcTop}><span className={`${styles.abcKey} ${styles.abcTopKey}`}>A</span><span className={`${styles.abcCount} ${styles.abcTopKey}`}>{String(abcSpotlightBandSummary.A.count || 0)}</span></div>
              <div className={styles.abcBottom}>
                <span className={styles.abcPct}>{`${asPct(abcSpotlightBandSummary.A.entityPct)} ${leaf0LabelPluralLower}`}</span>
                <span className={styles.abcPct}>{`${asPct(abcSpotlightBandSummary.A.metricPct)} ${abcSpotlightCurveType === "gross_margin" ? "margem bruta" : "receita"}`}</span>
              </div>
            </div>
            <div className={`${styles.abcColumn} ${styles.abcColumnSep} ${styles.abcToneB}`}>
              <div className={styles.abcTop}><span className={`${styles.abcKey} ${styles.abcTopKey}`}>B</span><span className={`${styles.abcCount} ${styles.abcTopKey}`}>{String(abcSpotlightBandSummary.B.count || 0)}</span></div>
              <div className={styles.abcBottom}>
                <span className={styles.abcPct}>{`${asPct(abcSpotlightBandSummary.B.entityPct)} ${leaf0LabelPluralLower}`}</span>
                <span className={styles.abcPct}>{`${asPct(abcSpotlightBandSummary.B.metricPct)} ${abcSpotlightCurveType === "gross_margin" ? "margem bruta" : "receita"}`}</span>
              </div>
            </div>
            <div className={`${styles.abcColumn} ${styles.abcToneC}`}>
              <div className={styles.abcTop}><span className={`${styles.abcKey} ${styles.abcTopKey}`}>C</span><span className={`${styles.abcCount} ${styles.abcTopKey}`}>{String(abcSpotlightBandSummary.C.count || 0)}</span></div>
              <div className={styles.abcBottom}>
                <span className={styles.abcPct}>{`${asPct(abcSpotlightBandSummary.C.entityPct)} ${leaf0LabelPluralLower}`}</span>
                <span className={styles.abcPct}>{`${asPct(abcSpotlightBandSummary.C.metricPct)} ${abcSpotlightCurveType === "gross_margin" ? "margem bruta" : "receita"}`}</span>
              </div>
            </div>
          </div>
        </section>

        <section className={styles.abcSpotlightTop}>
          <article className={styles.abcSpotlightParetoPanel} style={{ height: `${abcChartHeight}px` }}>
            <div className={styles.panelHeadInline}>
              <h3 className={styles.panelTitleMini}>Curva de Pareto</h3>
            </div>
            <div className={styles.abcSpotlightChart}>
              <div
                className={styles.abcSpotlightBandBackdrop}
                aria-hidden="true"
                style={{
                  inset: `${ABC_PARETO_MARGIN.top}px ${ABC_PARETO_Y_AXIS_WIDTH + ABC_PARETO_MARGIN.right}px ${ABC_PARETO_X_AXIS_HEIGHT + ABC_PARETO_MARGIN.bottom}px ${ABC_PARETO_Y_AXIS_WIDTH + ABC_PARETO_MARGIN.left}px`,
                }}
              >
                {abcParetoBandSpans.map((span) => (
                  <div
                    key={`abc-band-span-${span.band}`}
                    className={`${styles.abcSpotlightBandBackdropSegment} ${
                      span.band === "A"
                        ? styles.abcSpotlightBandBackdropA
                        : span.band === "B"
                        ? styles.abcSpotlightBandBackdropB
                        : styles.abcSpotlightBandBackdropC
                    }`}
                    style={{ left: `${span.leftPct}%`, width: `${span.widthPct}%` }}
                  >
                    <span>{span.band}</span>
                  </div>
                ))}
              </div>
              <ResponsiveContainer width="100%" height="100%">
                <ComposedChart data={abcParetoChartRows} margin={ABC_PARETO_MARGIN}>
                  <CartesianGrid stroke="rgba(148, 163, 184, 0.14)" vertical={false} />
                  <XAxis dataKey="node_short" tick={{ fill: "#64748b", fontSize: 10 }} axisLine={false} tickLine={false} interval={0} angle={-22} textAnchor="end" height={ABC_PARETO_X_AXIS_HEIGHT} />
                  <YAxis
                    yAxisId="left"
                    width={ABC_PARETO_Y_AXIS_WIDTH}
                    tick={{ fill: "#64748b", fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickCount={5}
                    domain={[0, 100]}
                    tickFormatter={(value) => `${Number(value || 0).toFixed(0)}%`}
                  />
                  <YAxis
                    yAxisId="right"
                    orientation="right"
                    width={ABC_PARETO_Y_AXIS_WIDTH}
                    tick={{ fill: "#64748b", fontSize: 11 }}
                    axisLine={false}
                    tickLine={false}
                    tickCount={5}
                    domain={[0, 100]}
                    tickFormatter={(value) => `${Number(value || 0).toFixed(0)}%`}
                  />
                  <ReTooltip
                    formatter={(value: number, name: string) => [`${Number(value || 0).toFixed(1)}%`, name === "cum_share_pct" ? "Acumulado" : "Share"]}
                    labelFormatter={(_, payload) => String((payload?.[0]?.payload as { node_name?: string } | undefined)?.node_name || "")}
                  />
                  <ReLegend wrapperStyle={{ fontSize: 11 }} />
                  <ReBar
                    yAxisId="left"
                    dataKey="share_pct"
                    name={abcSpotlightCurveType === "gross_margin" ? "Share da Margem" : "Share da Receita"}
                    fill="#8b1538"
                    radius={[3, 3, 0, 0]}
                  />
                  <ReLine yAxisId="right" dataKey="cum_share_pct" name="Acumulado" type="monotone" stroke="#059669" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 4 }} />
                </ComposedChart>
              </ResponsiveContainer>
            </div>
          </article>

          <aside ref={abcTopBandsRef} className={styles.abcSpotlightTopBands}>
            {(["A", "B", "C"] as const).map((band) => (
              <article key={`abc-top-${band}`} className={styles.abcSpotlightBandCard}>
                <h4>{`Top ${band}`}</h4>
                <div className={styles.abcSpotlightBandList}>
                  {(abcTopBands[band] || []).map((row, idx) => (
                    <div key={`abc-top-${band}-${row.node_id}`} className={styles.abcSpotlightBandRow}>
                      <span>{`${idx + 1}. ${String(row.node_name || "")}`}</span>
                      <strong>
                        {asBrlCompact(
                          Number(
                            abcSpotlightCurveType === "gross_margin"
                              ? row.gross_margin_brl
                              : row.revenue_brl
                          ) || 0
                        )}
                      </strong>
                    </div>
                  ))}
                  {!(abcTopBands[band] || []).length ? <p className={styles.spotlightEmpty}>Sem itens na faixa.</p> : null}
                </div>
              </article>
            ))}
          </aside>
        </section>

        <section className={styles.abcSpotlightTableWrap}>
          <div className={styles.panelHeadInline}>
            <h3 className={styles.panelTitleMini}>Pareto por {leaf0LabelLower}</h3>
          </div>
          <SpotlightDataTable
            rows={abcSpotlightRows}
            columns={abcSpotlightColumns}
            emptyText="Sem itens no recorte atual."
            rowKey={(row) => `abc-row-${row.node_id}`}
            defaultSort={{ id: "revenue_brl", desc: true }}
          />
        </section>
      </AnalyticsSpotlightDrawer>

      <section className={styles.bottomGrid}>
        <div className={styles.statCol}>
          <article
            className={`${styles.statCard} ${styles.statCardAbc} ${styles.statCardClickable}`}
            role="button"
            tabIndex={0}
            onClick={() => setIsAbcSpotlightOpen(true)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                setIsAbcSpotlightOpen(true);
              }
            }}
          >
            <p>Mix ABC</p>
            <div className={styles.abcColumns}>
              <div className={`${styles.abcColumn} ${styles.abcColumnSep} ${styles.abcToneA}`}>
                <div className={styles.abcTop}><span className={`${styles.abcKey} ${styles.abcTopKey}`}>A</span><span className={`${styles.abcCount} ${styles.abcTopKey}`}>{String(dto?.scope.analysis_cards?.abc_mix.a_count || 0)}</span></div>
                <div className={styles.abcBottom}>
                  <span className={styles.abcPct}>{`${asPct(abcCardGroupSharePct(dto?.scope.analysis_cards?.abc_mix.a_count || 0))} ${leaf0LabelPluralLower}`}</span>
                  <span className={styles.abcPct}>{`${asPct(dto?.scope.analysis_cards?.abc_mix.a_revenue_pct || 0)} receita`}</span>
                </div>
              </div>
              <div className={`${styles.abcColumn} ${styles.abcColumnSep} ${styles.abcToneB}`}>
                <div className={styles.abcTop}><span className={`${styles.abcKey} ${styles.abcTopKey}`}>B</span><span className={`${styles.abcCount} ${styles.abcTopKey}`}>{String(dto?.scope.analysis_cards?.abc_mix.b_count || 0)}</span></div>
                <div className={styles.abcBottom}>
                  <span className={styles.abcPct}>{`${asPct(abcCardGroupSharePct(dto?.scope.analysis_cards?.abc_mix.b_count || 0))} ${leaf0LabelPluralLower}`}</span>
                  <span className={styles.abcPct}>{`${asPct(dto?.scope.analysis_cards?.abc_mix.b_revenue_pct || 0)} receita`}</span>
                </div>
              </div>
              <div className={`${styles.abcColumn} ${styles.abcToneC}`}>
                <div className={styles.abcTop}><span className={`${styles.abcKey} ${styles.abcTopKey}`}>C</span><span className={`${styles.abcCount} ${styles.abcTopKey}`}>{String(dto?.scope.analysis_cards?.abc_mix.c_count || 0)}</span></div>
                <div className={styles.abcBottom}>
                  <span className={styles.abcPct}>{`${asPct(abcCardGroupSharePct(dto?.scope.analysis_cards?.abc_mix.c_count || 0))} ${leaf0LabelPluralLower}`}</span>
                  <span className={styles.abcPct}>{`${asPct(dto?.scope.analysis_cards?.abc_mix.c_revenue_pct || 0)} receita`}</span>
                </div>
              </div>
            </div>
          </article>
          <article className={styles.statCard}>
            <p>GMROI global</p>
            <strong>{dto?.scope.analysis_cards?.gmroi_global == null ? "-" : `${Number(dto.scope.analysis_cards.gmroi_global).toFixed(2).replace(".", ",")}x`}</strong>
            <small>{`Capital travado ${asBrlCompact(dto?.scope.analysis_cards?.capital_travado_brl || 0)}`}</small>
          </article>
          <article className={styles.statCard}>
            <p>Risco global</p>
            <strong>{asPct(dto?.scope.analysis_cards?.risco_global_pct || 0)}</strong>
            <small>{`Margem contrib ${asPct(dto?.scope.analysis_cards?.margem_contrib_real_pct || 0)}`}</small>
          </article>
        </div>

        <article
          className={`${styles.panel} ${styles.bottomTablePanel} ${styles.panelClickable}`}
          role="button"
          tabIndex={0}
          onClick={() => setIsTopMarginSpotlightOpen(true)}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              setIsTopMarginSpotlightOpen(true);
            }
          }}
        >
          <div className={styles.panelHeadInline}>
            <h2 className={styles.panelTitleMini}>{`Top ${leaf0LabelPluralLower} por margem`}</h2>
            <span className={styles.panelCta}>Ver Detalhes</span>
          </div>
          <table className={styles.table}>
            <thead><tr><th>#</th><th>{leaf0Label}</th><th>Margem %</th><th>Receita</th></tr></thead>
            <tbody>
              {topMarginRows.map((row, idx) => (
                <tr key={row.node_id}>
                  <td>{String(idx + 1).padStart(2, "0")}</td>
                  <td>{row.node_name}</td>
                  <td className={marginToneClass(row.margin_pct)}>
                    <div className={styles.tableMetricInline}>
                      {tableTrendPctCompact(row.trend?.margin_delta_mom_pp ?? null) ? (
                        <span className={`${trendClassBySign(row.trend?.margin_delta_mom_pp ?? null)} ${styles.tableTrendCompact}`}>{tableTrendPctCompact(row.trend?.margin_delta_mom_pp ?? null)}</span>
                      ) : null}
                      <span>{row.margin_pct == null ? "-" : asPct(row.margin_pct)}</span>
                    </div>
                  </td>
                  <td>
                    <div className={styles.tableMetricInline}>
                      {tableTrendPctCompact(row.trend?.revenue_delta_mom_pct ?? null) ? (
                        <span className={`${trendClassBySign(row.trend?.revenue_delta_mom_pct ?? null)} ${styles.tableTrendCompact}`}>{tableTrendPctCompact(row.trend?.revenue_delta_mom_pct ?? null)}</span>
                      ) : null}
                      <span>{asBrlCompact(row.revenue_brl)}</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </article>

        <article
          className={`${styles.panel} ${styles.bottomTablePanel} ${styles.panelClickable}`}
          role="button"
          tabIndex={0}
          onClick={() => setIsRiskSpotlightOpen(true)}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              setIsRiskSpotlightOpen(true);
            }
          }}
        >
          <div className={styles.panelHeadInline}>
            <h2 className={styles.panelTitleMini}>{`${leaf0LabelPlural} em risco`}</h2>
            <span className={styles.panelCta}>Ver Detalhes</span>
          </div>
          <table className={styles.table}>
            <thead><tr><th>{leaf0Label}</th><th>Risco</th><th>Receita</th><th>Capital em risco</th></tr></thead>
            <tbody>
              {nodesAtRiskRows.map((row) => (
                <tr key={row.node_id}>
                  <td>{row.node_name}</td>
                  <td><span className={row.risk_pct >= 30 ? styles.badgeRed : row.risk_pct >= 15 ? styles.badgeOrange : styles.badgeGray}>{asPct(row.risk_pct)}</span></td>
                  <td className={row.risk_pct >= 30 ? styles.bad : row.risk_pct >= 15 ? styles.warn : styles.good}>{asBrlCompact(row.revenue_brl || 0)}</td>
                  <td className={row.risk_pct >= 30 ? styles.bad : row.risk_pct >= 15 ? styles.warn : styles.good}>{asBrlCompact(row.capital_at_risk_brl)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </article>
      </section>

      <section className={styles.backlog}>
        <h2 className={styles.backlogTitle}>Backlog de execucao estrategica</h2>
        <div className={styles.backlogGrid}>
          {backlog.map((item) => (
            <article key={`${item.priority}:${item.title}`} className={`${styles.actionCard} ${styles[`tone_${item.tone}`]}`}>
              <div className={styles.actionMeta}>{item.priority}</div>
              <h3>{item.title}</h3>
              <p>{item.text}</p>
              <button type="button">{item.cta}</button>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
}

