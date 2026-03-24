import type { SkuStatus } from "../../contracts_products";
import styles from "../../analytics_products.module.css";
import { FilterDropdown, type SelectMenuOption } from "@metalshopping/ui";

type Filters = {
  brand: string[];
  status: SkuStatus[];
  taxonomyLeafName: string[];
};

type ProductsFiltersBarProps = {
  filters: Filters;
  brands: string[];
  statuses: { label: string; value: SkuStatus }[];
  taxonomyLeafs: string[];
  onChange: (next: Filters) => void;
};

function toggleMultiValue(current: string[], value: string): string[] {
  if (!value || value === "all") return [];
  if (current.includes(value)) return current.filter((item) => item !== value);
  return [...current, value];
}

export function ProductsFiltersBar({ filters, brands, statuses, taxonomyLeafs, onChange }: ProductsFiltersBarProps) {
  const brandOptions: SelectMenuOption[] = [{ label: "Todas Marcas", value: "all" }, ...brands.map((brand) => ({ label: brand, value: brand }))];
  const statusOptions: SelectMenuOption[] = [{ label: "Todos Status", value: "all" }, ...statuses.map((status) => ({ label: status.label, value: status.value }))];
  const taxonomyLeafOptions: SelectMenuOption[] = [{ label: "Todos Grupos", value: "all" }, ...taxonomyLeafs.map((taxonomyLeaf) => ({ label: taxonomyLeaf, value: taxonomyLeaf }))];

  return (
    <div className={styles.filtersBar}>
      <div className={styles.filterField}>
        <span className={styles.filterFieldLabel}>Marca</span>
        <FilterDropdown
          id="filter-brand"
          selectionMode="duo"
          value=""
          values={filters.brand}
          options={brandOptions}
          classNamesOverrides={{
            wrap: `spotlight-select-wrap ${styles.selectWrap}`,
            trigger: styles.filterSelectTrigger,
          }}
          onSelect={(value) => onChange({ ...filters, brand: toggleMultiValue(filters.brand, value) })}
        />
      </div>
      <div className={styles.filterField}>
        <span className={styles.filterFieldLabel}>Grupos</span>
        <FilterDropdown
          id="filter-taxonomy"
          selectionMode="duo"
          value=""
          values={filters.taxonomyLeafName}
          options={taxonomyLeafOptions}
          classNamesOverrides={{
            wrap: `spotlight-select-wrap ${styles.selectWrap}`,
            trigger: styles.filterSelectTrigger,
          }}
          onSelect={(value) => onChange({ ...filters, taxonomyLeafName: toggleMultiValue(filters.taxonomyLeafName, value) })}
        />
      </div>
      <div className={styles.filterField}>
        <span className={styles.filterFieldLabel}>Status</span>
        <FilterDropdown
          id="filter-status"
          selectionMode="duo"
          value=""
          values={filters.status}
          options={statusOptions}
          classNamesOverrides={{
            wrap: `spotlight-select-wrap ${styles.selectWrap}`,
            trigger: styles.filterSelectTrigger,
          }}
          onSelect={(value) => onChange({ ...filters, status: toggleMultiValue(filters.status, value) as SkuStatus[] })}
        />
      </div>
    </div>
  );
}
