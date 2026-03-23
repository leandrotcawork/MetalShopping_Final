// @ts-nocheck
import styles from "../analytics_home.module.css";

type SpotlightCalloutProps = {
  title: string;
  text: string;
};

export function SpotlightCallout({ title, text }: SpotlightCalloutProps) {
  return (
    <section className={styles.callout}>
      <h4>{title}</h4>
      <p>{text}</p>
    </section>
  );
}

