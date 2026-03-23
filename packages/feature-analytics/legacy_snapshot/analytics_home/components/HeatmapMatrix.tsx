import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import styles from "../analytics_home.module.css";

type Props = {
  onOpenSpotlight: (key: string) => void;
  values: AnalyticsHomeViewModel["heatmap"];
};

function heatClass(value: number): string {
  if (value >= 0.7) return styles.h5;
  if (value >= 0.55) return styles.h4;
  if (value >= 0.4) return styles.h3;
  if (value >= 0.25) return styles.h2;
  return styles.h1;
}

export function HeatmapMatrix({ onOpenSpotlight, values }: Props) {
  return (
    <div className={styles.heatWrap}>
      <div className={styles.heatLegend}>
        <span className={styles.leg}><i className={`${styles.sw} ${styles.levelNormal}`} /> normal</span>
        <span className={styles.leg}><i className={`${styles.sw} ${styles.levelAttention}`} /> atencao</span>
        <span className={styles.leg}><i className={`${styles.sw} ${styles.levelAlert}`} /> alerta</span>
        <span className={styles.leg}><i className={`${styles.sw} ${styles.levelCritical}`} /> critico</span>
      </div>

      <div className={styles.heatChart}>
        <div className={styles.xAxis}>
          <span className={styles.axisTag}>U1</span>
          <span className={styles.axisTag}>U2</span>
          <span className={styles.axisTag}>U3</span>
          <span className={styles.axisTag}>U4</span>
          <span className={styles.axisTag}>U5</span>
        </div>
        <div className={styles.yRail}>
          <div className={styles.yTitle}>Impacto</div>
          <div className={styles.yAxis}>
            <span className={styles.axisTag}>I5</span>
            <span className={styles.axisTag}>I4</span>
            <span className={styles.axisTag}>I3</span>
            <span className={styles.axisTag}>I2</span>
            <span className={styles.axisTag}>I1</span>
          </div>
        </div>
        <div className={styles.heat}>
          {values.flatMap((row, y) =>
            row.map((value, x) => (
              <button
                key={`${y}-${x}`}
                type="button"
                className={`${styles.cell} ${heatClass(value)}`}
                onClick={() => onOpenSpotlight(`heat-${y}-${x}`)}
                title={`x:${x + 1} y:${5 - y}`}
              />
            ))
          )}
        </div>
        <div className={styles.xTitle}>Urgencia</div>
      </div>
    </div>
  );
}
