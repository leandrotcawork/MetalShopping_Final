import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import { AnimatedNumber } from "./AnimatedNumber";
import styles from "../analytics_home.module.css";

type Props = {
  rows: AnalyticsHomeViewModel["topMetal"];
};

export function TopMetalTiles({ rows }: Props) {
  return (
    <div className={styles.topMetalGrid}>
      {rows.map((tile) => (
        <div key={tile.key} className={styles.tm} data-tone={tile.tone}>
          <div className={styles.tmTop}>
            <span className={styles.tmK}>{tile.k}</span>
            <span className={styles.badge}>M-1</span>
          </div>
          <div className={styles.tmMiddle}>
            <span className={styles.tmName}>{tile.name}</span>
          </div>
          <div className={styles.tmBottom}>
            <AnimatedNumber className={styles.tmVal} value={tile.val} />
            <span className={styles.tmSubVal}>{tile.subVal}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

