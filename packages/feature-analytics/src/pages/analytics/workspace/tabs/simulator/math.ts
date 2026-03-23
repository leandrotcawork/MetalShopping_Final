export type IndicatorTone = "good" | "warn" | "bad" | "neutral";

export type CurrentSnapshot = {
  pn: string;
  title: string;
  brand: string;
  taxonomyLeafName: string;
  priceCurrent: number | null;
  costAvg: number | null;
  costVariable: number | null;
  variableCostUnitAuto: number | null;
  variableCostSource: "AUTO_6M" | "AUTO_12M" | "AUTO" | "NONE_GROSS_FALLBACK";
  variableCostCoverageMonths: number;
  marginCalcMode: string;
  marginIsFallback: boolean;
  marketAvg: number | null;
  marketMin: number | null;
  marketMax: number | null;
  marginPctCurrent: number | null;
  markupPctCurrent: number | null;
  gmroiCurrent: number | null;
  gapPctCurrent: number | null;
};

export type SimulationInputs = {
  price: number | null;
  discountPct: number | null;
  targetMarginPct: number | null;
  freightAdjPct: number;
};

export type SimulationAlert = {
  level: "critical" | "warn" | "info";
  title: string;
  message: string;
};

export type SimulationResult = {
  priceNew: number | null;
  effectiveCost: number | null;
  discountPct: number | null;
  deltaPriceAbs: number | null;
  deltaPricePct: number | null;
  marginAmount: number | null;
  marginPct: number | null;
  markupPct: number | null;
  gmroi: number | null;
  gapPct: number | null;
  breakEvenPrice: number | null;
  marketPositionPct: number | null;
  tones: {
    margin: IndicatorTone;
    markup: IndicatorTone;
    gmroi: IndicatorTone;
    gap: IndicatorTone;
    discount: IndicatorTone;
  };
  alerts: SimulationAlert[];
  statusBadges: Array<{
    label: string;
    tone: "success" | "warning" | "info";
  }>;
};

type PriceBounds = {
  min: number;
  max: number;
};

function clamp(value: number, min: number, max: number): number {
  if (!Number.isFinite(value)) return min;
  return Math.max(min, Math.min(max, value));
}

function normalize(value: number | null): number | null {
  if (value == null || !Number.isFinite(value)) return null;
  return value;
}

function defaultFreightPctFromCurrent(current: CurrentSnapshot): number {
  if (
    current.costAvg == null ||
    !Number.isFinite(current.costAvg) ||
    current.costAvg <= 0 ||
    current.variableCostUnitAuto == null ||
    !Number.isFinite(current.variableCostUnitAuto)
  ) {
    return 0;
  }
  return clamp((current.variableCostUnitAuto / current.costAvg) * 100, 0, 300);
}

function effectiveCostFromInputs(current: CurrentSnapshot, freightAdjPct: number): number | null {
  if (current.costAvg == null) return null;
  const freightAdj = clamp(freightAdjPct, 0, 300) / 100;
  const freightCost = current.costAvg * freightAdj;
  return current.costAvg + freightCost;
}

function toneByBands(
  value: number | null,
  bands: { warn: number; good: number },
  options?: { higherIsBetter?: boolean },
): IndicatorTone {
  if (value == null || !Number.isFinite(value)) return "neutral";
  const higherIsBetter = options?.higherIsBetter !== false;
  if (higherIsBetter) {
    if (value >= bands.good) return "good";
    if (value >= bands.warn) return "warn";
    return "bad";
  }
  if (value <= bands.good) return "good";
  if (value <= bands.warn) return "warn";
  return "bad";
}

export function derivePriceBounds(current: CurrentSnapshot): PriceBounds {
  const candidates: number[] = [];
  if (current.marketMin != null) candidates.push(current.marketMin * 0.8);
  if (current.costAvg != null) {
    const freightPct = defaultFreightPctFromCurrent(current);
    candidates.push((current.costAvg * (1 + freightPct / 100)) * 0.9);
  }
  if (current.priceCurrent != null) candidates.push(current.priceCurrent * 0.65);
  const min = Math.max(1, Math.min(...(candidates.length ? candidates : [200])));

  const maxCandidates: number[] = [];
  if (current.marketMax != null) maxCandidates.push(current.marketMax * 1.2);
  if (current.priceCurrent != null) maxCandidates.push(current.priceCurrent * 1.35);
  if (current.costAvg != null) {
    const freightPct = defaultFreightPctFromCurrent(current);
    maxCandidates.push((current.costAvg * (1 + freightPct / 100)) * 2.4);
  }
  const max = Math.max(min + 50, Math.max(...(maxCandidates.length ? maxCandidates : [1200])));
  return { min, max };
}

export function deriveDiscountFromPrice(priceCurrent: number | null, price: number | null): number | null {
  if (priceCurrent == null || price == null || priceCurrent <= 0) return null;
  return ((priceCurrent - price) / priceCurrent) * 100;
}

export function derivePriceFromDiscount(priceCurrent: number | null, discountPct: number | null): number | null {
  if (priceCurrent == null || discountPct == null) return null;
  return priceCurrent * (1 - clamp(discountPct, 0, 90) / 100);
}

export function derivePriceFromMargin(
  costAvg: number | null,
  variableCostUnitAuto: number | null,
  freightAdjPct: number,
  marginPct: number | null,
): number | null {
  if (costAvg == null || marginPct == null) return null;
  const margin = clamp(marginPct, 0, 90);
  const freightCost = costAvg * (clamp(freightAdjPct, 0, 300) / 100);
  const effectiveCost = costAvg + freightCost;
  const denominator = 1 - margin / 100;
  if (denominator <= 0) return null;
  return effectiveCost / denominator;
}

export function deriveMarginFromPrice(
  costAvg: number | null,
  variableCostUnitAuto: number | null,
  freightAdjPct: number,
  price: number | null,
): number | null {
  if (costAvg == null || price == null || price <= 0) return null;
  const freightCost = costAvg * (clamp(freightAdjPct, 0, 300) / 100);
  const effectiveCost = costAvg + freightCost;
  return ((price - effectiveCost) / price) * 100;
}

export function defaultSimulationInputs(current: CurrentSnapshot): SimulationInputs {
  const freightAdjPct = defaultFreightPctFromCurrent(current);
  const targetMarginPct = deriveMarginFromPrice(
    current.costAvg,
    current.variableCostUnitAuto,
    freightAdjPct,
    current.priceCurrent,
  );
  return {
    price: normalize(current.priceCurrent),
    discountPct: 0,
    targetMarginPct: normalize(targetMarginPct),
    freightAdjPct,
  };
}

export function calculateSimulation(current: CurrentSnapshot, inputs: SimulationInputs): SimulationResult {
  const priceNew = normalize(inputs.price);
  const discountPct = normalize(inputs.discountPct);
  const effectiveCost = effectiveCostFromInputs(current, inputs.freightAdjPct);

  const marginAmount =
    priceNew != null && effectiveCost != null
      ? priceNew - effectiveCost
      : null;
  const marginPct =
    marginAmount != null && priceNew != null && priceNew > 0
      ? (marginAmount / priceNew) * 100
      : null;
  const markupPct =
    current.costAvg != null && priceNew != null && current.costAvg > 0
      ? ((priceNew / current.costAvg) - 1) * 100
      : null;
  const gmroi =
    current.gmroiCurrent != null &&
    current.marginPctCurrent != null &&
    current.marginPctCurrent > 0 &&
    marginPct != null
      ? current.gmroiCurrent * (marginPct / current.marginPctCurrent)
      : null;
  const gapPct =
    current.marketAvg != null && current.marketAvg > 0 && priceNew != null
      ? ((priceNew - current.marketAvg) / current.marketAvg) * 100
      : null;

  const deltaPriceAbs =
    priceNew != null && current.priceCurrent != null
      ? priceNew - current.priceCurrent
      : null;
  const deltaPricePct =
    deltaPriceAbs != null && current.priceCurrent != null && current.priceCurrent > 0
      ? (deltaPriceAbs / current.priceCurrent) * 100
      : null;

  const marketPositionPct =
    current.marketMin != null &&
    current.marketMax != null &&
    current.marketMax > current.marketMin &&
    priceNew != null
      ? clamp(((priceNew - current.marketMin) / (current.marketMax - current.marketMin)) * 100, 0, 100)
      : null;

  const breakEvenPrice = effectiveCost != null ? effectiveCost * 1.01 : null;

  const tones = {
    margin: toneByBands(marginPct, { warn: 15, good: 30 }),
    markup: toneByBands(markupPct, { warn: 20, good: 50 }),
    gmroi: toneByBands(gmroi, { warn: 0.9, good: 1.3 }),
    gap: toneByBands(Math.abs(gapPct ?? Number.NaN), { warn: 10, good: 4 }, { higherIsBetter: false }),
    discount: toneByBands(discountPct, { warn: 25, good: 12 }, { higherIsBetter: false }),
  } satisfies SimulationResult["tones"];

  const alerts: SimulationAlert[] = [];
  if (priceNew != null && effectiveCost != null && priceNew <= effectiveCost) {
    alerts.push({
      level: "critical",
      title: "Preco abaixo do custo",
      message: "O valor simulado nao cobre o custo total ajustado.",
    });
  }
  if (marginPct != null && marginPct < 15) {
    alerts.push({
      level: "warn",
      title: "Margem baixa",
      message: "A margem de contribuicao ficou abaixo de 15%.",
    });
  }
  if (discountPct != null && discountPct > 35) {
    alerts.push({
      level: "warn",
      title: "Desconto agressivo",
      message: "Desconto elevado pode comprometer rentabilidade.",
    });
  }
  if (gapPct != null && gapPct < -10) {
    alerts.push({
      level: "info",
      title: "Preco bem abaixo do mercado",
      message: "Posicionamento agressivo em relacao a media.",
    });
  }

  const statusBadges: SimulationResult["statusBadges"] = [];
  if (alerts.some((item) => item.level === "critical")) {
    statusBadges.push({ label: "Alerta critico", tone: "warning" });
  } else {
    statusBadges.push({ label: "Sem alertas criticos", tone: "info" });
  }

  if (gapPct != null && gapPct <= 0) {
    statusBadges.push({ label: "Mais competitivo", tone: "success" });
  } else {
    statusBadges.push({ label: "Acima do mercado", tone: "warning" });
  }

  return {
    priceNew,
    effectiveCost,
    discountPct,
    deltaPriceAbs,
    deltaPricePct,
    marginAmount,
    marginPct,
    markupPct,
    gmroi,
    gapPct,
    breakEvenPrice,
    marketPositionPct,
    tones,
    alerts,
    statusBadges,
  };
}
