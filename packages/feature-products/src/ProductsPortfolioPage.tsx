import { startTransition, useDeferredValue, useEffect, useMemo, useState } from "react";

import {
  AppFrame,
  FilterDropdown,
  type SelectMenuOption,
  SurfaceCard,
} from "@metalshopping/ui";

import { listProductsPortfolio, type ProductsPortfolioItem, type ProductsPortfolioQuery, type ProductsPortfolioResult } from "./api";
import styles from "./ProductsPortfolioPage.module.css";
import { buildProductsPortfolioSummary, formatCurrency, formatQuantity } from "./view-model";

const defaultQuery: ProductsPortfolioQuery = {
  search: "",
  brandName: "",
  taxonomyLeaf0Name: "",
  status: "",
  limit: 50,
  offset: 0,
};

const pageSizeOptions = [25, 50, 100];

type SortKey =
  | "pn_interno"
  | "name"
  | "brand_name"
  | "taxonomy_leaf0_name"
  | "product_status"
  | "current_price_amount"
  | "replacement_cost_amount"
  | "on_hand_quantity";

type SortState = {
  key: SortKey;
  direction: "asc" | "desc";
};

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

function normalizeText(value: string | null | undefined) {
  return (value ?? "").trim();
}

function compareNullableStrings(left: string | null | undefined, right: string | null | undefined) {
  return normalizeText(left).localeCompare(normalizeText(right), "pt-BR", {
    sensitivity: "base",
  });
}

function compareNullableNumbers(left: number | null | undefined, right: number | null | undefined) {
  return (left ?? Number.NEGATIVE_INFINITY) - (right ?? Number.NEGATIVE_INFINITY);
}

function sortRows(rows: ProductsPortfolioItem[], sort: SortState) {
  const next = [...rows];
  next.sort((left, right) => {
    let comparison = 0;

    switch (sort.key) {
      case "pn_interno":
        comparison = compareNullableStrings(left.pn_interno, right.pn_interno);
        break;
      case "name":
        comparison = compareNullableStrings(left.name, right.name);
        break;
      case "brand_name":
        comparison = compareNullableStrings(left.brand_name, right.brand_name);
        break;
      case "taxonomy_leaf0_name":
        comparison = compareNullableStrings(left.taxonomy_leaf0_name, right.taxonomy_leaf0_name);
        break;
      case "product_status":
        comparison = compareNullableStrings(left.product_status, right.product_status);
        break;
      case "current_price_amount":
        comparison = compareNullableNumbers(left.current_price_amount, right.current_price_amount);
        break;
      case "replacement_cost_amount":
        comparison = compareNullableNumbers(left.replacement_cost_amount, right.replacement_cost_amount);
        break;
      case "on_hand_quantity":
        comparison = compareNullableNumbers(left.on_hand_quantity, right.on_hand_quantity);
        break;
      default:
        comparison = 0;
        break;
    }

    if (comparison === 0) {
      comparison = compareNullableStrings(left.name, right.name);
    }

    return sort.direction === "asc" ? comparison : -comparison;
  });
  return next;
}

function statusTone(value: string) {
  const normalized = value.trim().toLowerCase();
  if (normalized === "active" || normalized === "ativo") {
    return "success";
  }
  if (normalized === "inactive" || normalized === "inativo") {
    return "muted";
  }
  return "neutral";
}

function statusLabel(value: string) {
  const normalized = value.trim().toLowerCase();
  if (normalized === "active") {
    return "Ativo";
  }
  if (normalized === "inactive") {
    return "Inativo";
  }
  return value;
}

function sortIndicator(sort: SortState, key: SortKey) {
  if (sort.key !== key) {
    return "↕";
  }
  return sort.direction === "asc" ? "↑" : "↓";
}

function ProductSelectionCheckbox(props: {
  checked: boolean;
  disabled?: boolean;
  onChange: () => void;
  label: string;
}) {
  return (
    <label className={styles.check} aria-label={props.label}>
      <input checked={props.checked} disabled={props.disabled} type="checkbox" onChange={props.onChange} />
      <span className={styles.checkUi} aria-hidden="true" />
    </label>
  );
}

export function ProductsPortfolioPage() {
  const [query, setQuery] = useState<ProductsPortfolioQuery>(() => readQueryFromUrl());
  const [searchDraft, setSearchDraft] = useState(() => readQueryFromUrl().search);
  const [result, setResult] = useState<ProductsPortfolioResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedProductIds, setSelectedProductIds] = useState<string[]>([]);
  const [selectionMode, setSelectionMode] = useState<"explicit" | "filtered">("explicit");
  const [sort, setSort] = useState<SortState>({ key: "pn_interno", direction: "asc" });
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
          setError(loadError instanceof Error ? loadError.message : "Falha ao carregar produtos.");
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

  const rows = useMemo(() => sortRows(result?.rows ?? [], sort), [result?.rows, sort]);
  const brands = result?.filters.brands ?? [];
  const taxonomyLeaf0Names = result?.filters.taxonomy_leaf0_names ?? [];
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
    query.brandName.trim() !== "" ? { key: "brandName", label: `Marca: ${query.brandName.trim()}` } : null,
    query.taxonomyLeaf0Name.trim() !== ""
      ? { key: "taxonomyLeaf0Name", label: `Grupo: ${query.taxonomyLeaf0Name.trim()}` }
      : null,
    query.status.trim() !== "" ? { key: "status", label: `Status: ${query.status.trim()}` } : null,
  ].filter((item): item is { key: string; label: string } => item !== null);

  const brandOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas as marcas", value: "" }, ...brands.map((brand) => ({ label: brand, value: brand }))],
    [brands],
  );
  const taxonomyOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "Todos os grupos", value: "" },
      ...taxonomyLeaf0Names.map((name) => ({ label: name, value: name })),
    ],
    [taxonomyLeaf0Names],
  );
  const statusOptions = useMemo<SelectMenuOption[]>(
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

  function updateSort(key: SortKey) {
    setSort((current) =>
      current.key === key
        ? {
            key,
            direction: current.direction === "asc" ? "desc" : "asc",
          }
        : {
            key,
            direction: "asc",
          },
    );
  }

  return (
    <AppFrame
      eyebrow="Products · Market Report"
      title="Relatório de preço de mercado por run"
      subtitle="Selecione produtos por filtros e exporte um XLSX comparativo com preço interno versus concorrentes."
      aside={
        <div className={styles.heroAside}>
          <div className={styles.heroChip}>
            <small>Na grade</small>
            <strong>{totalVisible}</strong>
          </div>
          <div className={styles.heroChip}>
            <small>Selecionados</small>
            <strong>{totalSelected}</strong>
          </div>
          <div className={styles.heroChip}>
            <small>Total base</small>
            <strong>{summary.totalProducts}</strong>
          </div>
          <div className={styles.heroChip}>
            <small>Runs</small>
            <strong>{totalRuns}</strong>
          </div>
          <div className={styles.heroActions}>
            <button type="button" className={styles.actionButton}>
              ⚙ Configurar relatório
            </button>
            <button type="button" className={`${styles.actionButton} ${styles.actionButtonPrimary}`} disabled>
              📤 Exportar relatório
            </button>
          </div>
          <div className={`${styles.heroStatusBanner} ${error ? styles.statusBannerError : styles.statusBannerSuccess}`}>
            <span>
              {error
                ? error
                : "Superfície operacional ativa para Catalog, Pricing e Inventory."}
            </span>
          </div>
        </div>
      }
    >
      <div className={styles.stack}>
        <SurfaceCard
          title="Filtros de Produtos"
          actions={<span className={styles.filterLogic}>Combinação lógica: AND entre filtros e OR dentro de cada multi-seleção.</span>}
          className={styles.filtersCard}
        >
          <div className={styles.toolbar}>
            <label className={styles.field}>
              <span className={styles.label}>Busca</span>
              <input
                className={styles.input}
                value={searchDraft}
                placeholder="PN, referência, EAN, descrição"
                onChange={(event) => setSearchDraft(event.target.value)}
              />
            </label>

            <div className={styles.field}>
              <span className={styles.label}>Marca</span>
              <FilterDropdown
                id="products-brand-filter"
                value={query.brandName}
                options={brandOptions}
                onSelect={(value) =>
                  setQuery((current) => ({
                    ...current,
                    brandName: value,
                    offset: 0,
                  }))
                }
              />
            </div>

            <div className={styles.field}>
              <span className={styles.label}>Status</span>
              <FilterDropdown
                id="products-status-filter"
                value={query.status}
                options={statusOptions}
                onSelect={(value) =>
                  setQuery((current) => ({
                    ...current,
                    status: value,
                    offset: 0,
                  }))
                }
              />
            </div>

            <div className={styles.field}>
              <span className={styles.label}>Grupo</span>
              <FilterDropdown
                id="products-taxonomy-filter"
                value={query.taxonomyLeaf0Name}
                options={taxonomyOptions}
                onSelect={(value) =>
                  setQuery((current) => ({
                    ...current,
                    taxonomyLeaf0Name: value,
                    offset: 0,
                  }))
                }
              />
            </div>
          </div>

          <div className={styles.filterFooter}>
            <div className={styles.filterChips}>
              {activeFilters.length > 0 ? (
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
                    <span aria-hidden="true">×</span>
                  </button>
                ))
              ) : (
                <span className={styles.filterHint}>Nenhum filtro ativo. Exibindo o portfólio completo visível no tenant.</span>
              )}
            </div>

            <button
              type="button"
              className={styles.clearButton}
              disabled={activeFilters.length === 0}
              onClick={() => {
                setSearchDraft("");
                setQuery({
                  ...defaultQuery,
                  limit: query.limit,
                });
              }}
            >
              Limpar filtros
            </button>
          </div>
        </SurfaceCard>

        <SurfaceCard
          title="Produtos Cadastrados"
          actions={
            <span className={styles.tableMeta}>
              {loading ? "Atualizando..." : `Mostrando ${result?.paging.returned ?? 0} de ${result?.paging.total ?? 0}`}
            </span>
          }
          className={styles.tableCard}
        >
          <div className={styles.tableActions}>
            <button type="button" className={styles.secondaryActionButton} disabled>
              Exportar selecionados
            </button>
            <button type="button" className={styles.secondaryActionButton} disabled={rows.length === 0} onClick={toggleCurrentPageSelection}>
              {allVisibleSelected ? "Desmarcar página" : "Selecionar página"}
            </button>
            <button type="button" className={styles.secondaryActionButton} disabled={rows.length === 0} onClick={selectFiltered}>
              Selecionar filtrados
            </button>
            <button type="button" className={styles.secondaryActionButton} disabled={totalSelected === 0} onClick={clearSelection}>
              Limpar
            </button>
          </div>

          <div className={styles.selectionRow}>
            <span>
              Modo: <strong>{selectionMode === "filtered" ? "Filtrados" : "Explícito"}</strong>
            </span>
            <span>
              Itens: <strong>{totalSelected}</strong>
            </span>
            <span>
              Preço médio visível: <strong>{summary.averageVisiblePrice}</strong>
            </span>
          </div>

          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th className={styles.checkboxColumn}>
                    <ProductSelectionCheckbox
                      checked={allVisibleSelected}
                      disabled={rows.length === 0 || selectionMode === "filtered"}
                      label="Selecionar produtos da página"
                      onChange={toggleCurrentPageSelection}
                    />
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("pn_interno")}>
                      <span>PN</span>
                      <span aria-hidden="true">{sortIndicator(sort, "pn_interno")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("name")}>
                      <span>Produto</span>
                      <span aria-hidden="true">{sortIndicator(sort, "name")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("brand_name")}>
                      <span>Marca</span>
                      <span aria-hidden="true">{sortIndicator(sort, "brand_name")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("taxonomy_leaf0_name")}>
                      <span>Grupo</span>
                      <span aria-hidden="true">{sortIndicator(sort, "taxonomy_leaf0_name")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("product_status")}>
                      <span>Status</span>
                      <span aria-hidden="true">{sortIndicator(sort, "product_status")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("current_price_amount")}>
                      <span>Nosso Preço</span>
                      <span aria-hidden="true">{sortIndicator(sort, "current_price_amount")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("replacement_cost_amount")}>
                      <span>Custos</span>
                      <span aria-hidden="true">{sortIndicator(sort, "replacement_cost_amount")}</span>
                    </button>
                  </th>
                  <th>
                    <button type="button" className={styles.sortButton} onClick={() => updateSort("on_hand_quantity")}>
                      <span>Estoque</span>
                      <span aria-hidden="true">{sortIndicator(sort, "on_hand_quantity")}</span>
                    </button>
                  </th>
                </tr>
              </thead>
              <tbody>
                {rows.length > 0 ? (
                  rows.map((row) => {
                    const checked = selectionMode === "filtered" || selectedProductIds.includes(row.product_id);
                    return (
                      <tr key={row.product_id}>
                        <td className={styles.checkboxColumn}>
                          <ProductSelectionCheckbox
                            checked={checked}
                            disabled={selectionMode === "filtered"}
                            label={`Selecionar produto ${row.name}`}
                            onChange={() => toggleRowSelection(row.product_id)}
                          />
                        </td>
                        <td>
                          <span className={styles.cellStrong}>{row.pn_interno ?? "—"}</span>
                          <span className={styles.cellMeta}>{row.sku}</span>
                        </td>
                        <td>
                          <span className={styles.cellStrong}>{row.name}</span>
                          <span className={styles.cellMeta}>{row.description ?? "Sem descrição cadastrada."}</span>
                          <span className={styles.cellSmall}>
                            Ref: {row.reference ?? "—"} · EAN: {row.ean ?? "—"}
                          </span>
                        </td>
                        <td>{row.brand_name ?? "—"}</td>
                        <td>
                          <span className={styles.cellStrong}>{row.taxonomy_leaf0_name ?? "—"}</span>
                          <span className={styles.cellMeta}>{row.taxonomy_leaf_name ?? "Sem folha definida."}</span>
                        </td>
                        <td>
                          <span className={`${styles.statusChip} ${styles[`statusChip${statusTone(row.product_status)}`]}`}>
                            {statusLabel(row.product_status)}
                          </span>
                        </td>
                        <td className={styles.moneyCell}>
                          <span className={styles.cellStrong}>{formatCurrency(row.current_price_amount)}</span>
                          <span className={styles.cellMeta}>{row.currency_code ?? "BRL"}</span>
                        </td>
                        <td>
                          <span className={styles.cellStrong}>Reposição: {formatCurrency(row.replacement_cost_amount)}</span>
                          <span className={styles.cellMeta}>Médio: {formatCurrency(row.average_cost_amount)}</span>
                        </td>
                        <td>
                          <span className={styles.cellStrong}>{formatQuantity(row.on_hand_quantity)}</span>
                          <span className={styles.cellMeta}>{row.inventory_position_status ?? "Sem posição"}</span>
                        </td>
                      </tr>
                    );
                  })
                ) : (
                  <tr>
                    <td className={styles.empty} colSpan={9}>
                      {loading ? "Carregando produtos..." : "Nenhum produto encontrado para o filtro atual."}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className={styles.paginationRow}>
            <div className={styles.paginationStatus}>
              <span>
                Página {currentPage} de {totalPages}
              </span>
              <span>{totalMatching} produtos encontrados</span>
            </div>

            <div className={styles.paginationActions}>
              <label className={styles.pageSizeField}>
                <span>Linhas</span>
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
                Anterior
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
                Próxima
              </button>
            </div>
          </div>
        </SurfaceCard>
      </div>
    </AppFrame>
  );
}
