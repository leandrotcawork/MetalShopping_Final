import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import styles from "../analytics_home.module.css";

type Props = {
  onOpenSpotlight: (key: string) => void;
  rows: AnalyticsHomeViewModel["actions"];
};

function priorityFinanceLabel(tier?: string): string | null {
  if (!tier) return null;
  if (tier === "P0") return "PF Critica";
  if (tier === "P1") return "PF Alta";
  if (tier === "P2") return "PF Media";
  if (tier === "P3") return "PF Baixa";
  return null;
}

function decisionStateLabel(state?: string): string | null {
  if (!state) return null;
  if (state === "READY") return "Pronto";
  if (state === "CAUTION") return "Cautela";
  if (state === "REVIEW") return "Revisao";
  if (state === "BLOCKED") return "Bloqueado";
  return null;
}

function dominantObjectiveLabel(objective?: string): string | null {
  if (!objective) return null;
  if (objective === "CAPITAL_RELIEF") return "Alivio de capital";
  if (objective === "DEMAND_GENERATION") return "Geracao de demanda";
  if (objective === "COMPETITIVE_POSITIONING") return "Posicionamento competitivo";
  if (objective === "MIX_OPTIMIZATION") return "Otimizacao de mix";
  if (objective === "NONE") return "Sem objetivo dominante";
  return null;
}

export function TopActionsList({ onOpenSpotlight, rows }: Props) {
  return (
    <div className={styles.actionsBox}>
      <div className={styles.actions}>
        {rows.map((row) => {
          const priorityLabel = priorityFinanceLabel(row.valuePriorityTier);
          const stateLabel = decisionStateLabel(row.decisionState);
          const objectiveLabel = dominantObjectiveLabel(row.dominantObjective);
          const desc = objectiveLabel
            ? `${row.desc} - Objetivo: ${objectiveLabel}`
            : row.desc;
          return (
          <button
            key={row.key}
            type="button"
            className={`${styles.action} ${styles[`signalCode_${row.actionCode}`] || styles[`signal_${row.signalClass}`]}`}
            onClick={() => onOpenSpotlight(row.key)}
          >
            <span className={styles.aTop}>
              <span className={styles.aLeft}>
                <span className={styles.aName}>{row.name}</span>
                <span className={styles.aDesc}>{desc}</span>
              </span>
              <span className={styles.aMeta}>
                <span className={`${styles.tag} ${styles[`tagCode_${row.actionCode}`] || styles[`tag_${row.signalClass}`]}`}>{row.skuCount}</span>
                {priorityLabel ? (
                  <span className={`${styles.tag} ${styles[`tagCode_${row.actionCode}`] || styles[`tag_${row.signalClass}`]}`}>{priorityLabel}</span>
                ) : null}
                {stateLabel ? (
                  <span className={`${styles.tag} ${styles[`tagCode_${row.actionCode}`] || styles[`tag_${row.signalClass}`]}`}>{stateLabel}</span>
                ) : null}
              </span>
            </span>
          </button>
          );
        })}
      </div>
    </div>
  );
}
