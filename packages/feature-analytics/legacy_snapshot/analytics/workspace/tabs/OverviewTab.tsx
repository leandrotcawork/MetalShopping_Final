import { useOutletContext } from "react-router-dom";

import { ColorScaleBar } from "../components/ColorScaleBar";
import { InfoTooltipLabel } from "../components/InfoTooltipLabel";
import { SemiGauge } from "../components/SemiGauge";
import { WorkspaceCard } from "../components/WorkspaceCard";
import type { ProductWorkspaceOutletContext } from "../ProductWorkspaceLayout";
import { ANALYTICS_HELP } from "../help_texts";
import styles from "../product_workspace.module.css";

function toneClass(tone: "default" | "positive" | "warn" | "negative" | "neutral"): string {
  if (tone === "positive") return styles.tonePositive;
  if (tone === "warn") return styles.toneWarn;
  if (tone === "negative") return styles.toneNegative;
  return styles.toneDefault;
}

function progressToneClass(tone: "positive" | "warn" | "negative" | "neutral"): string {
  if (tone === "positive") return styles.progressPositive;
  if (tone === "warn") return styles.progressWarn;
  if (tone === "negative") return styles.progressNegative;
  return styles.progressNeutral;
}

export function OverviewTab() {
  const { model } = useOutletContext<ProductWorkspaceOutletContext>();
  return (
    <div className={styles.mainGrid}>
      <WorkspaceCard icon="📦" title="Estoque & Vendas" tone="stock">
        <div className={`${styles.metricsRow} ${styles.stockMetricsRow}`}>
          <article className={styles.metricItem}>
            <div className={styles.metricLabel}>
              <InfoTooltipLabel label={model.stockSales.pme.label} help={ANALYTICS_HELP[model.stockSales.pme.label]} />
            </div>
            <div className={styles.metricValueSmall}>{model.stockSales.pme.value}</div>
          </article>
          <article className={styles.metricItem}>
            <div className={styles.metricLabel}>
              <InfoTooltipLabel label={model.stockSales.giro.label} help={ANALYTICS_HELP[model.stockSales.giro.label]} />
            </div>
            <div className={styles.metricValueSmall}>{model.stockSales.giro.value}</div>
          </article>
        </div>

        <div className={`${styles.gaugesRow} ${styles.stockGaugesRow}`}>
          <SemiGauge
            label={model.stockSales.gauges[0].label}
            value={model.stockSales.gauges[0].value}
            min={model.stockSales.gauges[0].min}
            max={model.stockSales.gauges[0].max}
            valueText={model.stockSales.gauges[0].valueText}
            ticks={model.stockSales.gauges[0].ticks}
            bands={model.stockSales.gauges[0].bands}
            gapDeg={1.2}
          />
          <SemiGauge
            label={model.stockSales.gauges[1].label}
            value={model.stockSales.gauges[1].value}
            min={model.stockSales.gauges[1].min}
            max={model.stockSales.gauges[1].max}
            valueText={model.stockSales.gauges[1].valueText}
            ticks={model.stockSales.gauges[1].ticks}
            bands={model.stockSales.gauges[1].bands}
            gapDeg={1.2}
          />
        </div>

        <div className={`${styles.barsRow} ${styles.stockBarsRow}`}>
          {model.stockSales.bars.map((bar) => (
            <article key={bar.label} className={styles.progressWrap}>
              <div className={styles.progressHeader}>
                <span className={styles.progressLabel}>
                  <InfoTooltipLabel label={bar.label} help={ANALYTICS_HELP[bar.label]} />
                </span>
                <span className={styles.progressValueStrong}>{bar.valueText}</span>
              </div>
              <div className={styles.progressTrack}>
                <div className={`${styles.progressFill} ${progressToneClass(bar.tone)}`} style={{ width: `${bar.percent}%` }} />
              </div>
            </article>
          ))}
        </div>
      </WorkspaceCard>

      <WorkspaceCard icon='💰' title='Rentabilidade' tone='profit'>
        <div className={styles.profitGrid}>
          <article className={styles.profitTile}>
            <div className={styles.profitTileHeader}>
              <span className={styles.metricLabel}>
                <InfoTooltipLabel label={model.profitability.metrics[0].label} help={ANALYTICS_HELP[model.profitability.metrics[0].label]} />
              </span>
              <span className={`${styles.profitTileValue} ${toneClass(model.profitability.metrics[0].tone || "default")}`}>{model.profitability.metrics[0].value}</span>
            </div>
            <ColorScaleBar
              scale={model.profitability.colorBars[0].scale}
              indicatorPct={model.profitability.colorBars[0].indicatorPct}
              fillPct={model.profitability.colorBars[0].fillPct}
              labels={model.profitability.colorBars[0].labels}
            />
          </article>

          <article className={styles.profitTile}>
            <div className={styles.profitTileHeader}>
              <span className={styles.metricLabel}>
                <InfoTooltipLabel label={model.profitability.metrics[1].label} help={ANALYTICS_HELP[model.profitability.metrics[1].label]} />
              </span>
              <span className={`${styles.profitTileValue} ${toneClass(model.profitability.metrics[1].tone || "default")}`}>{model.profitability.metrics[1].value}</span>
            </div>
            <ColorScaleBar
              scale={model.profitability.colorBars[1].scale}
              indicatorPct={model.profitability.colorBars[1].indicatorPct}
              fillPct={model.profitability.colorBars[1].fillPct}
              labels={model.profitability.colorBars[1].labels}
            />
          </article>

          <article className={styles.profitTile}>
            <div className={styles.profitTileHeader}>
              <span className={styles.metricLabel}>
                <InfoTooltipLabel label={model.profitability.lower[0].label} help={ANALYTICS_HELP[model.profitability.lower[0].label]} />
              </span>
              <span className={`${styles.profitTileValue} ${toneClass(model.profitability.lower[0].tone || "default")}`}>{model.profitability.lower[0].value}</span>
            </div>
            <ColorScaleBar
              scale={model.profitability.lower[1].scale}
              indicatorPct={model.profitability.lower[1].indicatorPct}
              fillPct={model.profitability.lower[1].fillPct}
              labels={model.profitability.lower[1].labels}
            />
          </article>

          <article className={styles.profitTile}>
            <div className={styles.profitTileHeader}>
              <span className={styles.metricLabel}>
                <InfoTooltipLabel label={model.profitability.cogs.label} help={ANALYTICS_HELP[model.profitability.cogs.label]} />
              </span>
              <span className={`${styles.profitTileValue} ${toneClass(model.profitability.cogs.tone || "default")}`}>{model.profitability.cogs.value}</span>
            </div>
            {model.profitability.cogsBar ? (
              <ColorScaleBar
                scale={model.profitability.cogsBar.scale}
                indicatorPct={model.profitability.cogsBar.indicatorPct}
                fillPct={model.profitability.cogsBar.fillPct}
                labels={model.profitability.cogsBar.labels}
              />
            ) : null}
          </article>
        </div>
      </WorkspaceCard>

      <WorkspaceCard icon="⚔️" title="Competitividade" tone="competition">
        <div className={styles.metricsRow}>
          {model.competitiveness.metrics.map((metric) => (
            <article key={metric.label} className={styles.progressWrap}>
              <div className={styles.progressHeader}>
                <span className={styles.progressLabel}>{metric.label}</span>
                <span className={`${styles.metricValueCompact} ${toneClass(metric.tone)}`}>{metric.valueText}</span>
              </div>
              <div className={styles.progressTrack}>
                <div className={`${styles.progressFill} ${progressToneClass(metric.tone)}`} style={{ width: `${metric.percent}%` }} />
              </div>
            </article>
          ))}
        </div>
        <div className={styles.metricsRow}>
          {model.competitiveness.lower.map((metric) => (
            <article key={metric.label} className={styles.progressWrap}>
              <div className={styles.progressHeader}>
                <span className={styles.progressLabel}>{metric.label}</span>
                <span className={`${styles.metricValueCompact} ${toneClass(metric.tone)}`}>{metric.valueText}</span>
              </div>
              <div className={styles.progressTrack}>
                <div className={`${styles.progressFill} ${progressToneClass(metric.tone)}`} style={{ width: `${metric.percent}%` }} />
              </div>
            </article>
          ))}
        </div>
      </WorkspaceCard>

      <WorkspaceCard icon="⚠️" title="Risco" tone="risk">
        <div className={styles.metricsRow}>
          {model.risk.metrics.map((metric) => (
            <article key={metric.label} className={styles.progressWrap}>
              <div className={styles.progressHeader}>
                <span className={styles.progressLabel}>{metric.label}</span>
                <span className={`${styles.metricValueCompact} ${toneClass(metric.tone)}`}>{metric.valueText}</span>
              </div>
              <div className={styles.progressTrack}>
                <div className={`${styles.progressFill} ${progressToneClass(metric.tone)}`} style={{ width: `${metric.percent}%` }} />
              </div>
            </article>
          ))}
        </div>
        <div className={styles.metricsRow}>
          <article className={styles.progressWrap}>
            <div className={styles.progressHeader}>
              <span className={styles.progressLabel}>{model.risk.lower[0].label}</span>
              <span className={`${styles.metricValueCompact} ${toneClass(model.risk.lower[0].tone)}`}>{model.risk.lower[0].valueText}</span>
            </div>
            <div className={styles.progressTrack}>
              <div className={`${styles.progressFill} ${progressToneClass(model.risk.lower[0].tone)}`} style={{ width: `${model.risk.lower[0].percent}%` }} />
            </div>
          </article>
          <article className={styles.metricItem}>
            <div className={styles.metricLabel}>{model.risk.lower[1].label}</div>
            <div className={`${styles.trendText} ${styles[`trend_${model.risk.lower[1].tone}`]}`}>{model.risk.lower[1].text}</div>
          </article>
        </div>
      </WorkspaceCard>
    </div>
  );
}




