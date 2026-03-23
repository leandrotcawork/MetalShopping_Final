import { Button } from "@metalshopping/ui";

import styles from "../ProductsPortfolioPage.module.css";

export function ProductsSelectionBar(props: {
  rowsCount: number;
  allVisibleSelected: boolean;
  totalSelected: number;
  selectedSuppliersLabel: string;
  selectionMode: "explicit" | "filtered";
  mode?: "actions" | "summary" | "full";
  exportDisabled: boolean;
  onExport: () => void;
  onToggleCurrentPage: () => void;
  onSelectFiltered: () => void;
  onClearSelection: () => void;
}) {
  const mode = props.mode ?? "full";

  return (
    <>
      {mode !== "summary" ? (
        <div className={styles.tableActions}>
          <Button
            className={styles.secondaryActionButton}
            variant="secondary"
            disabled={props.exportDisabled}
            onClick={props.onExport}
          >
            Exportar selecionados
          </Button>
          <Button
            className={styles.secondaryActionButton}
            variant="secondary"
            disabled={props.rowsCount === 0}
            onClick={props.onToggleCurrentPage}
          >
            {props.allVisibleSelected ? "Desmarcar página" : "Selecionar página"}
          </Button>
          <Button
            className={styles.secondaryActionButton}
            variant="secondary"
            disabled={props.rowsCount === 0}
            onClick={props.onSelectFiltered}
          >
            Selecionar filtrados
          </Button>
          <Button
            className={styles.secondaryActionButton}
            variant="secondary"
            disabled={props.totalSelected === 0}
            onClick={props.onClearSelection}
          >
            Limpar
          </Button>
        </div>
      ) : null}

      {mode !== "actions" ? (
        <div className={styles.selectionRow}>
          <span>
            Modo: <strong>{props.selectionMode === "filtered" ? "Filtrados" : "Explícito"}</strong>
          </span>
          <span>
            Itens: <strong>{props.totalSelected}</strong>
          </span>
          <span>
            Fornecedores: <strong>{props.selectedSuppliersLabel}</strong>
          </span>
        </div>
      ) : null}
    </>
  );
}
