import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import { AnimatedNumber } from "./AnimatedNumber";
import styles from "../analytics_home.module.css";

type Props = {
  rows: AnalyticsHomeViewModel["kpis"];
};

export function KpisPanel({ rows }: Props) {
  return (
    <div className={styles.kpiGrid}>
      {rows.map((kpi) => (
        <div key={kpi.key} className={styles.kpi}>
          <span className={styles.kTop}>
            <span className={styles.kLbl}>{kpi.label}</span>
            <span className={styles.badge}>{kpi.badge}</span>
          </span>
          <AnimatedNumber className={styles.kVal} value={kpi.value} />
          <span className={styles.kNote}>{kpi.note}</span>
          <span className={styles.spark} style={{ ["--bars-count" as string]: kpi.bars.length }}>
            {kpi.bars.map((bar) => (
              <span
                key={bar.key}
                className={styles.sparkCell}
                title={`${bar.tipLabel}: ${bar.tipValue}`}
                aria-label={`${bar.tipLabel}: ${bar.tipValue}`}
              >
                <i
                  className={`${styles.sbar} ${kpi.tone === "blue" ? styles.sbarBlue : ""}`}
                  style={{ height: `${bar.heightPct}%` }}
                />
                <span className={styles.sparkTip}>{`${bar.tipLabel}: ${bar.tipValue}`}</span>
              </span>
            ))}
          </span>
        </div>
      ))}
    </div>
  );
}

