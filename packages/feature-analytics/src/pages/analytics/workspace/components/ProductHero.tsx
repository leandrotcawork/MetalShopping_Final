import type { AnalyticsProductWorkspaceV1Dto } from "@metalshopping/feature-analytics";
import { HeroMetrics } from "./HeroMetrics";
import styles from "../product_workspace.module.css";

type ProductHeroProps = {
  model: AnalyticsProductWorkspaceV1Dto["model"];
};

function normalizeToken(value: string): string {
  return String(value || "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function pickMetricValue(
  items: Array<{ label: string; value: string }>,
  needle: string,
): string | null {
  const token = normalizeToken(needle);
  const found = items.find((item) => normalizeToken(item.label).includes(token));
  return found?.value ?? null;
}

export function ProductHero({ model }: ProductHeroProps) {
  const sourceMetrics = Array.isArray(model.heroMetrics) ? model.heroMetrics : [];
  const competitiveGapValue = model.competitiveness?.metrics?.[0]?.valueText || "-";
  const realPriceValue =
    pickMetricValue(sourceMetrics, "preco real efetivo") ||
    pickMetricValue(sourceMetrics, "preco real") ||
    "-";
  const variableSpendValue =
    pickMetricValue(sourceMetrics, "gasto var") ||
    pickMetricValue(sourceMetrics, "gasto variavel") ||
    (model.simulator?.variable_cost_unit_auto != null
      ? new Intl.NumberFormat("pt-BR", {
          style: "currency",
          currency: "BRL",
          maximumFractionDigits: 2,
        }).format(model.simulator.variable_cost_unit_auto)
      : "-");
  const orderedHeroMetrics = [
    { label: "Preco (nosso)", value: pickMetricValue(sourceMetrics, "preco (nosso)") || "-" },
    { label: "Mercado medio", value: pickMetricValue(sourceMetrics, "mercado medio") || "-" },
    { label: "Mercado max", value: pickMetricValue(sourceMetrics, "mercado max") || "-" },
    { label: "Custo medio", value: pickMetricValue(sourceMetrics, "custo medio") || "-" },
    { label: "Gasto var. (6m/un)", value: variableSpendValue },
    { label: "Preco real (efetivo)", value: realPriceValue },
    { label: "Gap real vs atual", value: competitiveGapValue },
    { label: "Mercado min", value: pickMetricValue(sourceMetrics, "mercado min") || "-" },
    { label: "Custo variavel", value: pickMetricValue(sourceMetrics, "custo variavel") || "-" },
    { label: "Estoque atual", value: pickMetricValue(sourceMetrics, "estoque atual") || "-" },
  ];

  return (
    <section className={styles.productHero}>
      <HeroMetrics items={orderedHeroMetrics} />
    </section>
  );
}
