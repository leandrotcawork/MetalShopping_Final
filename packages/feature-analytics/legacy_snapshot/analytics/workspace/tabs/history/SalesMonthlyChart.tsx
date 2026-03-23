import { useMemo, useState } from "react";
import type { WorkspaceHistorySalesPointV1 } from "@metalshopping/feature-analytics";
import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  Tooltip,
  type ChartData,
  type ChartOptions,
} from "chart.js";
import { Bar } from "react-chartjs-2";

import styles from "../history.module.css";

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip, Legend);

type SalesRange = "1m" | "3m" | "6m" | "12m";

type SalesMonthlyChartProps = {
  points: WorkspaceHistorySalesPointV1[];
};

function sliceByRange(points: WorkspaceHistorySalesPointV1[], range: SalesRange): WorkspaceHistorySalesPointV1[] {
  const sizeByRange: Record<SalesRange, number> = {
    "1m": 1,
    "3m": 3,
    "6m": 6,
    "12m": 12,
  };
  return points.slice(Math.max(0, points.length - sizeByRange[range]));
}

export function SalesMonthlyChart({ points }: SalesMonthlyChartProps) {
  const [range, setRange] = useState<SalesRange>("6m");
  const filtered = useMemo(() => sliceByRange(points, range), [points, range]);

  const labels = filtered.map((point) => point.month);
  const values = filtered.map((point) => point.units);

  const data: ChartData<"bar"> = {
    labels,
    datasets: [
      {
        label: "Vendas",
        data: values,
        backgroundColor: "rgba(139, 21, 56, 0.9)",
        borderColor: "#8b1538",
        borderWidth: 1.4,
        borderRadius: 6,
      },
    ],
  };

  const options: ChartOptions<"bar"> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { grid: { color: "rgba(148, 163, 184, 0.16)" }, ticks: { color: "#64748b" } },
      y: { grid: { color: "rgba(148, 163, 184, 0.16)" }, ticks: { color: "#64748b" } },
    },
  };

  const total = values.reduce((acc, cur) => acc + cur, 0);
  const avg = values.length ? total / values.length : 0;
  const peakIndex = values.reduce((best, cur, idx, list) => (cur > list[best] ? idx : best), 0);
  const peakUnits = values[peakIndex] ?? 0;
  const peakMonth = labels[peakIndex] ?? "-";

  return (
    <article className={styles.chartCard}>
      <header className={styles.chartHeader}>
        <div className={styles.chartTitleWrap}>
          <span className={styles.chartIcon} aria-hidden>📊</span>
          <h3 className={styles.chartTitle}>Vendas Mensais</h3>
        </div>
        <div className={styles.timeRange}>
          {(["1m", "3m", "6m", "12m"] as SalesRange[]).map((token) => (
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
          <span className={styles.statLabel}>Total vendas</span>
          <strong className={styles.statValue}>{Math.round(total)} un</strong>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Media mensal</span>
          <strong className={styles.statValue}>{avg.toFixed(1)} un</strong>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statLabel}>Pico</span>
          <strong className={styles.statValue}>{`${Math.round(peakUnits)} un (${peakMonth})`}</strong>
        </div>
      </div>

      <div className={styles.chartWrap}>
        <Bar data={data} options={options} />
      </div>
      <div className={styles.chartFooterSlot} aria-hidden />
    </article>
  );
}
