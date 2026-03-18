import { startTransition, useDeferredValue, useEffect, useMemo, useState } from "react";

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

const pageSizeOptions = [25, 50, 100];

function readQueryFromUrl(): ProductsPortfolioQuery {
  if (typeof window === "undefined") {
    return defaultQuery;
  }

  const params = new URLSearchParams(window.location.search);
  const limit = Number(params.get("limit") ?? defaultQuery.limit);
  const offset = Number(params.get("offset") ?? defaultQuery.offset);

  return {
    search: params.get("search") ?? defaultQuery.search,
    brandName: params.get("brand_name") ?? defaultQuery.brandName,
    taxonomyLeaf0Name: params.get("taxonomy_leaf0_name") ?? defaultQuery.taxonomyLeaf0Name,
    status: params.get("status") ?? defaultQuery.status,
    limit: Number.isFinite(limit) && limit > 0 ? limit : defaultQuery.limit,
    offset: Number.isFinite(offset) && offset >= 0 ? offset : defaultQuery.offset,
  };
}

function writeQueryToUrl(query: ProductsPortfolioQuery) {
  if (typeof window === "undefined") {
    return;
  }

  const params = new URLSearchParams();
  if (query.search.trim() !== "") params.set("search", query.search.trim());
  if (query.brandName.trim() !== "") params.set("brand_name", query.brandName.trim());
  if (query.taxonomyLeaf0Name.trim() !== "") params.set("taxonomy_leaf0_name", query.taxonomyLeaf0Name.trim());
  if (query.status.trim() !== "") params.set("status", query.status.trim());
  if (query.limit !== defaultQuery.limit) params.set("limit", String(query.limit));
  if (query.offset !== defaultQuery.offset) params.set("offset", String(query.offset));

  const nextSearch = params.toString();
  const nextUrl = `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ""}`;
  window.history.replaceState(null, "", nextUrl);
}

export function ProductsPortfolioPage() {
  const [query, setQuery] = useState<ProductsPortfolioQuery>(() => readQueryFromUrl());
  const [searchDraft, setSearchDraft] = useState(() => readQueryFromUrl().search);
  const [result, setResult] = useState<ProductsPortfolioResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const deferredSearch = useDeferredValue(searchDraft);

  useEffect(() => {
    startTransition(() => {
      setQuery((current) => {
        if (current.search === deferredSearch) {
          return current;
        }

        return {
          ...current,
          search: deferredSearch,
          offset: 0,
        };
      });
    });
  }, [deferredSearch]);

  useEffect(() => {
    writeQueryToUrl(query);
  }, [query]);

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
  const currentPage = Math.floor(query.offset / query.limit) + 1;
  const totalPages = totalMatching === 0 ? 1 : Math.max(1, Math.ceil(totalMatching / query.limit));
  const canGoPrevious = query.offset > 0;
  const canGoNext = query.offset + query.limit < totalMatching;

  const activeFilters = [
    query.search.trim() !== "" ? { key: "search", label: `Busca: ${query.search.trim()}` } : null,
    query.brandName.trim() !== "" ? { key: "brandName", label: `Marca: ${query.brandName.trim()}` } : null,
    query.taxonomyLeaf0Name.trim() !== ""
      ? { key: "taxonomyLeaf0Name", label: `Taxonomia: ${query.taxonomyLeaf0Name.trim()}` }
      : null,
    query.status.trim() !== "" ? { key: "status", label: `Status: ${query.status.trim()}` } : null,
  ].filter((item): item is { key: string; label: string } => item !== null);

  const hasActiveFilters = activeFilters.length > 0;

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
                value={searchDraft}
                placeholder="SKU, description, pn_interno, EAN or reference"
                onChange={(event) => setSearchDraft(event.target.value)}
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

          <div className={styles.filterFooter}>
            <div className={styles.filterChips}>
              {hasActiveFilters ? (
                activeFilters.map((filter) => (
                  <button
                    key={filter.key}
                    type="button"
                    className={styles.filterChip}
                    onClick={() => {
                      if (filter.key === "search") {
                        setSearchDraft("");
                      }

                      setQuery((current) => ({
                        ...current,
                        [filter.key]: "",
                        offset: 0,
                      }));
                    }}
                  >
                    <span>{filter.label}</span>
                    <span aria-hidden="true">x</span>
                  </button>
                ))
              ) : (
                <span className={styles.filterHint}>No active filters. The full visible tenant portfolio is being shown.</span>
              )}
            </div>

            <button
              type="button"
              className={styles.clearButton}
              disabled={!hasActiveFilters}
              onClick={() => {
                setSearchDraft("");
                setQuery({
                  ...defaultQuery,
                  limit: query.limit,
                });
              }}
            >
              Clear filters
            </button>
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

          <div className={styles.paginationRow}>
            <div className={styles.paginationStatus}>
              <span>Page {currentPage} of {totalPages}</span>
              <span>{totalMatching} matching products</span>
            </div>

            <div className={styles.paginationActions}>
              <label className={styles.pageSizeField}>
                <span>Rows</span>
                <select
                  className={styles.pageSizeSelect}
                  value={query.limit}
                  onChange={(event) =>
                    setQuery((current) => ({
                      ...current,
                      limit: Number(event.target.value),
                      offset: 0,
                    }))
                  }
                >
                  {pageSizeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
              </label>

              <button
                type="button"
                className={styles.paginationButton}
                disabled={!canGoPrevious}
                onClick={() =>
                  setQuery((current) => ({
                    ...current,
                    offset: Math.max(0, current.offset - current.limit),
                  }))
                }
              >
                Previous
              </button>
              <button
                type="button"
                className={`${styles.paginationButton} ${styles.paginationButtonPrimary}`}
                disabled={!canGoNext}
                onClick={() =>
                  setQuery((current) => ({
                    ...current,
                    offset: current.offset + current.limit,
                  }))
                }
              >
                Next
              </button>
            </div>
          </div>
        </SurfaceCard>
      </div>
    </AppFrame>
  );
}
