import type { CSSProperties } from "react";

import styles from "../product_workspace.module.css";

export type WorkspaceMetricScale = {
  min_value: number;
  max_value: number;
  low_cut: number;
  high_cut: number;
  higher_is_worse: boolean;
  min_visible_pct: number;
};

type ColorScaleBarProps = {
  scale: WorkspaceMetricScale;
  indicatorPct: number;
  fillPct?: number;
  labels: readonly string[];
  /** show debug points/lines (DEV only) */
  debug?: boolean;
};

type ScaleCssVars = CSSProperties & {
  "--t1"?: string;
  "--t2"?: string;
  "--fill"?: string;
  "--c1-weak"?: string;
  "--c2-weak"?: string;
  "--c3-weak"?: string;
  "--c1-strong"?: string;
  "--c2-strong"?: string;
  "--c3-strong"?: string;
};

function clamp01(x: number) {
  return Math.max(0, Math.min(1, x));
}

function toPct(value: number, min: number, max: number) {
  const ratio = (value - min) / Math.max(1e-6, max - min);
  return clamp01(ratio) * 100;
}

export function ColorScaleBar({ scale, indicatorPct, fillPct, labels, debug = false }: ColorScaleBarProps) {
  const min = scale.min_value;
  const max = scale.max_value;

  // thresholds in %
  const t1 = toPct(scale.low_cut, min, max);
  const t2 = toPct(scale.high_cut, min, max);

  // marker + fill clamped
  const markerPct = clamp01(indicatorPct / 100) * 100;
  const fill = typeof fillPct === "number" ? clamp01(fillPct / 100) * 100 : undefined;

  // Colors: keep your existing tokens (these match your design system)
  const weakZone1 = "rgba(220, 38, 38, 0.22)";
  const weakZone2 = "rgba(234, 88, 12, 0.22)";
  const weakZone3 = "rgba(22, 163, 74, 0.22)";

  const strongZone1 = "rgba(220, 38, 38, 1)";
  const strongZone2 = "rgba(234, 88, 12, 1)";
  const strongZone3 = "rgba(22, 163, 74, 1)";

  const labelsWithPlus =
    labels.length >= 2 ? labels : [String(scale.min_value), String(scale.max_value)];

  const cssVars: ScaleCssVars = {
    "--t1": `${t1}%`,
    "--t2": `${t2}%`,
    "--fill": typeof fill === "number" ? `${fill}%` : undefined,

    "--c1-weak": weakZone1,
    "--c2-weak": weakZone2,
    "--c3-weak": weakZone3,
    "--c1-strong": strongZone1,
    "--c2-strong": strongZone2,
    "--c3-strong": strongZone3,
  };

  const debugEnabled = debug && import.meta.env.DEV;

  return (
    <>
      {/* OUTER wrapper keeps overflow visible so the triangle arrow is never clipped */}
      <div className={styles.colorTrackOuter} style={cssVars}>
        {/* INNER track handles radius + (optional) fill clipping */}
        <div className={styles.colorTrackInner}>
          {/* optional fill layer */}
          {typeof fill === "number" && <div className={styles.colorFill} />}
        </div>

        {/* marker: thin stem + triangle arrow (handled by CSS ::after) */}
        <div className={styles.colorIndicator} style={{ left: `${markerPct}%` }} />

        {debugEnabled && (
          <div className={styles.colorDebug}>
            <div className={styles.colorDebugLine} style={{ left: `${t1}%` }} />
            <div className={styles.colorDebugLine} style={{ left: `${t2}%` }} />
          </div>
        )}
      </div>

      {/* labels positioned at 0, t1, t2, 100 via CSS vars */}
      <div className={styles.colorLabels} style={{ "--t1": `${t1}%`, "--t2": `${t2}%` } as ScaleCssVars}>
        {labelsWithPlus.map((text, index) => {
          // expected: [min, low_cut, high_cut, max+]
          const cls =
            index === 0
              ? styles.colorLabelStart
              : index === labelsWithPlus.length - 1
                ? styles.colorLabelEnd
                : styles.colorLabelMid;

          const style =
            index === 1
              ? ({ left: "var(--t1)" } as CSSProperties)
              : index === 2
                ? ({ left: "var(--t2)" } as CSSProperties)
                : undefined;

          return (
            <span key={`${text}-${index}`} className={`${styles.colorLabel} ${cls}`} style={style}>
              {text}
            </span>
          );
        })}
      </div>
    </>
  );
}
