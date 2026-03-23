import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import { AnimatedNumber } from "./AnimatedNumber";
import styles from "../analytics_home.module.css";

type Props = {
  onOpenSpotlight: (key: string) => void;
  rows: AnalyticsHomeViewModel["portfolio"];
};

export function PortfolioDistribution({ onOpenSpotlight, rows }: Props) {
  return (
    <div className={styles.dist}>
      {rows.map((row) => (
        <button key={row.key} type="button" className={styles.barRow} onClick={() => onOpenSpotlight(row.key)}>
          <span className={`${styles.ico} ${styles[`ico_${row.iconStyle}`]}`}>{row.icon}</span>
          <span className={styles.barBody}>
            <span className={styles.barHdr}>
              <span className={styles.barLbl}>{row.label}</span>
              <AnimatedNumber className={`${styles.barVal} ${styles[`barVal_${row.iconStyle}`]}`} value={row.value} />
            </span>
            <span className={styles.track}>
              <span className={`${styles.fill} ${styles[`fill_${row.fillStyle}`]}`} style={{ width: `${row.pct}%` }} />
            </span>
          </span>
        </button>
      ))}
    </div>
  );
}
