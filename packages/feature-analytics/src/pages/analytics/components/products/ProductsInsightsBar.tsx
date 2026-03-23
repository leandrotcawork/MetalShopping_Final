import styles from "../../analytics_products.module.css";

export type Insight = {
  label: string;
  value: string;
  tone: "wine" | "blue" | "green" | "warn";
  icon: string;
  spark?: number[];
};

type ProductsInsightsBarProps = {
  insights: Insight[];
};

export function ProductsInsightsBar({ insights }: ProductsInsightsBarProps) {
  return (
    <div className={styles.insightsBar}>
      {insights.map((insight) => (
        <div
          key={insight.label}
          className={styles.insightItem}
          data-tone={insight.tone}
          data-label={insight.label.toLowerCase()}
        >
          <div className={`${styles.insightIcon} ${styles[`icon_${insight.tone}`]}`}>{insight.icon}</div>
          <div className={styles.insightContent}>
            <h4>{insight.label}</h4>
            <p>{insight.value}</p>
          </div>
          {insight.spark ? (
            <div className={`${styles.insightSparkline} ${styles[`spark_${insight.tone}`]}`}>
              {insight.spark.map((height, index) => (
                <div key={`${insight.label}-spark-${index}`} className={styles.spark} style={{ height: `${height}%` }} />
              ))}
            </div>
          ) : null}
        </div>
      ))}
    </div>
  );
}
