import type { PropsWithChildren } from "react";

import styles from "./MetricChip.module.css";

export function MetricChip(props: PropsWithChildren<{ label: string }>) {
  return (
    <div className={styles.chip}>
      <small>{props.label}</small>
      <strong>{props.children}</strong>
    </div>
  );
}
