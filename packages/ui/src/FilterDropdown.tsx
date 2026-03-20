import { useMemo } from "react";

import {
  SelectMenu,
  type SelectMenuClassNames,
  type SelectMenuOption,
} from "./SelectMenu";
import { createSpotlightSelectClassNames } from "./spotlightSelect";

type FilterDropdownProps = {
  id: string;
  options: SelectMenuOption[];
  onSelect: (value: string) => void;
  label?: string;
  allLabel?: string;
  value?: string;
  values?: string[];
  selectionMode?: "one" | "duo";
  disabled?: boolean;
  searchThreshold?: number;
  closeOnSelectInDuo?: boolean;
  chevronStrokeWidth?: number;
  classNamesOverrides?: Partial<SelectMenuClassNames>;
};

export type { SelectMenuOption };

export function FilterDropdown({
  id,
  options,
  onSelect,
  label,
  allLabel,
  value = "",
  values = [],
  selectionMode = "one",
  disabled = false,
  searchThreshold = 10,
  closeOnSelectInDuo = false,
  chevronStrokeWidth = 2,
  classNamesOverrides = {},
}: FilterDropdownProps) {
  const resolvedOptions = useMemo(() => {
    if (!allLabel || options.some((option) => option.value === "all")) {
      return options;
    }
    return [{ value: "all", label: allLabel }, ...options];
  }, [allLabel, options]);

  const classNames = useMemo(
    () => createSpotlightSelectClassNames(classNamesOverrides),
    [classNamesOverrides],
  );

  return (
    <SelectMenu
      id={id}
      label={label}
      options={resolvedOptions}
      onSelect={onSelect}
      value={value}
      values={values}
      mode={selectionMode === "duo" ? "multi" : "single"}
      classNames={classNames}
      disabled={disabled}
      searchThreshold={searchThreshold}
      closeOnMultiSelect={closeOnSelectInDuo}
      chevronStrokeWidth={chevronStrokeWidth}
    />
  );
}
