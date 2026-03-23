export type WorkspaceBadge = {
  label: string;
  tone: "class" | "health" | "action";
};

export type WorkspaceHeroMetric = {
  label: string;
  value: string;
};

export type WorkspaceMetricPair = {
  label: string;
  value: string;
  tone?: "default" | "positive" | "warn" | "negative";
};

export type WorkspaceGauge = {
  label: string;
  valueText: string;
  value: number;
  min: number;
  max: number;
  ticks: string[];
  bands: Array<{
    to: number;
    color: string;
  }>;
};

export type WorkspaceProgress = {
  label: string;
  valueText: string;
  percent: number;
  tone: "positive" | "warn" | "negative" | "neutral";
};

export type WorkspaceColorBar = {
  label: string;
  valueText: string;
  scale: {
    min_value: number;
    max_value: number;
    low_cut: number;
    high_cut: number;
    higher_is_worse: boolean;
    min_visible_pct: number;
  };
  indicatorPct: number;
  fillPct: number;
  labels: [string, string, string, string];
};

export type WorkspaceTrend = {
  label: string;
  text: string;
  tone: "up" | "down" | "neutral";
};

export type ProductWorkspaceModel = {
  pn: string;
  brand: string;
  taxonomyLeafName: string;
  title: string;
  subtitle: string;
  updatedAt: string;
  badges: WorkspaceBadge[];
  heroMetrics: WorkspaceHeroMetric[];
  stockSales: {
    pme: WorkspaceMetricPair;
    giro: WorkspaceMetricPair;
    gauges: [WorkspaceGauge, WorkspaceGauge];
    bars: [WorkspaceProgress, WorkspaceProgress];
  };
  profitability: {
    metrics: [WorkspaceMetricPair, WorkspaceMetricPair];
    colorBars: [WorkspaceColorBar, WorkspaceColorBar];
    lower: [WorkspaceMetricPair, WorkspaceColorBar];
    cogs: WorkspaceMetricPair;
  };
  competitiveness: {
    metrics: [WorkspaceProgress, WorkspaceProgress];
    lower: [WorkspaceProgress, WorkspaceProgress];
  };
  risk: {
    metrics: [WorkspaceProgress, WorkspaceProgress];
    lower: [WorkspaceProgress, WorkspaceTrend];
  };
  history: {
    indicators: Array<{
      label: string;
      value: string;
      fill_pct: number;
    }>;
    meta: {
      price_window: string;
      sales_window: string;
      last_updated: string;
      coverage: string;
    };
    price_series: Array<{
      date: string;
      our_price: number | null;
      market_mean: number | null;
      suppliers: Record<string, number | null>;
    }>;
    supplier_links: Record<string, { label: string; url?: string | null }>;
    sales_monthly: Array<{
      date: string;
      units: number | null;
      revenue: number | null;
    }>;
  };
  simulator?: {
    variable_cost_unit_auto?: number | null;
    variable_cost_source?: string;
    variable_cost_coverage_months?: number;
    margin_calc_mode?: string;
    margin_is_fallback?: boolean;
    cost_avg_date?: string;
  };
};

function clamp(n: number, low: number, high: number): number {
  return Math.max(low, Math.min(high, n));
}

function indicatorPct(
  value: number,
  scale: {
    min_value: number;
    max_value: number;
  },
): number {
  const span = Math.max(1e-6, scale.max_value - scale.min_value);
  const clamped = clamp(value, scale.min_value, scale.max_value);
  return clamp(((clamped - scale.min_value) / span) * 100, 0, 100);
}

function fillPct(
  indicator: number,
  scale: {
    min_visible_pct: number;
  },
): number {
  if (indicator <= 0) return 0;
  const minVisible = clamp(scale.min_visible_pct, 0, 100);
  return indicator < minVisible ? minVisible : indicator;
}

function numFromPn(pn: string): number {
  const token = String(pn || "").replace(/\D/g, "");
  const seed = Number(token || "33584");
  return Number.isFinite(seed) && seed > 0 ? seed : 33584;
}

function brl(value: number): string {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    maximumFractionDigits: 2,
  }).format(value);
}

export function getWorkspaceModel(pn: string): ProductWorkspaceModel {
  const seed = numFromPn(pn);
  const drift = (seed % 17) - 8;
  const drift2 = (seed % 13) - 6;
  const pme = clamp(88 + drift, 28, 190);
  const giro = clamp(2.09 + drift2 * 0.07, 0.3, 3.2);
  const dos = clamp(20 + drift, 4, 140);
  const pressao = clamp(95 + drift2 * 2, 35, 130);
  const margem = clamp(48.4 + drift * 0.6, 8, 72);
  const markup = clamp(94 + drift2 * 1.8, 20, 140);
  const gmroi = clamp(1.33 + drift2 * 0.03, 0.1, 2.6);
  const gapPct = clamp(14.2 + drift * 0.4, -14, 42);
  const daysNoSales = clamp(3 + drift2, 0, 90);
  const marketSignal = clamp(100 + drift, 8, 100);
  const dispersao = clamp(71.8 + drift2 * 1.4, 8, 100);
  const riskCapital = clamp(100 + drift2 * 2, 20, 100);
  const dataQ = clamp(67 + drift * 1.3, 10, 100);
  const maturity = clamp(85 + drift2 * 1.3, 20, 100);
  const slope = -2.343 + drift2 * 0.12;

  const priceOur = 729 + drift * 8;
  const marketAvg = 598.9 + drift2 * 5;
  const marketMin = 369.9 + drift2 * 3.1;
  const marketMax = 827.9 + drift * 6.8;
  const costAvg = 375.91 + drift * 4;
  const costVar = 358.71 + drift2 * 3.8;
  const stock = clamp(3 + (seed % 7), 0, 28);

  const classId = ["A1", "A2", "B1", "B2", "C1", "C2"][seed % 6];
  const health = riskCapital > 70 && dataQ > 60 ? "Saudavel" : riskCapital > 45 ? "Moderado" : "Atencao";
  const action = gapPct > 10 ? "Expandir" : gapPct < -4 ? "Monitorar" : "Ajustar";
  const trendTone: WorkspaceTrend["tone"] = slope > 0.3 ? "up" : slope < -0.3 ? "down" : "neutral";
  const trendText =
    trendTone === "up"
      ? `Subindo (+${slope.toFixed(3)} un/mes)`
      : trendTone === "down"
        ? `Caindo (${slope.toFixed(3)} un/mes)`
        : `Estavel (${slope.toFixed(3)} un/mes)`;
  const margemScale = {
    min_value: 0,
    max_value: 60,
    low_cut: 15,
    high_cut: 30,
    higher_is_worse: false,
    min_visible_pct: 3,
  };
  const markupScale = {
    min_value: 0,
    max_value: 100,
    low_cut: 20,
    high_cut: 50,
    higher_is_worse: false,
    min_visible_pct: 3,
  };
  const gmroiScale = {
    min_value: 0,
    max_value: 2,
    low_cut: 0.9,
    high_cut: 1.3,
    higher_is_worse: false,
    min_visible_pct: 3,
  };
  const margemIndicator = indicatorPct(margem, margemScale);
  const markupIndicator = indicatorPct(markup, markupScale);
  const gmroiIndicator = indicatorPct(gmroi, gmroiScale);

  const baseDate = new Date("2026-03-22T00:00:00Z");
  const priceSeries = Array.from({ length: 30 }, (_, idx) => {
    const offset = 29 - idx;
    const date = new Date(baseDate.getTime() - offset * 24 * 60 * 60 * 1000);
    const driftLocal = drift + (idx % 5) - 2;
    const ourPrice = Math.max(100, priceOur + driftLocal * 3);
    const marketMean = Math.max(90, marketAvg + driftLocal * 2.6);
    return {
      date: date.toISOString().slice(0, 10),
      our_price: Number(ourPrice.toFixed(2)),
      market_mean: Number(marketMean.toFixed(2)),
      suppliers: {
        DEXCO: Number((marketMean * 0.96).toFixed(2)),
        LEROY: Number((marketMean * 1.03).toFixed(2)),
      },
    };
  });

  const salesMonthly = Array.from({ length: 12 }, (_, idx) => {
    const date = new Date(baseDate.getTime());
    date.setUTCMonth(date.getUTCMonth() - (11 - idx));
    const units = Math.max(0, Math.round(40 + drift * 2 + idx * 3));
    const revenue = Math.max(0, units * Math.max(90, priceOur * 0.92));
    const month = `${String(date.getUTCMonth() + 1).padStart(2, "0")}/${date.getUTCFullYear()}`;
    return {
      month,
      date: month,
      units,
      revenue: Number(revenue.toFixed(2)),
    };
  });

  return {
    pn: String(pn || seed),
    brand: "DECA LOUCA",
    taxonomyLeafName: "ASSENTO TERMOFIXO",
    title: `${String(pn || seed)} â€¢ DECA LOUCA â€¢ ASSENTO TERMOFIXO`,
    subtitle: "ASS.TERMOFIXO SLOW POLO/UNI/AXI GELO",
    updatedAt: "13/02/2026 01:38",
    badges: [
      { label: classId, tone: "class" },
      { label: health, tone: "health" },
      { label: action, tone: "action" },
    ],
    heroMetrics: [
      { label: "Preco (nosso)", value: brl(priceOur) },
      { label: "Mercado medio", value: brl(marketAvg) },
      { label: "Mercado min", value: brl(marketMin) },
      { label: "Mercado max", value: brl(marketMax) },
      { label: "Custo medio", value: brl(costAvg) },
      { label: "Custo variavel", value: brl(costVar) },
      { label: "Estoque atual", value: `${stock}` },
    ],
    stockSales: {
      pme: { label: "PME (dias)", value: `${Math.round(pme)} d` },
      giro: { label: "Giro (6M)", value: giro.toFixed(2) },
      gauges: [
        {
          label: "0 <- 60 -> 120 -> 180+",
          valueText: `${Math.round(pme)}d`,
          value: pme,
          min: 0,
          max: 180,
          ticks: ["0", "60", "120", "180+"],
          bands: [
            { to: 60, color: "#DC2626" },
            { to: 120, color: "#CA8A04" },
            { to: 180, color: "#16A34A" },
          ],
        },
        {
          label: "0.0 <- 0.7 -> 1.4 -> 3.0+",
          valueText: giro.toFixed(2),
          value: giro,
          min: 0,
          max: 3,
          ticks: ["0.0", "0.7", "1.4", "3.0+"],
          bands: [
            { to: 0.7, color: "#DC2626" },
            { to: 1.4, color: "#CA8A04" },
            { to: 3, color: "#16A34A" },
          ],
        },
      ],
      bars: [
        {
          label: "DOS (dias)",
          valueText: `${Math.round(dos)} d`,
          percent: clamp((dos / 180) * 100, 0, 100),
          tone: dos < 45 ? "positive" : dos < 120 ? "warn" : "negative",
        },
        {
          label: "Pressao de reposicao",
          valueText: `${(pressao / 100).toFixed(2)}x`,
          percent: clamp(pressao, 0, 100),
          tone: pressao < 85 ? "positive" : pressao < 110 ? "warn" : "negative",
        },
      ],
    },
    profitability: {
      metrics: [
        { label: "Margem contrib. (%)", value: `${margem.toFixed(1)}%`, tone: margem > 26 ? "positive" : margem > 15 ? "warn" : "negative" },
        { label: "Markup", value: `${Math.round(markup)}%`, tone: markup > 50 ? "positive" : markup > 20 ? "warn" : "negative" },
      ],
      colorBars: [
        {
          label: "Margem contrib. (%)",
          valueText: `${margem.toFixed(1)}%`,
          scale: margemScale,
          indicatorPct: margemIndicator,
          fillPct: fillPct(margemIndicator, margemScale),
          labels: ["0%", "15%", "30%", "60%"],
        },
        {
          label: "Markup",
          valueText: `${Math.round(markup)}%`,
          scale: markupScale,
          indicatorPct: markupIndicator,
          fillPct: fillPct(markupIndicator, markupScale),
          labels: ["0%", "20%", "50%", "100%"],
        },
      ],
      lower: [
        { label: "GMROI (6M)", value: gmroi.toFixed(2), tone: gmroi > 1.3 ? "positive" : gmroi > 0.9 ? "warn" : "negative" },
        {
          label: "GMROI escala",
          valueText: gmroi.toFixed(2),
          scale: gmroiScale,
          indicatorPct: gmroiIndicator,
          fillPct: fillPct(gmroiIndicator, gmroiScale),
          labels: ["0.0", "0.9", "1.3", "2.0"],
        },
      ],
      cogs: { label: "COGS (6M)", value: brl(costAvg) },
    },
    competitiveness: {
      metrics: [
        { label: "Gap vs mercado (%)", valueText: `${gapPct.toFixed(1)}%`, percent: clamp(Math.abs(gapPct), 0, 100), tone: gapPct <= 0 ? "positive" : gapPct < 10 ? "warn" : "negative" },
        { label: "Dias sem venda", valueText: `${Math.round(daysNoSales)} d`, percent: clamp((daysNoSales / 60) * 100, 0, 100), tone: daysNoSales <= 7 ? "positive" : daysNoSales <= 20 ? "warn" : "negative" },
      ],
      lower: [
        { label: "Sinal de mercado", valueText: `${Math.round(marketSignal)}`, percent: clamp(marketSignal, 0, 100), tone: marketSignal >= 70 ? "positive" : marketSignal >= 40 ? "warn" : "negative" },
        { label: "Dispersao mercado (%)", valueText: `${dispersao.toFixed(1)}%`, percent: clamp(dispersao, 0, 100), tone: dispersao < 35 ? "positive" : dispersao < 70 ? "warn" : "negative" },
      ],
    },
    risk: {
      metrics: [
        { label: "Risco capital parado", valueText: `${Math.round(riskCapital)}`, percent: riskCapital, tone: riskCapital > 70 ? "positive" : riskCapital > 45 ? "warn" : "negative" },
        { label: "Data quality", valueText: `${Math.round(dataQ)}`, percent: dataQ, tone: dataQ > 70 ? "positive" : dataQ > 40 ? "warn" : "negative" },
      ],
      lower: [
        { label: "Maturidade", valueText: `${Math.round(maturity)}`, percent: maturity, tone: maturity > 70 ? "positive" : maturity > 45 ? "warn" : "negative" },
        { label: "Tendencia da demanda", text: trendText, tone: trendTone },
      ],
    },
    history: {
      indicators: [
        { label: "Tendencia", value: trendText, fill_pct: clamp(50 + slope * 10, 5, 95) },
        { label: "Volatilidade", value: "Media", fill_pct: clamp(55 + drift2 * 3, 5, 95) },
        { label: "XYZ", value: "Y", fill_pct: clamp(42 + drift, 5, 95) },
        { label: "Forca demanda", value: `${Math.round(marketSignal)}%`, fill_pct: clamp(marketSignal, 5, 95) },
      ],
      meta: {
        price_window: "30 dias",
        sales_window: "12 meses",
        last_updated: "22/03/2026 10:14",
        coverage: "83%",
      },
      price_series: priceSeries,
      supplier_links: {
        DEXCO: { label: "Dexco", url: null },
        LEROY: { label: "Leroy", url: null },
      },
      sales_monthly: salesMonthly,
    },
    simulator: {
      variable_cost_unit_auto: Math.max(0, costVar),
      variable_cost_source: "AUTO",
      variable_cost_coverage_months: 6,
      margin_calc_mode: "GROSS_FALLBACK_NO_GV",
      margin_is_fallback: false,
      cost_avg_date: "2026-03-10",
    },
  };
}

