import { useMemo, useState } from "react";
import type { WorkspaceHistoryPricePointV1, WorkspaceHistorySupplierLinkV1 } from "@metalshopping/feature-analytics";

import styles from "../history.module.css";

type PriceRange = "7d" | "30d" | "90d" | "6m" | "12m";

type PriceMarketChartProps = {
  points: WorkspaceHistoryPricePointV1[];
  supplierLinks?: Record<string, WorkspaceHistorySupplierLinkV1>;
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
];

function sliceByRange(points: WorkspaceHistoryPricePointV1[], range: PriceRange): WorkspaceHistoryPricePointV1[] {
  const sizeByRange: Record<PriceRange, number> = {
    "7d": 7,
    "30d": 30,
    "90d": 90,
    "6m": 180,
    "12m": 365,
  };
  return points.slice(Math.max(0, points.length - sizeByRange[range]));
}

function asCurrency(value: number | null): string {
  if (value == null || !Number.isFinite(value)) return "—";
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    maximumFractionDigits: 2,
  }).format(value);
}

function avg(values: Array<number | null>): number {
  const valid = values.filter((item): item is number => item != null && Number.isFinite(item));
  if (!valid.length) return 0;
  return valid.reduce((acc, cur) => acc + cur, 0) / valid.length;
}

function buildPath(values: Array<number | null>, min: number, max: number): string {
  const span = Math.max(1, max - min);
  const width = 100;
  const height = 40;
  const step = values.length > 1 ? width / (values.length - 1) : width;
  return values
    .map((value, index) => {
      const x = index * step;
      const normalized = value == null ? 0 : (value - min) / span;
      const y = height - normalized * height;
      return `${index === 0 ? "M" : "L"} ${x.toFixed(2)} ${y.toFixed(2)}`;
    })
    .join(" ");
}

export function PriceMarketChart({ points, supplierLinks }: PriceMarketChartProps) {
  const [range, setRange] = useState<PriceRange>("30d");
  const filtered = useMemo(() => sliceByRange(points, range), [points, range]);

  const supplierKeys = useMemo(() => {
    const set = new Set<string>();
    for (const point of filtered) {
      const suppliers = point.suppliers || {};
      Object.keys(suppliers).forEach((key) => set.add(key));
    }
    return Array.from(set).slice(0, 2);
  }, [filtered]);

  const ourSeries = filtered.map((point) => (typeof point.our_price === "number" ? point.our_price : null));
  const marketSeries = filtered.map((point) => (typeof point.market_mean === "number" ? point.market_mean : null));
  const supplierSeries = supplierKeys.map((key) =>
    filtered.map((point) => {
      const raw = point.suppliers?.[key];
      return typeof raw === "number" ? raw : null;
    }),
  );

  const allValues = [...ourSeries, ...marketSeries, ...supplierSeries.flat()].filter(
    (value): value is number => value != null && Number.isFinite(value),
  );
  const min = allValues.length ? Math.min(...allValues) : 0;
  const max = allValues.length ? Math.max(...allValues) : 1;

  const avgOur = avg(ourSeries);
  const avgMarket = avg(marketSeries);
  const lastOur = ourSeries.length ? ourSeries[ourSeries.length - 1] : null;
  const firstOur = ourSeries.length ? ourSeries[0] : null;
  const delta = firstOur && lastOur ? ((lastOur - firstOur) / firstOur) * 100 : 0;

  return (
    <article className={styles.chartCard}>
      <header className={styles.chartHeader}>
        <div className={styles.chartTitleWrap}>
          <span className={styles.chartIcon} aria-hidden>MS</span>
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
          <span className={styles.statLabel}>Preco medio (nosso)</span>
          <strong className={styles.statValue}>{asCurrency(avgOur)}</strong>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Preco medio (mercado)</span>
          <strong className={styles.statValue}>{asCurrency(avgMarket)}</strong>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Variacao no periodo</span>
          <strong className={`${styles.statValue} ${delta >= 0 ? styles.statPositive : styles.statNegative}`}>
            {`${delta >= 0 ? "+" : ""}${delta.toFixed(1)}%`}
          </strong>
        </div>
      </div>

      <div className={styles.chartWrap}>
        <svg viewBox="0 0 100 40" preserveAspectRatio="none" width="100%" height="100%">
          <path d={buildPath(marketSeries, min, max)} fill="none" stroke={MARKET_MEAN_COLOR} strokeWidth="1.6" />
          <path d={buildPath(ourSeries, min, max)} fill="none" stroke={OUR_PRICE_COLOR} strokeWidth="2.2" />
          {supplierSeries.map((series, index) => (
            <path
              key={`supplier-${supplierKeys[index]}`}
              d={buildPath(series, min, max)}
              fill="none"
              stroke={FALLBACK_SUPPLIER_COLORS[index % FALLBACK_SUPPLIER_COLORS.length]}
              strokeWidth="1.2"
              strokeDasharray="4 4"
            />
          ))}
        </svg>
      </div>

      <div className={styles.chartFooterSlot}>
        <div className={styles.priceLegend}>
          <span className={styles.priceLegendItem}>
            <span className={styles.priceLegendSwatch} style={{ ["--legend-color" as string]: OUR_PRICE_COLOR }} />
            <span className={styles.priceLegendLabel}>Nosso preco</span>
          </span>
          <span className={styles.priceLegendItem}>
            <span className={styles.priceLegendSwatch} style={{ ["--legend-color" as string]: MARKET_MEAN_COLOR }} />
            <span className={styles.priceLegendLabel}>Mercado medio</span>
          </span>
          {supplierKeys.map((key, index) => {
            const link = supplierLinks?.[key];
            const color = FALLBACK_SUPPLIER_COLORS[index % FALLBACK_SUPPLIER_COLORS.length];
            return (
              <span key={key} className={styles.priceLegendItem}>
                <span
                  className={`${styles.priceLegendSwatch} ${styles.priceLegendSwatchDashed}`}
                  style={{ ["--legend-color" as string]: color }}
                />
                {link?.url ? (
                  <a className={`${styles.priceLegendLabel} ${styles.priceLegendLink}`} href={link.url} target="_blank" rel="noreferrer">
                    {link.label || key}
                  </a>
                ) : (
                  <span className={styles.priceLegendLabel}>{link?.label || key}</span>
                )}
              </span>
            );
          })}
        </div>
      </div>
    </article>
  );
}
