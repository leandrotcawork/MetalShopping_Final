import styles from "./products_overview.module.css";

type SparklineMiniProps = {
  points: number[];
  tone: "positive" | "negative";
};

function buildPoints(points: number[], width: number, height: number): string {
  if (!points.length) return "";
  const max = Math.max(...points);
  const min = Math.min(...points);
  const span = Math.max(1, max - min);
  return points
    .map((value, index) => {
      const x = (index / Math.max(1, points.length - 1)) * width;
      const y = height - ((value - min) / span) * height;
      return `${x},${y}`;
    })
    .join(" ");
}

export function SparklineMini({ points, tone }: SparklineMiniProps) {
  const safe = points.length > 1 ? points : [48, 52, 50, 54, 53, 56];
  const polyline = buildPoints(safe, 100, 36);

  return (
    <div className={styles.sparkWrap} aria-hidden>
      <svg viewBox="0 0 100 36" className={`${styles.sparkSvg} ${styles[`sparkSvg_${tone}`]}`}>
        <polyline points={polyline} fill="none" />
      </svg>
    </div>
  );
}

