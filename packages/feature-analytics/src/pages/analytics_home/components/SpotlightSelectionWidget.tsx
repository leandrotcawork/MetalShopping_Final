// @ts-nocheck
import { useMemo, useState } from "react";
import styles from "../analytics_home.module.css";
import { SpotlightSkuTable } from "./SpotlightSkuTable";

type SpotlightSelectionItem = {
  key: string;
  title: string;
  meta: string;
  className?: string;
};

type SpotlightSelectionWidgetProps = {
  listTitle?: string;
  items?: SpotlightSelectionItem[];
  onSelectItem?: (key: string) => void;
  tableTitle?: string;
  tableThirdHeader?: string;
  tableRows?: Array<{
    pn: string;
    description: string;
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    stockValue?: string;
    stockValueNumeric?: number | null;
    stockQty?: string;
    stockQtyNumeric?: number | null;
    financialPriority?: string;
    financialPriorityScore?: number | null;
    third: string;
    thirdNumeric?: number | null;
    details?: Array<{ label: string; value: string }>;
  }>;
  tableEmptyText?: string;
  tableDefaultSort?: {
    key: "pn" | "description" | "brand" | "taxonomy" | "financialPriority" | "stockValue" | "stockQty" | "third" | "details";
    dir: "asc" | "desc";
  };
  tableMode?: "standard" | "details_only";
  tableDetailsHeader?: string;
  tableInitialFilters?: {
    brandFilter?: string[];
    taxonomyFilter?: string[];
    stockTypeFilter?: string[];
    pageSize?: number;
    pageIndex?: number;
    sortKey?: string;
    sortDir?: "asc" | "desc";
    stockOp?: "gte" | "eq" | "lte";
    stockValue?: number | null;
  };
  onOpenSku?: (
    pn: string,
    context?: {
      brandFilter: string[];
      taxonomyFilter: string[];
      stockTypeFilter: string[];
      pageSize: number;
      pageIndex?: number;
      sortKey?: string;
      sortDir?: "asc" | "desc";
      stockOp?: "gte" | "eq" | "lte";
      stockValue?: number;
    }
  ) => void;
  onTableStateChange?: (context: {
    brandFilter: string[];
    taxonomyFilter: string[];
    stockTypeFilter: string[];
    pageSize: number;
    pageIndex?: number;
    sortKey?: string;
    sortDir?: "asc" | "desc";
    stockOp?: "gte" | "eq" | "lte";
    stockValue?: number;
  }) => void;
};

export function SpotlightSelectionWidget({
  listTitle,
  items = [],
  onSelectItem,
  tableTitle,
  tableThirdHeader,
  tableRows = [],
  tableEmptyText,
  tableDefaultSort,
  tableMode = "standard",
  tableDetailsHeader = "Detalhes",
  tableInitialFilters,
  onOpenSku,
  onTableStateChange,
}: SpotlightSelectionWidgetProps) {
  const [listQuery, setListQuery] = useState("");
  const filteredItems = useMemo(() => {
    const query = String(listQuery || "").trim().toLowerCase();
    if (!query) return items;
    return items.filter((item) => {
      const title = String(item.title || "").toLowerCase();
      const meta = String(item.meta || "").toLowerCase();
      return title.includes(query) || meta.includes(query);
    });
  }, [items, listQuery]);

  return (
    <>
      {listTitle ? (
        <section className={styles.spotSection}>
          <h4>{listTitle}</h4>
          <div className={styles.spotlightListSearch}>
            <input
              value={listQuery}
              onChange={(event) => setListQuery(event.target.value)}
              placeholder="Buscar no spotlight..."
              aria-label="Buscar itens do spotlight"
            />
          </div>
          <div className={styles.spotlightActionList}>
            {filteredItems.map((item) => (
              <button
                key={item.key}
                type="button"
                className={`${styles.spotlightActionItem} ${item.className || ""}`}
                onClick={() => onSelectItem?.(item.key)}
              >
                <span className={styles.spotlightActionTitle}>{item.title}</span>
                <span className={styles.spotlightActionMeta}>{item.meta}</span>
              </button>
            ))}
          </div>
        </section>
      ) : null}

      {tableTitle && tableThirdHeader && tableEmptyText && onOpenSku ? (
        <SpotlightSkuTable
          key={`table-${tableTitle}`}
          title={tableTitle}
          thirdHeader={tableThirdHeader}
          rows={tableRows}
          emptyText={tableEmptyText}
          defaultSort={tableDefaultSort}
          mode={tableMode}
          detailsHeader={tableDetailsHeader}
          initialFilters={tableInitialFilters}
          onTableStateChange={onTableStateChange}
          onOpenSku={onOpenSku}
        />
      ) : null}
    </>
  );
}

