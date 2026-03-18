import type { ButtonHTMLAttributes, PropsWithChildren } from "react";

import styles from "./Button.module.css";

type ButtonVariant = "primary" | "secondary" | "quiet";

type ButtonProps = PropsWithChildren<
  ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: ButtonVariant;
  }
>;

export function Button({ variant = "secondary", className = "", children, ...props }: ButtonProps) {
  const variantClass = styles[variant];
  return (
    <button
      type="button"
      {...props}
      className={`${styles.button} ${variantClass} ${className}`.trim()}
    >
      {children}
    </button>
  );
}
