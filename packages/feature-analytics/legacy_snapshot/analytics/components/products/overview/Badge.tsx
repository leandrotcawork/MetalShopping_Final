import styles from "./products_overview.module.css";

type BadgeTone = "critical" | "attention" | "opportunity" | "focus";

type BadgeProps = {
  left: string;
  right: string;
  tone: BadgeTone;
};

export function Badge({ left, right, tone }: BadgeProps) {
  return (
    <div className={`${styles.actionBadge} ${styles[`actionBadge_${tone}`]}`}>
      <span>{left || "-"}</span>
      <span aria-hidden>•</span>
      <span>{right || "-"}</span>
    </div>
  );
}

