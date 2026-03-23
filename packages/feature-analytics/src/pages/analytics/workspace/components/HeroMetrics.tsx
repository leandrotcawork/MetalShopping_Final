import type { WorkspaceHeroMetricV1 } from "@metalshopping/feature-analytics";
import styles from "../product_workspace.module.css";

type HeroMetricsProps = {
  items: WorkspaceHeroMetricV1[];
};

export function HeroMetrics({ items }: HeroMetricsProps) {
  return (
    <div className={styles.heroMetrics}>
      {items.map((item) => (
        <article key={item.label} className={styles.heroMetric}>
          <div className={styles.heroMetricLabel}>{item.label}</div>
          <div className={styles.heroMetricValue}>{item.value}</div>
        </article>
      ))}
    </div>
  );
}
