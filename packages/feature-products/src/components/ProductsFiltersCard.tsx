import { Button, FilterDropdown, type SelectMenuOption, SurfaceCard } from "@metalshopping/ui";

import type { ProductsPortfolioQuery } from "../api";
import styles from "../ProductsPortfolioPage.module.css";

function toggleMultiSelection(current: string[], next: string): string[] {
  if (next === "") {
    return [];
  }
  if (current.includes(next)) {
    return current.filter((value) => value !== next);
  }
  return [...current, next];
}

export function ProductsFiltersCard(props: {
  searchDraft: string;
  onSearchDraftChange: (value: string) => void;
  query: ProductsPortfolioQuery;
  taxonomyLeaf0Label: string;
  brandOptions: SelectMenuOption[];
  taxonomyOptions: SelectMenuOption[];
  statusOptions: SelectMenuOption[];
  activeFilters: Array<{ key: string; label: string }>;
  onChangeQuery: (next: ProductsPortfolioQuery) => void;
  onClearAll: () => void;
}) {
  return (
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
            value={props.searchDraft}
            placeholder="PN, referência, EAN, descrição"
            onChange={(event) => props.onSearchDraftChange(event.target.value)}
          />
        </label>

        <div className={styles.field}>
          <span className={styles.label}>Marca</span>
          <FilterDropdown
            id="products-brand-filter"
            values={props.query.brand_name}
            selectionMode="duo"
            options={props.brandOptions}
            onSelect={(value) =>
              props.onChangeQuery({
                ...props.query,
                brand_name: toggleMultiSelection(props.query.brand_name, value),
                offset: 0,
              })
            }
          />
        </div>

        <div className={styles.field}>
          <span className={styles.label}>Status</span>
          <FilterDropdown
            id="products-status-filter"
            values={props.query.status}
            selectionMode="duo"
            options={props.statusOptions}
            onSelect={(value) =>
              props.onChangeQuery({
                ...props.query,
                status: toggleMultiSelection(props.query.status, value),
                offset: 0,
              })
            }
          />
        </div>

        <div className={styles.field}>
          <span className={styles.label}>{props.taxonomyLeaf0Label}</span>
          <FilterDropdown
            id="products-taxonomy-filter"
            values={props.query.taxonomy_leaf0_name}
            selectionMode="duo"
            options={props.taxonomyOptions}
            onSelect={(value) =>
              props.onChangeQuery({
                ...props.query,
                taxonomy_leaf0_name: toggleMultiSelection(props.query.taxonomy_leaf0_name, value),
                offset: 0,
              })
            }
          />
        </div>
      </div>

      <div className={styles.filterFooter}>
        <div className={styles.filterChips}>
          {props.activeFilters.length > 0 ? (
            props.activeFilters.map((filter) => (
              <button
                key={filter.key}
                type="button"
                className={styles.filterChip}
                onClick={() => {
                  if (filter.key === "search") {
                    props.onSearchDraftChange("");
                  }

                  props.onChangeQuery({
                    ...props.query,
                    [filter.key]: filter.key === "search" ? "" : [],
                    offset: 0,
                  } as ProductsPortfolioQuery);
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

        <Button
          className={styles.clearButton}
          variant="quiet"
          disabled={props.activeFilters.length === 0}
          onClick={props.onClearAll}
        >
          Limpar filtros
        </Button>
      </div>
    </SurfaceCard>
  );
}
