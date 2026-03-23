// @ts-nocheck
import { ApiClientError } from "../../app/apiClient";
import {
  makeAnalyticsProductsIndexV1Dto,
  makeAnalyticsProductsOverviewV1Dto,
  type AnalyticsProductsOverviewV1Dto,
  type ProductsAnalyticsIndexRowV1,
} from "../../legacy_products_dto";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import { useAppSession } from "../../app/providers/AppProviders";
import { ProductsAnalyticsView } from "./components/products/ProductsAnalyticsView";
import { ProductsDensityView } from "./components/products/ProductsDensityView";
import { ProductsFiltersBar } from "./components/products/ProductsFiltersBar";
import { ProductsInsightsBar, type Insight } from "./components/products/ProductsInsightsBar";
import { ProductsToolbar } from "./components/products/ProductsToolbar";
import type { AnalyticsSkuRow, SkuAction, SkuStatus } from "./contracts_products";
import styles from "./analytics_products.module.css";

type ProductFilters = {
  taxonomyLeafName: string[];
  brand: string[];
  status: SkuStatus[];
};

type ColumnSortKey =
  | "pn"
  | "product"
  | "brand"
  | "price"
  | "market"
  | "gap"
  | "margin"
  | "trend"
  | "class"
  | "stock";

type ColumnSort = {
  key: ColumnSortKey;
  direction: "asc" | "desc" | "";
};

const PAGE_SIZE = 200;
const VALID_STATUSES: ReadonlySet<SkuStatus> = new Set(["crit", "warn", "ok", "info"]);

function parseMultiParam(params: URLSearchParams, key: string): string[] {
  const fromAll = params
    .getAll(key)
    .flatMap((item) => String(item || "").split(","))
    .map((item) => item.trim())
    .filter(Boolean);
  if (fromAll.length > 0) return Array.from(new Set(fromAll));

  const raw = params.get(key);
  if (!raw) return [];
  return Array.from(
    new Set(
      String(raw)
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
    ),
  );
}

function toCsv(values: string[]): string | undefined {
  const normalized = Array.from(
    new Set(
      values
        .map((item) => String(item || "").trim())
        .filter(Boolean),
    ),
  );
  if (normalized.length === 0) return undefined;
  return normalized.join(",");
}

function parseInitialFilters(search: string): ProductFilters {
  const params = new URLSearchParams(search);
  const status = parseMultiParam(params, "status")
    .map((item) => item.toLowerCase())
    .filter((item): item is SkuStatus => VALID_STATUSES.has(item as SkuStatus));
  return {
    brand: parseMultiParam(params, "marca"),
    taxonomyLeafName: parseMultiParam(params, "taxonomy_leaf_name"),
    status,
  };
}

function parseInitialSort(search: string): ColumnSort {
  const params = new URLSearchParams(search);
  const keyRaw = params.get("sort_key") || "";
  const dirRaw = params.get("sort_dir") || "";
  const key: ColumnSortKey | "" = (
    keyRaw === "pn" ||
    keyRaw === "product" ||
    keyRaw === "brand" ||
    keyRaw === "price" ||
    keyRaw === "market" ||
    keyRaw === "gap" ||
    keyRaw === "margin" ||
    keyRaw === "trend" ||
    keyRaw === "class" ||
    keyRaw === "stock"
  ) ? keyRaw : "";
  const direction: ColumnSort["direction"] = dirRaw === "asc" || dirRaw === "desc" ? dirRaw : "";
  if (!key || !direction) return { key: "price", direction: "" };
  return { key, direction };
}

function asNumber(value: unknown, fallback = 0): number {
  const num = Number(value);
  return Number.isFinite(num) ? num : fallback;
}

function sortSegmentIds(ids: string[]): string[] {
  const rank = (value: string): number => {
    const token = String(value || "").trim().toUpperCase();
    const match = /^([ABC])([12])$/.exec(token);
    if (!match) return 10_000;
    const letter = match[1];
    const digit = Number(match[2]);
    const base = letter === "A" ? 0 : letter === "B" ? 100 : 200;
    return base + digit;
  };
  return [...ids].sort((a, b) => {
    const ra = rank(a);
    const rb = rank(b);
    if (ra !== rb) return ra - rb;
    return String(a).localeCompare(String(b));
  });
}

function asStatus(value: unknown): SkuStatus {
  const token = String(value || "").toLowerCase();
  if (token === "crit" || token === "warn" || token === "ok" || token === "info") return token;
  return "info";
}

function asAction(value: unknown): SkuAction {
  const token = String(value || "").toUpperCase();
  if (token === "PRUNE" || token === "EXPAND" || token === "MONITOR") return token;
  return "MONITOR";
}

function buildProductsOverviewCacheKey(query: string, filters: ProductFilters): string {
  return JSON.stringify({
    query: query.trim(),
    brand: [...filters.brand].sort(),
    taxonomyLeafName: [...filters.taxonomyLeafName].sort(),
    status: [...filters.status].sort(),
  });
}

function toTrendSpark(trend: number): number[] {
  if (trend > 0.5) return [35, 55, 72, 92];
  if (trend < -0.5) return [92, 74, 58, 38];
  return [50, 52, 50, 51];
}

function asClassLetterFromSegment(segmentId: string): "A" | "B" | "C" {
  const token = String(segmentId || "").trim().toUpperCase();
  const letter = token.slice(0, 1);
  if (letter === "A" || letter === "B" || letter === "C") return letter;
  return "C";
}

function asClassLetterFromXyz(xyz: string): "A" | "B" | "C" {
  const token = String(xyz || "").trim().toUpperCase();
  if (token === "X") return "A";
  if (token === "Y") return "B";
  return "C";
}

function mapRow(row: ProductsAnalyticsIndexRowV1): AnalyticsSkuRow {
  const gap = asNumber(row.kpis?.gap_pct, 0);
  const margin = asNumber(row.kpis?.margem_contrib_pct, 0);
  const slope = asNumber(row.kpis?.slope, 0);
  const meanUnits = asNumber(row.kpis?.mean_units_6m, 0);
  const trendPctMom = row.kpis?.trend_pct_mom ?? null;
  const xyz = String(row.kpis?.xyz || "Z").toUpperCase();
  const segmentId = row.segment_id != null ? String(row.segment_id) : "";
  const trendPct = trendPctMom != null && Number.isFinite(Number(trendPctMom))
    ? Number(Number(trendPctMom).toFixed(1))
    : 0;
  const trendTone = trendPct > 0.5 ? "up" : trendPct < -0.5 ? "down" : "neutral";
  const className = segmentId ? asClassLetterFromSegment(segmentId) : asClassLetterFromXyz(xyz);
  const classLabel = segmentId ? segmentId : xyz;
  const trendLabel = trendPctMom == null
    ? (meanUnits > 0 ? "Baixo volume" : "N/D")
    : `${trendPct >= 0 ? "+" : ""}${trendPct.toFixed(1)}%/mes`;
  const trendMeta = `Slope ${slope.toFixed(2)} un/mes | Media ${meanUnits.toFixed(1)} un/mes`;
  return {
    pn: String(row.pn || ""),
    ean: String(row.ean || ""),
    description: String(row.description || row.pn || ""),
    taxonomyLeafName: String(row.taxonomy_leaf_name || "SEM_CLASSIFICACAO"),
    brand: String(row.brand || "SEM_MARCA"),
    status: asStatus(row.status),
    action: asAction(row.action),
    className,
    trend: trendTone === "up" ? "up" : trendTone === "down" ? "down" : "flat",
    trendPct,
    trendLabel,
    trendTone,
    trendColor: trendTone === "up" ? "trendGreen" : trendTone === "down" ? "trendRed" : "trendNeutral",
    trendSpark: toTrendSpark(trendPctMom != null ? trendPct : 0),
    trendMeta,
    classLabel,
    classTone: className === "A" ? "classA1" : className === "B" ? "classB1" : "classB2",
    stock: Math.max(0, Math.round(asNumber(row.kpis?.estoque_un, 0))),
    price: asNumber(row.kpis?.our_price, 0),
    marketPrice: asNumber(row.kpis?.comp_price_mean, 0),
    gapLabel: `${gap > 0 ? "+" : ""}${gap.toFixed(1)}%`,
    gapTone: gap > 1 ? "gapNegative" : gap < -1 ? "gapPositive" : "gapNeutral",
    marginTone: margin >= 22 ? "gapPositive" : margin <= 15 ? "gapNegative" : "gapNeutral",
    alertsCount: Array.isArray(row.alerts) ? row.alerts.length : 0,
    metrics: {
      gapPct: gap,
      marginPct: margin,
      pme6: asNumber(row.kpis?.pme, 0),
      giro6: asNumber(row.kpis?.giro_6m ?? row.kpis?.giro, 0),
      dos6: Math.round(asNumber(row.kpis?.dos, 0)),
      gmroi6: asNumber(row.kpis?.gmroi, 0),
      slope6: slope,
      cv6: asNumber(row.kpis?.cv, 0),
      xyz6: xyz === "X" || xyz === "Y" || xyz === "Z" ? xyz : "Z",
      dataQuality: 80,
      maturity: 65,
    },
    short: {
      dem3: Math.max(0, Math.round(asNumber(row.kpis?.venda_1m_un, 0))),
      slope3: slope,
      xyz3: xyz === "X" || xyz === "Y" || xyz === "Z" ? xyz : "Z",
    },
  };
}

export function AnalyticsProductsPage() {
  const { api, getProductsOverviewSnapshot, setProductsOverviewSnapshot } = useAppSession();
  const navigate = useNavigate();
  const location = useLocation();
  const [query, setQuery] = useState(() => new URLSearchParams(location.search).get("search") || "");
  const [filters, setFilters] = useState<ProductFilters>(() => parseInitialFilters(location.search));
  const [rows, setRows] = useState<AnalyticsSkuRow[]>([]);
  const [columnSort, setColumnSort] = useState<ColumnSort>(() => parseInitialSort(location.search));
  const [brands, setBrands] = useState<string[]>([]);
  const [taxonomyLeafs, setTaxonomyLeafs] = useState<string[]>([]);
  const [paging, setPaging] = useState({ offset: 0, limit: PAGE_SIZE, returned: 0, total: 0 });
  const [snapshotAsOf, setSnapshotAsOf] = useState<string | null>(null);
  const [windowPrimary, setWindowPrimary] = useState<number | null>(null);
  const [windowShort, setWindowShort] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const overviewCacheKey = useMemo(
    () => buildProductsOverviewCacheKey(query, filters),
    [filters, query],
  );
  const brandFilterParam = useMemo(() => toCsv(filters.brand), [filters.brand]);
  const taxonomyFilterParam = useMemo(() => toCsv(filters.taxonomyLeafName), [filters.taxonomyLeafName]);
  const statusFilterParam = useMemo(() => toCsv(filters.status), [filters.status]);
  const [overviewDto, setOverviewDto] = useState<AnalyticsProductsOverviewV1Dto | null>(
    () => getProductsOverviewSnapshot(overviewCacheKey)?.data || null,
  );
  const [overviewError, setOverviewError] = useState("");
  const viewMode: "density" | "analytics" = useMemo(() => {
    const params = new URLSearchParams(location.search);
    return params.get("view") === "density" ? "density" : "analytics";
  }, [location.search]);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const nextSearch = query.trim();
    if (nextSearch) params.set("search", nextSearch);
    else params.delete("search");

    params.delete("marca");
    for (const value of filters.brand) params.append("marca", value);
    params.delete("taxonomy_leaf_name");
    for (const value of filters.taxonomyLeafName) params.append("taxonomy_leaf_name", value);
    params.delete("status");
    for (const value of filters.status) params.append("status", value);

    if (columnSort.direction) {
      params.set("sort_key", columnSort.key);
      params.set("sort_dir", columnSort.direction);
    } else {
      params.delete("sort_key");
      params.delete("sort_dir");
    }

    const current = location.search.startsWith("?") ? location.search.slice(1) : location.search;
    const next = params.toString();
    if (next === current) return;

    navigate(
      {
        pathname: location.pathname,
        search: next ? `?${next}` : "",
      },
      { replace: true },
    );
  }, [columnSort, filters.brand, filters.taxonomyLeafName, filters.status, location.pathname, location.search, navigate, query]);

  const handleViewChange = useCallback(
    (next: "density" | "analytics") => {
      const params = new URLSearchParams(location.search);
      if (next === "density") {
        params.set("view", "density");
      } else {
        params.delete("view");
      }
      const nextSearch = params.toString();
      navigate(
        {
          pathname: location.pathname,
          search: nextSearch ? `?${nextSearch}` : "",
        },
        { replace: true },
      );
    },
    [location.pathname, location.search, navigate],
  );

  useEffect(() => {
    let disposed = false;
    async function loadManifest() {
      try {
        const env = await api.analytics.meta();
        if (disposed) return;
        const windows = env.data?.windows && typeof env.data.windows === "object" ? env.data.windows : null;
        const asOf = env.data?.as_of ? String(env.data.as_of) : null;
        setSnapshotAsOf((current) => current || asOf);
        setWindowPrimary(
          windows && Number.isFinite(Number(windows.primary_months)) ? Number(windows.primary_months) : null,
        );
        setWindowShort(
          windows && Number.isFinite(Number(windows.short_months)) ? Number(windows.short_months) : null,
        );
      } catch {
        // Optional metadata only.
      }
    }
    void loadManifest();
    return () => {
      disposed = true;
    };
  }, [api.analytics]);

  useEffect(() => {
    let disposed = false;
    async function load() {
      setIsLoading(true);
      setErrorMessage("");
      try {
        const env = await api.products.analyticsIndex({
          search: query.trim() || undefined,
          marca: brandFilterParam,
          taxonomyLeafName: taxonomyFilterParam,
          status: statusFilterParam,
          limit: PAGE_SIZE,
          offset: 0,
        });
        const dto = makeAnalyticsProductsIndexV1Dto(env.data, "current");
        if (disposed) return;
        setRows(dto.rows.map(mapRow));
        setBrands(dto.filters.marcas);
        setTaxonomyLeafs(dto.filters.taxonomy_leaf_names);
        setPaging(dto.paging);
        setSnapshotAsOf(dto.snapshot.as_of || null);
      } catch (err) {
        if (disposed) return;
        const apiErr = err instanceof ApiClientError ? err : null;
        setErrorMessage(apiErr?.message || (err instanceof Error ? err.message : String(err)));
        setRows([]);
      } finally {
        if (!disposed) setIsLoading(false);
      }
    }
    void load();
    return () => {
      disposed = true;
    };
  }, [api, brandFilterParam, taxonomyFilterParam, statusFilterParam, query]);

  useEffect(() => {
    const cached = getProductsOverviewSnapshot(overviewCacheKey);
    if (cached) {
      setOverviewDto(cached.data);
    }
  }, [getProductsOverviewSnapshot, overviewCacheKey]);

  useEffect(() => {
    let disposed = false;
    async function loadOverview() {
      setOverviewError("");
      try {
        const env = await api.products.analyticsOverview({
          search: query.trim() || undefined,
          marca: brandFilterParam,
          taxonomyLeafName: taxonomyFilterParam,
          status: statusFilterParam,
        });
        if (disposed) return;
        const dto = makeAnalyticsProductsOverviewV1Dto(env.data, "current");
        setOverviewDto(dto);
        setProductsOverviewSnapshot(overviewCacheKey, dto);
      } catch (err) {
        if (disposed) return;
        const apiErr = err instanceof ApiClientError ? err : null;
        setOverviewError(apiErr?.message || (err instanceof Error ? err.message : String(err)));
      }
    }
    void loadOverview();
    return () => {
      disposed = true;
    };
  }, [api, brandFilterParam, taxonomyFilterParam, statusFilterParam, overviewCacheKey, query, setProductsOverviewSnapshot]);

  const visibleRows = useMemo(() => {
    if (!columnSort.direction) return rows;
    const dir = columnSort.direction === "asc" ? 1 : -1;
    const rankClass = (value: string): number => {
      const token = String(value || "").trim().toUpperCase();
      const match = /^([ABC])([12])$/.exec(token);
      if (!match) return 10_000;
      const base = match[1] === "A" ? 0 : match[1] === "B" ? 100 : 200;
      return base + Number(match[2]);
    };
    const toNullable = (value: number | null | undefined) =>
      value == null || Number.isNaN(value) ? null : Number(value);
    const sorted = [...rows].sort((a, b) => {
      switch (columnSort.key) {
        case "pn": {
          const na = Number(a.pn);
          const nb = Number(b.pn);
          if (Number.isFinite(na) && Number.isFinite(nb)) return (na - nb) * dir;
          return String(a.pn).localeCompare(String(b.pn)) * dir;
        }
        case "product":
          return String(a.description).localeCompare(String(b.description)) * dir;
        case "brand":
          return String(a.brand).localeCompare(String(b.brand)) * dir;
        case "price":
          return (a.price - b.price) * dir;
        case "market":
          return (a.marketPrice - b.marketPrice) * dir;
        case "gap":
          return (a.metrics.gapPct - b.metrics.gapPct) * dir;
        case "margin":
          return (a.metrics.marginPct - b.metrics.marginPct) * dir;
        case "stock":
          return (a.stock - b.stock) * dir;
        case "trend": {
          const va = toNullable(a.trendLabel === "Baixo volume" || a.trendLabel === "N/D" ? null : a.trendPct);
          const vb = toNullable(b.trendLabel === "Baixo volume" || b.trendLabel === "N/D" ? null : b.trendPct);
          if (va == null && vb == null) return 0;
          if (va == null) return 1;
          if (vb == null) return -1;
          return (va - vb) * dir;
        }
        case "class":
          if (columnSort.direction === "asc") {
            return rankClass(b.classLabel) - rankClass(a.classLabel);
          }
          return rankClass(a.classLabel) - rankClass(b.classLabel);
        default:
          return 0;
      }
    });
    return sorted;
  }, [columnSort, rows]);

  async function loadMore(): Promise<void> {
    if (isLoadingMore || paging.offset + paging.returned >= paging.total) return;
    setIsLoadingMore(true);
    setErrorMessage("");
    const nextOffset = paging.offset + paging.returned;
    try {
      const env = await api.products.analyticsIndex({
        search: query.trim() || undefined,
        marca: brandFilterParam,
        taxonomyLeafName: taxonomyFilterParam,
        status: statusFilterParam,
        limit: PAGE_SIZE,
        offset: nextOffset,
      });
      const dto = makeAnalyticsProductsIndexV1Dto(env.data, "current");
      const nextRows = dto.rows.map(mapRow);
      setRows((prev) => {
        const known = new Set(prev.map((row) => row.pn));
        return [...prev, ...nextRows.filter((row) => !known.has(row.pn))];
      });
      setPaging(dto.paging);
    } catch (err) {
      const apiErr = err instanceof ApiClientError ? err : null;
      setErrorMessage(apiErr?.message || (err instanceof Error ? err.message : String(err)));
    } finally {
      setIsLoadingMore(false);
    }
  }

  function openWorkspace(pn: string): void {
    const token = encodeURIComponent(String(pn || "").trim());
    if (!token) return;
    navigate(`/analytics/products/${token}/overview`, {
      state: {
        from: `${location.pathname}${location.search}`,
      },
    });
  }

  const insights = useMemo<Insight[]>(() => {
    const totalProducts = asNumber(overviewDto?.kpis.portfolio_active.value, 0);
    const stockValue = asNumber(overviewDto?.kpis.capital_brl.value, 0);
    const marginAvg = asNumber(overviewDto?.kpis.weighted_margin_pct.value, 0);
    const alerts = overviewDto?.alerts
      ? asNumber(overviewDto.alerts.products_with_alerts, 0)
      : (
        asNumber(overviewDto?.matrix.attention.count, 0) +
        asNumber(overviewDto?.matrix.critical.count, 0)
      );
    const stockValueLabel = new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: "BRL",
      notation: "compact",
      maximumFractionDigits: 1,
    }).format(stockValue);
    return [
      { label: "Total Produtos", value: totalProducts.toLocaleString("pt-BR"), tone: "wine", icon: "\uD83D\uDCE6" },
      { label: "Estoque Total", value: stockValueLabel, tone: "blue", icon: "\uD83D\uDCB0", spark: [40, 62, 76, 92] },
      { label: "Margem Media", value: `${marginAvg.toFixed(1)}%`, tone: "green", icon: "\uD83D\uDCC8" },
      { label: "Alertas", value: alerts.toString(), tone: "warn", icon: "\u26A0\uFE0F" },
    ];
  }, [overviewDto]);

  const toolbarMetaPrimary = snapshotAsOf ? `Atualizado: ${snapshotAsOf.slice(0, 10)}` : null;
  const primaryMonths = windowPrimary ?? 6;
  const shortMonths = windowShort ?? 3;
  const toolbarMetaSecondary = `Janela: ${primaryMonths}M / ${shortMonths}M`;

  return (
    <section className={styles.page}>
      <div className={styles.productsContainer}>
        <ProductsToolbar
          query={query}
          onQueryChange={setQuery}
          viewMode={viewMode}
          onViewChange={handleViewChange}
          metaPrimary={toolbarMetaPrimary}
          metaSecondary={toolbarMetaSecondary}
          compact={viewMode === "analytics"}
        />
        {viewMode === "density" ? (
          <ProductsFiltersBar
            filters={filters}
            brands={brands}
            taxonomyLeafs={taxonomyLeafs}
            statuses={[
              { label: "Critico", value: "crit" },
              { label: "Atencao", value: "warn" },
              { label: "OK", value: "ok" },
              { label: "Info", value: "info" },
            ]}
            onChange={setFilters}
          />
        ) : null}
        {viewMode === "density" ? <ProductsInsightsBar insights={insights} /> : null}
        {errorMessage ? <div className={styles.productsNotice}>{errorMessage}</div> : null}
        {overviewError ? <div className={styles.productsNotice}>{overviewError}</div> : null}
        {isLoading && rows.length === 0 && !overviewDto ? <div className={styles.productsNotice}>Carregando analytics...</div> : null}
        {viewMode === "density" ? (
          <ProductsDensityView
            rows={visibleRows}
            onRowClick={openWorkspace}
            columnSort={columnSort}
            onColumnSortChange={setColumnSort}
          />
        ) : (
          overviewDto ? <ProductsAnalyticsView overview={overviewDto} /> : <div className={styles.productsNotice}>Carregando overview...</div>
        )}
        {viewMode === "density" && rows.length > 0 && paging.offset + paging.returned < paging.total ? (
          <div className={styles.loadMoreRow}>
            <button type="button" className={styles.loadMoreBtn} onClick={() => void loadMore()} disabled={isLoadingMore}>
              {isLoadingMore ? "Carregando..." : "Carregar mais"}
            </button>
          </div>
        ) : null}
      </div>

    </section>
  );
}
