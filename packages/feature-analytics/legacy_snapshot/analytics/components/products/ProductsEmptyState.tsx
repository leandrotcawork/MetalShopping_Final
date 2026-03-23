import styles from "../../analytics_products.module.css";

type ProductsEmptyStateProps = {
  onReset: () => void;
};

export function ProductsEmptyState({ onReset }: ProductsEmptyStateProps) {
  return (
    <section className={styles.empty}>
      <h3>No SKU matched the current filters</h3>
      <p>Try clearing search or adjusting taxonomy, brand, status and action filters.</p>
      <div>
        <button type="button" className={styles.clearBtn} onClick={onReset}>
          Reset filters
        </button>
      </div>
    </section>
  );
}
