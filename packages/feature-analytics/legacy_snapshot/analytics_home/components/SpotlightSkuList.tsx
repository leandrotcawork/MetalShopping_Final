import styles from "../analytics_home.module.css";

type SpotlightSkuListProps = {
  skus: string[];
};

export function SpotlightSkuList({ skus }: SpotlightSkuListProps) {
  return (
    <section className={styles.spotSection}>
      <h4>SKUs relacionados</h4>
      <div className={styles.drawerSkus}>
        {skus.length ? (
          skus.map((sku) => (
            <span key={sku} className={styles.skuPill}>
              SKU {sku}
            </span>
          ))
        ) : (
          <small>Sem SKUs relacionados.</small>
        )}
      </div>
    </section>
  );
}
