import { Button } from "@metalshopping/ui";

import styles from "../ProductsPortfolioPage.module.css";

export function ProductsPaginationBar(props: {
  currentPage: number;
  totalPages: number;
  totalMatching: number;
  limit: number;
  pageSizeOptions: number[];
  canGoPrevious: boolean;
  canGoNext: boolean;
  onChangeLimit: (limit: number) => void;
  onPrevious: () => void;
  onNext: () => void;
}) {
  return (
    <div className={styles.paginationRow}>
      <div className={styles.paginationStatus}>
        <span>
          Página {props.currentPage} de {props.totalPages}
        </span>
        <span>{props.totalMatching} produtos encontrados</span>
      </div>

      <div className={styles.paginationActions}>
        <label className={styles.pageSizeField}>
          <span>Linhas</span>
          <select
            className={styles.pageSizeSelect}
            value={props.limit}
            onChange={(event) => props.onChangeLimit(Number(event.target.value))}
          >
            {props.pageSizeOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>

        <Button className={styles.paginationButton} variant="secondary" disabled={!props.canGoPrevious} onClick={props.onPrevious}>
          Anterior
        </Button>
        <Button className={`${styles.paginationButton} ${styles.paginationButtonPrimary}`} variant="primary" disabled={!props.canGoNext} onClick={props.onNext}>
          Próxima
        </Button>
      </div>
    </div>
  );
}
