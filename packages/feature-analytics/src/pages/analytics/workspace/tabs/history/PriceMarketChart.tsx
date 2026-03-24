import { useMemo, useRef, useState, type CSSProperties } from "react";
import type { WorkspaceHistoryPricePointV1, WorkspaceHistorySupplierLinkV1 } from "@metalshopping/feature-analytics";
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
  type ChartData,
  type ChartOptions,
  type TooltipModel,
} from "chart.js";
import { Line } from "react-chartjs-2";

import styles from "../history.module.css";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler);

type PriceRange = "7d" | "30d" | "90d" | "6m" | "12m";

type PriceMarketChartProps = {
  points: WorkspaceHistoryPricePointV1[];
  supplierLinks?: Record<string, WorkspaceHistorySupplierLinkV1>;
};

type TooltipRow = {
  label: string;
  value: string;
  color: string;
};

type ExternalTooltipState = {
  visible: boolean;
  x: number;
  y: number;
  title: string;
  rows: TooltipRow[];
  index: number | null;
};

type LegendItem = {
  key: string;
  label: string;
  color: string;
  dashed?: boolean;
  values: Array<number | null>;
  href?: string | null;
};

const OUR_PRICE_COLOR = "#8b1538";
const MARKET_MEAN_COLOR = "#6b7280";
const FALLBACK_SUPPLIER_COLORS = [
  "#2563eb",
  "#dc2626",
  "#ea580c",
  "#16a34a",
  "#9333ea",
  "#0891b2",
  "#db2777",
  "#0d9488",
  "#7c3aed",
  "#0284c7",
  "#f59e0b",
  "#65a30d",
];

const FIXED_SUPPLIER_COLORS: Record<string, string> = {
  CONDEC: "#2563eb",
  DEXCO: "#dc2626",
  LEROY: "#ea580c",
  "LEROY MERLIN": "#ea580c",
  SUVINIL: "#9333ea",
  TIGRE: "#16a34a",
  AKZO: "#0891b2",
  AMANCO: "#0d9488",
};

function asCurrency(value: number): string {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    maximumFractionDigits: 2,
  }).format(value);
}

function sliceByRange(points: WorkspaceHistoryPricePointV1[], range: PriceRange): WorkspaceHistoryPricePointV1[] {
  return points.slice(Math.max(0, points.length - rangeWindowSize(range)));
}

function rangeWindowSize(range: PriceRange): number {
  const sizeByRange: Record<PriceRange, number> = {
    "7d": 7,
    "30d": 30,
    "90d": 90,
    "6m": 180,
    "12m": 365,
  };
  return sizeByRange[range];
}

function avg(values: Array<number | null>): number {
  const valid = values.filter((item): item is number => item != null && Number.isFinite(item));
  if (!valid.length) return 0;
  return valid.reduce((acc, cur) => acc + cur, 0) / valid.length;
}

function formatDeltaPct(start: number | null, end: number | null): { text: string; positive: boolean; available: boolean } {
  if (start == null || end == null || start <= 0) return { text: "—", positive: true, available: false };
  const pct = ((end - start) / start) * 100;
  return { text: `${pct >= 0 ? "+" : ""}${pct.toFixed(1)}%`, positive: pct >= 0, available: true };
}

function normalizeColor(raw: unknown): string {
  if (Array.isArray(raw) && raw.length) return String(raw[0]);
  if (typeof raw === "string" && raw.trim()) return raw;
  return "#64748b";
}

function normalizeSupplierToken(value: string): string {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toUpperCase();
}

function hashToken(token: string): number {
  let hash = 0;
  for (let index = 0; index < token.length; index += 1) {
    hash = (hash * 31 + token.charCodeAt(index)) >>> 0;
  }
  return hash;
}

function resolveFixedSupplierColor(name: string): string | null {
  const token = normalizeSupplierToken(name);
  for (const [key, color] of Object.entries(FIXED_SUPPLIER_COLORS)) {
    if (token === key || token.includes(key)) return color;
  }
  return null;
}

function buildSupplierColorMap(names: string[]): Record<string, string> {
  const usedColors = new Set<string>();
  const map: Record<string, string> = {};

  for (const name of names) {
    const fixed = resolveFixedSupplierColor(name);
    if (fixed) {
      map[name] = fixed;
      usedColors.add(fixed);
      continue;
    }

    const token = normalizeSupplierToken(name) || name;
    let slot = hashToken(token) % FALLBACK_SUPPLIER_COLORS.length;
    let color = FALLBACK_SUPPLIER_COLORS[slot];
    let guard = 0;
    while (usedColors.has(color) && guard < FALLBACK_SUPPLIER_COLORS.length) {
      slot = (slot + 1) % FALLBACK_SUPPLIER_COLORS.length;
      color = FALLBACK_SUPPLIER_COLORS[slot];
      guard += 1;
    }
    map[name] = color;
    usedColors.add(color);
  }

  return map;
}

function carryForward(values: Array<number | null>): Array<number | null> {
  let last: number | null = null;
  return values.map((value) => {
    if (value != null && Number.isFinite(value)) {
      last = value;
      return value;
    }
    return last;
  });
}

export function PriceMarketChart({ points, supplierLinks = {} }: PriceMarketChartProps) {
  const [range, setRange] = useState<PriceRange>("90d");
  const [activeIndex, setActiveIndex] = useState<number | null>(null);
  const activeIndexRef = useRef<number | null>(null);
  const tooltipIndexRef = useRef<number | null>(null);
  const [tooltipState, setTooltipState] = useState<ExternalTooltipState>({
    visible: false,
    x: 0,
    y: 0,
    title: "",
    rows: [],
    index: null,
  });

  const filtered = useMemo(() => sliceByRange(points, range), [points, range]);

  const chartLabels = filtered.map((point) => String(point.date || "").slice(5));
  const ourData = filtered.map((point) => point.our_price ?? null);

  const competitorNames = Array.from(
    new Set(filtered.flatMap((point) => Object.keys(point.suppliers || {}))),
  );

  const supplierColorMap = useMemo(() => buildSupplierColorMap(competitorNames), [competitorNames]);
  const supplierLinksByToken = useMemo(() => {
    const next: Record<string, string> = {};
    Object.entries(supplierLinks || {}).forEach(([supplier, item]) => {
      const token = normalizeSupplierToken(supplier);
      const url = String(item?.url || "").trim();
      if (token && url) next[token] = url;
    });
    return next;
  }, [supplierLinks]);
  const competitorSeriesByName = useMemo<Record<string, Array<number | null>>>(() => {
    const next: Record<string, Array<number | null>> = {};
    competitorNames.forEach((name) => {
      const rawSeries = filtered.map((point) => point.suppliers?.[name] ?? null);
      next[name] = carryForward(rawSeries);
    });
    return next;
  }, [competitorNames, filtered]);
  const marketData = useMemo<Array<number | null>>(
    () =>
      filtered.map((point, index) => {
        const seriesValues = competitorNames
          .map((name) => competitorSeriesByName[name]?.[index] ?? null)
          .filter((value): value is number => value != null && Number.isFinite(value));
        if (seriesValues.length) {
          return seriesValues.reduce((acc, cur) => acc + cur, 0) / seriesValues.length;
        }
        return point.market_mean ?? null;
      }),
    [competitorNames, competitorSeriesByName, filtered],
  );

  const legendItems = useMemo<LegendItem[]>(
    () => [
      {
        key: "our-price",
        label: "Nosso preco",
        color: OUR_PRICE_COLOR,
        values: ourData,
      },
      {
        key: "market-mean",
        label: "Mercado medio",
        color: MARKET_MEAN_COLOR,
        dashed: true,
        values: marketData,
      },
      ...competitorNames.map((name) => ({
        key: `supplier-${name}`,
        label: name,
        color: supplierColorMap[name] || "#64748b",
        values: competitorSeriesByName[name] || [],
        href: supplierLinksByToken[normalizeSupplierToken(name)] || null,
      })),
    ],
    [competitorNames, competitorSeriesByName, marketData, ourData, supplierColorMap, supplierLinksByToken],
  );

  const competitorDatasets = competitorNames.map((name, index) => {
    const color = supplierColorMap[name] || FALLBACK_SUPPLIER_COLORS[index % FALLBACK_SUPPLIER_COLORS.length];
    return {
      label: name,
      data: competitorSeriesByName[name] || [],
      borderColor: color,
      backgroundColor: color,
      borderWidth: 1.1,
      pointRadius: (ctx: { dataIndex: number }) => (ctx.dataIndex === activeIndex ? 4.8 : 3.2),
      pointHoverRadius: (ctx: { dataIndex: number }) => (ctx.dataIndex === activeIndex ? 5.4 : 3.8),
      fill: false,
      tension: 0.24,
      spanGaps: true,
    };
  });

  const data: ChartData<"line"> = {
    labels: chartLabels,
    datasets: [
      {
        label: "Nosso preco",
        data: ourData,
        borderColor: OUR_PRICE_COLOR,
        backgroundColor: "rgba(139, 21, 56, 0.20)",
        borderWidth: 2.8,
        pointRadius: (ctx: { dataIndex: number }) => (ctx.dataIndex === activeIndex ? 5.8 : 0),
        pointHoverRadius: (ctx: { dataIndex: number }) => (ctx.dataIndex === activeIndex ? 6.2 : 0),
        pointBackgroundColor: OUR_PRICE_COLOR,
        pointBorderColor: "#ffffff",
        pointBorderWidth: 1.8,
        fill: true,
        tension: 0.28,
      },
      {
        label: "Mercado medio",
        data: marketData,
        borderColor: MARKET_MEAN_COLOR,
        borderDash: [6, 4],
        borderWidth: 2,
        pointRadius: (ctx: { dataIndex: number }) => (ctx.dataIndex === activeIndex ? 4.2 : 0),
        pointHoverRadius: (ctx: { dataIndex: number }) => (ctx.dataIndex === activeIndex ? 4.8 : 0),
        pointBackgroundColor: MARKET_MEAN_COLOR,
        pointBorderColor: "#ffffff",
        pointBorderWidth: 1.4,
        fill: false,
        tension: 0.2,
      },
      ...competitorDatasets,
    ],
  };

  const options = useMemo<ChartOptions<"line">>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    animation: {
      duration: 700,
      easing: "easeOutCubic",
    },
    transitions: {
      active: { animation: { duration: 500, easing: "easeOutCubic" } },
      show: { animation: { duration: 500, easing: "easeOutCubic" } },
      hide: { animation: { duration: 200, easing: "easeOutCubic" } },
    },
    interaction: {
      mode: "index",
      intersect: false,
    },
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        enabled: false,
        external: ({ chart, tooltip }: { chart: ChartJS<"line">; tooltip: TooltipModel<"line"> }) => {
          if (!tooltip || tooltip.opacity === 0 || !tooltip.dataPoints.length) {
            setTooltipState((prev) => (prev.visible ? { ...prev, visible: false, index: null } : prev));
            tooltipIndexRef.current = null;
            if (activeIndexRef.current != null) {
              activeIndexRef.current = null;
              setActiveIndex(null);
            }
            return;
          }

          const hoveredIndex = tooltip.dataPoints[0]?.dataIndex;
          const safeIndex = typeof hoveredIndex === "number" ? hoveredIndex : 0;
          if (activeIndexRef.current !== safeIndex) {
            activeIndexRef.current = safeIndex;
            setActiveIndex(safeIndex);
          }

          if (tooltipIndexRef.current === safeIndex) {
            return;
          }
          tooltipIndexRef.current = safeIndex;

          const rows: TooltipRow[] = legendItems.map((item) => {
            const yValue = item.values[safeIndex] ?? null;
            return {
              label: item.label,
              value: yValue == null ? "—" : asCurrency(yValue),
              color: item.color,
            };
          });

          const title = tooltip.title?.[0] || "";
          const caretX = tooltip.caretX || 0;
          const caretY = tooltip.caretY || 0;
          const chartWidth = chart.width || 0;
          const chartHeight = chart.height || 0;
          const tooltipWidth = 218;
          const tooltipHeight = Math.min(236, 34 + rows.length * 22);
          const horizontalOffset = 16;
          const verticalOffset = 14;
          let x = chart.canvas.offsetLeft + caretX + horizontalOffset;
          let y = chart.canvas.offsetTop + caretY - tooltipHeight - verticalOffset;

          if (caretX > chartWidth * 0.66) {
            x = chart.canvas.offsetLeft + caretX - tooltipWidth - horizontalOffset;
          }
          if (x < chart.canvas.offsetLeft + 8) {
            x = chart.canvas.offsetLeft + 8;
          }
          if (x + tooltipWidth > chart.canvas.offsetLeft + chartWidth - 8) {
            x = chart.canvas.offsetLeft + chartWidth - tooltipWidth - 8;
          }

          if (caretY < chartHeight * 0.4) {
            y = chart.canvas.offsetTop + caretY + verticalOffset;
          }
          y += 10;
          if (y < chart.canvas.offsetTop + 8) {
            y = chart.canvas.offsetTop + 8;
          }
          if (y + tooltipHeight > chart.canvas.offsetTop + chartHeight - 8) {
            y = chart.canvas.offsetTop + chartHeight - tooltipHeight - 8;
          }

          setTooltipState({
            visible: true,
            x,
            y,
            title,
            rows,
            index: safeIndex,
          });
        },
      },
    },
    onHover: (_event, elements) => {
      const nextIndex = elements.length ? elements[0].index : null;
      if (activeIndexRef.current !== nextIndex) {
        activeIndexRef.current = nextIndex;
        setActiveIndex(nextIndex);
      }
    },
    scales: {
      x: { grid: { display: false }, ticks: { color: "#64748b", maxTicksLimit: 8 } },
      y: { grid: { color: "rgba(148, 163, 184, 0.16)" }, ticks: { color: "#64748b" } },
    },
  }), [legendItems]);

  const avgOur = avg(ourData);
  const avgMarket = avg(marketData);
  const validMarketSeries = marketData.filter((value): value is number => value != null && Number.isFinite(value));
  const firstMarket = validMarketSeries.length ? validMarketSeries[0] : null;
  const lastMarket = validMarketSeries.length ? validMarketSeries[validMarketSeries.length - 1] : null;
  const delta = formatDeltaPct(firstMarket, lastMarket);

  return (
    <article className={styles.chartCard}>
      <header className={styles.chartHeader}>
        <div className={styles.chartTitleWrap}>
          <span className={styles.chartIcon} aria-hidden>📉</span>
          <h3 className={styles.chartTitle}>Preco vs Mercado</h3>
        </div>
        <div className={styles.timeRange}>
          {(["7d", "30d", "90d", "6m", "12m"] as PriceRange[]).map((token) => (
            <button
              key={token}
              type="button"
              className={`${styles.timeBtn} ${range === token ? styles.timeBtnActive : ""}`}
              onClick={() => setRange(token)}
            >
              {token.toUpperCase()}
            </button>
          ))}
        </div>
      </header>

      <div className={styles.statsRow}>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Preco medio (periodo)</span>
          <strong className={styles.statValue}>{asCurrency(avgOur)}</strong>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Variacao do mercado (intervalo exibido)</span>
          <strong className={`${styles.statValue} ${delta.positive ? styles.statPositive : styles.statNegative}`}>{delta.text}</strong>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Mercado medio</span>
          <strong className={styles.statValue}>{asCurrency(avgMarket)}</strong>
        </div>
      </div>

      <div className={styles.chartWrap}>
        <Line data={data} options={options} />
        <div
          className={`${styles.chartTooltip} ${tooltipState.visible ? styles.chartTooltipVisible : ""}`}
          style={{ left: `${tooltipState.x}px`, top: `${tooltipState.y}px` }}
        >
          <div className={styles.tooltipDate}>{tooltipState.title}</div>
          {tooltipState.rows.map((row) => (
            <div key={`${row.label}-${row.value}`} className={styles.tooltipRow}>
              <span className={styles.tooltipLabel}>
                <span className={styles.tooltipDot} style={{ backgroundColor: row.color }} />
                {row.label}
              </span>
              <span className={`${styles.tooltipValue} ${styles.mono}`}>{row.value}</span>
            </div>
          ))}
        </div>
      </div>
      <div className={styles.chartFooterSlot}>
        <div className={styles.priceLegend}>
          {legendItems.map((item) => (
            <div key={item.key} className={styles.priceLegendItem}>
              <span
                className={`${styles.priceLegendSwatch} ${item.dashed ? styles.priceLegendSwatchDashed : ""}`}
                style={{ "--legend-color": normalizeColor(item.color) } as CSSProperties}
              />
              {item.href ? (
                <button
                  type="button"
                  className={`${styles.priceLegendLabel} ${styles.priceLegendLink}`}
                  onClick={() => window.open(item.href || "", "_blank", "noopener,noreferrer")}
                  title={`Abrir ${item.label} no navegador`}
                >
                  {item.label}
                </button>
              ) : (
                <span className={styles.priceLegendLabel}>{item.label}</span>
              )}
            </div>
          ))}
        </div>
      </div>

    </article>
  );
}
