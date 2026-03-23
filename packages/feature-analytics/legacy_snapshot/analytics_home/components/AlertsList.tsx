import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import styles from "../analytics_home.module.css";

type Props = {
  onOpenSpotlight: (key: string) => void;
  rows: AnalyticsHomeViewModel["alerts"];
};

export function AlertsList({ onOpenSpotlight, rows }: Props) {
  return (
    <div className={styles.alerts}>
      {rows.map((row) => (
        <button
          key={row.key}
          type="button"
          className={`${styles.alert} ${styles[`alertCode_${String(row.code || "").toUpperCase()}`] || styles[row.toneClass]}`}
          onClick={() => onOpenSpotlight(row.key)}
        >
          <span className={styles.alertTop}>
            <span className={styles.alertName}>{row.name}</span>
            <span className={styles.alertStockChip}>Estoque: {row.stockTotalLabel || "-"}</span>
            <span className={styles.count}>{row.count}</span>
          </span>
          <span className={styles.alertDesc}>{row.desc}</span>
        </button>
      ))}
    </div>
  );
}
