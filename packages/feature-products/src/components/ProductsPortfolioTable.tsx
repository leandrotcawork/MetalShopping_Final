import type { ReactNode } from "react";

import { SortHeaderButton, SurfaceCard } from "@metalshopping/ui";

import type { ProductsPortfolioItem, ProductsPortfolioSortKey } from "../api";
import { formatCurrency } from "../view-model";
import styles from "../ProductsPortfolioPage.module.css";

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

type SortColumn = {
  key: ProductsPortfolioSortKey;
  label: string;
};

const sortColumns: SortColumn[] = [
  { key: "pn_interno", label: "PN" },
  { key: "name", label: "Produto" },
  { key: "brand_name", label: "Marca" },
  { key: "taxonomy_leaf0_name", label: "Taxonomia" },
  { key: "product_status", label: "Status" },
  { key: "current_price_amount", label: "Nosso Preço" },
  { key: "replacement_cost_amount", label: "Custos" },
];

export function ProductsPortfolioTable(props: {
  rows: ProductsPortfolioItem[];
  loading: boolean;
  taxonomyLeaf0Label: string;
  sortIndicator: (key: ProductsPortfolioSortKey) => string;
  onSort: (key: ProductsPortfolioSortKey) => void;
  allVisibleSelected: boolean;
  selectionMode: "explicit" | "filtered";
  selectedProductIds: string[];
  onToggleCurrentPage: () => void;
  onToggleRow: (productId: string) => void;
  actions: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <SurfaceCard title="Produtos Cadastrados" actions={props.actions} className={styles.tableCard}>
      {props.children}

      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.checkboxColumn}>
                <ProductSelectionCheckbox
                  checked={props.allVisibleSelected}
                  disabled={props.rows.length === 0 || props.selectionMode === "filtered"}
                  label="Selecionar produtos da página"
                  onChange={props.onToggleCurrentPage}
                />
              </th>
              {sortColumns.map((column) => (
                <th key={column.key}>
                  <SortHeaderButton indicator={props.sortIndicator(column.key)} onClick={() => props.onSort(column.key)}>
                    {column.key === "taxonomy_leaf0_name" ? props.taxonomyLeaf0Label : column.label}
                  </SortHeaderButton>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {props.rows.length > 0 ? (
              props.rows.map((row) => {
                const checked = props.selectionMode === "filtered" || props.selectedProductIds.includes(row.product_id);
                return (
                  <tr key={row.product_id}>
                    <td className={styles.checkboxColumn}>
                      <ProductSelectionCheckbox
                        checked={checked}
                        disabled={props.selectionMode === "filtered"}
                        label={`Selecionar produto ${row.name}`}
                        onChange={() => props.onToggleRow(row.product_id)}
                      />
                    </td>
                    <td>
                      <span className={styles.cellStrong}>{row.pn_interno ?? "—"}</span>
                    </td>
                    <td>
                      <span className={styles.cellStrong}>{row.name}</span>
                      <span className={styles.cellSmall}>Ref: {row.reference ?? "—"} · EAN: {row.ean ?? "—"}</span>
                    </td>
                    <td>{row.brand_name ?? "—"}</td>
                    <td>
                      <span className={styles.cellStrong}>{row.taxonomy_leaf0_name ?? "—"}</span>
                      <span className={styles.cellMeta}>{row.taxonomy_leaf_name ?? "Sem folha definida."}</span>
                    </td>
                    <td>
                      <span className={styles.cellStrong}>
                        {row.product_status ? (
                          <span className={`${styles.statusChip} ${styles[`statusChip${statusTone(row.product_status)}`]}`}>
                            {statusLabel(row.product_status)}
                          </span>
                        ) : (
                          "—"
                        )}
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
                  </tr>
                );
              })
            ) : (
              <tr>
                <td className={styles.empty} colSpan={7}>
                  {props.loading ? "Carregando produtos..." : "Nenhum produto encontrado para o filtro atual."}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {props.footer}
    </SurfaceCard>
  );
}
