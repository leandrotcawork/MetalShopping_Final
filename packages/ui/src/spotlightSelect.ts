import type { SelectMenuClassNames } from "./SelectMenu";
import styles from "./spotlightSelect.module.css";

function mergeClassName(base: string, override?: string): string {
  const extra = String(override || "").trim();
  return extra ? `${base} ${extra}` : base;
}

export function createSpotlightSelectClassNames(
  overrides: Partial<SelectMenuClassNames> = {},
): SelectMenuClassNames {
  return {
    wrap: mergeClassName(styles.wrap, overrides.wrap),
    trigger: mergeClassName(styles.trigger, overrides.trigger),
    value: mergeClassName(styles.value, overrides.value),
    chevron: mergeClassName(styles.chevron, overrides.chevron),
    menu: mergeClassName(styles.menu, overrides.menu),
    option: mergeClassName(styles.option, overrides.option),
    optionActive: mergeClassName(styles.optionActive, overrides.optionActive),
    searchWrap: mergeClassName(styles.searchWrap, overrides.searchWrap),
    searchInput: mergeClassName(styles.searchInput, overrides.searchInput),
    emptyState: mergeClassName(styles.emptyState, overrides.emptyState),
    label: overrides.label,
  };
}
