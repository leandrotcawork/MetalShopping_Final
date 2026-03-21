import { useEffect, useMemo, useState } from "react";

import type { ServerCoreSdk, ShoppingRunStatus } from "@metalshopping/sdk-runtime";
import type {
  ProductsPortfolioItemV1,
  ProductsPortfolioListV1,
  ShoppingBootstrapV1,
  ShoppingCreateRunRequestV1,
  ShoppingRunRequestV1,
  ShoppingRunV1,
  ShoppingManualUrlCandidateV1,
} from "@metalshopping/sdk-types";
import { AppFrame, Checkbox, FilterDropdown, type SelectMenuOption } from "@metalshopping/ui";

import styles from "./ShoppingPage.module.css";

type ShoppingPageProps = {
  shoppingApi: ServerCoreSdk["shopping"];
  productsApi: ServerCoreSdk["products"];
};

type WizardStep = 1 | 2 | 3;
type InputMode = "xlsx" | "catalog";

const catalogPageLimit = 30;
const manualPageLimit = 10;
const allOptionValue = "all";

const statusOptions: Array<{ value: ShoppingRunStatus | "all"; label: string }> = [
  { value: "all", label: "Todos" },
  { value: "queued", label: "Queued" },
  { value: "running", label: "Running" },
  { value: "completed", label: "Completed" },
  { value: "failed", label: "Failed" },
];

function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return "--";
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) {
    return value;
  }
  return parsed.toLocaleString("pt-BR");
}

function formatMoney(value: number | null | undefined, currencyCode: string | null | undefined) {
  if (value === null || value === undefined || Number.isNaN(value)) {
    return "--";
  }
  const currency = (currencyCode || "BRL").trim() || "BRL";
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency }).format(value);
}

function statusClass(stylesheet: Record<string, string>, status: string) {
  switch (status) {
    case "completed":
      return stylesheet.statusCompleted;
    case "failed":
      return stylesheet.statusFailed;
    case "running":
      return stylesheet.statusRunning;
    case "queued":
      return stylesheet.statusQueued;
    default:
      return "";
  }
}

function toSelectOptions(items: string[], allLabel: string): SelectMenuOption[] {
  return [{ value: allOptionValue, label: allLabel }, ...items.map((item) => ({ value: item, label: item }))];
}

function StepPill(props: { step: WizardStep; activeStep: WizardStep; label: string; onClick: () => void }) {
  const isCompleted = props.activeStep > props.step;
  const isActive = props.activeStep === props.step;
  return (
    <button
      type="button"
      className={`${styles.step} ${isCompleted ? styles.stepCompleted : ""} ${isActive ? styles.stepActive : ""}`.trim()}
      onClick={props.onClick}
    >
      <span className={styles.stepNumber}>{isCompleted ? "✓" : props.step}</span>
      {props.label}
    </button>
  );
}

export function ShoppingPage({ shoppingApi, productsApi }: ShoppingPageProps) {
  const [step, setStep] = useState<WizardStep>(1);
  const [inputMode, setInputMode] = useState<InputMode>("xlsx");
  const [selectedStatus, setSelectedStatus] = useState<ShoppingRunStatus | "all">("all");
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [showLog, setShowLog] = useState(false);
  const [showManualUrlPanel, setShowManualUrlPanel] = useState(false);
  const [reloadTick, setReloadTick] = useState(0);

  const [summary, setSummary] = useState<Awaited<ReturnType<ServerCoreSdk["shopping"]["getSummary"]>> | null>(null);
  const [bootstrap, setBootstrap] = useState<ShoppingBootstrapV1 | null>(null);
  const [runs, setRuns] = useState<ShoppingRunV1[]>([]);
  const [selectedRun, setSelectedRun] = useState<ShoppingRunV1 | null>(null);
  const [loadingShopping, setLoadingShopping] = useState(true);
  const [creatingRun, setCreatingRun] = useState(false);
  const [createRunInfo, setCreateRunInfo] = useState<string | null>(null);
  const [createdRunRequestId, setCreatedRunRequestId] = useState<string | null>(null);
  const [runRequest, setRunRequest] = useState<ShoppingRunRequestV1 | null>(null);
  const [xlsxFilePath, setXlsxFilePath] = useState("");
  const [xlsxScopeText, setXlsxScopeText] = useState("");
  const [showUploadAdvanced, setShowUploadAdvanced] = useState(false);
  const [xlsxSelectedName, setXlsxSelectedName] = useState("");
  const [supplierCodes, setSupplierCodes] = useState<string[]>([]);
  const [advancedTimeout, setAdvancedTimeout] = useState(60);
  const [advancedHttpWorkers, setAdvancedHttpWorkers] = useState(10);
  const [advancedPlaywrightWorkers, setAdvancedPlaywrightWorkers] = useState(7);
  const [advancedTopN, setAdvancedTopN] = useState(5);
  const [manualSignalSaving, setManualSignalSaving] = useState(false);
  const [manualCandidates, setManualCandidates] = useState<ShoppingManualUrlCandidateV1[]>([]);
  const [manualSearch, setManualSearch] = useState("");
  const [manualSupplierCode, setManualSupplierCode] = useState(allOptionValue);
  const [manualBrand, setManualBrand] = useState(allOptionValue);
  const [manualTaxonomy, setManualTaxonomy] = useState(allOptionValue);
  const [manualShowExisting, setManualShowExisting] = useState(true);
  const [manualOffset, setManualOffset] = useState(0);
  const [manualTotal, setManualTotal] = useState(0);
  const [manualReturned, setManualReturned] = useState(0);
  const [manualEditUrls, setManualEditUrls] = useState<Record<string, string>>({});
  const [manualLoading, setManualLoading] = useState(false);
  const [manualLoadError, setManualLoadError] = useState<string | null>(null);

  const [catalogSearch, setCatalogSearch] = useState("");
  const [catalogBrand, setCatalogBrand] = useState(allOptionValue);
  const [catalogLeaf0, setCatalogLeaf0] = useState(allOptionValue);
  const [catalogStatus, setCatalogStatus] = useState(allOptionValue);
  const [catalogOffset, setCatalogOffset] = useState(0);
  const [catalogRows, setCatalogRows] = useState<ProductsPortfolioItemV1[]>([]);
  const [catalogTotal, setCatalogTotal] = useState(0);
  const [catalogReturned, setCatalogReturned] = useState(0);
  const [catalogLoading, setCatalogLoading] = useState(false);
  const [selectedProductIds, setSelectedProductIds] = useState<string[]>([]);
  const [catalogLeaf0Label, setCatalogLeaf0Label] = useState("Grupo");
  const [catalogBrandOptions, setCatalogBrandOptions] = useState<SelectMenuOption[]>([
    { value: allOptionValue, label: "Todas as marcas" },
  ]);
  const [catalogLeaf0Options, setCatalogLeaf0Options] = useState<SelectMenuOption[]>([
    { value: allOptionValue, label: "Todos os grupos" },
  ]);
  const [catalogStatusOptions, setCatalogStatusOptions] = useState<SelectMenuOption[]>([
    { value: allOptionValue, label: "Todos os status" },
  ]);

  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadShoppingSurface() {
      setLoadingShopping(true);
      setError(null);
      try {
        const [nextSummary, nextBootstrap, nextRuns] = await Promise.all([
          shoppingApi.getSummary(),
          shoppingApi.getBootstrap(),
          shoppingApi.listRuns(selectedStatus === "all" ? {} : { status: selectedStatus, limit: 20, offset: 0 }),
        ]);
        if (cancelled) {
          return;
        }
        setSummary(nextSummary);
        setBootstrap(nextBootstrap);
        setRuns(nextRuns.rows);

        setAdvancedTimeout(nextBootstrap.advancedDefaults.timeoutSeconds);
        setAdvancedHttpWorkers(nextBootstrap.advancedDefaults.httpWorkers);
        setAdvancedPlaywrightWorkers(nextBootstrap.advancedDefaults.playwrightWorkers);
        setAdvancedTopN(nextBootstrap.advancedDefaults.topN);

        if (nextRuns.rows.length > 0) {
          const firstRun = await shoppingApi.getRun(nextRuns.rows[0].runId);
          if (!cancelled) {
            setSelectedRun(firstRun);
          }
        } else {
          setSelectedRun(null);
        }
      } catch (loadError) {
        if (!cancelled) {
          const message = loadError instanceof Error ? loadError.message : "Falha ao carregar Shopping.";
          setError(message);
        }
      } finally {
        if (!cancelled) {
          setLoadingShopping(false);
        }
      }
    }

    void loadShoppingSurface();
    return () => {
      cancelled = true;
    };
  }, [shoppingApi, selectedStatus, reloadTick]);

  useEffect(() => {
    let cancelled = false;

    async function loadCandidates() {
      const searchValue = manualSearch.trim();
      const enabledSupplierCodes = (bootstrap?.suppliers ?? [])
        .filter((supplier) => supplier.enabled)
        .map((supplier) => supplier.supplierCode);
      if (manualSupplierCode === allOptionValue) {
        if (!searchValue) {
          setManualCandidates([]);
          setManualTotal(0);
          setManualReturned(0);
          setManualLoadError(null);
          return;
        }

        if (enabledSupplierCodes.length === 0) {
          setManualCandidates([]);
          setManualTotal(0);
          setManualReturned(0);
          setManualLoadError("Nenhum fornecedor habilitado para consulta.");
          return;
        }
      }

      setManualLoading(true);
      setManualLoadError(null);
      try {
        if (manualSupplierCode === allOptionValue) {
          const limit = manualOffset + manualPageLimit;
          const responses = await Promise.all(
            enabledSupplierCodes.map((supplierCode) =>
              shoppingApi.listManualUrlCandidates({
                supplierCode,
                search: searchValue || undefined,
                brandName: manualBrand === allOptionValue ? undefined : manualBrand,
                taxonomyLeaf0Name: manualTaxonomy === allOptionValue ? undefined : manualTaxonomy,
                includeExisting: manualShowExisting,
                limit,
                offset: 0,
              }),
            ),
          );
          if (!cancelled) {
            const mergedRows = responses
              .flatMap((response) => response.rows)
              .sort((left, right) => {
                const supplierCompare = left.supplierCode.localeCompare(right.supplierCode);
                if (supplierCompare !== 0) {
                  return supplierCompare;
                }
                const skuCompare = (left.sku ?? "").localeCompare(right.sku ?? "");
                if (skuCompare !== 0) {
                  return skuCompare;
                }
                return left.productId.localeCompare(right.productId);
              });
            const total = responses.reduce((sum, response) => sum + response.paging.total, 0);
            const slice = mergedRows.slice(manualOffset, manualOffset + manualPageLimit);
            setManualCandidates(slice);
            setManualTotal(total);
            setManualReturned(slice.length);
          }
          return;
        }

        const list = await shoppingApi.listManualUrlCandidates({
          supplierCode: manualSupplierCode,
          search: searchValue || undefined,
          brandName: manualBrand === allOptionValue ? undefined : manualBrand,
          taxonomyLeaf0Name: manualTaxonomy === allOptionValue ? undefined : manualTaxonomy,
          includeExisting: manualShowExisting,
          limit: manualPageLimit,
          offset: manualOffset,
        });
        if (!cancelled) {
          setManualCandidates(list.rows);
          setManualTotal(list.paging.total);
          setManualReturned(list.paging.returned);
        }
      } catch (loadError) {
        if (!cancelled) {
          const message =
            loadError instanceof Error ? loadError.message : "Falha ao carregar candidatos de URL manual.";
          setManualLoadError(message);
          setManualCandidates([]);
          setManualTotal(0);
          setManualReturned(0);
        }
      } finally {
        if (!cancelled) {
          setManualLoading(false);
        }
      }
    }
    void loadCandidates();
    return () => {
      cancelled = true;
    };
  }, [
    shoppingApi,
    manualSupplierCode,
    manualSearch,
    manualBrand,
    manualTaxonomy,
    manualShowExisting,
    manualOffset,
    bootstrap,
    reloadTick,
  ]);

  useEffect(() => {
    if (inputMode !== "catalog") {
      return;
    }
    let cancelled = false;

    async function loadCatalogProducts() {
      setCatalogLoading(true);
      setError(null);
      try {
        const response = await productsApi.listProductsPortfolio({
          search: catalogSearch || undefined,
          brand_name: catalogBrand === allOptionValue ? undefined : catalogBrand,
          taxonomy_leaf0_name: catalogLeaf0 === allOptionValue ? undefined : catalogLeaf0,
          status: catalogStatus === allOptionValue ? undefined : catalogStatus,
          limit: catalogPageLimit,
          offset: catalogOffset,
        });
        if (cancelled) {
          return;
        }
        applyCatalogResponse(response);
      } catch (catalogError) {
        if (!cancelled) {
          const message = catalogError instanceof Error ? catalogError.message : "Falha ao carregar produtos cadastrados.";
          setError(message);
        }
      } finally {
        if (!cancelled) {
          setCatalogLoading(false);
        }
      }
    }

    void loadCatalogProducts();
    return () => {
      cancelled = true;
    };
  }, [productsApi, inputMode, catalogSearch, catalogBrand, catalogLeaf0, catalogStatus, catalogOffset]);

  useEffect(() => {
    const runRequestId = createdRunRequestId;
    if (!runRequestId) {
      return;
    }
    const stableRunRequestId: string = runRequestId;
    let cancelled = false;
    let timer: number | null = null;

    async function poll() {
      try {
        const next = await shoppingApi.getRunRequest(stableRunRequestId);
        if (cancelled) {
          return;
        }
        setRunRequest(next);

        if (["completed", "failed", "cancelled"].includes(next.status)) {
          return;
        }
      } catch (pollError) {
        if (!cancelled) {
          const message = pollError instanceof Error ? pollError.message : "Falha ao consultar status da run request.";
          setError(message);
        }
        return;
      }

      timer = window.setTimeout(poll, 1500);
    }

    void poll();
    return () => {
      cancelled = true;
      if (timer !== null) {
        window.clearTimeout(timer);
      }
    };
  }, [shoppingApi, createdRunRequestId]);

  function applyCatalogResponse(response: ProductsPortfolioListV1) {
    const filters = response.filters;
    const leaf0Label = filters.taxonomy_leaf0_label.trim() || "Grupo";

    setCatalogRows(response.rows ?? []);
    setCatalogTotal(response.paging.total);
    setCatalogReturned(response.paging.returned);
    setCatalogLeaf0Label(leaf0Label);
    setCatalogBrandOptions(toSelectOptions(filters.brands, "Todas as marcas"));
    setCatalogLeaf0Options(toSelectOptions(filters.taxonomy_leaf0_names, `Todos os ${leaf0Label.toLowerCase()}s`));
    setCatalogStatusOptions(toSelectOptions(filters.status, "Todos os status"));
  }

  function handleXlsxSelection(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;
    if (!file) {
      setXlsxSelectedName("");
      return;
    }
    const maybePath = (file as File & { path?: string }).path;
    setXlsxFilePath((maybePath || file.name || "").trim());
    setXlsxSelectedName(file.name || "");
  }

  const runProgressPct = useMemo(() => {
    if (!selectedRun || selectedRun.totalItems <= 0) {
      return 0;
    }
    return Math.max(0, Math.min(100, Math.round((selectedRun.processedItems / selectedRun.totalItems) * 100)));
  }, [selectedRun]);

  const kpis = useMemo(
    () => ({
      ok: summary?.completedRuns ?? 0,
      nf: summary ? Math.max((summary.totalRuns ?? 0) - (summary.completedRuns ?? 0) - (summary.failedRuns ?? 0), 0) : 0,
      amb: summary?.runningRuns ?? 0,
      err: summary?.failedRuns ?? 0,
    }),
    [summary],
  );

  const currentPageIds = useMemo(() => catalogRows.map((row) => row.product_id), [catalogRows]);
  const allCurrentPageSelected = useMemo(
    () => currentPageIds.length > 0 && currentPageIds.every((id) => selectedProductIds.includes(id)),
    [currentPageIds, selectedProductIds],
  );

  const modeSummary =
    inputMode === "xlsx"
      ? "Fonte selecionada: XLSX (Atual)"
      : `Fonte selecionada: Produtos Cadastrados (${selectedProductIds.length} selecionados)`;

  const manualSupplierOptions = useMemo(
    () =>
      (bootstrap?.suppliers ?? [])
        .filter((supplier) => supplier.enabled)
        .map((supplier) => ({
          value: supplier.supplierCode,
          label: `${supplier.supplierLabel} (${supplier.supplierCode})`,
        })),
    [bootstrap],
  );


  const manualBrandOptions = useMemo(
    () => catalogBrandOptions.filter((option) => option.value !== allOptionValue),
    [catalogBrandOptions],
  );

  const manualTaxonomyOptions = useMemo(
    () => catalogLeaf0Options.filter((option) => option.value !== allOptionValue),
    [catalogLeaf0Options],
  );

  const manualRows = useMemo(() => manualCandidates, [manualCandidates]);

  async function handleRunSelect(runId: string) {
    setError(null);
    try {
      const run = await shoppingApi.getRun(runId);
      setSelectedRun(run);
      setStep(3);
    } catch (selectError) {
      const message = selectError instanceof Error ? selectError.message : "Falha ao carregar detalhe do run.";
      setError(message);
    }
  }

  function toggleProduct(productId: string) {
    setSelectedProductIds((current) =>
      current.includes(productId) ? current.filter((value) => value !== productId) : [...current, productId],
    );
  }

  function toggleCurrentPage() {
    if (allCurrentPageSelected) {
      setSelectedProductIds((current) => current.filter((id) => !currentPageIds.includes(id)));
      return;
    }
    setSelectedProductIds((current) => Array.from(new Set([...current, ...currentPageIds])));
  }

  function resetCatalogPaging() {
    setCatalogOffset(0);
  }

  function clearSelection() {
    setSelectedProductIds([]);
  }

  function toggleSupplier(code: string) {
    setSupplierCodes((current) =>
      current.includes(code) ? current.filter((value) => value !== code) : [...current, code],
    );
  }

  async function createRun() {
    setError(null);
    setCreateRunInfo(null);
    const payload: ShoppingCreateRunRequestV1 = {
      inputMode,
      supplierCodes: supplierCodes.length > 0 ? supplierCodes : undefined,
      advanced: {
        timeoutSeconds: advancedTimeout,
        httpWorkers: advancedHttpWorkers,
        playwrightWorkers: advancedPlaywrightWorkers,
        topN: advancedTopN,
      },
    };

    if (inputMode === "catalog") {
      payload.catalogProductIds = selectedProductIds;
    } else {
      payload.xlsxFilePath = xlsxFilePath.trim() || undefined;
      const xlsxScopeIdentifiers = Array.from(
        new Set(
          xlsxScopeText
            .split(/[\n,;]+/g)
            .map((item) => item.trim())
            .filter((item) => item.length > 0),
        ),
      );
      if (xlsxScopeIdentifiers.length > 0) {
        payload.xlsxScopeIdentifiers = xlsxScopeIdentifiers;
      }
    }

    setCreatingRun(true);
    try {
      const created = await shoppingApi.createRunRequest(payload);
      setCreateRunInfo(`Run solicitada: ${created.runRequestId} (${created.status})`);
      setCreatedRunRequestId(created.runRequestId);
      setRunRequest(null);
      setStep(3);
      setReloadTick((current) => current + 1);
    } catch (requestError) {
      const message = requestError instanceof Error ? requestError.message : "Falha ao criar run.";
      setError(message);
    } finally {
      setCreatingRun(false);
    }
  }

  function manualRowKey(productId: string, supplierCode: string) {
    return `${productId}::${supplierCode.toUpperCase()}`;
  }

  async function saveManualSignalRow(productId: string, supplierCode: string) {
    const key = manualRowKey(productId, supplierCode);
    const nextUrl = (manualEditUrls[key] ?? "").trim();
    if (nextUrl && !nextUrl.startsWith("http")) {
      setError("A URL manual deve iniciar com http ou https.");
      return;
    }
    setManualSignalSaving(true);
    setError(null);
    try {
      await shoppingApi.upsertSupplierSignal({
        productId,
        supplierCode,
        productUrl: nextUrl || null,
        lookupMode: "REFERENCE",
        urlStatus: nextUrl ? "ACTIVE" : "STALE",
        manualOverride: true,
      });
      setReloadTick((current) => current + 1);
    } catch (saveError) {
      const message = saveError instanceof Error ? saveError.message : "Falha ao salvar configuracao manual de URL.";
      setError(message);
    } finally {
      setManualSignalSaving(false);
    }
  }

  return (
    <AppFrame
      eyebrow="MetalShopping"
      title="Shopping de Precos"
      subtitle="Fluxo legado preservado, consumo via SDK oficial."
      hideHero
    >
      <section className={styles.shopping}>
        <header className={styles.header}>
          <h1>
            Shopping <span>de Precos</span>
          </h1>
          <p>Upload, configuracao e execucao com trilha operacional.</p>
        </header>

        <div className={styles.steps}>
          <StepPill step={1} activeStep={step} label="Upload" onClick={() => setStep(1)} />
          <span className={styles.connector} />
          <StepPill step={2} activeStep={step} label="Configurar" onClick={() => setStep(2)} />
          <span className={styles.connector} />
          <StepPill step={3} activeStep={step} label="Executar" onClick={() => setStep(3)} />
        </div>

        {error ? <p className={styles.error}>{error}</p> : null}

        <article className={`${styles.panel} ${step === 1 ? styles.panelActive : ""}`}>
          <h2 className={styles.panelTitle}>Selecione a entrada de produtos</h2>
          <p className={styles.panelSubtitle}>Buscar precos de mercado nos fornecedores selecionados.</p>

          <div className={styles.inputModeToggle}>
            <button
              type="button"
              className={`${styles.btn} ${inputMode === "xlsx" ? styles.btnPrimary : styles.btnSecondary}`.trim()}
              onClick={() => setInputMode("xlsx")}
            >
              XLSX (Atual)
            </button>
            <button
              type="button"
              className={`${styles.btn} ${inputMode === "catalog" ? styles.btnPrimary : styles.btnSecondary}`.trim()}
              onClick={() => setInputMode("catalog")}
            >
              Produtos Cadastrados
            </button>
          </div>

          {inputMode === "xlsx" ? (
            <div className={styles.catalogBlock}>
              <label className={styles.uploadZone}>
                <span className={styles.uploadIcon}>PASTA</span>
                <strong>Arraste um arquivo XLSX ou clique para selecionar</strong>
                <span>{xlsxSelectedName ? `Arquivo: ${xlsxSelectedName}` : "Nenhum arquivo selecionado"}</span>
                <input type="file" className={styles.hiddenInput} onChange={handleXlsxSelection} />
              </label>
              <div className={styles.uploadAdvancedRow}>
                <button
                  type="button"
                  className={styles.linkButton}
                  onClick={() => setShowUploadAdvanced((value) => !value)}
                >
                  {showUploadAdvanced ? "Ocultar configuracao tecnica" : "Mostrar configuracao tecnica"}
                </button>
              </div>
              {showUploadAdvanced ? (
                <div className={styles.uploadAdvancedPanel}>
                  <label className={styles.fieldLabel}>
                    Caminho do arquivo XLSX (backend)
                    <input
                      type="text"
                      value={xlsxFilePath}
                      onChange={(event) => setXlsxFilePath(event.target.value)}
                      placeholder="ex: C:\\imports\\shopping\\atual.xlsx"
                    />
                  </label>
                  <label className={styles.fieldLabel}>
                    Identificadores de escopo (um por linha: SKU/EAN/Referencia)
                    <textarea
                      value={xlsxScopeText}
                      onChange={(event) => setXlsxScopeText(event.target.value)}
                      placeholder={"ex:\n7891234567890\nREF-001\nSKU-XYZ"}
                    />
                  </label>
                </div>
              ) : null}
            </div>
          ) : (
            <div className={styles.catalogBlock}>
              <div className={styles.catalogFilters}>
                <label>
                  Buscar
                  <input
                    className={styles.filterWidgetInput}
                    type="text"
                    value={catalogSearch}
                    onChange={(event) => {
                      setCatalogSearch(event.target.value);
                      resetCatalogPaging();
                    }}
                    placeholder="SKU, nome, EAN, referencia..."
                  />
                </label>

                <div className={styles.catalogFilterField}>
                  <span>Marca</span>
                  <FilterDropdown
                    id="shopping-filter-brand"
                    options={catalogBrandOptions}
                    value={catalogBrand}
                    onSelect={(value) => {
                      setCatalogBrand(value);
                      resetCatalogPaging();
                    }}
                    classNamesOverrides={{
                      wrap: styles.filterWidgetDropdownWrap,
                      trigger: styles.filterWidgetDropdownTrigger,
                      value: styles.filterWidgetDropdownValue,
                    }}
                  />
                </div>

                <div className={styles.catalogFilterField}>
                  <span>{catalogLeaf0Label}</span>
                  <FilterDropdown
                    id="shopping-filter-taxonomy"
                    options={catalogLeaf0Options}
                    value={catalogLeaf0}
                    onSelect={(value) => {
                      setCatalogLeaf0(value);
                      resetCatalogPaging();
                    }}
                    classNamesOverrides={{
                      wrap: styles.filterWidgetDropdownWrap,
                      trigger: styles.filterWidgetDropdownTrigger,
                      value: styles.filterWidgetDropdownValue,
                    }}
                  />
                </div>

                <div className={styles.catalogFilterField}>
                  <span>Status</span>
                  <FilterDropdown
                    id="shopping-filter-status"
                    options={catalogStatusOptions}
                    value={catalogStatus}
                    onSelect={(value) => {
                      setCatalogStatus(value);
                      resetCatalogPaging();
                    }}
                    classNamesOverrides={{
                      wrap: styles.filterWidgetDropdownWrap,
                      trigger: styles.filterWidgetDropdownTrigger,
                      value: styles.filterWidgetDropdownValue,
                    }}
                  />
                </div>
              </div>

              <div className={styles.selectRow}>
                <div className={styles.selectionActions}>
                  <button type="button" onClick={toggleCurrentPage}>
                    {allCurrentPageSelected ? "Desmarcar pagina" : "Selecionar pagina"}
                  </button>
                  <button type="button" onClick={clearSelection} disabled={selectedProductIds.length === 0}>
                    Limpar selecao
                  </button>
                </div>
                <span>
                  {selectedProductIds.length} selecionados | {catalogReturned} de {catalogTotal} itens
                </span>
              </div>

              <div className={styles.catalogTableWrap}>
                <table className={styles.catalogTable}>
                  <thead>
                    <tr>
                      <th>
                        <Checkbox
                          checked={allCurrentPageSelected}
                          onChange={() => toggleCurrentPage()}
                          ariaLabel="Selecionar pagina atual"
                        />
                      </th>
                      <th>SKU</th>
                      <th>Produto</th>
                      <th>Marca</th>
                      <th>{catalogLeaf0Label}</th>
                      <th>Preco</th>
                      <th>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {catalogLoading ? (
                      <tr>
                        <td colSpan={7} className={styles.empty}>
                          Carregando produtos cadastrados...
                        </td>
                      </tr>
                    ) : catalogRows.length === 0 ? (
                      <tr>
                        <td colSpan={7} className={styles.empty}>
                          Nenhum produto encontrado para os filtros atuais.
                        </td>
                      </tr>
                    ) : (
                      catalogRows.map((row) => (
                        <tr key={row.product_id}>
                          <td>
                            <Checkbox
                              checked={selectedProductIds.includes(row.product_id)}
                              onChange={() => toggleProduct(row.product_id)}
                              ariaLabel={`Selecionar ${row.name}`}
                            />
                          </td>
                          <td>{row.sku}</td>
                          <td>{row.name}</td>
                          <td>{row.brand_name ?? "--"}</td>
                          <td>{row.taxonomy_leaf0_name ?? "--"}</td>
                          <td>{formatMoney(row.current_price_amount, row.currency_code)}</td>
                          <td>{row.product_status}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className={styles.selectRow}>
                <button
                  type="button"
                  disabled={catalogOffset <= 0 || catalogLoading}
                  onClick={() => setCatalogOffset((current) => Math.max(0, current - catalogPageLimit))}
                >
                  Pagina anterior
                </button>
                <button
                  type="button"
                  disabled={catalogOffset + catalogReturned >= catalogTotal || catalogLoading}
                  onClick={() => setCatalogOffset((current) => current + catalogPageLimit)}
                >
                  Proxima pagina
                </button>
              </div>
            </div>
          )}

          <div className={styles.btnRow}>
            <span className={styles.modeSummary}>{modeSummary}</span>
            <button type="button" className={`${styles.btn} ${styles.btnPrimary}`} onClick={() => setStep(2)}>
              Continuar
            </button>
          </div>  
        </article>

        <article className={`${styles.panel} ${step === 2 ? styles.panelActive : ""}`}>
          <h2 className={styles.panelTitle}>Selecionar fornecedores</h2>
          <p className={styles.panelSubtitle}>Selecione fornecedores e ajuste limites conforme necessario.</p>

          {bootstrap && bootstrap.suppliers.length > 0 ? (
            <div className={styles.catalogBlock}>
              <div className={styles.supplierHeader}>
                <button type="button" className={`${styles.btn} ${styles.btnGhost}`} onClick={() => setSupplierCodes([])}>
                  Desmarcar todos
                </button>
                <span className={styles.supplierCounter}>
                  {supplierCodes.length} de {bootstrap.suppliers.filter((supplier) => supplier.enabled).length} selecionados
                </span>
              </div>
              <div className={styles.supplierGrid}>
                {bootstrap.suppliers
                  .filter((supplier) => supplier.enabled)
                  .map((supplier) => {
                    const selected = supplierCodes.includes(supplier.supplierCode);
                    return (
                      <button
                        key={supplier.supplierCode}
                        type="button"
                        className={`${styles.supplierCard} ${selected ? styles.supplierCardSelected : ""}`.trim()}
                        onClick={() => toggleSupplier(supplier.supplierCode)}
                      >
                        <span className={`${styles.supplierCheck} ${selected ? styles.supplierCheckSelected : ""}`.trim()}>
                          {selected ? "✓" : ""}
                        </span>
                        <div className={styles.supplierMeta}>
                          <strong>{supplier.supplierLabel}</strong>
                          <small>
                            <span className={styles.supplierMetaDot} />
                            {supplier.supplierCode} - {supplier.executionKind}
                          </small>
                        </div>
                      </button>
                    );
                  })}
              </div>
            </div>
          ) : null}

          <div className={styles.manualPanelCard}>
            <div className={styles.manualPanelHeader}>
              <div>
                <h3 className={styles.manualPanelTitle}>Configurar URLs manuais</h3>
                <p className={styles.manualPanelSubtitle}>Abra apenas quando precisar preencher links faltantes.</p>
              </div>
              <button
                type="button"
                className={`${styles.btn} ${styles.btnPrimary} ${styles.manualPanelTrigger}`}
                onClick={() => setShowManualUrlPanel((value) => !value)}
              >
                {showManualUrlPanel ? "Fechar URLs" : "Configurar URLs"}
              </button>
            </div>
            {showManualUrlPanel ? (
  <div className={styles.manualPanel}>
    <div className={styles.manualFiltersCard}>
      <div className={styles.manualFilters}>
        <label className={styles.manualSearch}>
          Buscar
          <input
            className={styles.filterWidgetInput}
            type="text"
            value={manualSearch}
            onChange={(event) => {
              setManualSearch(event.target.value);
              setManualOffset(0);
            }}
            placeholder="Produto, SKU ou referencia"
          />
        </label>
        <div className={styles.manualFilterField}>
          <span>Fornecedor</span>
                      <FilterDropdown
                        id="manual-filter-supplier"
                        options={manualSupplierOptions}
                        allLabel="Todos fornecedores"
                        value={manualSupplierCode}
                        selectionMode="one"
                        onSelect={(value) => {
                          if (value === allOptionValue) {
                            setManualSupplierCode(allOptionValue);
                            setManualOffset(0);
                            return;
                          }
                          setManualSupplierCode(value);
                          setManualOffset(0);
                        }}
                        classNamesOverrides={{
                          wrap: styles.filterWidgetDropdownWrap,
                          trigger: styles.filterWidgetDropdownTrigger,
                          value: styles.filterWidgetDropdownValue,
                        }}
                      />
        </div>
        <div className={styles.manualFilterField}>
          <span>Marca</span>
                      <FilterDropdown
                        id="manual-filter-brand"
                        options={manualBrandOptions}
                        allLabel="Todas marcas"
                        value={manualBrand}
                        selectionMode="one"
                        onSelect={(value) => {
                          if (value === allOptionValue) {
                            setManualBrand(allOptionValue);
                            setManualOffset(0);
                            return;
                          }
                          setManualBrand(value);
                          setManualOffset(0);
                        }}
                        classNamesOverrides={{
                          wrap: styles.filterWidgetDropdownWrap,
                          trigger: styles.filterWidgetDropdownTrigger,
                          value: styles.filterWidgetDropdownValue,
                        }}
                      />
        </div>
        <div className={styles.manualFilterField}>
          <span>{catalogLeaf0Label}</span>
                      <FilterDropdown
                        id="manual-filter-taxonomy"
                        options={manualTaxonomyOptions}
                        allLabel={`Todos os ${catalogLeaf0Label.toLowerCase()}s`}
                        value={manualTaxonomy}
                        selectionMode="one"
                        onSelect={(value) => {
                          if (value === allOptionValue) {
                            setManualTaxonomy(allOptionValue);
                            setManualOffset(0);
                            return;
                          }
                          setManualTaxonomy(value);
                          setManualOffset(0);
                        }}
                        classNamesOverrides={{
                          wrap: styles.filterWidgetDropdownWrap,
                          trigger: styles.filterWidgetDropdownTrigger,
                          value: styles.filterWidgetDropdownValue,
                        }}
                      />
        </div>
      </div>
      <div className={styles.manualFiltersRow}>
        <label className={styles.manualToggleSwitch}>
          <input
            type="checkbox"
            checked={manualShowExisting}
            onChange={() => setManualShowExisting((value) => !value)}
          />
          <span className={styles.manualToggleSlider} />
          <span className={styles.manualToggleLabel}>Mostrar tambem URLs ja cadastradas</span>
        </label>
        <button
          type="button"
          className={`${styles.btn} ${styles.btnSecondary}`}
          onClick={() => {
            setManualOffset(0);
            setReloadTick((current) => current + 1);
          }}
        >
          Atualizar lista
        </button>
      </div>
    </div>

    <div className={styles.manualTableWrap}>
      <table className={styles.manualTable}>
        <thead>
          <tr>
            <th>Produto</th>
            <th>Fornecedor</th>
            <th>URL</th>
            <th>Status</th>
            <th>Cooldown</th>
          </tr>
        </thead>
        <tbody>
          {manualLoading ? (
            <tr>
              <td colSpan={5} className={styles.manualEmpty}>
                Carregando candidatos...
              </td>
            </tr>
          ) : manualLoadError ? (
            <tr>
              <td colSpan={5} className={styles.manualEmpty}>
                {manualLoadError}
              </td>
            </tr>
          ) : manualSupplierCode === allOptionValue ? (
            <tr>
              <td colSpan={5} className={styles.manualEmpty}>
                {manualSearch.trim()
                  ? "Busque um SKU para consultar todos os fornecedores."
                  : "Informe um SKU ou referencia para buscar em todos os fornecedores."}
              </td>
            </tr>
          ) : manualRows.length === 0 ? (
            <tr>
              <td colSpan={5} className={styles.manualEmpty}>
                Nenhum sinal encontrado.
              </td>
            </tr>
          ) : (
            manualRows.map((candidate) => {
              const rowKey = manualRowKey(candidate.productId, candidate.supplierCode);
              const draftUrl = manualEditUrls[rowKey] ?? candidate.productUrl ?? "";
              const nextDiscovery = candidate.nextDiscoveryAt ? formatDateTime(candidate.nextDiscoveryAt) : "--";
              const notFoundCount = candidate.notFoundCount ?? 0;
              return (
                <tr key={rowKey}>
                  <td>
                    <div className={styles.manualProductCell}>
                      <strong>{candidate.name}</strong>
                      <small>
                        {candidate.sku} | NotFound {notFoundCount}
                      </small>
                    </div>
                  </td>
                  <td>{candidate.supplierCode}</td>
                  <td>
                    <input
                      type="text"
                      value={draftUrl}
                      onChange={(event) =>
                        setManualEditUrls((current) => ({
                          ...current,
                          [rowKey]: event.target.value,
                        }))
                      }
                      placeholder="https://fornecedor/pdp/produto"
                    />
                  </td>
                  <td>{candidate.urlStatus}</td>
                  <td>{nextDiscovery}</td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>

    <div className={styles.manualFooterRow}>
      <span className={styles.manualFooterSummary}>
        Mostrando {manualReturned} de {manualTotal}
      </span>
      <div className={styles.manualFooterPagination}>
        <button
          type="button"
          className={`${styles.btn} ${styles.btnGhost}`}
          disabled={manualOffset <= 0}
          onClick={() => setManualOffset((current) => Math.max(0, current - manualPageLimit))}
        >
          ← Pagina anterior
        </button>
        <button
          type="button"
          className={`${styles.btn} ${styles.btnGhost}`}
          disabled={manualOffset + manualReturned >= manualTotal}
          onClick={() => setManualOffset((current) => current + manualPageLimit)}
        >
          Proxima pagina →
        </button>
      </div>
      <button
        type="button"
        className={`${styles.btn} ${styles.btnPrimary} ${styles.btnCompact} ${styles.manualFooterSave}`}
        disabled
        title="Salvar em lote em breve"
      >
        Salvar
      </button>
    </div>
  </div>
) : null}
          </div>

          <div className={styles.advancedWrap}>
            <div className={styles.advancedHeader}>
              <div>
                <h3 className={styles.advancedTitle}>Configuracoes avancadas</h3>
                <p className={styles.advancedSubtitle}>Ajuste limites e gere relatorio de execucao quando necessario.</p>
              </div>
              <div className={styles.advancedActions}>
                <button
                  type="button"
                  className={`${styles.btn} ${styles.btnPrimary} ${styles.btnCompact} ${styles.manualPanelTrigger}`}
                  onClick={() => setShowAdvanced((value) => !value)}
                >
                  {showAdvanced ? "Ocultar avancado" : "Exibir avancado"}
                </button>
              </div>
            </div>
            {showAdvanced ? (
              <div className={styles.advancedGrid}>
                <label>
                  Timeout (s)
                  <input
                    type="number"
                    min={1}
                    value={advancedTimeout}
                    onChange={(event) => setAdvancedTimeout(Number(event.target.value) || 1)}
                  />
                </label>
                <label>
                  HTTP workers
                  <input
                    type="number"
                    min={1}
                    value={advancedHttpWorkers}
                    onChange={(event) => setAdvancedHttpWorkers(Number(event.target.value) || 1)}
                  />
                </label>
                <label>
                  Playwright workers
                  <input
                    type="number"
                    min={1}
                    value={advancedPlaywrightWorkers}
                    onChange={(event) => setAdvancedPlaywrightWorkers(Number(event.target.value) || 1)}
                  />
                </label>
                <label>
                  Top N
                  <input
                    type="number"
                    min={1}
                    value={advancedTopN}
                    onChange={(event) => setAdvancedTopN(Number(event.target.value) || 1)}
                  />
                </label>
              </div>
            ) : null}
          </div>

          {createRunInfo ? <p className={styles.modeSummary}>{createRunInfo}</p> : null}

          <div className={styles.btnRow}>
            <button
              type="button"
              className={`${styles.btn} ${styles.btnGhost} ${styles.btnGhostUnderline}`}
              onClick={() => setStep(1)}
            >
              ← Voltar
            </button>
            <button
              type="button"
              className={`${styles.btn} ${styles.btnPrimary} ${styles.btnRunPrimary} ${styles.btnCompact}`}
              onClick={() => void createRun()}
              disabled={creatingRun}
            >
              {creatingRun ? "Solicitando..." : "Iniciar run ⚡"}
            </button>
          </div>
        </article>

        <article className={`${styles.panel} ${step === 3 ? styles.panelActive : ""}`}>
          <div className={styles.execHeader}>
            <span className={`${styles.execBadge} ${loadingShopping ? styles.execBadgeRunning : styles.execBadgeDone}`.trim()}>
              {loadingShopping ? "EXECUTANDO" : "PRONTO"}
            </span>
            <strong>
              {runRequest?.runId
                ? `Run ${runRequest.runId}`
                : selectedRun?.runId
                  ? `Run ${selectedRun.runId}`
                  : createdRunRequestId
                    ? `Request ${createdRunRequestId}`
                    : "Sem run ativa"}
            </strong>
            <span>{formatDateTime(summary?.lastRunAt)}</span>
          </div>

          {createdRunRequestId ? (
            <p className={styles.modeSummary}>
              Status request: {runRequest?.status ?? "queued"}{" "}
              {runRequest?.workerId ? `| worker=${runRequest.workerId}` : ""}{" "}
              {runRequest?.errorMessage ? `| erro=${runRequest.errorMessage}` : ""}
            </p>
          ) : null}

          <div className={styles.progress}>
            <div className={styles.progressTrack}>
              <span style={{ width: `${runProgressPct}%` }} />
            </div>
            <div className={styles.progressMeta}>
              <strong>{runProgressPct}%</strong>
              <small>{selectedRun ? `${selectedRun.processedItems}/${selectedRun.totalItems} itens` : "Aguardando run"}</small>
            </div>
          </div>

          <div className={styles.current}>
            <p>{loadingShopping ? "Atualizando resumo e historico..." : modeSummary}</p>
          </div>

          <div className={styles.kpis}>
            <div className={styles.ok}>
              <strong>{kpis.ok}</strong>
              <span>OK</span>
            </div>
            <div className={styles.nf}>
              <strong>{kpis.nf}</strong>
              <span>Not Found</span>
            </div>
            <div className={styles.amb}>
              <strong>{kpis.amb}</strong>
              <span>Ambiguous</span>
            </div>
            <div className={styles.err}>
              <strong>{kpis.err}</strong>
              <span>Error</span>
            </div>
          </div>

          <div className={styles.filterBar}>
            {statusOptions.map((option) => (
              <button
                key={option.value}
                type="button"
                className={`${styles.filterButton} ${selectedStatus === option.value ? styles.filterActive : ""}`.trim()}
                onClick={() => setSelectedStatus(option.value)}
              >
                {option.label}
              </button>
            ))}
          </div>

          <div className={styles.runGrid}>
            <div className={styles.runListWrap}>
              <h3>Historico recente</h3>
              {runs.length === 0 ? (
                <p className={styles.empty}>Nenhum run encontrado para o filtro atual.</p>
              ) : (
                <ul className={styles.runList}>
                  {runs.map((run) => (
                    <li key={run.runId}>
                      <button type="button" className={styles.runButton} onClick={() => void handleRunSelect(run.runId)}>
                        <span className={styles.runMain}>
                          <strong className={styles.runId}>{run.runId}</strong>
                          <small className={styles.runTime}>{formatDateTime(run.startedAt)}</small>
                        </span>
                        <span className={`${styles.statusPill} ${statusClass(styles, run.status)}`.trim()}>{run.status}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <div className={styles.detailWrap}>
              <h3>Detalhe do run</h3>
              {!selectedRun ? (
                <p className={styles.empty}>Selecione um run para visualizar os detalhes.</p>
              ) : (
                <dl className={styles.detailGrid}>
                  <div>
                    <dt>Run ID</dt>
                    <dd>{selectedRun.runId}</dd>
                  </div>
                  <div>
                    <dt>Status</dt>
                    <dd>{selectedRun.status}</dd>
                  </div>
                  <div>
                    <dt>Inicio</dt>
                    <dd>{formatDateTime(selectedRun.startedAt)}</dd>
                  </div>
                  <div>
                    <dt>Fim</dt>
                    <dd>{formatDateTime(selectedRun.finishedAt ?? null)}</dd>
                  </div>
                  <div>
                    <dt>Itens processados</dt>
                    <dd>{selectedRun.processedItems}</dd>
                  </div>
                  <div>
                    <dt>Total de itens</dt>
                    <dd>{selectedRun.totalItems}</dd>
                  </div>
                </dl>
              )}
            </div>
          </div>

          <div className={styles.btnRow}>
            <button type="button" className={`${styles.btn} ${styles.btnGhost}`} onClick={() => setShowLog((value) => !value)}>
              {showLog ? "Ocultar log" : "Ver log detalhado"}
            </button>
            <button type="button" className={`${styles.btn} ${styles.btnSecondary}`} onClick={() => setStep(2)}>
              Voltar para configuracao
            </button>
          </div>

          {showLog ? (
            <div className={styles.log}>
              <p>[STATE] loading={String(loadingShopping)} | runs={runs.length}</p>
              <p>
                [SUMMARY] total={summary?.totalRuns ?? 0} | running={summary?.runningRuns ?? 0} | failed=
                {summary?.failedRuns ?? 0}
              </p>
              <p>[SOURCE] {modeSummary}</p>
              {selectedRun ? (
                <p>
                  [RUN] id={selectedRun.runId} status={selectedRun.status} processed={selectedRun.processedItems}/
                  {selectedRun.totalItems}
                </p>
              ) : (
                <p>[RUN] sem run selecionado</p>
              )}
            </div>
          ) : null}
        </article>
      </section>
    </AppFrame>
  );
}
