import type { MouseEvent, ReactNode } from "react";

import { Card } from "../../../components/ui";
import styles from "../analytics_home.module.css";

type CardShellProps = {
  emoji?: string;
  title: string;
  subtitle?: string;
  actionLabel?: string;
  onAction?: () => void;
  children: ReactNode;
  className?: string;
};

export function CardShell({ emoji, title, subtitle, actionLabel, onAction, children, className = "" }: CardShellProps) {
  const clickable = typeof onAction === "function";
  return (
    <div
      onClick={(event: MouseEvent<HTMLDivElement>) => {
        if (!onAction) return;
        const target = event.target as HTMLElement | null;
        if (target && target.closest("button, a, input, select, textarea, [role='button']")) return;
        onAction();
      }}
    >
      <Card variant="glass" className={`${styles.card} ${clickable ? styles.cardClickable : ""} ${className}`.trim()}>
      <header className={styles.cardHead}>
        <div>
          <h3 className={styles.cardTitle}>{emoji ? <span className={styles.emoji}>{emoji}</span> : null}{title}</h3>
          {subtitle ? <p className={styles.cardSub}>{subtitle}</p> : null}
        </div>
        {actionLabel ? <button type="button" className={styles.ghostBtn} onClick={onAction}>{actionLabel}</button> : null}
      </header>
      {children}
      </Card>
    </div>
  );
}

