import { useMemo } from "react";
import styles from "../product_workspace.module.css";

type GaugeBand = {
  to: number;
  color: string;
};

type SemiGaugeProps = {
  value: number;
  min: number;
  max: number;
  valueText: string;
  ticks: string[];
  label: string;
  bands?: GaugeBand[];
  gapDeg?: number;
  debug?: boolean;
  debugLabels?: boolean;
};

type LabelPoint = {
  text: string;
  x: number;
  y: number;
  p: { x: number; y: number };
};

const VIEW_W = 200;
const VIEW_H = 134;

const CX = 100;
const CY = 100;

const ARC_R = 80;
const ARC_STROKE = 20;

const BASELINE_Y = 114;
const EDGE_LABEL_R = ARC_R + 18;
const EDGE_OFFSET_DOWN = 0;

// marcador do valor (sua linha grossa)
const VALUE_MARK_INNER_R = 66;
const VALUE_MARK_OUTER_R = 90;

function clamp01(value: number): number {
  return Math.max(0, Math.min(1, value));
}

function valueToRatio(value: number, min: number, max: number): number {
  return clamp01((value - min) / Math.max(1e-6, max - min));
}

/** ratio 0 -> 180deg (esquerda), ratio 1 -> 0deg (direita) */
function ratioToDeg(ratio: number): number {
  return 180 - clamp01(ratio) * 180;
}

function pointOnArc(radius: number, deg: number) {
  const rad = (deg * Math.PI) / 180;
  return {
    x: CX + radius * Math.cos(rad),
    y: CY - radius * Math.sin(rad),
  };
}

function arcPath(startDeg: number, endDeg: number, radius = ARC_R): string {
  const start = pointOnArc(radius, startDeg);
  const end = pointOnArc(radius, endDeg);
  return `M ${start.x} ${start.y} A ${radius} ${radius} 0 0 1 ${end.x} ${end.y}`;
}

function normalizeBands(min: number, max: number, bands?: GaugeBand[]) {
  const fallback: GaugeBand[] = [
    { to: min + (max - min) / 3, color: "#DC2626" },
    { to: min + ((max - min) * 2) / 3, color: "#CA8A04" },
    { to: max, color: "#16A34A" },
  ];

  const raw = (bands && bands.length ? bands : fallback)
    .map((band) => ({ to: Math.max(min, Math.min(max, band.to)), color: band.color }))
    .sort((left, right) => left.to - right.to);

  const segments: Array<{ from: number; to: number; color: string }> = [];
  let from = min;
  raw.forEach((band) => {
    segments.push({ from, to: band.to, color: band.color });
    from = band.to;
  });
  return segments;
}

export function SemiGauge({
  value,
  min,
  max,
  valueText,
  ticks,
  label,
  bands,
  gapDeg = 1.4,
  debug = false,
  debugLabels = false,
}: SemiGaugeProps) {
  const debugEnabled = (debug || debugLabels) && import.meta.env.DEV;

  const segments = useMemo(() => normalizeBands(min, max, bands), [bands, max, min]);
  const boundaries = useMemo(() => segments.map((segment) => segment.to), [segments]);

  const baseLabels = useMemo<LabelPoint[]>(() => {
    if (ticks.length === 0) return [];
    const firstText = ticks[0];
    const lastText = ticks[ticks.length - 1];
    const leftPoint = pointOnArc(ARC_R, 180);
    const rightPoint = pointOnArc(ARC_R, 0);
    return [
      { text: firstText, x: leftPoint.x, y: BASELINE_Y, p: leftPoint },
      { text: lastText, x: rightPoint.x, y: BASELINE_Y, p: rightPoint },
    ];
  }, [ticks]);

  const edgeLabels = useMemo<LabelPoint[]>(() => {
    if (ticks.length <= 2) return [];
    return boundaries.slice(0, -1).map((boundary, index) => {
      const text = ticks[index + 1] ?? `${boundary}`;
      const ratio = valueToRatio(boundary, min, max);
      const deg = ratioToDeg(ratio);
      const p = pointOnArc(EDGE_LABEL_R, deg);
      return { text, x: p.x, y: p.y + EDGE_OFFSET_DOWN, p };
    });
  }, [boundaries, max, min, ticks]);

  // =========================
  // VALUE MARKER / NEEDLE
  // =========================
  const valueRatio = valueToRatio(value, min, max);
  const valueDeg = ratioToDeg(valueRatio);
  const valueMarkerOuter = pointOnArc(VALUE_MARK_OUTER_R, valueDeg);
  const valueMarkerInner = pointOnArc(VALUE_MARK_INNER_R, valueDeg);
  const needleAngle = -90 + valueRatio * 180;

  // No label measurements needed when labels are hidden

  return (
    <div className={styles.gaugeContainer}>
      <div className={styles.gauge}>
        <svg className={styles.gaugeSvg} viewBox={`0 0 ${VIEW_W} ${VIEW_H}`} role="img" aria-label={`${label}: ${valueText}`}>
          {/* trilho */}
          <path d={arcPath(180, 0)} fill="none" stroke="#E2E8F0" strokeWidth={ARC_STROKE} strokeLinecap="butt" />

          {/* base bands */}
          {segments.map((segment, index) => {
            const startDeg = ratioToDeg(valueToRatio(segment.from, min, max)) - (index > 0 ? gapDeg / 2 : 0);
            const endDeg = ratioToDeg(valueToRatio(segment.to, min, max)) + (index < segments.length - 1 ? gapDeg / 2 : 0);
            return (
              <path
                key={`seg-base-${segment.color}-${segment.to}-${index}`}
                d={arcPath(startDeg, endDeg)}
                fill="none"
                stroke={segment.color}
                strokeOpacity={0.6}
                strokeWidth={ARC_STROKE}
                strokeLinecap="butt"
              />
            );
          })}

          {/* fill atÃ© o value */}
          {segments.map((segment, index) => {
            if (value <= segment.from) return null;
            const endValue = Math.min(value, segment.to);
            const startDeg = ratioToDeg(valueToRatio(segment.from, min, max)) - (index > 0 ? gapDeg / 2 : 0);
            const endDeg = ratioToDeg(valueToRatio(endValue, min, max)) + (index < segments.length - 1 ? gapDeg / 2 : 0);
            return (
              <path
                key={`seg-fill-${segment.color}-${segment.to}-${index}`}
                d={arcPath(startDeg, endDeg)}
                fill="none"
                stroke={segment.color}
                strokeWidth={ARC_STROKE}
                strokeLinecap="butt"
              />
            );
          })}

          {/* labels base (extremos) */}
          {baseLabels.map((tick, index) => (
            <text
              key={`base-label-${index}`}
              x={tick.x}
              y={tick.y}
              className={styles.gaugeTick}
              textAnchor="middle"
              dominantBaseline="middle"
            >
              {tick.text}
            </text>
          ))}

          {/* labels edge (divisoes) */}
          {edgeLabels.map((tick, index) => (
            <text
              key={`edge-label-${index}`}
              x={tick.x}
              y={tick.y}
              className={styles.gaugeTick}
              textAnchor="middle"
              dominantBaseline="middle"
            >
              {tick.text}
            </text>
          ))}

          {debugEnabled && (
            <g className={styles.gaugeDebug}>
              {baseLabels.map((tick, index) => (
                <circle key={`dbg-base-${index}`} cx={tick.x} cy={tick.y} r="1.6" />
              ))}
              {edgeLabels.map((tick, index) => (
                <g key={`dbg-edge-${index}`}>
                  <line x1={CX} y1={CY} x2={tick.p.x} y2={tick.p.y} stroke="red" strokeWidth="1" />
                  <circle cx={tick.p.x} cy={tick.p.y} r="1.6" />
                </g>
              ))}
            </g>
          )}

          {/* marcador do valor */}
          <line x1={valueMarkerOuter.x} y1={valueMarkerOuter.y} x2={valueMarkerInner.x} y2={valueMarkerInner.y} stroke="#F8FAFC" strokeWidth="8" strokeLinecap="round" />
          <line x1={valueMarkerOuter.x} y1={valueMarkerOuter.y} x2={valueMarkerInner.x} y2={valueMarkerInner.y} stroke="#0F172A" strokeWidth="4" strokeLinecap="round" />

          {/* agulha */}
          <g transform={`rotate(${needleAngle} ${CX} ${CY})`}>
            <line x1={CX} y1={CY} x2={CX} y2={48} stroke="#F8FAFC" strokeWidth="6" strokeLinecap="round" />
            <line x1={CX} y1={CY} x2={CX} y2={48} stroke="#0F172A" strokeWidth="3.2" strokeLinecap="round" />
          </g>
          <circle cx={CX} cy={CY} r="7" fill="#F8FAFC" />
          <circle cx={CX} cy={CY} r="4.6" fill="#0F172A" />
        </svg>
      </div>
    </div>
  );
}

