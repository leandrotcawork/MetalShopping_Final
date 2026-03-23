// @ts-nocheck
import { useMemo, useState } from "react";

import { FilterDropdown, type SelectMenuOption } from "../../../components/ui/FilterDropdown";
import styles from "../analytics_home.module.css";

type SpotlightSkuTableRow = {
  pn: string;
  description: string;
  brand?: string;
  taxonomyLeafName?: string;
  stockType?: string;
  stockValue?: string;
  stockQty?: string;
  third: string;
  details?: Array<{ label: string; value: string }>;
};

type SpotlightTableContext = {
  brandFilter: string[];
  taxonomyFilter: string[];
  stockTypeFilter: string[];
  pageSize: number;
  pageIndex?: number;
  sortKey?: string;
  sortDir?: "asc" | "desc";
  stockOp?: "gte" | "eq" | "lte";
  stockValue?: number;
};

type SpotlightSkuTableProps = {
  title: string;
  thirdHeader: string;
  rows: SpotlightSkuTableRow[];
  emptyText: string;
  onOpenSku: (pn: string, context?: SpotlightTableContext) => void;
  onTableStateChange?: (context: SpotlightTableContext) => void;
  mode?: "standard" | "details_only";
  defaultSort?: { key: string; dir: "asc" | "desc" };
  detailsHeader?: string;
  initialFilters?: Record<string, unknown>;
};

function uniqueSorted(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean))).sort((a, b) => a.localeCompare(b, "pt-BR"));
}

export function SpotlightSkuTable({
  title,
  thirdHeader,
  rows,
  emptyText,
  onOpenSku,
  onTableStateChange,
  mode = "standard",
}: SpotlightSkuTableProps) {
  const [search, setSearch] = useState("");
  const [brandFilter, setBrandFilter] = useState<string[]>([]);
  const [taxonomyFilter, setTaxonomyFilter] = useState<string[]>([]);
  const [stockTypeFilter, setStockTypeFilter] = useState<string[]>([]);

  const brandOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas", value: "all" }, ...uniqueSorted(rows.map((row) => String(row.brand || ""))).map((item) => ({ label: item, value: item }))],
    [rows],
  );
  const taxonomyOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas", value: "all" }, ...uniqueSorted(rows.map((row) => String(row.taxonomyLeafName || ""))).map((item) => ({ label: item, value: item }))],
    [rows],
  );
  const stockTypeOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas", value: "all" }, ...uniqueSorted(rows.map((row) => String(row.stockType || ""))).map((item) => ({ label: item, value: item }))],
    [rows],
  );

  const filteredRows = useMemo(() => {
    const token = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (brandFilter.length && !brandFilter.includes(String(row.brand || "").trim())) return false;
      if (taxonomyFilter.length && !taxonomyFilter.includes(String(row.taxonomyLeafName || "").trim())) return false;
      if (stockTypeFilter.length && !stockTypeFilter.includes(String(row.stockType || "").trim())) return false;
      if (!token) return true;
      const haystack = [
        row.pn,
        row.description,
        row.brand,
        row.taxonomyLeafName,
        row.stockType,
        row.stockValue,
        row.stockQty,
        row.third,
      ]
        .map((value) => String(value || "").toLowerCase())
        .join(" ");
      return haystack.includes(token);
    });
  }, [brandFilter, rows, search, stockTypeFilter, taxonomyFilter]);

  const context = useMemo<SpotlightTableContext>(
    () => ({
      brandFilter,
      taxonomyFilter,
      stockTypeFilter,
      pageSize: 20,
      pageIndex: 0,
      sortKey: "pn",
      sortDir: "asc",
    }),
    [brandFilter, stockTypeFilter, taxonomyFilter],
  );

  return (
    <section className={styles.spotSection}>
      <h4>{title}</h4>
      <div className={styles.spotlightSkuToolbar}>
        <label className={styles.spotlightSkuFilter}>
          <span>Buscar</span>
          <input
            className={styles.spotlightSkuSearchInput}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="SKU, produto, marca, grupo..."
          />
        </label>
        <div className={styles.spotlightSkuFilter}>
          <span>Marca</span>
          <FilterDropdown
            id="spotlight-brand-filter"
            selectionMode="duo"
            value=""
            values={brandFilter}
            options={brandOptions}
            onSelect={(value) => setBrandFilter(!value || value === "all" ? [] : [value])}
          />
        </div>
        <div className={styles.spotlightSkuFilter}>
          <span>Grupo</span>
          <FilterDropdown
            id="spotlight-tax-filter"
            selectionMode="duo"
            value=""
            values={taxonomyFilter}
            options={taxonomyOptions}
            onSelect={(value) => setTaxonomyFilter(!value || value === "all" ? [] : [value])}
          />
        </div>
        <div className={styles.spotlightSkuFilter}>
          <span>Class. Estoque</span>
          <FilterDropdown
            id="spotlight-stocktype-filter"
            selectionMode="duo"
            value=""
            values={stockTypeFilter}
            options={stockTypeOptions}
            onSelect={(value) => setStockTypeFilter(!value || value === "all" ? [] : [value])}
          />
        </div>
      </div>

      {filteredRows.length === 0 ? <p className={styles.spotlightSkuEmpty}>{emptyText}</p> : null}

      {filteredRows.length > 0 ? (
        <div className={styles.spotlightSkuTable}>
          <table className={`${styles.spotlightSkuNativeTable} ${styles.spotlightSkuCenteredTable}`}>
            <thead className={styles.spotlightSkuNativeHead}>
              <tr>
                <th>SKU</th>
                <th>Produto</th>
                {mode === "details_only" ? <th>Estoque (R$)</th> : null}
                {mode === "details_only" ? <th>Estoque (UN)</th> : null}
                <th>{mode === "details_only" ? "Detalhes" : thirdHeader}</th>
              </tr>
            </thead>
            <tbody className={styles.spotlightSkuNativeBody}>
              {filteredRows.slice(0, 50).map((row) => (
                <tr
                  key={`${row.pn}-${row.description}`}
                  onClick={() => {
                    onOpenSku(row.pn, context);
                    onTableStateChange?.(context);
                  }}
                >
                  <td>{row.pn}</td>
                  <td>
                    <div className={styles.spotlightSkuName}>
                      <span className={styles.spotlightSkuNamePrimary}>{row.description}</span>
                      <span className={styles.spotlightSkuNameMeta}>
                        {`${row.brand || "-"} • ${row.taxonomyLeafName || "-"}`}
                      </span>
                    </div>
                  </td>
                  {mode === "details_only" ? <td>{row.stockValue || "-"}</td> : null}
                  {mode === "details_only" ? <td>{row.stockQty || "-"}</td> : null}
                  <td>
                    {mode === "details_only" && row.details?.length
                      ? row.details.map((detail) => `${detail.label}: ${detail.value}`).join(" | ")
                      : row.third}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </section>
  );
}

