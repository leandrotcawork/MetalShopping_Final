import styles from "./analytics_widgets.module.css";
import type { AnalyticsWidgetSpotlight } from "./AnalyticsPerformanceSection";

export type AnalyticsEfficiencyItem = {
  id: string;
  name: string;
  supportingText: string;
  progressPct: number;
  valueLabel: string;
};

type AnalyticsEfficiencySectionProps = {
  title: string;
  hint: string;
  summaryLabel?: string;
  spotlight?: AnalyticsWidgetSpotlight;
  items: AnalyticsEfficiencyItem[];
};

export function AnalyticsEfficiencySection({
  title,
  hint,
  summaryLabel,
  spotlight,
  items,
}: AnalyticsEfficiencySectionProps) {
  return (
    <article className={styles.panel}>
      <div className={styles.panelHeadRow}>
        <div className={styles.panelHead}>
          <h2 className={styles.panelTitle}>{title}</h2>
          <p className={styles.panelSub}>{hint}</p>
        </div>
        <div className={styles.panelHeadActions}>
          {summaryLabel ? <span className={styles.panelActionText}>{summaryLabel}</span> : null}
          {spotlight ? (
            spotlight.onClick ? (
              <button type="button" className={styles.panelAction} onClick={spotlight.onClick}>
                {spotlight.label}
              </button>
            ) : (
              <span className={styles.panelActionText}>{spotlight.label}</span>
            )
          ) : null}
        </div>
      </div>
      <div className={styles.efficiencyList}>
        {items.map((item) => (
          <div key={item.id} className={styles.efficiencyRow}>
            <div className={styles.efficiencyMeta}>
              <strong>{item.name}</strong>
              <span>{item.supportingText}</span>
            </div>
            <div className={styles.efficiencyTrack}>
              <div className={styles.efficiencyFill} style={{ width: `${Math.min(100, Math.max(0, item.progressPct))}%` }} />
            </div>
            <strong className={styles.efficiencyValue}>{item.valueLabel}</strong>
          </div>
        ))}
      </div>
    </article>
  );
}
