import styles from "./analytics_widgets.module.css";
import type { AnalyticsWidgetSpotlight } from "./AnalyticsPerformanceSection";

export type AnalyticsConcentrationStat = {
  label: string;
  value: string;
  valueTone?: "positive" | "negative" | "neutral";
};

export type AnalyticsConcentrationBar = {
  label: string;
  valueText: string;
  percent: number;
  tone: "top3" | "top5" | "top10";
};

type AnalyticsConcentrationSectionProps = {
  title: string;
  hint: string;
  spotlight?: AnalyticsWidgetSpotlight;
  riskBadge?: string;
  riskTone?: "positive" | "negative" | "neutral";
  stats: AnalyticsConcentrationStat[];
  bars: AnalyticsConcentrationBar[];
  quote: string;
};

export function AnalyticsConcentrationSection({
  title,
  hint,
  spotlight,
  riskBadge,
  riskTone = "neutral",
  stats,
  bars,
  quote,
}: AnalyticsConcentrationSectionProps) {
  const riskClass =
    riskTone === "negative" ? styles.riskBadgeNegative : riskTone === "positive" ? styles.riskBadgePositive : styles.riskBadgeNeutral;

  return (
    <article className={styles.panel}>
      <div className={styles.panelHeadRow}>
        <div className={styles.panelHead}>
          <h2 className={styles.panelTitle}>{title}</h2>
          <p className={`${styles.panelHint} ${styles.textWarn}`}>{hint}</p>
        </div>
        <div className={styles.panelHeadActions}>
          {riskBadge ? <span className={`${styles.riskBadge} ${riskClass}`}>{riskBadge}</span> : null}
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
      <div className={styles.concStats}>
        {stats.map((stat) => (
          <article key={stat.label} className={styles.concStatCard}>
            <span>{stat.label}</span>
            <strong
              className={
                stat.valueTone === "negative"
                  ? styles.kpiWarn
                  : stat.valueTone === "positive"
                    ? styles.kpiPositive
                    : undefined
              }
            >
              {stat.value}
            </strong>
          </article>
        ))}
      </div>
      <div className={styles.concentrationList}>
        {bars.map((bar) => (
          <div key={bar.label} className={styles.concItem}>
            <div className={styles.concMeta}>
              <span>{bar.label}</span>
              <span>{bar.valueText}</span>
            </div>
            <div className={styles.concTrack}>
              <div
                className={`${styles.concFill} ${bar.tone === "top3" ? styles.concTop3 : bar.tone === "top5" ? styles.concTop5 : styles.concTop10}`}
                style={{ width: `${Math.min(100, Math.max(0, bar.percent))}%` }}
              />
            </div>
          </div>
        ))}
      </div>
      <p className={styles.quote}>{quote}</p>
    </article>
  );
}
