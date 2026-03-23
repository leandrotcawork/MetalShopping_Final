import styles from "../analytics_home.module.css";

type SpotlightNextStepsProps = {
  items: string[];
};

export function SpotlightNextSteps({ items }: SpotlightNextStepsProps) {
  const bullets = ["1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"];
  return (
    <section className={styles.spotSection}>
      <h4>Proximos passos</h4>
      <div className={styles.nextStepsList}>
        {items.map((item, idx) => (
          <div key={`${idx}-${item}`} className={styles.nextStepItem}>
            <span className={styles.nextStepBadge} aria-hidden>{bullets[idx] || "•"}</span>
            <span className={styles.nextStepText}>{item}</span>
          </div>
        ))}
      </div>
    </section>
  );
}
