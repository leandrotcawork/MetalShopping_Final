export type SkuStatus = "crit" | "warn" | "ok" | "info";
export type SkuAction = "PRUNE" | "EXPAND" | "MONITOR";
export type SkuClass = "A" | "B" | "C";
export type SkuTrend = "up" | "down" | "flat";
export type SkuXYZ = "X" | "Y" | "Z";

export type SkuMetricsPrimary = {
  gapPct: number;
  marginPct: number;
  pme6: number;
  giro6: number;
  dos6: number;
  gmroi6: number;
  slope6: number;
  cv6: number;
  xyz6: SkuXYZ;
  dataQuality: number;
  maturity: number;
};

export type SkuSignalsShort = {
  dem3: number;
  slope3: number;
  xyz3: SkuXYZ;
};

export type AnalyticsSkuRow = {
  pn: string;
  ean: string;
  description: string;
  taxonomyLeafName: string;
  brand: string;
  status: SkuStatus;
  action: SkuAction;
  className: SkuClass;
  trend: SkuTrend;
  trendPct: number;
  trendLabel: string;
  trendTone: "up" | "down" | "neutral";
  trendColor: "trendGreen" | "trendRed" | "trendNeutral";
  trendSpark: number[];
  trendMeta?: string;
  classLabel: string;
  classTone: "classA1" | "classA2" | "classB1" | "classB2";
  stock: number;
  price: number;
  marketPrice: number;
  gapLabel: string;
  gapTone: "gapPositive" | "gapNegative" | "gapNeutral";
  marginTone: "gapPositive" | "gapNegative" | "gapNeutral";
  alertsCount: number;
  metrics: SkuMetricsPrimary;
  short: SkuSignalsShort;
};

export type SkuSpotlightModel = {
  pn: string;
  description: string;
  taxonomyLeafName: string;
  brand: string;
  status: SkuStatus;
  action: SkuAction;
  whyMatters: string;
  recommendations: string[];
  nextSteps: string[];
  competition: {
    min: number;
    avg: number;
    max: number;
  };
  metrics: SkuMetricsPrimary;
  short: SkuSignalsShort;
};
