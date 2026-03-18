import type { PropsWithChildren } from "react";

import styles from "./SortHeaderButton.module.css";

export function SortHeaderButton(
  props: PropsWithChildren<{
    indicator: string;
    onClick: () => void;
  }>,
) {
  return (
    <button type="button" className={styles.button} onClick={props.onClick}>
      <span>{props.children}</span>
      <span className={styles.indicator} aria-hidden="true">{props.indicator}</span>
    </button>
  );
}
