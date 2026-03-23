import type { PropsWithChildren } from "react";

type ChipProps = {
  label: string;
  tone?: "neutral" | "critical" | "warning" | "inform" | string;
};

export function Chip({ label }: ChipProps) {
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        borderRadius: 999,
        border: "1px solid rgba(148,163,184,0.45)",
        padding: "4px 10px",
        fontSize: 12,
        fontWeight: 700,
      }}
    >
      {label}
    </span>
  );
}

type CardProps = PropsWithChildren<{
  variant?: "default" | "glass";
  className?: string;
}>;

export function Card({ children, className = "" }: CardProps) {
  return <section className={className}>{children}</section>;
}
