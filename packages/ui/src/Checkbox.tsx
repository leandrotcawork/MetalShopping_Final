import styles from "./Checkbox.module.css";

type CheckboxProps = {
  checked: boolean;
  onChange: (nextChecked?: boolean) => void;
  label?: string;
  ariaLabel?: string;
  disabled?: boolean;
  className?: string;
  id?: string;
};

export function Checkbox({
  checked,
  onChange,
  label,
  ariaLabel,
  disabled = false,
  className = "",
  id,
}: CheckboxProps) {
  const wrapperClassName = `${styles.checkbox} ${disabled ? styles.disabled : ""} ${className}`.trim();

  return (
    <label className={wrapperClassName}>
      <input
        id={id}
        type="checkbox"
        className={styles.input}
        checked={checked}
        disabled={disabled}
        aria-label={ariaLabel}
        onChange={(event) => onChange(event.target.checked)}
      />
      {label ? <span className={styles.label}>{label}</span> : null}
    </label>
  );
}
