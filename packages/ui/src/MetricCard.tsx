import type { ReactNode } from "react";

import styles from "./MetricCard.module.css";

type MetricCardProps = {
  label: string;
  value: ReactNode;
  hint: ReactNode;
};

export function MetricCard({ label, value, hint }: MetricCardProps) {
  return (
    <section className={styles.card}>
      <p className={styles.label}>{label}</p>
      <p className={styles.value}>{value}</p>
      <p className={styles.hint}>{hint}</p>
    </section>
  );
}
