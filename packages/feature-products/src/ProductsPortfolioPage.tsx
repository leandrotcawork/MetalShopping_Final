import { startTransition, useDeferredValue, useEffect, useMemo, useState } from "react";

import type {
  ProductsPortfolioApi,
  ProductsPortfolioQuery,
  ProductsPortfolioResult,
  ProductsPortfolioSortKey,
  ProductsShoppingApi,
} from "./api";
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
import { Button, FilterDropdown, StatusBanner, type SelectMenuOption } from "@metalshopping/ui";
import type { ShoppingBootstrapV1, ShoppingMarketReportExportXlsxResponseV1, ShoppingRunV1 } from "@metalshopping/sdk-types";

const pageSizeOptions = [25, 50, 100];
const exportBatchLimit = 250;
const exportMaxRows = 5000;
const exportDraftStorageKey = "ms.products.marketReportExport.v1";
const allSuppliersValue = "all";

type ExportDraft = {
  runId: string;
  supplierCodes: string[];
  outputFilePath: string;
};

function readExportDraft(): ExportDraft {
  if (typeof window === "undefined") {
    return { runId: "", supplierCodes: [], outputFilePath: "" };
  }
  try {
    const raw = window.sessionStorage.getItem(exportDraftStorageKey);
    if (!raw) {
      return { runId: "", supplierCodes: [], outputFilePath: "" };
    }
    const parsed = JSON.parse(raw) as ExportDraft;
    return {
      runId: typeof parsed.runId === "string" ? parsed.runId : "",
      supplierCodes: Array.isArray(parsed.supplierCodes)
        ? parsed.supplierCodes.filter((value) => typeof value === "string")
        : [],
      outputFilePath: typeof parsed.outputFilePath === "string" ? parsed.outputFilePath : "",
    };
  } catch {
    return { runId: "", supplierCodes: [], outputFilePath: "" };
  }
}

function writeExportDraft(next: ExportDraft) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.sessionStorage.setItem(exportDraftStorageKey, JSON.stringify(next));
  } catch {
  }
}

function toggleMultiSelection(current: string[], next: string): string[] {
  if (next === allSuppliersValue) {
    return [];
  }
  if (current.includes(next)) {
    return current.filter((value) => value !== next);
  }
  return [...current, next];
}

export function ProductsPortfolioPage(props: { api: ProductsPortfolioApi; shoppingApi: ProductsShoppingApi }) {
  const [query, setQuery] = useState<ProductsPortfolioQuery>(() => readProductsPortfolioQueryFromUrl());
  const [searchDraft, setSearchDraft] = useState(() => readProductsPortfolioQueryFromUrl().search);
  const [result, setResult] = useState<ProductsPortfolioResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedProductIds, setSelectedProductIds] = useState<string[]>([]);
  const [selectionMode, setSelectionMode] = useState<"explicit" | "filtered">("explicit");
  const deferredSearch = useDeferredValue(searchDraft);
  const [exportModalOpen, setExportModalOpen] = useState(false);
  const [exportRunId, setExportRunId] = useState(() => readExportDraft().runId);
  const [exportSupplierCodes, setExportSupplierCodes] = useState(() => readExportDraft().supplierCodes);
  const [exportOutputPath, setExportOutputPath] = useState(() => readExportDraft().outputFilePath);
  const [exportConfigLoading, setExportConfigLoading] = useState(false);
  const [exportConfigError, setExportConfigError] = useState<string | null>(null);
  const [exportRuns, setExportRuns] = useState<ShoppingRunV1[]>([]);
  const [exportSuppliers, setExportSuppliers] = useState<ShoppingBootstrapV1["suppliers"]>([]);
  const [exporting, setExporting] = useState(false);
  const [exportStatus, setExportStatus] = useState<{ tone: "success" | "error"; message: string } | null>(null);
  const [exportResult, setExportResult] = useState<ShoppingMarketReportExportXlsxResponseV1 | null>(null);

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

  useEffect(() => {
    writeExportDraft({
      runId: exportRunId,
      supplierCodes: exportSupplierCodes,
      outputFilePath: exportOutputPath,
    });
  }, [exportRunId, exportSupplierCodes, exportOutputPath]);

  useEffect(() => {
    if (!exportModalOpen) {
      return;
    }
    let cancelled = false;

    async function loadExportConfig() {
      setExportConfigLoading(true);
      setExportConfigError(null);
      try {
        const [bootstrap, runs] = await Promise.all([
          props.shoppingApi.getBootstrap(),
          props.shoppingApi.listRuns({ limit: 50, offset: 0 }),
        ]);
        if (cancelled) {
          return;
        }
        setExportSuppliers(bootstrap.suppliers);
        setExportRuns(runs.rows);
        if (!exportRunId && runs.rows.length > 0) {
          setExportRunId(runs.rows[0].runId);
        }
      } catch (loadError) {
        if (!cancelled) {
          const message = loadError instanceof Error ? loadError.message : "Falha ao carregar configuracoes do relatorio.";
          setExportConfigError(message);
        }
      } finally {
        if (!cancelled) {
          setExportConfigLoading(false);
        }
      }
    }

    void loadExportConfig();
    return () => {
      cancelled = true;
    };
  }, [props.shoppingApi, exportModalOpen, exportRunId]);

  useEffect(() => {
    if (!exportModalOpen) {
      return;
    }
    setExportResult(null);
    setExportStatus(null);
  }, [exportRunId, exportSupplierCodes, exportOutputPath, exportModalOpen]);

  const summary = useMemo(() => buildProductsPortfolioSummary(result), [result]);
  const rows = result?.rows ?? [];
  const brands = result?.filters.brands ?? [];
  const taxonomyLeaf0Names = result?.filters.taxonomy_leaf0_names ?? [];
  const taxonomyLeaf0Label = result?.filters.taxonomy_leaf0_label?.trim() || "Taxonomia";
  const statuses = result?.filters.status ?? [];
  const totalVisible = result?.paging.returned ?? 0;
  const totalMatching = result?.paging.total ?? 0;
  const totalSelected = selectionMode === "filtered" ? totalMatching : selectedProductIds.length;
  const totalRuns = exportRuns.length;
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
  const exportRunOptions = useMemo<SelectMenuOption[]>(
    () =>
      exportRuns.map((run) => ({
        value: run.runId,
        label: `${run.runId} \u2022 ${run.status}`,
      })),
    [exportRuns],
  );
  const exportSupplierOptions = useMemo<SelectMenuOption[]>(
    () =>
      exportSuppliers.map((supplier) => ({
        value: supplier.supplierCode,
        label: supplier.supplierLabel || supplier.supplierCode,
      })),
    [exportSuppliers],
  );
  const selectedSuppliersLabel =
    exportSupplierOptions.length === 0 ? "--" : exportSupplierCodes.length === 0 ? "Todos" : `${exportSupplierCodes.length}`;
  const exportSelectionCount = totalSelected;
  const exportOverLimit = exportSelectionCount > exportMaxRows;
  const exportDisabled = exportSelectionCount === 0 || exporting;

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

  function openExportModal() {
    setExportModalOpen(true);
  }

  function closeExportModal() {
    setExportModalOpen(false);
  }

  async function resolveExportProductIds() {
    if (selectionMode === "explicit") {
      return Array.from(new Set(selectedProductIds));
    }
    if (!result) {
      return [];
    }

    const targetTotal = Math.min(totalMatching, exportMaxRows);
    const collectedIds: string[] = [];
    let offset = 0;

    while (offset < targetTotal) {
      const response = await props.api.listProductsPortfolio({
        search: query.search.trim() || undefined,
        brand_name: query.brand_name.length > 0 ? query.brand_name : undefined,
        taxonomy_leaf0_name: query.taxonomy_leaf0_name.length > 0 ? query.taxonomy_leaf0_name : undefined,
        status: query.status.length > 0 ? query.status : undefined,
        sort_key: query.sort_key,
        sort_direction: query.sort_direction,
        limit: exportBatchLimit,
        offset,
      });
      collectedIds.push(...response.rows.map((row) => row.product_id));
      if (response.rows.length === 0) {
        break;
      }
      offset += response.rows.length;
    }

    return Array.from(new Set(collectedIds)).slice(0, targetTotal);
  }

  async function handleExport() {
    if (exporting) {
      return;
    }
    setExportModalOpen(true);
    setExportStatus(null);
    setExportResult(null);

    if (exportSelectionCount === 0) {
      setExportStatus({ tone: "error", message: "Selecione produtos para exportar o relatÃ³rio." });
      return;
    }
    if (exportOverLimit) {
      setExportStatus({
        tone: "error",
        message: `SeleÃ§Ã£o acima de ${exportMaxRows} itens. Ajuste os filtros antes de exportar.`,
      });
      return;
    }
    if (selectionMode === "filtered" && !result) {
      setExportStatus({ tone: "error", message: "Aguarde o carregamento da lista antes de exportar." });
      return;
    }

    const runId = exportRunId.trim();
    if (!runId) {
      setExportStatus({ tone: "error", message: "Selecione uma run para exportar o relatÃ³rio." });
      return;
    }
    const outputFilePath = exportOutputPath.trim();
    if (!outputFilePath) {
      setExportStatus({ tone: "error", message: "Informe o caminho de destino do arquivo XLSX." });
      return;
    }
    const supplierCodes =
      exportSupplierCodes.length > 0
        ? exportSupplierCodes
        : exportSuppliers.map((supplier) => supplier.supplierCode);
    if (supplierCodes.length === 0) {
      setExportStatus({ tone: "error", message: "Selecione ao menos um fornecedor para exportar." });
      return;
    }

    setExporting(true);
    try {
      const productIds = await resolveExportProductIds();
      if (productIds.length === 0) {
        setExportStatus({ tone: "error", message: "Nenhum produto encontrado para exportaÃ§Ã£o." });
        return;
      }
      const response = await props.shoppingApi.exportMarketReportXlsx(runId, {
        productIds,
        supplierCodes,
        outputFilePath,
      });
      setExportResult(response);
      setExportStatus({ tone: "success", message: `RelatÃ³rio exportado em ${response.outputFilePath}.` });
    } catch (exportError) {
      const message = exportError instanceof Error ? exportError.message : "Falha ao exportar relatÃ³rio.";
      setExportStatus({ tone: "error", message });
    } finally {
      setExporting(false);
    }
  }

  return (
    <div className={styles.stack}>
      <ProductsHero
        totalVisible={totalVisible}
        totalSelected={totalSelected}
        totalProducts={summary.totalProducts}
        totalRuns={totalRuns}
        error={error}
        exportDisabled={exportDisabled}
        exportStatus={exportStatus}
        onConfigureReport={openExportModal}
        onExportReport={() => void handleExport()}
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
            selectedSuppliersLabel={selectedSuppliersLabel}
            selectionMode={selectionMode}
            mode="actions"
            exportDisabled={exportDisabled}
            onExport={() => void handleExport()}
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
          selectedSuppliersLabel={selectedSuppliersLabel}
          selectionMode={selectionMode}
          mode="summary"
          exportDisabled={exportDisabled}
          onExport={() => void handleExport()}
          onToggleCurrentPage={toggleCurrentPageSelection}
          onSelectFiltered={selectFiltered}
          onClearSelection={clearSelection}
        />
      </ProductsPortfolioTable>

      {exportModalOpen ? (
        <div className={styles.exportBackdrop} role="presentation" onClick={closeExportModal}>
          <div className={styles.exportModal} role="dialog" aria-modal="true" onClick={(event) => event.stopPropagation()}>
            <div className={styles.exportHeader}>
              <div>
                <h3 className={styles.exportTitle}>Configurar relatÃ³rio de mercado</h3>
                <p className={styles.exportSubtitle}>
                  Escolha a run, fornecedores e o caminho do XLSX para exportar o comparativo de preÃ§os.
                </p>
              </div>
              <Button className={styles.secondaryActionButton} variant="secondary" onClick={closeExportModal}>
                Fechar
              </Button>
            </div>

            <div className={styles.exportSummary}>
              <span>
                Origem: <strong>{selectionMode === "filtered" ? "Filtros atuais" : "SeleÃ§Ã£o manual"}</strong>
              </span>
              <span>
                Produtos: <strong>{exportSelectionCount}</strong>
              </span>
              <span>
                Fornecedores: <strong>{selectedSuppliersLabel}</strong>
              </span>
            </div>

            {exportConfigLoading ? <p className={styles.exportMeta}>Carregando configuracoes...</p> : null}
            {exportConfigError ? (
              <StatusBanner className={styles.exportStatus} tone="error">
                {exportConfigError}
              </StatusBanner>
            ) : null}
            {exportOverLimit ? (
              <StatusBanner className={styles.exportStatus} tone="error">
                SelecÃ£o acima de {exportMaxRows} itens. Ajuste os filtros antes de exportar.
              </StatusBanner>
            ) : null}
            {exportStatus ? (
              <StatusBanner className={styles.exportStatus} tone={exportStatus.tone}>
                {exportStatus.message}
              </StatusBanner>
            ) : null}

            <div className={styles.exportGrid}>
              <label className={styles.exportField}>
                Run
                <FilterDropdown
                  id="products-export-run"
                  options={exportRunOptions}
                  value={exportRunId}
                  onSelect={setExportRunId}
                  disabled={exportConfigLoading || exportRunOptions.length === 0}
                />
              </label>
              <label className={styles.exportField}>
                Fornecedores
                <FilterDropdown
                  id="products-export-suppliers"
                  options={exportSupplierOptions}
                  values={exportSupplierCodes}
                  selectionMode="duo"
                  allLabel="Todos fornecedores"
                  onSelect={(value) => setExportSupplierCodes((current) => toggleMultiSelection(current, value))}
                  disabled={exportConfigLoading || exportSupplierOptions.length === 0}
                />
              </label>
              <label className={styles.exportField}>
                Caminho do arquivo XLSX
                <input
                  className={styles.exportInput}
                  type="text"
                  value={exportOutputPath}
                  onChange={(event) => setExportOutputPath(event.target.value)}
                  placeholder="C:\\Users\\leandro.theodoro\\Documents\\Export\\market-report.xlsx"
                />
              </label>
            </div>

            <div className={styles.exportFooter}>
              <span className={styles.exportMeta}>
                {exportResult
                  ? `Exportado em ${exportResult.outputFilePath} (${exportResult.totalProducts} produtos).`
                  : "O arquivo serÃ¡ gravado no servidor conforme o caminho informado."}
              </span>
              <div className={styles.exportActions}>
                <Button className={styles.secondaryActionButton} variant="secondary" onClick={closeExportModal}>
                  Cancelar
                </Button>
                <Button
                  className={styles.actionButtonPrimary}
                  variant="primary"
                  disabled={exporting || exportConfigLoading}
                  onClick={() => void handleExport()}
                >
                  {exporting ? "Exportando..." : "Exportar XLSX"}
                </Button>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
