import styles from "../insights.module.css";
import type { InsightDomain, InsightSeverity } from "./types";

const DOMAIN_LABEL: Record<InsightDomain, string> = {
  COMPETITIVIDADE: "PRECO",
  ESTOQUE: "ESTOQUE",
  RENTABILIDADE: "PORTFOLIO",
  RISCO: "RISCO",
  DADOS: "CLASSIFICACAO",
  MERCADO: "CAMPANHA",
};

type InsightBadgeProps =
  | {
      variant: "domain";
      domain: InsightDomain;
      className?: string;
    }
  | {
      variant: "severity";
      severity: InsightSeverity;
      className?: string;
    };

function severityLabel(severity: InsightSeverity): string {
  if (severity === "CRITICAL") return "Critico";
  if (severity === "WARN") return "Atencao";
  if (severity === "GOOD") return "Bom";
  return "Info";
}

export function InsightBadge(props: InsightBadgeProps) {
  if (props.variant === "domain") {
    return (
      <span className={`${styles.badge} ${styles[`badgeDomain_${props.domain}`]}${props.className ? ` ${props.className}` : ""}`}>
        {DOMAIN_LABEL[props.domain]}
      </span>
    );
  }
  return (
    <span className={`${styles.badge} ${styles[`badgeSeverity_${props.severity}`]}${props.className ? ` ${props.className}` : ""}`}>
      {severityLabel(props.severity)}
    </span>
  );
}
