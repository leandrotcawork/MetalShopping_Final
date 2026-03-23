import styles from "../history.module.css";

type InfoItem = {
  label: string;
  value: string;
};

type InfoBarProps = {
  items: [InfoItem, InfoItem, InfoItem, InfoItem];
};

export function InfoBar({ items }: InfoBarProps) {
  return (
    <section className={styles.historyInfoBar}>
      {items.map((item) => (
        <article key={item.label} className={styles.infoItem}>
          <span className={styles.infoLabel}>{item.label}</span>
          <span className={styles.infoValue}>{item.value}</span>
        </article>
      ))}
    </section>
  );
}
