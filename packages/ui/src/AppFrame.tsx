import type { PropsWithChildren, ReactNode } from "react";

import styles from "./AppFrame.module.css";

type AppFrameProps = PropsWithChildren<{
  eyebrow: string;
  title: string;
  subtitle: ReactNode;
  aside?: ReactNode;
}>;

export function AppFrame({ eyebrow, title, subtitle, aside, children }: AppFrameProps) {
  return (
    <div className={styles.frame}>
      <header className={styles.hero}>
        <div className={styles.heroMain}>
          <span className={styles.eyebrow}>{eyebrow}</span>
          <h1 className={styles.title}>{title}</h1>
          <p className={styles.subtitle}>{subtitle}</p>
        </div>
        {aside ? <div className={styles.heroAside}>{aside}</div> : null}
      </header>
      <main className={styles.content}>{children}</main>
    </div>
  );
}
