import styles from "../../analytics_products.module.css";

type ProductsHeaderProps = {
  totalSkus: number;
  updatedAtLabel: string;
  onExport: () => void;
};

export function ProductsHeader({ totalSkus, updatedAtLabel, onExport }: ProductsHeaderProps) {
  return (
    <header className={styles.productsHeader}>
      <div>
        <h2 className={styles.productsTitle}>Produtos Analytics</h2>
        <p className={styles.productsSub}>Operating Window 6M - Supporting 3M</p>
      </div>
      <div className={styles.productsMeta}>
        <span className={styles.metaPill}>{totalSkus} produtos</span>
        <span className={styles.metaPill}>Atualizado: {updatedAtLabel}</span>
        <button type="button" className={styles.ghostBtn} onClick={onExport}>
          Exportar
        </button>
      </div>
    </header>
  );
}
