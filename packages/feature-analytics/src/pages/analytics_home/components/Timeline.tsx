// @ts-nocheck
import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import styles from "../analytics_home.module.css";

type Props = {
  rows: AnalyticsHomeViewModel["timeline"];
};

export function Timeline({ rows }: Props) {
  return (
    <div className={styles.timeline}>
      {rows.map((evt) => (
        <div key={evt.key} className={styles.evt}>
          <span className={`${styles.pin} ${styles[`pin_${evt.pin}`]}`} />
          <span className={styles.evtBody}>
            <span className={styles.evtName}>{evt.name}</span>
            <span className={styles.evtDesc}>{evt.desc}</span>
          </span>
          <span className={styles.evtTime}>{evt.time}</span>
        </div>
      ))}
    </div>
  );
}

