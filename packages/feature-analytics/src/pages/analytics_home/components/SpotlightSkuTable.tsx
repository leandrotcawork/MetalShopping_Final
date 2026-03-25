import { useEffect, useMemo, useRef, useState } from "react";
import {
  ColumnDef,
  PaginationState,
  SortingState,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { FilterDropdown, type SelectMenuOption } from "@metalshopping/ui";
import styles from "../analytics_home.module.css";

type SpotlightSkuTableRow = {
  pn: string;
  description: string;
  brand?: string;
  taxonomyLeafName?: string;
  stockType?: string;
  financialPriority?: string;
  financialPriorityScore?: number | null;
  stockValue?: string;
  stockValueNumeric?: number | null;
  stockQty?: string;
  stockQtyNumeric?: number | null;
  third: string;
  thirdNumeric?: number | null;
  details?: Array<{ label: string; value: string }>;
};

type SortKey =
  | "pn"
  | "description"
  | "brand"
  | "taxonomy"
  | "financialPriority"
  | "stockValue"
  | "stockQty"
  | "third"
  | "details";
type StockFilterOp = "gte" | "eq" | "lte";

type SpotlightTableContext = {
  brandFilter: string[];
  taxonomyFilter: string[];
  stockTypeFilter: string[];
  pageSize: number;
  pageIndex?: number;
  sortKey?: string;
  sortDir?: "asc" | "desc";
  stockOp?: StockFilterOp;
  stockValue?: number;
};

type SpotlightSkuTableProps = {
  title: string;
  thirdHeader: string;
  rows: SpotlightSkuTableRow[];
  emptyText: string;
  onOpenSku: (pn: string, context?: SpotlightTableContext) => void;
  onTableStateChange?: (context: SpotlightTableContext) => void;
  defaultSort?: { key: SortKey; dir: "asc" | "desc" };
  mode?: "standard" | "details_only";
  detailsHeader?: string;
  initialFilters?: {
    brandFilter?: string[];
    taxonomyFilter?: string[];
    stockTypeFilter?: string[];
    pageSize?: number;
    pageIndex?: number;
    sortKey?: string;
    sortDir?: "asc" | "desc";
    stockOp?: StockFilterOp;
    stockValue?: number | null;
  };
};

export function SpotlightSkuTable({
  title,
  thirdHeader,
  rows,
  emptyText,
  onOpenSku,
  onTableStateChange,
  defaultSort = { key: "pn", dir: "asc" },
  mode = "standard",
  detailsHeader = "Detalhes",
  initialFilters,
}: SpotlightSkuTableProps) {
  const isValidSortKey = (value: unknown): value is SortKey =>
    [
      "pn",
      "description",
      "brand",
      "taxonomy",
      "financialPriority",
      "stockValue",
      "stockQty",
      "third",
      "details",
    ].includes(String(value || ""));

  const initialSort: { key: SortKey; dir: "asc" | "desc" } = (() => {
    const key = initialFilters?.sortKey;
    const dir = initialFilters?.sortDir;
    if (isValidSortKey(key) && (dir === "asc" || dir === "desc")) {
      return { key, dir };
    }
    return defaultSort;
  })();

  const clamp = (value: number, min: number, max: number): number =>
    Math.max(min, Math.min(max, value));

  const estimateWidthPx = (
    values: Array<string | number | null | undefined>,
    minPx: number,
    maxPx: number,
    pxPerChar: number
  ): number => {
    const longest = values.reduce<number>((maxLen, raw) => {
      const size = String(raw ?? "").trim().length;
      return size > maxLen ? size : maxLen;
    }, 0);
    const estimated = Math.round(longest * pxPerChar + 28);
    return clamp(estimated, minPx, maxPx);
  };

  const [sorting, setSorting] = useState<SortingState>([
    { id: initialSort.key, desc: initialSort.dir === "desc" },
  ]);
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex:
      Number.isFinite(Number(initialFilters?.pageIndex)) && Number(initialFilters?.pageIndex) >= 0
        ? Math.floor(Number(initialFilters?.pageIndex))
        : 0,
    pageSize: [10, 20, 50].includes(Number(initialFilters?.pageSize))
      ? Number(initialFilters?.pageSize)
      : 20,
  });
  const normalizeMultiFilter = (values: string[] | undefined): string[] =>
    Array.from(
      new Set(
        (Array.isArray(values) ? values : [])
          .map((value) => String(value || "").trim())
          .filter((value) => value && value !== "all")
      )
    );

  const [brandFilter, setBrandFilter] = useState<string[]>(() => normalizeMultiFilter(initialFilters?.brandFilter));
  const [taxonomyFilter, setTaxonomyFilter] = useState<string[]>(() => normalizeMultiFilter(initialFilters?.taxonomyFilter));
  const [stockTypeFilter, setStockTypeFilter] = useState<string[]>(() => normalizeMultiFilter(initialFilters?.stockTypeFilter));
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [stockFilterOp, setStockFilterOp] = useState<StockFilterOp>(
    initialFilters?.stockOp === "eq" || initialFilters?.stockOp === "lte" ? initialFilters.stockOp : "gte"
  );
  const [stockFilterInput, setStockFilterInput] = useState<string>(
    initialFilters?.stockValue != null && Number.isFinite(Number(initialFilters.stockValue))
      ? String(initialFilters.stockValue)
      : ""
  );

  useEffect(() => {
    const nextPageSize = [10, 20, 50].includes(Number(initialFilters?.pageSize))
      ? Number(initialFilters?.pageSize)
      : 20;
    const rawPageIndex = Number(initialFilters?.pageIndex);
    const nextPageIndex = Number.isFinite(rawPageIndex) && rawPageIndex >= 0
      ? Math.floor(rawPageIndex)
      : 0;
    setPagination((prev) => {
      if (prev.pageSize === nextPageSize && prev.pageIndex === nextPageIndex) return prev;
      return { pageSize: nextPageSize, pageIndex: nextPageIndex };
    });
  }, [initialFilters?.pageIndex, initialFilters?.pageSize]);

  const hasDetailsColumn = rows.some(
    (row) => Array.isArray(row.details) && row.details.length > 0
  );
  const isDetailsOnly = mode === "details_only";
  const hasOpsColumns =
    isDetailsOnly &&
    rows.some(
      (row) =>
        String(row.brand || "").trim() ||
        String(row.taxonomyLeafName || "").trim() ||
        row.stockValueNumeric != null ||
        row.stockQtyNumeric != null
    );
  const hasStockQtyColumn = hasOpsColumns;

  function parseLocaleNumber(raw: string): number | null {
    const source = String(raw || "").trim();
    if (!source) return null;
    const cleaned = source
      .replace(/[^\d,.\-]/g, "")
      .replace(/\.(?=\d{3}\b)/g, "")
      .replace(",", ".");
    if (!cleaned || cleaned === "-" || cleaned === "," || cleaned === ".") return null;
    const value = Number(cleaned);
    return Number.isFinite(value) ? value : null;
  }

  function detailNumeric(row: SpotlightSkuTableRow): number | null {
    const details = Array.isArray(row.details) ? row.details : [];
    for (const detail of details) {
      const parsed = parseLocaleNumber(String(detail.value || ""));
      if (parsed != null) return parsed;
    }
    return null;
  }

  function detailText(row: SpotlightSkuTableRow): string {
    const details = Array.isArray(row.details) ? row.details : [];
    return details
      .map((item) => {
        const label = String(item.label || "").trim();
        const value = String(item.value || "").trim();
        if (!label) return value;
        return `${label}: ${value}`;
      })
      .join(" | ");
  }

  const pageSizeOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: "10", value: "10" },
      { label: "20", value: "20" },
      { label: "50", value: "50" },
    ],
    []
  );
  const stockOperatorOptions = useMemo<SelectMenuOption[]>(
    () => [
      { label: ">=", value: "gte" },
      { label: "=", value: "eq" },
      { label: "<=", value: "lte" },
    ],
    []
  );
  const stockFilterValue = useMemo(() => parseLocaleNumber(stockFilterInput), [stockFilterInput]);
  const normalizedSearchQuery = useMemo(
    () => String(searchQuery || "").trim().toLowerCase(),
    [searchQuery]
  );

  const matchesSearchQuery = (row: SpotlightSkuTableRow, query: string): boolean => {
    if (!query) return true;
    const detailsText = Array.isArray(row.details)
      ? row.details.map((item) => `${item.label || ""} ${item.value || ""}`).join(" ")
      : "";
    const searchable = [
      row.pn,
      row.description,
      row.brand,
      row.taxonomyLeafName,
      row.stockType,
      row.third,
      row.stockValue,
      row.stockQty,
      detailsText,
    ]
      .map((value) => String(value || "").toLowerCase())
      .join(" ");
    return searchable.includes(query);
  };

  const stockQtyResolved = (row: SpotlightSkuTableRow): number | null =>
    typeof row.stockQtyNumeric === "number" && Number.isFinite(row.stockQtyNumeric)
      ? row.stockQtyNumeric
      : parseLocaleNumber(String(row.stockQty || ""));

  const matchesStockFilter = (
    row: SpotlightSkuTableRow,
    threshold: number | null,
    op: StockFilterOp
  ): boolean => {
    if (threshold == null) return true;
    const qtyValue = stockQtyResolved(row);
    if (qtyValue == null) return false;
    if (op === "gte") return qtyValue >= threshold;
    if (op === "lte") return qtyValue <= threshold;
    return Math.abs(qtyValue - threshold) <= 1e-9;
  };

  const brandOptions = useMemo(
    () =>
      Array.from(
        new Set(
          rows
            .filter((row) =>
              (taxonomyFilter.length === 0 || taxonomyFilter.includes(String(row.taxonomyLeafName || "").trim())) &&
              (stockTypeFilter.length === 0 || stockTypeFilter.includes(String(row.stockType || "").trim())) &&
              matchesStockFilter(row, stockFilterValue, stockFilterOp) &&
              matchesSearchQuery(row, normalizedSearchQuery)
            )
            .map((row) => String(row.brand || "").trim())
            .filter(Boolean)
        )
      ).sort((a, b) => a.localeCompare(b, "pt-BR")),
    [rows, taxonomyFilter, stockTypeFilter, stockFilterValue, stockFilterOp, normalizedSearchQuery]
  );

  const taxonomyOptions = useMemo(
    () =>
      Array.from(
        new Set(
          rows
            .filter((row) =>
              (brandFilter.length === 0 || brandFilter.includes(String(row.brand || "").trim())) &&
              (stockTypeFilter.length === 0 || stockTypeFilter.includes(String(row.stockType || "").trim())) &&
              matchesStockFilter(row, stockFilterValue, stockFilterOp) &&
              matchesSearchQuery(row, normalizedSearchQuery)
            )
            .map((row) => String(row.taxonomyLeafName || "").trim())
            .filter(Boolean)
        )
      ).sort((a, b) => a.localeCompare(b, "pt-BR")),
    [rows, brandFilter, stockTypeFilter, stockFilterValue, stockFilterOp, normalizedSearchQuery]
  );

  const stockTypeOptions = useMemo(
    () =>
      Array.from(
        new Set(
          rows
            .filter((row) =>
              (brandFilter.length === 0 || brandFilter.includes(String(row.brand || "").trim())) &&
              (taxonomyFilter.length === 0 || taxonomyFilter.includes(String(row.taxonomyLeafName || "").trim())) &&
              matchesStockFilter(row, stockFilterValue, stockFilterOp) &&
              matchesSearchQuery(row, normalizedSearchQuery)
            )
            .map((row) => String(row.stockType || "").trim())
            .filter(Boolean)
        )
      ).sort((a, b) => a.localeCompare(b, "pt-BR")),
    [rows, brandFilter, taxonomyFilter, stockFilterValue, stockFilterOp, normalizedSearchQuery]
  );

  const brandSelectOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas", value: "all" }, ...brandOptions.map((brand) => ({ label: brand, value: brand }))],
    [brandOptions]
  );

  const taxonomySelectOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas", value: "all" }, ...taxonomyOptions.map((taxonomy) => ({ label: taxonomy, value: taxonomy }))],
    [taxonomyOptions]
  );
  const stockTypeSelectOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas", value: "all" }, ...stockTypeOptions.map((stockType) => ({ label: stockType, value: stockType }))],
    [stockTypeOptions]
  );

  useEffect(() => {
    setBrandFilter((prev) => {
      const next = prev.filter((value) => brandOptions.includes(value));
      if (next.length === prev.length && next.every((value, index) => value === prev[index])) return prev;
      return next;
    });
  }, [brandFilter, brandOptions]);

  useEffect(() => {
    setTaxonomyFilter((prev) => {
      const next = prev.filter((value) => taxonomyOptions.includes(value));
      if (next.length === prev.length && next.every((value, index) => value === prev[index])) return prev;
      return next;
    });
  }, [taxonomyFilter, taxonomyOptions]);

  useEffect(() => {
    setStockTypeFilter((prev) => {
      const next = prev.filter((value) => stockTypeOptions.includes(value));
      if (next.length === prev.length && next.every((value, index) => value === prev[index])) return prev;
      return next;
    });
  }, [stockTypeFilter, stockTypeOptions]);

  const filteredRows = useMemo(() => {
    return rows.filter((row) => {
      const brandMatch =
        brandFilter.length === 0 || brandFilter.includes(String(row.brand || "").trim());
      const taxonomyMatch =
        taxonomyFilter.length === 0 ||
        taxonomyFilter.includes(String(row.taxonomyLeafName || "").trim());
      const stockTypeMatch =
        stockTypeFilter.length === 0 ||
        stockTypeFilter.includes(String(row.stockType || "").trim());
      if (!brandMatch || !taxonomyMatch || !stockTypeMatch) return false;
      if (!matchesStockFilter(row, stockFilterValue, stockFilterOp)) return false;
      return matchesSearchQuery(row, normalizedSearchQuery);
    });
  }, [rows, brandFilter, taxonomyFilter, stockTypeFilter, stockFilterOp, stockFilterValue, normalizedSearchQuery]);

  const lastFilterStateSignatureRef = useRef<string | null>(null);
  useEffect(() => {
    const currentSignature = JSON.stringify([
      [...brandFilter].sort(),
      [...taxonomyFilter].sort(),
      [...stockTypeFilter].sort(),
      stockFilterOp,
      stockFilterInput,
      searchQuery,
    ]);
    if (lastFilterStateSignatureRef.current == null) {
      lastFilterStateSignatureRef.current = currentSignature;
      return;
    }
    if (lastFilterStateSignatureRef.current === currentSignature) {
      return;
    }
    lastFilterStateSignatureRef.current = currentSignature;
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  }, [brandFilter, taxonomyFilter, stockTypeFilter, stockFilterOp, stockFilterInput, searchQuery]);

  const columns = useMemo<ColumnDef<SpotlightSkuTableRow>[]>(() => {
    const numberOrParsed = (
      explicit: number | null | undefined,
      label: string | undefined
    ): number | null => {
      if (typeof explicit === "number" && Number.isFinite(explicit)) return explicit;
      return parseLocaleNumber(String(label || ""));
    };

    const base: ColumnDef<SpotlightSkuTableRow>[] = [
      {
        id: "pn",
        header: "SKU",
        accessorFn: (row) => String(row.pn || ""),
        cell: ({ row }) => (
          <span className={styles.spotlightSkuPn}>{row.original.pn}</span>
        ),
        sortingFn: (a, b) =>
          String(a.original.pn || "").localeCompare(String(b.original.pn || ""), "pt-BR", {
            numeric: true,
          }),
      },
      {
        id: "description",
        header: "Produto",
        accessorFn: (row) => String(row.description || ""),
        cell: ({ row }) => (
          <span className={styles.spotlightSkuName}>
            <span className={styles.spotlightSkuNamePrimary}>
              {row.original.description || "-"}
            </span>
            {hasOpsColumns ? (
              <span className={styles.spotlightSkuNameMeta}>
                {`${row.original.brand || "-"} • ${row.original.taxonomyLeafName || "-"}`}
              </span>
            ) : null}
          </span>
        ),
        sortingFn: (a, b) =>
          String(a.original.description || "").localeCompare(
            String(b.original.description || ""),
            "pt-BR"
          ),
      },
    ];

    if (hasOpsColumns) {
      base.push({
        id: "stockValue",
        header: "Estoque (R$)",
        accessorFn: (row) => String(row.stockValue || ""),
        cell: ({ row }) => (
          <span className={styles.spotlightSkuUrgency}>{row.original.stockValue || "-"}</span>
        ),
        sortingFn: (a, b) => {
          const aNum = numberOrParsed(a.original.stockValueNumeric, a.original.stockValue);
          const bNum = numberOrParsed(b.original.stockValueNumeric, b.original.stockValue);
          if (aNum != null && bNum != null && aNum !== bNum) return aNum - bNum;
          return String(a.original.stockValue || "").localeCompare(
            String(b.original.stockValue || ""),
            "pt-BR"
          );
        },
      });

      if (hasStockQtyColumn) {
        base.push({
          id: "stockQty",
          header: "Estoque (UN)",
          accessorFn: (row) => String(row.stockQty || ""),
          cell: ({ row }) => (
            <span className={styles.spotlightSkuCellMuted}>{row.original.stockQty || "-"}</span>
          ),
          sortingFn: (a, b) => {
            const aNum = numberOrParsed(a.original.stockQtyNumeric, a.original.stockQty);
            const bNum = numberOrParsed(b.original.stockQtyNumeric, b.original.stockQty);
            if (aNum != null && bNum != null && aNum !== bNum) return aNum - bNum;
            return String(a.original.stockQty || "").localeCompare(
              String(b.original.stockQty || ""),
              "pt-BR",
              { numeric: true }
            );
          },
        });
      }
    }

    if (!isDetailsOnly) {
      base.push({
        id: "third",
        header: thirdHeader,
        accessorFn: (row) => String(row.third || ""),
        cell: ({ row }) => (
          <span className={styles.spotlightSkuUrgency}>{row.original.third || "-"}</span>
        ),
        sortingFn: (a, b) => {
          const aNum = numberOrParsed(a.original.thirdNumeric, a.original.third);
          const bNum = numberOrParsed(b.original.thirdNumeric, b.original.third);
          if (aNum != null && bNum != null && aNum !== bNum) return aNum - bNum;
          return String(a.original.third || "").localeCompare(String(b.original.third || ""), "pt-BR", {
            numeric: true,
          });
        },
      });
    }

    if (hasDetailsColumn) {
      base.push({
        id: "details",
        header: detailsHeader,
        accessorFn: (row) => detailText(row),
        cell: ({ row }) => (
          <span className={styles.spotlightSkuDetailsText}>
            {row.original.details && row.original.details.length
              ? row.original.details
                  .map((item) => {
                    const label = String(item.label || "").trim();
                    const value = String(item.value || "").trim();
                    if (!label) return value;
                    return `${label}: ${value}`;
                  })
                  .join(" | ")
              : "-"}
          </span>
        ),
        sortingFn: (a, b) => {
          const aNum = detailNumeric(a.original);
          const bNum = detailNumeric(b.original);
          if (aNum != null && bNum != null && aNum !== bNum) return aNum - bNum;
          return detailText(a.original).localeCompare(detailText(b.original), "pt-BR", {
            numeric: true,
          });
        },
      });
    }

    return base;
  }, [hasOpsColumns, hasStockQtyColumn, isDetailsOnly, hasDetailsColumn, thirdHeader, detailsHeader]);

  useEffect(() => {
    const activeSortId = String(sorting[0]?.id || "");
    const available = new Set(columns.map((column) => String(column.id || "")));
    if (!activeSortId || available.has(activeSortId)) return;
    setSorting([{ id: defaultSort.key, desc: defaultSort.dir === "desc" }]);
  }, [columns, defaultSort.dir, defaultSort.key, sorting]);

  const table = useReactTable({
    data: filteredRows,
    columns,
    state: { sorting, pagination },
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    enableSortingRemoval: false,
    autoResetPageIndex: false,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  });

  const columnWidthStyle = useMemo(() => {
    const ids = columns.map((column) => String(column.id));
    const byId: Record<string, string | undefined> = {};

    for (const id of ids) {
      if (id === "pn") {
        byId[id] = `${estimateWidthPx(
          ["SKU", ...filteredRows.map((row) => row.pn)],
          76,
          108,
          7
        )}px`;
        continue;
      }

      if (id === "stockValue") {
        byId[id] = `${estimateWidthPx(
          ["Estoque (R$)", ...filteredRows.map((row) => row.stockValue || "-")],
          118,
          160,
          7.1
        )}px`;
        continue;
      }

      if (id === "stockQty") {
        byId[id] = `${estimateWidthPx(
          ["Estoque (UN)", ...filteredRows.map((row) => row.stockQty || "-")],
          110,
          146,
          7
        )}px`;
        continue;
      }

      if (id === "third") {
        byId[id] = `${estimateWidthPx(
          [thirdHeader, ...filteredRows.map((row) => row.third || "-")],
          132,
          188,
          7.1
        )}px`;
        continue;
      }

      if (id === "details") {
        byId[id] = `${estimateWidthPx(
          [
            detailsHeader,
            ...filteredRows.map((row) =>
              row.details && row.details.length
                ? row.details.map((item) => `${item.label}: ${item.value}`).join(" | ")
                : "-"
            ),
          ],
          210,
          260,
          6.8
        )}px`;
        continue;
      }

      if (id === "description") {
        byId[id] = `${estimateWidthPx(
          ["Produto", ...filteredRows.map((row) => row.description || "-")],
          300,
          420,
          7
        )}px`;
        continue;
      }

      byId[id] = `${estimateWidthPx(
        [String(id)],
        120,
        180,
        7
      )}px`;
    }

    return ids.map((id) => byId[id]);
  }, [columns, filteredRows, thirdHeader, detailsHeader]);

  const totalRows = filteredRows.length;
  const totalPages = Math.max(1, table.getPageCount());
  const currentPage = table.getState().pagination.pageIndex + 1;
  const didMountTableStateSyncRef = useRef(false);

  useEffect(() => {
    if (!onTableStateChange) return;
    if (!didMountTableStateSyncRef.current) {
      didMountTableStateSyncRef.current = true;
      return;
    }
    const context = {
      brandFilter,
      taxonomyFilter,
      stockTypeFilter,
      pageSize: pagination.pageSize,
      pageIndex: pagination.pageIndex,
      sortKey: isValidSortKey(sorting[0]?.id) ? sorting[0]?.id : undefined,
      sortDir: sorting[0]?.desc ? "desc" : "asc",
      stockOp: stockFilterValue != null ? stockFilterOp : undefined,
      stockValue: stockFilterValue != null ? stockFilterValue : undefined,
    } satisfies SpotlightTableContext;
    onTableStateChange(context);
  }, [
    onTableStateChange,
    brandFilter,
    taxonomyFilter,
    stockTypeFilter,
    pagination.pageSize,
    pagination.pageIndex,
    sorting,
    stockFilterOp,
    stockFilterValue,
  ]);

  return (
    <section className={styles.spotSection}>
      <h4>{title}</h4>

      {hasOpsColumns ? (
        <div className={styles.spotlightSkuToolbar}>
          <label className={styles.spotlightSkuFilter}>
            <span>Buscar</span>
            <input
              className={styles.spotlightSkuSearchInput}
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder="SKU, produto, marca, grupo..."
              aria-label="Buscar SKUs no spotlight"
            />
          </label>

          <div className={styles.spotlightSkuFilter}>
            <span>Marca</span>
            <FilterDropdown
              id="spotlight-filter-brand"
              selectionMode="duo"
              value=""
              values={brandFilter}
              options={brandSelectOptions}
              classNamesOverrides={{ wrap: `spotlight-select-wrap ${styles.spotlightSelectWrap}` }}
              onSelect={(value) =>
                setBrandFilter((prev) => {
                  if (!value || value === "all") return [];
                  return prev.includes(value) ? prev.filter((item) => item !== value) : [...prev, value];
                })
              }
            />
          </div>

          <div className={styles.spotlightSkuFilter}>
            <span>Grupos</span>
            <FilterDropdown
              id="spotlight-filter-group"
              selectionMode="duo"
              value=""
              values={taxonomyFilter}
              options={taxonomySelectOptions}
              classNamesOverrides={{ wrap: `spotlight-select-wrap ${styles.spotlightSelectWrap}` }}
              onSelect={(value) =>
                setTaxonomyFilter((prev) => {
                  if (!value || value === "all") return [];
                  return prev.includes(value) ? prev.filter((item) => item !== value) : [...prev, value];
                })
              }
            />
          </div>
          <div className={styles.spotlightSkuFilter}>
            <span>Class. Estoque</span>
            <FilterDropdown
              id="spotlight-filter-stock-type"
              selectionMode="duo"
              value=""
              values={stockTypeFilter}
              options={stockTypeSelectOptions}
              classNamesOverrides={{ wrap: `spotlight-select-wrap ${styles.spotlightSelectWrap}` }}
              onSelect={(value) =>
                setStockTypeFilter((prev) => {
                  if (!value || value === "all") return [];
                  return prev.includes(value) ? prev.filter((item) => item !== value) : [...prev, value];
                })
              }
            />
          </div>
          <div className={`${styles.spotlightSkuFilter} ${styles.spotlightSkuStockFilter}`}>
            <span>Estoque (UN)</span>
            <FilterDropdown
              id="spotlight-filter-stock-op"
              value={stockFilterOp}
              options={stockOperatorOptions}
              classNamesOverrides={{ wrap: `spotlight-select-wrap ${styles.spotlightSelectWrap}` }}
              onSelect={(value) => {
                const next = value === "eq" || value === "lte" ? value : "gte";
                setStockFilterOp(next);
              }}
            />
            <input
              className={styles.spotlightSkuStockValueInput}
              value={stockFilterInput}
              onChange={(event) => setStockFilterInput(event.target.value)}
              placeholder="Valor"
              aria-label="Valor do filtro de estoque"
              inputMode="decimal"
            />
          </div>

          <div className={styles.spotlightSkuFilter}>
            <span>Itens/pagina</span>
            <FilterDropdown
              id="spotlight-filter-page-size"
              value={String(table.getState().pagination.pageSize)}
              options={pageSizeOptions}
              classNamesOverrides={{ wrap: `spotlight-select-wrap ${styles.spotlightSelectWrap}` }}
              onSelect={(value) => table.setPageSize(Math.max(5, Number(value) || 20))}
            />
          </div>
        </div>
      ) : (
        <div className={styles.spotlightSkuToolbar}>
          <label className={styles.spotlightSkuFilter}>
            <span>Buscar</span>
            <input
              className={styles.spotlightSkuSearchInput}
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder="SKU, produto, detalhes..."
              aria-label="Buscar SKUs no spotlight"
            />
          </label>
        </div>
      )}

      <div className={styles.spotlightSkuTable} style={{ minWidth: 0 }}>
        {totalRows > 0 ? (
          <table
            className={`${styles.spotlightSkuNativeTable} ${styles.spotlightSkuCenteredTable}`}
            style={{
              width: "100%",        // ? ocupa o container inteiro
              tableLayout: "fixed", // ? respeita colgroup e evita shrink-to-fit
            }}
          >
            <colgroup>
              {columnWidthStyle.map((width, idx) => (
                <col key={`col-${idx}`} style={width ? { width } : undefined} />
              ))}
            </colgroup>

            <thead className={styles.spotlightSkuNativeHead}>
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className={styles[`spotlightColHeader_${header.column.id}` as keyof typeof styles]}
                    >
                      {header.isPlaceholder ? null : (
                        <button
                          type="button"
                          className={styles.spotlightSkuSortBtn}
                          onClick={header.column.getToggleSortingHandler()}
                        >
                          <span className={styles.spotlightSkuHeadLabel}>
                            {flexRender(header.column.columnDef.header, header.getContext())}
                          </span>
                        </button>
                      )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>

            <tbody className={styles.spotlightSkuNativeBody}>
              {table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  onClick={() => {
                    const context = {
                      brandFilter,
                      taxonomyFilter,
                      stockTypeFilter,
                      pageSize: table.getState().pagination.pageSize,
                      pageIndex: table.getState().pagination.pageIndex,
                      sortKey: isValidSortKey(table.getState().sorting[0]?.id)
                        ? table.getState().sorting[0]?.id
                        : undefined,
                      sortDir: table.getState().sorting[0]?.desc ? "desc" : "asc",
                      stockOp: stockFilterValue != null ? stockFilterOp : undefined,
                      stockValue: stockFilterValue != null ? stockFilterValue : undefined,
                    } satisfies SpotlightTableContext;
                    onOpenSku(row.original.pn, context);
                  }}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td
                      key={cell.id}
                      className={styles[`spotlightColCell_${cell.column.id}` as keyof typeof styles]}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className={styles.spotlightSkuEmpty}>{emptyText}</div>
        )}
      </div>

      {totalRows > 0 ? (
        <div className={styles.spotlightSkuPager}>
          <span>
            Pagina {currentPage} de {totalPages} ({totalRows} itens)
          </span>
          <div className={styles.spotlightSkuPagerActions}>
            <button type="button" onClick={() => table.previousPage()} disabled={!table.getCanPreviousPage()}>
              Anterior
            </button>
            <button type="button" onClick={() => table.nextPage()} disabled={!table.getCanNextPage()}>
              Proxima
            </button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
