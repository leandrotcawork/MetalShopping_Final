import { useEffect, useMemo, useState } from "react";

type AnimatedNumberProps = {
  value: string | number;
  className?: string;
  durationMs?: number;
  as?: "span" | "strong";
};

type ParsedValue = {
  prefix: string;
  suffix: string;
  numberValue: number;
  decimals: number;
  decimalSep: "." | ",";
};

function parseNumeric(value: string | number): ParsedValue | null {
  if (typeof value === "number" && Number.isFinite(value)) {
    return { prefix: "", suffix: "", numberValue: value, decimals: 0, decimalSep: "." };
  }

  const text = String(value);
  const match = text.match(/-?\d[\d.,]*/);
  if (!match || match.index == null) return null;

  const token = match[0];
  const prefix = text.slice(0, match.index);
  const suffix = text.slice(match.index + token.length);

  const hasDot = token.includes(".");
  const hasComma = token.includes(",");
  const decimalSep: "." | "," = hasComma && !hasDot ? "," : ".";

  let decimals = 0;
  if (hasComma && hasDot) {
    const commaIndex = token.lastIndexOf(",");
    decimals = token.length - commaIndex - 1;
  } else if (decimalSep === "," && hasComma) {
    decimals = token.length - token.lastIndexOf(",") - 1;
  } else if (decimalSep === "." && hasDot) {
    decimals = token.length - token.lastIndexOf(".") - 1;
  }

  let normalized = token;
  if (hasComma && hasDot) {
    normalized = token.replace(/\./g, "").replace(",", ".");
  } else if (decimalSep === ",") {
    normalized = token.replace(/\./g, "").replace(",", ".");
  } else {
    normalized = token.replace(/,/g, "");
  }

  const numberValue = Number.parseFloat(normalized);
  if (!Number.isFinite(numberValue)) return null;

  return { prefix, suffix, numberValue, decimals, decimalSep };
}

function formatNumeric(value: number, parsed: ParsedValue): string {
  const fixed = parsed.decimals > 0 ? value.toFixed(parsed.decimals) : Math.round(value).toString();
  const withSep = parsed.decimalSep === "," ? fixed.replace(".", ",") : fixed;
  return `${parsed.prefix}${withSep}${parsed.suffix}`;
}

export function AnimatedNumber({ value, className, durationMs = 800, as = "span" }: AnimatedNumberProps) {
  const parsed = useMemo(() => parseNumeric(value), [value]);
  const [display, setDisplay] = useState(() => (parsed ? formatNumeric(0, parsed) : String(value)));

  useEffect(() => {
    if (!parsed) {
      setDisplay(String(value));
      return;
    }

    const prefersReduce = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
    if (prefersReduce) {
      setDisplay(formatNumeric(parsed.numberValue, parsed));
      return;
    }

    let raf = 0;
    const start = performance.now();

    const tick = (now: number) => {
      const elapsed = now - start;
      const t = Math.min(1, elapsed / durationMs);
      const eased = 1 - Math.pow(1 - t, 3);
      const current = parsed.numberValue * eased;
      setDisplay(formatNumeric(current, parsed));
      if (t < 1) raf = window.requestAnimationFrame(tick);
    };

    raf = window.requestAnimationFrame(tick);
    return () => {
      if (raf) window.cancelAnimationFrame(raf);
    };
  }, [parsed, durationMs, value]);

  if (as === "strong") {
    return <strong className={className}>{display}</strong>;
  }

  return <span className={className}>{display}</span>;
}

