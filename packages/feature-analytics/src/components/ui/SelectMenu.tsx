export type SelectMenuOption = {
  label: string;
  value: string;
};

type SelectMenuProps = {
  id?: string;
  value?: string;
  values?: string[];
  options: SelectMenuOption[];
  mode?: "single" | "multi";
  onSelect: (value: string) => void;
  classNames?: {
    wrap?: string;
  };
};

export function SelectMenu({
  id,
  value = "",
  values = [],
  options,
  mode = "single",
  onSelect,
  classNames,
}: SelectMenuProps) {
  const selectedValue = mode === "multi" ? values[0] || "all" : value || "all";

  return (
    <div className={classNames?.wrap}>
      <select id={id} value={selectedValue} onChange={(event) => onSelect(event.target.value)} style={{ width: "100%" }}>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
}
