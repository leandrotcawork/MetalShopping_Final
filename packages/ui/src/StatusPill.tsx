import styles from "./StatusPill.module.css";

type StatusPillTone = "success" | "neutral" | "muted";

type StatusPillProps = {
  label: string;
  tone?: StatusPillTone;
};

export function StatusPill({ label, tone = "neutral" }: StatusPillProps) {
  const toneClass =
    tone === "success"
      ? styles.success
      : tone === "muted"
        ? styles.muted
        : styles.neutral;

  return <span className={`${styles.pill} ${toneClass}`}>{label}</span>;
}
