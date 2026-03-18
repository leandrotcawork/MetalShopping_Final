import styles from "./AuthStatusScreen.module.css";

export function AuthBootstrapScreen() {
  return (
    <div className={styles.screen}>
      <div className={`${styles.panel} ${styles.panelCompact}`.trim()}>
        <strong className={styles.brand}>MetalShopping</strong>
        <span className={`${styles.message} ${styles.messageCompact}`.trim()}>
          Validando sessao ativa...
        </span>
      </div>
    </div>
  );
}
