import type { PropsWithChildren, ReactNode } from "react";

import styles from "./SurfaceCard.module.css";

type SurfaceCardProps = PropsWithChildren<{
  title?: ReactNode;
  subtitle?: ReactNode;
  actions?: ReactNode;
  tone?: "default" | "soft";
  className?: string;
}>;

export function SurfaceCard({
  title,
  subtitle,
  actions,
  tone = "default",
  className,
  children,
}: SurfaceCardProps) {
  const classNames = [styles.card, tone === "soft" ? styles.soft : "", className ?? ""]
    .filter(Boolean)
    .join(" ");

  return (
    <section className={classNames}>
      {title || subtitle || actions ? (
        <header className={styles.head}>
          <div className={styles.headMain}>
            {title ? <h2 className={styles.title}>{title}</h2> : null}
            {subtitle ? <p className={styles.subtitle}>{subtitle}</p> : null}
          </div>
          {actions ? <div className={styles.actions}>{actions}</div> : null}
        </header>
      ) : null}
      {children}
    </section>
  );
}
