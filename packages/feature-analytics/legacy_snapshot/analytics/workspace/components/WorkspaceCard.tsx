import type { ReactNode } from "react";

import styles from "../product_workspace.module.css";

type WorkspaceCardProps = {
  icon: string;
  title: string;
  tone: "stock" | "profit" | "competition" | "risk";
  children: ReactNode;
};

export function WorkspaceCard({ icon, title, tone, children }: WorkspaceCardProps) {
  return (
    <article className={styles.glassCard}>
      <header className={styles.cardHeader}>
        <div className={`${styles.cardIcon} ${styles[`cardIcon_${tone}`]}`}>{icon}</div>
        <h2 className={styles.cardTitle}>{title}</h2>
      </header>
      {children}
    </article>
  );
}

