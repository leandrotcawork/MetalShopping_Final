import { startTransition, useDeferredValue, useEffect, useMemo, useState } from "react";

import type { ProductsPortfolioApi, ProductsPortfolioQuery, ProductsPortfolioResult, ProductsPortfolioSortKey } from "./api";
import styles from "./ProductsPortfolioPage.module.css";
import {
  defaultProductsPortfolioQuery,
  readProductsPortfolioQueryFromUrl,
  sortIndicator,
  toggleProductsPortfolioSort,
  writeProductsPortfolioQueryToUrl,
} from "./query-state";
import { buildProductsPortfolioSummary } from "./view-model";
import { ProductsFiltersCard } from "./components/ProductsFiltersCard";
import { ProductsHero } from "./components/ProductsHero";
import { ProductsPaginationBar } from "./components/ProductsPaginationBar";
import { ProductsPortfolioTable } from "./components/ProductsPortfolioTable";
import { ProductsSelectionBar } from "./components/ProductsSelectionBar";

const pageSizeOptions = [25, 50, 100];

export function ProductsPortfolioPage(props: { api: ProductsPortfolioApi }) {
  const [query, setQuery] = useState<ProductsPortfolioQuery>(() => readProductsPortfolioQueryFromUrl());
  const [searchDraft, setSearchDraft] = useState(() => readProductsPortfolioQueryFromUrl().search);
  const [result, setResult] = useState<ProductsPortfolioResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedProductIds, setSelectedProductIds] = useState<string[]>([]);
  const [selectionMode, setSelectionMode] = useState<"explicit" | "filtered">("explicit");
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
    writeProductsPortfolioQueryToUrl(query);
  }, [query]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);

      try {
        const next = await props.api.listProductsPortfolio(query);
        if (!cancelled) {
          setResult(next);
        }
      } catch (loadError) {
        if (!cancelled) {
          const message = loadError instanceof Error ? loadError.message : "Falha ao carregar produtos.";
          setError(
            message === "Failed to fetch"
              ? "Falha ao carregar a superfície. Verifique se o backend está ativo."
              : message,
          );
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
  }, [props.api, query]);

  const summary = useMemo(() => buildProductsPortfolioSummary(result), [result]);
  const rows = result?.rows ?? [];
  const brands = result?.filters.brands ?? [];
  const taxonomyLeaf0Names = result?.filters.taxonomy_leaf0_names ?? [];
  const taxonomyLeaf0Label = result?.filters.taxonomy_leaf0_label?.trim() || "Taxonomia";
  const statuses = result?.filters.status ?? [];
  const totalVisible = result?.paging.returned ?? 0;
  const totalMatching = result?.paging.total ?? 0;
  const totalSelected = selectionMode === "filtered" ? totalMatching : selectedProductIds.length;
  const totalRuns = 0;
  const totalPages = totalMatching === 0 ? 1 : Math.max(1, Math.ceil(totalMatching / query.limit));
  const currentPage = Math.floor(query.offset / query.limit) + 1;
  const canGoPrevious = query.offset > 0;
  const canGoNext = query.offset + query.limit < totalMatching;
  const allVisibleSelected =
    selectionMode === "filtered" ||
    (rows.length > 0 && rows.every((row) => selectedProductIds.includes(row.product_id)));

  const activeFilters = [
    query.search.trim() !== "" ? { key: "search", label: `Busca: ${query.search.trim()}` } : null,
    query.brand_name.length > 0 ? { key: "brand_name", label: `Marca: ${query.brand_name.join(", ")}` } : null,
    query.taxonomy_leaf0_name.length > 0
      ? { key: "taxonomy_leaf0_name", label: `${taxonomyLeaf0Label}: ${query.taxonomy_leaf0_name.join(", ")}` }
      : null,
    query.status.length > 0 ? { key: "status", label: `Status: ${query.status.join(", ")}` } : null,
  ].filter((item): item is { key: string; label: string } => item !== null);

  const brandOptions = useMemo(
    () => [{ label: "Todas as marcas", value: "" }, ...brands.map((brand) => ({ label: brand, value: brand }))],
    [brands],
  );
  const taxonomyOptions = useMemo(
    () => [
      { label: `Todos os ${taxonomyLeaf0Label.toLocaleLowerCase("pt-BR")}s`, value: "" },
      ...taxonomyLeaf0Names.map((name) => ({ label: name, value: name })),
    ],
    [taxonomyLeaf0Label, taxonomyLeaf0Names],
  );
  const statusOptions = useMemo(
    () => [{ label: "Todos os status", value: "" }, ...statuses.map((status) => ({ label: status, value: status }))],
    [statuses],
  );

  function clearSelection() {
    setSelectionMode("explicit");
    setSelectedProductIds([]);
  }

  function toggleRowSelection(productId: string) {
    setSelectionMode("explicit");
    setSelectedProductIds((current) =>
      current.includes(productId) ? current.filter((item) => item !== productId) : [...current, productId],
    );
  }

  function toggleCurrentPageSelection() {
    setSelectionMode("explicit");
    setSelectedProductIds((current) => {
      const visibleIds = rows.map((row) => row.product_id);
      const shouldSelect = visibleIds.some((id) => !current.includes(id));
      if (shouldSelect) {
        return Array.from(new Set([...current, ...visibleIds]));
      }
      return current.filter((id) => !visibleIds.includes(id));
    });
  }

  function selectFiltered() {
    setSelectionMode("filtered");
    setSelectedProductIds([]);
  }

  function updateSort(key: ProductsPortfolioSortKey) {
    setQuery((current) => toggleProductsPortfolioSort(current, key));
  }

  return (
    <div className={styles.stack}>
      <ProductsHero
        totalVisible={totalVisible}
        totalSelected={totalSelected}
        totalProducts={summary.totalProducts}
        totalRuns={totalRuns}
        error={error}
      />

      <ProductsFiltersCard
        searchDraft={searchDraft}
        onSearchDraftChange={setSearchDraft}
        query={query}
        taxonomyLeaf0Label={taxonomyLeaf0Label}
        brandOptions={brandOptions}
        taxonomyOptions={taxonomyOptions}
        statusOptions={statusOptions}
        activeFilters={activeFilters}
        onChangeQuery={setQuery}
        onClearAll={() => {
          setSearchDraft("");
          setQuery({
            ...defaultProductsPortfolioQuery,
            limit: query.limit,
          });
        }}
      />

      <ProductsPortfolioTable
        rows={rows}
        loading={loading}
        taxonomyLeaf0Label={taxonomyLeaf0Label}
        sortIndicator={(key) => sortIndicator(query, key)}
        onSort={updateSort}
        allVisibleSelected={allVisibleSelected}
        selectionMode={selectionMode}
        selectedProductIds={selectedProductIds}
        onToggleCurrentPage={toggleCurrentPageSelection}
        onToggleRow={toggleRowSelection}
        actions={
          <ProductsSelectionBar
            rowsCount={rows.length}
            allVisibleSelected={allVisibleSelected}
            totalSelected={totalSelected}
            selectionMode={selectionMode}
            mode="actions"
            onToggleCurrentPage={toggleCurrentPageSelection}
            onSelectFiltered={selectFiltered}
            onClearSelection={clearSelection}
          />
        }
        footer={
          <ProductsPaginationBar
            currentPage={currentPage}
            totalPages={totalPages}
            totalMatching={totalMatching}
            limit={query.limit}
            pageSizeOptions={pageSizeOptions}
            canGoPrevious={canGoPrevious}
            canGoNext={canGoNext}
            onChangeLimit={(limit) =>
              setQuery((current) => ({
                ...current,
                limit,
                offset: 0,
              }))
            }
            onPrevious={() =>
              setQuery((current) => ({
                ...current,
                offset: Math.max(0, current.offset - current.limit),
              }))
            }
            onNext={() =>
              setQuery((current) => ({
                ...current,
                offset: current.offset + current.limit,
              }))
            }
          />
        }
      >
        <ProductsSelectionBar
          rowsCount={rows.length}
          allVisibleSelected={allVisibleSelected}
          totalSelected={totalSelected}
          selectionMode={selectionMode}
          mode="summary"
          onToggleCurrentPage={toggleCurrentPageSelection}
          onSelectFiltered={selectFiltered}
          onClearSelection={clearSelection}
        />
      </ProductsPortfolioTable>
    </div>
  );
}
