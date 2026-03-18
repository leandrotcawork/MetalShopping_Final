import { useEffect, useMemo, useState } from "react";

import { AppFrame, MetricCard, StatusPill, SurfaceCard } from "@metalshopping/ui";

import { listProductsPortfolio, type ProductsPortfolioQuery, type ProductsPortfolioResult } from "./api";
import styles from "./ProductsPortfolioPage.module.css";
import {
  buildProductsPortfolioSummary,
  formatCurrency,
  formatQuantity,
  rowHasLiveCommercialState,
} from "./view-model";

const defaultQuery: ProductsPortfolioQuery = {
  search: "",
  brandName: "",
  taxonomyLeaf0Name: "",
  status: "",
  limit: 50,
  offset: 0,
};

export function ProductsPortfolioPage() {
  const [query, setQuery] = useState<ProductsPortfolioQuery>(defaultQuery);
  const [result, setResult] = useState<ProductsPortfolioResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);

      try {
        const next = await listProductsPortfolio(query);
        if (!cancelled) {
          setResult(next);
        }
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError instanceof Error ? loadError.message : "Failed to load products.");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [query]);

  const summary = useMemo(() => buildProductsPortfolioSummary(result), [result]);

  const brands = result?.filters.brands ?? [];
  const taxonomyLeaf0Names = result?.filters.taxonomy_leaf0_names ?? [];
  const statuses = result?.filters.status ?? [];
  const totalVisible = result?.paging.returned ?? 0;
  const totalMatching = result?.paging.total ?? 0;
  const liveRows = result?.rows.filter((row) => rowHasLiveCommercialState(row)).length ?? 0;
  const inventoryLive = result?.rows.filter((row) => (row.on_hand_quantity ?? 0) > 0).length ?? 0;

  return (
    <AppFrame
      eyebrow="Operational Surface"
      title="Products"
      subtitle="Portfolio visibility rebuilt on top of the canonical backend. Identity comes from Catalog, price from Pricing, and stock from Inventory through a single backend-owned read surface."
      aside={
        <div className={styles.heroAside}>
          <MetricCard
            label="Visible now"
            value={totalVisible}
            hint={`${totalMatching} matching products in the current filter scope.`}
          />
          <MetricCard
            label="Commercially live"
            value={liveRows}
            hint="Rows with current price or stock visibility."
          />
          <MetricCard
            label="Inventory live"
            value={inventoryLive}
            hint="Rows currently exposing on-hand quantity above zero."
          />
          <MetricCard
            label="Avg visible price"
            value={summary.averageVisiblePrice}
            hint="Average over products with a current effective price."
          />
        </div>
      }
    >
      <div className={styles.stack}>
        <div className={styles.bannerRow}>
          <div className={`${styles.statusBanner} ${error ? styles.statusBannerError : styles.statusBannerSuccess}`}>
            <strong>{error ? "Surface degraded" : "Read surface online"}</strong>
            <span>
              {error
                ? error
                : `Catalog, Pricing, and Inventory are composing ${summary.totalProducts} canonical products for the current tenant.`}
            </span>
          </div>
          <div className={styles.quickActions}>
            <button type="button" className={`${styles.actionButton} ${styles.actionButtonPrimary}`}>
              Products live
            </button>
            <button type="button" className={styles.actionButton} disabled>
              Shopping soon
            </button>
          </div>
        </div>

        <SurfaceCard
          title="Portfolio filters"
          subtitle="Search and narrow the active portfolio the same way the legacy Products workspace did, but now through a backend-owned read surface."
          className={styles.filtersCard}
        >
          <div className={styles.toolbar}>
            <label className={styles.field}>
              <span className={styles.label}>Search</span>
              <input
                className={styles.input}
                value={query.search}
                placeholder="SKU, description, pn_interno, EAN or reference"
                onChange={(event) =>
                  setQuery((current) => ({
                    ...current,
                    search: event.target.value,
                    offset: 0,
                  }))
                }
              />
            </label>

            <label className={styles.field}>
              <span className={styles.label}>Brand</span>
              <select
                className={styles.select}
                value={query.brandName}
                onChange={(event) =>
                  setQuery((current) => ({
                    ...current,
                    brandName: event.target.value,
                    offset: 0,
                  }))
                }
              >
                <option value="">All brands</option>
                {brands.map((brand) => (
                  <option key={brand} value={brand}>
                    {brand}
                  </option>
                ))}
              </select>
            </label>

            <label className={styles.field}>
              <span className={styles.label}>Taxonomy</span>
              <select
                className={styles.select}
                value={query.taxonomyLeaf0Name}
                onChange={(event) =>
                  setQuery((current) => ({
                    ...current,
                    taxonomyLeaf0Name: event.target.value,
                    offset: 0,
                  }))
                }
              >
                <option value="">All taxonomy roots</option>
                {taxonomyLeaf0Names.map((name) => (
                  <option key={name} value={name}>
                    {name}
                  </option>
                ))}
              </select>
            </label>

            <label className={styles.field}>
              <span className={styles.label}>Status</span>
              <select
                className={styles.select}
                value={query.status}
                onChange={(event) =>
                  setQuery((current) => ({
                    ...current,
                    status: event.target.value,
                    offset: 0,
                  }))
                }
              >
                <option value="">All statuses</option>
                {statuses.map((status) => (
                  <option key={status} value={status}>
                    {status}
                  </option>
                ))}
              </select>
            </label>
          </div>
        </SurfaceCard>

        <SurfaceCard
          title="Portfolio table"
          subtitle="Current product identity, current pricing state, and current stock position in one operational workspace."
          actions={
            <span className={styles.tableMeta}>
              {loading
                ? "Refreshing..."
                : `Showing ${result?.paging.returned ?? 0} of ${result?.paging.total ?? 0}`}
            </span>
          }
          className={styles.tableCard}
        >
          <div className={styles.statusRow}>
            <span>Current workspace status</span>
            <span>{loading ? "Syncing visible rows..." : "Portfolio synchronized."}</span>
          </div>

          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Product</th>
                  <th>Identifiers</th>
                  <th>Classification</th>
                  <th>Current Price</th>
                  <th>Current Costs</th>
                  <th>Stock</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {result?.rows.length ? (
                  result.rows.map((row) => (
                    <tr key={row.product_id}>
                      <td>
                        <div className={styles.identity}>
                          <p className={styles.name}>{row.name}</p>
                          <p className={styles.meta}>{row.sku}</p>
                          <p className={styles.meta}>{row.description ?? "No description yet."}</p>
                        </div>
                      </td>
                      <td>
                        <div className={styles.secondary}>
                          <span>PN: {row.pn_interno ?? "—"}</span>
                          <span>Ref: {row.reference ?? "—"}</span>
                          <span>EAN: {row.ean ?? "—"}</span>
                        </div>
                      </td>
                      <td>
                        <div className={styles.secondary}>
                          <span>Brand: {row.brand_name ?? "—"}</span>
                          <span>Taxonomy: {row.taxonomy_leaf_name ?? "—"}</span>
                          <span>Root: {row.taxonomy_leaf0_name ?? "—"}</span>
                        </div>
                      </td>
                      <td className={styles.money}>{formatCurrency(row.current_price_amount)}</td>
                      <td>
                        <div className={styles.secondary}>
                          <span>Replacement: {formatCurrency(row.replacement_cost_amount)}</span>
                          <span>Average: {formatCurrency(row.average_cost_amount)}</span>
                        </div>
                      </td>
                      <td>
                        <div className={styles.secondary}>
                          <span>{formatQuantity(row.on_hand_quantity)}</span>
                          <span>{row.inventory_position_status ?? "No position"}</span>
                        </div>
                      </td>
                      <td>
                        <StatusPill
                          label={row.product_status}
                          tone={
                            row.product_status === "active" && rowHasLiveCommercialState(row)
                              ? "success"
                              : row.product_status === "active"
                                ? "neutral"
                                : "muted"
                          }
                        />
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td className={styles.empty} colSpan={7}>
                      {loading
                        ? "Loading products..."
                        : "No products matched the current portfolio filters."}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </SurfaceCard>
      </div>
    </AppFrame>
  );
}
