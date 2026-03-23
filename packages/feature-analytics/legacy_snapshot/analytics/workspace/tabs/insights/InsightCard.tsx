import { InsightBadge } from "./InsightBadge";
import type { Insight } from "./types";
import styles from "../insights.module.css";
import { chipIconKey, friendlyChipText } from "./chips_mapper";
import type { CSSProperties, KeyboardEvent } from "react";
import { getInsightIconAsset } from "../../../registry/analyticsRegistry";

type InsightCardProps = {
  insight: Insight;
  onClick?: () => void;
};

type ThemeKey = "pricing" | "campaign" | "portfolio" | "risk" | "classification";

function domainTheme(domain: Insight["domain"]): ThemeKey {
  if (domain === "COMPETITIVIDADE") return "pricing";
  if (domain === "MERCADO") return "campaign";
  if (domain === "ESTOQUE") return "portfolio";
  if (domain === "RENTABILIDADE" || domain === "DADOS") return "classification";
  if (domain === "RISCO") return "risk";
  return "classification";
}

function chipListForInsight(insight: Insight): string[] {
  const fallbackTags = (insight.evidence || []).map((row) => `${row.label}: ${row.value || "--"}`);
  const tags = insight.tags && insight.tags.length > 0 ? insight.tags : fallbackTags;
  const alerts = insight.alerts || [];
  const actions = insight.actionTags || [];
  if (insight.domain === "COMPETITIVIDADE") {
    return [...alerts, ...tags, ...actions];
  }
  return [...tags, ...alerts, ...actions];
}

function chipIcon(tokenText: string): string | null {
  const key = chipIconKey(tokenText);
  if (!key) return null;
  return getInsightIconAsset(key);
}

export function InsightCard({ insight, onClick }: InsightCardProps) {
  const theme = domainTheme(insight.domain);
  const priorityLabel = insight.severity === "CRITICAL" ? "Alta" : insight.severity === "WARN" ? "Media" : "Baixa";
  const priorityClass =
    insight.severity === "CRITICAL"
      ? styles.priorityHigh
      : insight.severity === "WARN"
        ? styles.priorityMedium
        : styles.priorityLow;
  const confidenceClass =
    insight.confidence == null
      ? ""
      : insight.confidence <= 32
        ? styles.confidenceLow
        : insight.confidence <= 66
          ? styles.confidenceMedium
          : styles.confidenceHigh;

  const chips = chipListForInsight(insight)
    .map((item) => friendlyChipText(item))
    .filter((item) => String(item || "").trim().length > 0);
  const visibleChips = chips.slice(0, 4);
  const hiddenCount = Math.max(0, chips.length - visibleChips.length);

  function handleKeyDown(event: KeyboardEvent<HTMLElement>) {
    if (!onClick) return;
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      onClick();
    }
  }

  return (
    <article
      className={`${styles.recoCard} ${styles[`recoCardTheme_${theme}`]}${onClick ? ` ${styles.recoCardInteractive}` : ""}`}
      onClick={onClick}
      onKeyDown={handleKeyDown}
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      aria-label={onClick ? `Abrir detalhes de ${insight.title}` : undefined}
    >
      <div className={styles.recoHeader}>
        <InsightBadge variant="domain" domain={insight.domain} className={styles.recoBadge} />
        <div className={styles.recoHeaderBody}>
          <h3 className={`${styles.recoTitle} ${styles[`recoTitleTheme_${theme}`]}`}>{insight.title}</h3>
          <div className={styles.recoMetaRow}>
            <span className={styles.recoPriority}>
              Prioridade: <strong className={priorityClass}>{priorityLabel}</strong>
            </span>
            <span className={styles.recoDivider}>|</span>
            <span className={styles.recoConfidenceLine}>
              Confianca: <strong className={confidenceClass}>{insight.confidence != null ? `${Math.round(insight.confidence)}%` : "--"}</strong>
            </span>
          </div>
        </div>
      </div>

      <div className={styles.recoDividerLine} />

      <div className={styles.recoActionRow}>
        <span className={`${styles.recoActionIcon} ${styles[`recoActionIconTheme_${theme}`]}`} aria-hidden />
        <span className={styles.recoActionText}>{insight.summary}</span>
      </div>

      {visibleChips.length > 0 ? (
        <div className={styles.recoChipRow}>
          {visibleChips.map((chip) => {
            const icon = chipIcon(chip);
            return (
              <span key={`${insight.id}-${chip}`} className={`${styles.recoChip} ${styles[`recoChipTheme_${theme}`]}`}>
                {icon ? (
                  <span
                    className={styles.recoChipIcon}
                    style={{ "--chip-icon-url": `url(${icon})` } as CSSProperties}
                    aria-hidden
                  />
                ) : null}
                {chip}
              </span>
            );
          })}
          {hiddenCount > 0 ? <span className={styles.recoChipMore}>+{hiddenCount}</span> : null}
        </div>
      ) : null}
    </article>
  );
}
