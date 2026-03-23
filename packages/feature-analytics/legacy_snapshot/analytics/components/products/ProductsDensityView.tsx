import { useMemo, useState, type FocusEvent, type ReactNode } from "react";

import type { AnalyticsSkuRow } from "../../contracts_products";
import styles from "../../analytics_products.module.css";

type ProductsDensityViewProps = {
  rows: AnalyticsSkuRow[];
  onRowClick: (pn: string) => void;
  columnSort: {
    key:
      | "pn"
      | "product"
      | "brand"
      | "price"
      | "market"
      | "gap"
      | "margin"
      | "trend"
      | "class"
      | "stock";
    direction: "asc" | "desc" | "";
  };
  onColumnSortChange: (next: ProductsDensityViewProps["columnSort"]) => void;
};

function formatCurrency(value: number): string {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    maximumFractionDigits: 2
  }).format(value);
}

type FilterPopoverProps = {
  label: string;
  active: boolean;
  align?: "start" | "end";
  children: (helpers: { close: () => void }) => ReactNode;
};

function FilterPopover({ label, active, align = "end", children }: FilterPopoverProps) {
  const [open, setOpen] = useState(false);
  const close = () => setOpen(false);

  function handleBlur(event: FocusEvent<HTMLDivElement>) {
    const nextTarget = event.relatedTarget as Node | null;
    if (!nextTarget || !event.currentTarget.contains(nextTarget)) {
      setOpen(false);
    }
  }

  return (
    <div className={styles.colFilterWrap} onBlur={handleBlur}>
      <button
        type="button"
        className={`${styles.colFilterBtn}${active ? ` ${styles.colFilterBtnActive}` : ""}`}
        onClick={() => setOpen((prev) => !prev)}
        aria-haspopup="dialog"
        aria-expanded={open}
        title={`Ordenar ${label}`}
      >
        <svg viewBox="0 0 20 20" fill="none" aria-hidden>
          <path d="M6 4h8M8 10h4M9.5 16h1" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </button>

      {open ? (
        <div
          className={`${styles.colFilterPopover} ${align === "start" ? styles.colFilterPopoverStart : styles.colFilterPopoverEnd}`}
        >
          {children({ close })}
        </div>
      ) : null}
    </div>
  );
}

type Option = { label: string; value: string };

type OptionsFilterProps = {
  label: string;
  value: string;
  options: Option[];
  onSelect: (value: string) => void;
  onClose?: () => void;
};

function OptionsFilter({ label, value, options, onSelect, onClose }: OptionsFilterProps) {
  return (
    <div className={styles.colFilterContent}>
      <div className={styles.colFilterTitle}>{label}</div>
      <div className={styles.colFilterOptions} role="listbox" aria-label={label}>
        {options.map((opt) => (
          <button
            key={`${label}-${opt.value || "all"}`}
            type="button"
            className={`${styles.colFilterOption}${opt.value === value ? ` ${styles.colFilterOptionActive}` : ""}`}
            role="option"
            aria-selected={opt.value === value}
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => {
              onSelect(opt.value);
              onClose?.();
            }}
          >
            {opt.label}
          </button>
        ))}
      </div>
    </div>
  );
}

export function ProductsDensityView({
  rows,
  onRowClick,
  columnSort,
  onColumnSortChange,
}: ProductsDensityViewProps) {
  const trendOptions = useMemo<Option[]>(() => {
    return [
      { label: "Menor → maior", value: "asc" },
      { label: "Maior → menor", value: "desc" },
      { label: "Limpar", value: "" },
    ];
  }, []);

  const alphaOptions = useMemo<Option[]>(() => {
    return [
      { label: "A → Z", value: "asc" },
      { label: "Z → A", value: "desc" },
      { label: "Limpar", value: "" },
    ];
  }, []);

  const numericOptions = trendOptions;

  function setSort(key: ProductsDensityViewProps["columnSort"]["key"], direction: string) {
    onColumnSortChange({
      key,
      direction: direction === "asc" || direction === "desc" ? direction : "",
    });
  }

  const isActive = (key: ProductsDensityViewProps["columnSort"]["key"]) =>
    columnSort.key === key && !!columnSort.direction;

  const getDirection = (key: ProductsDensityViewProps["columnSort"]["key"]) =>
    columnSort.key === key ? columnSort.direction : "";
  return (
    <div className={styles.densityView}>
      <div className={styles.tableHeader}>
        <div className={styles.thCell}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>PN</span>
          <FilterPopover label="PN" active={isActive("pn")} align="start">
            {({ close }) => (
              <OptionsFilter
                label="PN"
                value={getDirection("pn")}
                options={alphaOptions}
                onSelect={(value) => setSort("pn", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={styles.thCell}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Produto</span>
          <FilterPopover label="Produto" active={isActive("product") || isActive("brand")} align="start">
            {({ close }) => (
              <OptionsFilter
                label="Produto"
                value={getDirection("product")}
                options={[
                  { label: "Produto A → Z", value: "asc" },
                  { label: "Produto Z → A", value: "desc" },
                  { label: "Marca A → Z", value: "brand:asc" },
                  { label: "Marca Z → A", value: "brand:desc" },
                  { label: "Limpar", value: "" },
                ]}
                onSelect={(value) => {
                  if (value.startsWith("brand:")) {
                    setSort("brand", value.endsWith("desc") ? "desc" : "asc");
                    return;
                  }
                  setSort("product", value);
                }}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={styles.thCell}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Preco</span>
          <FilterPopover label="Preco" active={isActive("price")}>
            {({ close }) => (
              <OptionsFilter
                label="Preco"
                value={getDirection("price")}
                options={numericOptions}
                onSelect={(value) => setSort("price", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={styles.thCell}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Mercado</span>
          <FilterPopover label="Mercado" active={isActive("market")}>
            {({ close }) => (
              <OptionsFilter
                label="Mercado"
                value={getDirection("market")}
                options={numericOptions}
                onSelect={(value) => setSort("market", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={`${styles.thCell} ${styles.colGap}`}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Gap %</span>
          <FilterPopover label="Gap" active={isActive("gap")}>
            {({ close }) => (
              <OptionsFilter
                label="Gap"
                value={getDirection("gap")}
                options={numericOptions}
                onSelect={(value) => setSort("gap", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={`${styles.thCell} ${styles.colMargin}`}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Margem</span>
          <FilterPopover label="Margem" active={isActive("margin")}>
            {({ close }) => (
              <OptionsFilter
                label="Margem"
                value={getDirection("margin")}
                options={numericOptions}
                onSelect={(value) => setSort("margin", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={styles.thCell}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Tendencia</span>
          <FilterPopover label="Tendencia" active={isActive("trend")}>
            {({ close }) => (
              <OptionsFilter
                label="Tendencia"
                value={getDirection("trend")}
                options={numericOptions}
                onSelect={(value) => setSort("trend", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={`${styles.thCell} ${styles.colClass}`}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Classe</span>
          <FilterPopover label="Classe" active={isActive("class")}>
            {({ close }) => (
              <OptionsFilter
                label="Classe"
                value={getDirection("class")}
                options={numericOptions}
                onSelect={(value) => setSort("class", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
        <div className={styles.thCell}>
          <div className={styles.thLabelRow}>
            <span className={styles.thLabel}>Estoque</span>
          <FilterPopover label="Estoque" active={isActive("stock")}>
            {({ close }) => (
              <OptionsFilter
                label="Estoque"
                value={getDirection("stock")}
                options={numericOptions}
                onSelect={(value) => setSort("stock", value)}
                onClose={close}
              />
            )}
          </FilterPopover>
          </div>
        </div>
      </div>
      <div>
        {rows.map((row) => (
          <button
            key={row.pn}
            type="button"
            className={styles.tableRow}
            onClick={() => onRowClick(row.pn)}
          >
            <div className={styles.productPn}>{row.pn}</div>
            <div>
              <div className={styles.productName}>{row.description}</div>
              <div className={styles.productDesc}>{row.brand} • {row.taxonomyLeafName}</div>
            </div>
            <div className={styles.productPrice}>{formatCurrency(row.price)}</div>
            <div className={styles.productPriceMuted}>{formatCurrency(row.marketPrice)}</div>
            <div className={`${styles.gapText} ${styles[row.gapTone]}`}>{row.gapLabel}</div>
            <div
              className={`${styles.marginText} ${
                row.metrics.marginPct > 26
                  ? styles.marginPositive
                  : row.metrics.marginPct >= 15
                    ? styles.marginWarn
                    : styles.marginNegative
              }`}
            >
              {row.metrics.marginPct.toFixed(1)}%
            </div>
            <div className={styles.productTrend}>
              <span className={`${styles.trendBadge} ${styles[row.trendTone]}`} title={row.trendMeta}>
                {row.trendLabel}
              </span>
              <div className={`${styles.miniSparkline} ${styles[row.trendColor]}`}>
                {row.trendSpark.map((height, index) => (
                  <div
                    key={`${row.pn}-spark-${index}`}
                    className={styles.miniSpark}
                    style={{ height: `${height}%` }}
                  />
                ))}
              </div>
            </div>
            <div>
              <span
                className={`${styles.classeBadge} ${
                  row.className.startsWith("A")
                    ? styles.classToneA
                    : row.className.startsWith("B")
                      ? styles.classToneB
                      : styles.classToneC
                }`}
              >
                {row.classLabel}
              </span>
            </div>
            <div className={styles.stockValue}>{row.stock} un</div>
          </button>
        ))}
      </div>
    </div>
  );
}
