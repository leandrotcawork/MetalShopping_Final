// @ts-nocheck
import styles from "../analytics_home.module.css";

type HeroCommandCenterProps = {
  onOpenSpotlight: (key: string) => void;
  updatedAtLabel?: string;
};

export function HeroCommandCenter({ onOpenSpotlight, updatedAtLabel }: HeroCommandCenterProps) {
  const updatedLabel = updatedAtLabel && updatedAtLabel !== "N/D" ? updatedAtLabel : "N/D";

  return (
    <section className={styles.heroMain}>
      <div className={styles.heroTitle}>{"Vis\u00e3o Inteligente do Portfolio"}</div>
      <div className={styles.heroSub}>
        {"Insights acion\u00e1veis \u2022 janela "}
        <b>6M</b>
        {" (principal) + sinal "}
        <b>3M</b>
        {" \u2022 atualizado "}
        {updatedLabel}
      </div>
      <div className={styles.cmdRow}>
        <button type="button" className={styles.cmd} onClick={() => onOpenSpotlight("cmd-run")}>
          <span className={styles.k}>R</span>
          <span>
            <b className={styles.t}>Nova Run</b>
            <small className={styles.d}>Atualizar scraping e metricas</small>
          </span>
        </button>
        <button type="button" className={styles.cmd} onClick={() => onOpenSpotlight("cmd-review")}>
          <span className={styles.k}>A</span>
          <span>
            <b className={styles.t}>{"Ver A\u00e7\u00f5es"}</b>
            <small className={styles.d}>{"Fila / Kanban \u2022 urg\u00eancia"}</small>
          </span>
        </button>
        <button type="button" className={styles.cmd} onClick={() => onOpenSpotlight("cmd-changes")}>
          <span className={styles.k}>C</span>
          <span>
            <b className={styles.t}>O que mudou</b>
            <small className={styles.d}>Deltas desde a semana passada</small>
          </span>
        </button>
        <button type="button" className={styles.cmd} onClick={() => onOpenSpotlight("cmd-alerts")}>
          <span className={styles.k}>L</span>
          <span>
            <b className={styles.t}>Alertas</b>
            <small className={styles.d}>Quebra de banda / estoque</small>
          </span>
        </button>
      </div>
    </section>
  );
}

