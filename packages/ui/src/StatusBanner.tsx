import type { PropsWithChildren } from "react";

import styles from "./StatusBanner.module.css";

export function StatusBanner(props: PropsWithChildren<{ tone?: "success" | "error"; className?: string }>) {
  const tone = props.tone ?? "success";
  return <div className={`${styles.banner} ${styles[tone]} ${props.className ?? ""}`.trim()}>{props.children}</div>;
}
