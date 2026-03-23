import styles from "../../analytics_products.module.css";

type ProductsToolbarProps = {
  query: string;
  onQueryChange: (value: string) => void;
  viewMode: "density" | "analytics";
  onViewChange: (next: "density" | "analytics") => void;
  metaPrimary?: string | null;
  metaSecondary?: string | null;
  compact?: boolean;
};

export function ProductsToolbar({
  query,
  onQueryChange,
  viewMode,
  onViewChange,
  metaPrimary,
  metaSecondary,
  compact = false,
}: ProductsToolbarProps) {
  if (compact) {
    return (
      <div className={styles.productsToolbar}>
        <div className={styles.headerRight}>
          <div className={styles.viewSwitcher}>
            <button
              type="button"
              className={`${styles.viewBtn} ${viewMode === "analytics" ? styles.active : ""}`}
              onClick={() => onViewChange("analytics")}
            >
              Analytics
            </button>
            <button
              type="button"
              className={`${styles.viewBtn} ${viewMode === "density" ? styles.active : ""}`}
              onClick={() => onViewChange("density")}
            >
              Busca
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.productsToolbar}>
      <label className={styles.searchContainer}>
        <span className={styles.searchIcon} aria-hidden>🔍</span>
        <input
          id="productsSearch"
          type="text"
          className={styles.searchBar}
          placeholder="Buscar produtos por PN, EAN ou descricao..."
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
        />
        <span className={styles.searchKbd}>⌘K</span>
      </label>
      <div className={styles.headerRight}>
        {(metaPrimary || metaSecondary) ? (
          <div className={styles.toolbarMeta}>
            {metaPrimary ? <span className={styles.toolbarMetaPrimary}>{metaPrimary}</span> : null}
            {metaSecondary ? <span className={styles.toolbarMetaSecondary}>{metaSecondary}</span> : null}
          </div>
        ) : null}
        <div className={styles.viewSwitcher}>
          <button
            type="button"
            className={`${styles.viewBtn} ${viewMode === "analytics" ? styles.active : ""}`}
            onClick={() => onViewChange("analytics")}
          >
            Analytics
          </button>
          <button
            type="button"
            className={`${styles.viewBtn} ${viewMode === "density" ? styles.active : ""}`}
            onClick={() => onViewChange("density")}
          >
            Busca
          </button>
        </div>
      </div>
    </div>
  );
}
