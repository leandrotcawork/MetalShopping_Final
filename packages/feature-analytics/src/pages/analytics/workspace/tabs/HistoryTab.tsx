// @ts-nocheck
import { useOutletContext } from "react-router-dom";

import type { ProductWorkspaceOutletContext } from "../ProductWorkspaceLayout";
import { PriceMarketChart } from "./history/PriceMarketChart";
import { SalesMonthlyChart } from "./history/SalesMonthlyChart";
import styles from "./history.module.css";

type IndicatorTone = "good" | "warn" | "bad" | "neutral";

function getIndicatorTone(label: string, value: string): IndicatorTone {
  const token = String(value || "").trim().toLowerCase();
  const rawLabel = String(label || "").trim().toLowerCase();

  if (rawLabel.includes("tendencia")) {
    if (token.includes("subindo") || token.includes("up")) return "good";
    if (token.includes("caindo") || token.includes("down")) return "bad";
    return "neutral";
  }

  if (rawLabel.includes("volatilidade")) {
    if (token.includes("baixa")) return "good";
    if (token.includes("media") || token.includes("média")) return "warn";
    if (token.includes("alta")) return "bad";
    return "neutral";
  }

  if (rawLabel === "xyz") {
    if (token === "x") return "good";
    if (token === "y") return "warn";
    if (token === "z") return "bad";
    return "neutral";
  }

  if (rawLabel.includes("forca") || rawLabel.includes("força")) {
    const num = Number(token.replace("%", "").replace(",", "."));
    if (!Number.isFinite(num)) return "neutral";
    if (num >= 75) return "good";
    if (num >= 45) return "warn";
    return "bad";
  }

  return "neutral";
}

function toneTextClass(tone: IndicatorTone): string {
  if (tone === "good") return styles.toneGood;
  if (tone === "warn") return styles.toneWarn;
  if (tone === "bad") return styles.toneBad;
  return styles.toneNeutral;
}

function toneFillClass(tone: IndicatorTone): string {
  if (tone === "good") return styles.fillGood;
  if (tone === "warn") return styles.fillWarn;
  if (tone === "bad") return styles.fillBad;
  return styles.fillNeutral;
}

export function HistoryTab() {
  const { model } = useOutletContext<ProductWorkspaceOutletContext>();
  const history = model.history;

  return (
    <section className={styles.historyRoot}>
      <section className={styles.historyIndicators}>
        {history.indicators.map((item) => {
          const tone = getIndicatorTone(item.label, item.value);
          return (
            <article key={item.label} className={styles.indicatorCard}>
              <div className={styles.indicatorLabel}>{item.label}</div>
              <div className={`${styles.indicatorValue} ${styles.mono} ${toneTextClass(tone)}`}>{item.value}</div>
              <div className={styles.indicatorBar}>
                <div
                  className={`${styles.indicatorFill} ${toneFillClass(tone)}`}
                  style={{ width: `${item.fill_pct}%` }}
                />
              </div>
            </article>
          );
        })}
      </section>

      <section className={styles.historyInfoBar}>
        {[
          { label: "Janela Preco", value: history.meta.price_window },
          { label: "Janela Vendas", value: history.meta.sales_window },
          { label: "Ultima atualizacao", value: history.meta.last_updated },
          { label: "Cobertura", value: history.meta.coverage },
        ].map((item) => (
          <article key={item.label} className={styles.infoItem}>
            <span className={styles.infoLabel}>{item.label}</span>
            <span className={`${styles.infoValue} ${styles.mono}`}>{item.value}</span>
          </article>
        ))}
      </section>

      <section className={styles.chartsGrid}>
        <PriceMarketChart points={history.price_series} supplierLinks={history.supplier_links} />
        <SalesMonthlyChart points={history.sales_monthly} />
      </section>
    </section>
  );
}
