import { useMemo } from "react";

export type SelectMenuOption = {
  label: string;
  value: string;
};

type FilterDropdownProps = {
  id?: string;
  value?: string;
  values?: string[];
  options: SelectMenuOption[];
  selectionMode?: "single" | "duo";
  onSelect: (value: string) => void;
  classNamesOverrides?: {
    wrap?: string;
    trigger?: string;
  };
};

export function FilterDropdown({
  id,
  value = "",
  values = [],
  options,
  selectionMode = "single",
  onSelect,
  classNamesOverrides,
}: FilterDropdownProps) {
  const selectedValue = useMemo(() => {
    if (selectionMode === "duo") {
      return values[0] || "all";
    }
    return value || "all";
  }, [selectionMode, value, values]);

  return (
    <div className={classNamesOverrides?.wrap}>
      <select
        id={id}
        value={selectedValue}
        onChange={(event) => onSelect(event.target.value)}
        className={classNamesOverrides?.trigger}
        style={{ width: "100%" }}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
}
