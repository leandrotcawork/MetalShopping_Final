import type { WorkspaceHistoryIndicatorV1 } from "@metalshopping/feature-analytics";

import styles from "../history.module.css";

type IndicatorCardProps = {
  item: WorkspaceHistoryIndicatorV1;
};

export function IndicatorCard({ item }: IndicatorCardProps) {
  return (
    <article className={styles.indicatorCard}>
      <div className={styles.indicatorLabel}>{item.label}</div>
      <div className={styles.indicatorValue}>{item.value}</div>
      <div className={styles.indicatorBar}>
        <div className={styles.indicatorFill} style={{ width: `${item.fill_pct}%` }} />
      </div>
    </article>
  );
}
