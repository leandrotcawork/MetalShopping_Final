import { InfoTooltipLabel } from "../../../workspace/components/InfoTooltipLabel";
import styles from "./products_overview.module.css";

type KpiCardProps = {
  label: string;
  value: string;
  change: string;
  tone: "positive" | "negative";
  helpItems: string[];
  showChangeArrow?: boolean;
};

export function KpiCard({ label, value, change, tone, helpItems, showChangeArrow = true }: KpiCardProps) {
  const help = helpItems.length
    ? {
        title: label,
        items: helpItems,
      }
    : null;
  const hasChange = String(change || "").trim().length > 0;

  return (
    <article className={styles.kpiCard}>
      <div className={styles.kpiCardSurface}>
        <InfoTooltipLabel label={label} help={help} className={styles.kpiLabelWithInfo} />
        <p className={styles.kpiValue}>{value || "-"}</p>
        {hasChange ? (
          <p className={`${styles.kpiChange} ${styles[`kpiChange_${tone}`]}`}>
            {showChangeArrow ? <span aria-hidden>{tone === "positive" ? "\u2197" : "\u2198"}</span> : null}
            <span>{change}</span>
          </p>
        ) : null}
      </div>
    </article>
  );
}
