import type { ReactNode } from "react";

import styles from "./analytics_widgets.module.css";

export type AnalyticsWidgetSpotlight = {
  label: string;
  onClick?: () => void;
};

export type AnalyticsPerformanceMetric = {
  label: string;
  value: string;
  delta?: string | null;
  deltaTone?: "positive" | "negative" | "neutral";
  subValue?: string | null;
};

export type AnalyticsPerformanceItem = {
  id: string;
  name: string;
  icon?: ReactNode;
  metrics: AnalyticsPerformanceMetric[];
};

type AnalyticsPerformanceSectionProps = {
  title: string;
  hint: string;
  spotlight?: AnalyticsWidgetSpotlight;
  items: AnalyticsPerformanceItem[];
  solid?: boolean;
};

export function AnalyticsPerformanceSection({
  title,
  hint,
  spotlight,
  items,
  solid = false,
}: AnalyticsPerformanceSectionProps) {
  return (
    <article className={`${styles.panel} ${solid ? styles.panelSolid : ""}`}>
      <div className={styles.panelHeadRow}>
        <div className={styles.panelHead}>
          <h2 className={styles.panelTitle}>{title}</h2>
          <p className={styles.panelHint}>{hint}</p>
        </div>
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
      <div className={styles.categoryGrid}>
        {items.map((item) => (
          <article key={item.id} className={styles.categoryCard}>
            <div className={styles.categoryHeader}>
              {item.icon ? <div className={styles.categoryIcon}>{item.icon}</div> : null}
              <div className={styles.categoryName}>{item.name}</div>
            </div>
            <div className={styles.categoryMetrics}>
              {item.metrics.map((metric) => (
                <div key={`${item.id}:${metric.label}`} className={styles.metricItem}>
                  <div className={styles.metricLabel}>{metric.label}</div>
                  <div className={styles.metricValue}>{metric.value}</div>
                  {metric.delta ? (
                    <div
                      className={
                        metric.deltaTone === "negative"
                          ? styles.kpiWarn
                          : metric.deltaTone === "positive"
                            ? styles.kpiPositive
                            : styles.metricSub
                      }
                    >
                      {metric.delta}
                    </div>
                  ) : metric.subValue ? (
                    <div className={styles.metricSub}>{metric.subValue}</div>
                  ) : null}
                </div>
              ))}
            </div>
          </article>
        ))}
      </div>
    </article>
  );
}
