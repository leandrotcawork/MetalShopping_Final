import type { AnalyticsSkuRow, SkuAction, SkuClass, SkuSpotlightModel, SkuStatus, SkuTrend, SkuXYZ } from "../contracts_products";

const taxonomyLeafCatalog = ["Perfis", "Chaparia", "Fixacao", "Ferragens", "Acabamentos", "Acessorios"];
const brands = ["MetalNobre", "Atlas", "Vexus", "PrimeFix", "Orbital", "Linea"];
const descriptors = ["Premium", "Reforcado", "Leve", "Industrial", "Compacto", "Ultra"];
const materials = ["Aco", "Inox", "Aluminio", "Liga", "Composto"];

function pick<T>(list: readonly T[], index: number): T {
  return list[index % list.length];
}

function statusFromIndex(index: number): SkuStatus {
  const cycle = index % 10;
  if (cycle <= 1) return "crit";
  if (cycle <= 4) return "warn";
  if (cycle <= 7) return "ok";
  return "info";
}

function actionFromStatus(status: SkuStatus, index: number): SkuAction {
  if (status === "crit") return index % 2 === 0 ? "PRUNE" : "MONITOR";
  if (status === "warn") return index % 2 === 0 ? "MONITOR" : "EXPAND";
  if (status === "ok") return index % 3 === 0 ? "EXPAND" : "MONITOR";
  return "MONITOR";
}

function trendFromIndex(index: number): SkuTrend {
  const cycle = index % 3;
  if (cycle === 0) return "up";
  if (cycle === 1) return "down";
  return "flat";
}

function xyzFromIndex(index: number): SkuXYZ {
  const cycle = index % 3;
  if (cycle === 0) return "X";
  if (cycle === 1) return "Y";
  return "Z";
}

function classFromIndex(index: number): SkuClass {
  const cycle = index % 3;
  if (cycle === 0) return "A";
  if (cycle === 1) return "B";
  return "C";
}

function buildRow(index: number): AnalyticsSkuRow {
  const ordinal = index + 1;
  const status = statusFromIndex(ordinal);
  const action = actionFromStatus(status, ordinal);
  const taxonomyLeafName = pick(taxonomyLeafCatalog, ordinal);
  const brand = pick(brands, ordinal * 2);
  const pn = `${18000 + ordinal}`;
  const ean = `7894200${(757700 + ordinal).toString()}`;
  const trend = trendFromIndex(ordinal);
  const className = classFromIndex(ordinal);
  const xyz6 = xyzFromIndex(ordinal);
  const xyz3 = xyzFromIndex(ordinal + 1);
  const trendPct = Number((((ordinal % 12) - 5) * 2.1).toFixed(1));

  const gapPct = Number((((ordinal % 13) - 5) * 1.7).toFixed(1));
  const marginPct = Number((19 + (ordinal % 9) * 1.35).toFixed(1));
  const pme6 = Number((20 + (ordinal % 17) * 3.1).toFixed(1));
  const giro6 = Number((0.7 + (ordinal % 12) * 0.24).toFixed(2));
  const dos6 = Math.round(18 + (ordinal % 20) * 4.4);
  const gmroi6 = Number((1.05 + (ordinal % 11) * 0.19).toFixed(2));
  const slope6 = Number((((ordinal % 15) - 7) * 0.12).toFixed(2));
  const cv6 = Number((0.11 + (ordinal % 14) * 0.045).toFixed(2));
  const dataQuality = Number((74 + (ordinal % 9) * 2.6).toFixed(1));
  const maturity = Number((58 + (ordinal % 13) * 2.3).toFixed(1));
  const dem3 = Math.round(26 + (ordinal % 18) * 5.2);
  const slope3 = Number((((ordinal % 11) - 5) * 0.14).toFixed(2));
  const stock = Math.max(0, Math.round(40 + (ordinal % 23) * 14 - (status === "crit" ? 70 : 0)));
  const price = Number((80 + (ordinal % 30) * 32.5 + (ordinal % 4) * 2.3).toFixed(2));
  const marketPrice = Number((price * (1 + gapPct / 100)).toFixed(2));
  const gapTone = gapPct < -1 ? "gapPositive" : gapPct > 1 ? "gapNegative" : "gapNeutral";
  const marginTone = marginPct >= 22 ? "gapPositive" : marginPct <= 15 ? "gapNegative" : "gapNeutral";
  const trendTone = trendPct > 1 ? "up" : trendPct < -1 ? "down" : "neutral";
  const trendColor = trendTone === "up" ? "trendGreen" : trendTone === "down" ? "trendRed" : "trendNeutral";
  const trendLabel = trendTone === "neutral" ? `→${Math.abs(trendPct).toFixed(1)}%` : `${trendPct > 0 ? "+" : ""}${trendPct.toFixed(1)}%`;
  const trendSpark = trendTone === "up" ? [30, 50, 70, 90] : trendTone === "down" ? [95, 75, 55, 35] : [50, 55, 50, 52];
  const classLabel = className === "A" ? (ordinal % 2 === 0 ? "A1" : "A2") : className === "B" ? (ordinal % 2 === 0 ? "B1" : "B2") : "C1";
  const classTone = classLabel.startsWith("A") ? (classLabel.endsWith("1") ? "classA1" : "classA2") : (classLabel.startsWith("B") ? (classLabel.endsWith("1") ? "classB1" : "classB2") : "classB2");

  return {
    pn,
    ean,
    description: `${pick(materials, ordinal)} ${pick(descriptors, ordinal + 1)} ${taxonomyLeafName} ${ordinal}`,
    taxonomyLeafName,
    brand,
    status,
    action,
    className,
    trend,
    trendPct,
    trendLabel,
    trendTone,
    trendColor,
    trendSpark,
    classLabel,
    classTone,
    stock,
    price,
    marketPrice,
    gapLabel: `${gapPct > 0 ? "+" : ""}${gapPct.toFixed(1)}%`,
    gapTone,
    marginTone,
    alertsCount: status === "crit" ? 2 : status === "warn" ? 1 : 0,
    metrics: {
      gapPct,
      marginPct,
      pme6,
      giro6,
      dos6,
      gmroi6,
      slope6,
      cv6,
      xyz6,
      dataQuality,
      maturity
    },
    short: {
      dem3,
      slope3,
      xyz3
    }
  };
}

function buildSpotlight(row: AnalyticsSkuRow): SkuSpotlightModel {
  const marketAvg = Number((row.marketPrice * (1 + row.metrics.gapPct / 100)).toFixed(2));
  const min = Number((marketAvg * 0.91).toFixed(2));
  const max = Number((marketAvg * 1.12).toFixed(2));

  return {
    pn: row.pn,
    description: row.description,
    taxonomyLeafName: row.taxonomyLeafName,
    brand: row.brand,
    status: row.status,
    action: row.action,
    whyMatters: `${row.pn} combina gap ${row.metrics.gapPct}% com DOS ${row.metrics.dos6}, impactando margem e capital imobilizado no ciclo 6M.`,
    recommendations: [
      `Ajustar faixa de preco em ate ${Math.max(1, Math.round(Math.abs(row.metrics.gapPct) * 0.6))}% por cluster competitivo.`,
      `Rebalancear cobertura para DOS alvo de ${Math.max(22, Math.round(row.metrics.dos6 * 0.78))} dias com reposicao semanal.`,
      `Priorizar monitoramento da hierarquia ${row.taxonomyLeafName} com foco em ${row.short.xyz3}-class no curto prazo.`
    ],
    nextSteps: [
      "Validar estrategia com comercial e compras.",
      "Publicar ajuste em lote no proximo ciclo.",
      "Revisar indicador apos 7 dias uteis."
    ],
    competition: {
      min,
      avg: marketAvg,
      max
    },
    metrics: row.metrics,
    short: row.short
  };
}

export const mockProducts: AnalyticsSkuRow[] = Array.from({ length: 60 }, (_, index) => buildRow(index));

export const mockProductsByPn: Record<string, SkuSpotlightModel> = Object.fromEntries(
  mockProducts.map((row) => [row.pn, buildSpotlight(row)])
);
