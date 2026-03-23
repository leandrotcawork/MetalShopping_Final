// @ts-nocheck
import type { AnalyticsHomeV2Dto } from "../../legacy_dto";
import { resolveRegistryText } from "../analytics/registry/analyticsRegistry";

export type ActionRow = {
  key: string;
  actionCode: string;
  name: string;
  desc: string;
  skuCount: string;
  skus: string[];
  stockTotalRs?: number | null;
  stockTotalLabel?: string;
  skuDetails?: Array<{
    pn: string;
    description: string;
    urgencyScore: number;
    urgencyLabel: string;
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    stockValue?: string;
    stockValueNumeric?: number | null;
    stockQty?: string;
    stockQtyNumeric?: number | null;
    decisionState?: string;
    valuePriorityTier?: string;
    riskAdjustedExposureRs?: number | null;
    dominantObjective?: string;
    arbitrationReasonCodes?: string[];
  }>;
  signalLabel: string;
  signalClass: "price" | "capital" | "stock" | "prune";
  decisionState?: string;
  valuePriorityTier?: string;
  riskAdjustedExposureRs?: number | null;
  dominantObjective?: string;
  arbitrationReasonCodes?: string[];
};

export type AlertRow = {
  key: string;
  code: string;
  name: string;
  desc: string;
  count: string;
  stockTotalRs?: number | null;
  stockTotalLabel?: string;
  toneClass: "critical" | "warning" | "inform";
  skuDetails?: Array<{
    pn: string;
    description: string;
    details: string;
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    stockValue?: string;
    stockValueNumeric?: number | null;
    stockQty?: string;
    stockQtyNumeric?: number | null;
    financialPriority?: string;
    financialPriorityScore?: number | null;
  }>;
};

export type PortfolioRow = {
  key: string;
  icon: string;
  iconStyle: "err" | "ok" | "info";
  label: string;
  value: string;
  countSkus: number;
  pct: number;
  fillStyle: "err" | "ok" | "info";
  drivers: string[];
  skus: string[];
  skuDetails: Array<{ pn: string; description: string; details: string }>;
};

export type TopMetalRow = {
  key: string;
  tone: "sku" | "brand" | "taxonomy";
  k: string;
  name: string;
  val: string;
  subVal: string;
};

export type TimelineRow = {
  key: string;
  pin: "green" | "blue" | "wine";
  name: string;
  desc: string;
  time: string;
};

export type KpiRow = {
  key: string;
  label: string;
  badge: string;
  value: string;
  note: string;
  bars: Array<{ key: string; heightPct: number; tipLabel: string; tipValue: string }>;
  tone: "default" | "blue";
};

export type HeatCellDetail = {
  key: string;
  impact: string;
  urgency: string;
  count: number;
  topDrivers: string[];
  skuDetails: Array<{
    pn: string;
    description: string;
    details: string;
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    stockValue?: string;
    stockValueNumeric?: number | null;
    stockQty?: string;
    stockQtyNumeric?: number | null;
  }>;
};

export type AnalyticsHomeViewModel = {
  actions: ActionRow[];
  allActions: ActionRow[];
  alerts: AlertRow[];
  allAlerts: AlertRow[];
  heatmap: number[][];
  heatCells: Record<string, HeatCellDetail>;
  portfolio: PortfolioRow[];
  topMetal: TopMetalRow[];
  timeline: TimelineRow[];
  kpis: KpiRow[];
  miniStats: Array<{ key: string; label: string; value: string; sub: string; badge: string }>;
};

type HomeActionCatalogItem = {
  actionCode: string;
  name: string;
  signalClass: ActionRow["signalClass"];
  defaultDesc: string;
};

const HOME_ACTION_CATALOG: HomeActionCatalogItem[] = [
  { actionCode: "AJUSTAR_PRECO_BAIXAR", name: "Ajustar preco (baixar)", signalClass: "price", defaultDesc: "Ganhar competitividade com guardrails de margem." },
  { actionCode: "AJUSTAR_PRECO_SUBIR", name: "Ajustar preco (subir)", signalClass: "price", defaultDesc: "Recuperar margem quando houver espaco competitivo." },
  { actionCode: "TESTAR_CAMPANHA", name: "Testar campanha", signalClass: "price", defaultDesc: "Ativar demanda com oferta controlada." },
  { actionCode: "LIBERAR_CAPITAL", name: "Liberar capital", signalClass: "capital", defaultDesc: "Converter estoque em caixa para realocacao." },
  { actionCode: "REDUZIR_IMOBILIZACAO", name: "Reduzir imobilizacao", signalClass: "capital", defaultDesc: "Reduzir capital parado em SKUs de baixa tracao." },
  { actionCode: "REVER_COMPRAS", name: "Rever compras", signalClass: "stock", defaultDesc: "Recalibrar reposicao e cobertura de estoque." },
  { actionCode: "AUMENTAR_EXPOSICAO", name: "Aumentar exposicao", signalClass: "stock", defaultDesc: "Melhorar visibilidade comercial dos SKUs." },
  { actionCode: "AJUSTAR_ESTRATEGIA", name: "Ajustar estrategia", signalClass: "prune", defaultDesc: "Revisar estrategia integrada de preco, estoque e campanha." },
];

const ALERT_CATALOG = [
  "capital_imobilizado",
  "estoque_parado",
  "giro_critico",
  "margem_apertada",
  "overstock",
  "price_above_market",
  "price_below_market",
] as const;

function uniqueSkuList(values: unknown[]): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const value of values) {
    const token = String(value || "").trim();
    if (!token || seen.has(token)) continue;
    seen.add(token);
    out.push(token);
  }
  return out;
}

function toList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map((it) => String(it || "").trim()).filter(Boolean);
}

function toPnList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value
    .map((item) => {
      if (item && typeof item === "object") {
        const row = item as Record<string, unknown>;
        return String(row.pn || row.sku || row.code || "").trim();
      }
      return String(item || "").trim();
    })
    .filter(Boolean);
}

function normalizeUrgencyLabel(value: unknown, score: number): string {
  const token = String(value || "").trim().toUpperCase();
  if (token) return token;
  if (score >= 80) return "HIGH";
  if (score >= 55) return "MEDIUM";
  return "LOW";
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

function asArray(value: unknown): Record<string, unknown>[] {
  return Array.isArray(value)
    ? value.filter((row) => row && typeof row === "object") as Record<string, unknown>[]
    : [];
}

function asArrayOrObjectValues(value: unknown): Record<string, unknown>[] {
  const fromArray = asArray(value);
  if (fromArray.length) return fromArray;
  const record = asRecord(value);
  if (!Object.keys(record).length) return [];
  return Object.values(record)
    .filter((row) => row && typeof row === "object") as Record<string, unknown>[];
}

function asNumber(value: unknown, fallback = 0): number {
  const num = Number(value);
  return Number.isFinite(num) ? num : fallback;
}

function asNumberOrNull(value: unknown): number | null {
  if (value == null || value === "") return null;
  const num = Number(value);
  return Number.isFinite(num) ? num : null;
}

function asPct(value: unknown): string {
  const num = asNumber(value, 0);
  return `${num >= 0 ? "+" : ""}${num.toFixed(1)}%`;
}

function asCurrency(value: unknown): string {
  const num = asNumber(value, 0);
  const abs = Math.abs(num);
  if (abs >= 1_000_000) return `R$ ${(num / 1_000_000).toFixed(1)}M`;
  if (abs >= 1_000) return `R$ ${(num / 1_000).toFixed(1)}k`;
  return `R$ ${num.toFixed(0)}`;
}

function asQty(value: unknown): string {
  const num = asNumber(value, Number.NaN);
  if (!Number.isFinite(num)) return "-";
  return num.toLocaleString("pt-BR", {
    minimumFractionDigits: Number.isInteger(num) ? 0 : 1,
    maximumFractionDigits: Number.isInteger(num) ? 0 : 2,
  });
}

function monthLabel(yyyymm: string): string {
  const token = String(yyyymm || "").trim();
  if (!/^\d{4}-\d{2}$/.test(token)) return token;
  return `${token.slice(5, 7)}/${token.slice(2, 4)}`;
}

function dateLabel(isoDate: string): string {
  const token = String(isoDate || "").trim();
  if (!/^\d{4}-\d{2}-\d{2}$/.test(token)) return token;
  return `${token.slice(8, 10)}/${token.slice(5, 7)}`;
}

function normalizeBarHeights(values: number[]): number[] {
  if (!values.length) return [];
  const safe = values.map((v) => (Number.isFinite(v) ? v : 0));
  const min = Math.min(...safe);
  const max = Math.max(...safe);
  if (max <= 0) return safe.map(() => 16);
  if (max === min) return safe.map(() => 72);
  return safe.map((v) => 16 + (((v - min) / (max - min)) * 84));
}

function mapSignalClass(actionCode: string): ActionRow["signalClass"] {
  const token = String(actionCode || "").toUpperCase();
  if (token.includes("PRUNE") || token.includes("REDUZ")) return "prune";
  if (token.includes("CAPITAL") || token.includes("EXPAND")) return "capital";
  if (token.includes("PRECO") || token.includes("PRICE")) return "price";
  return "stock";
}

function mapAlertTone(code: string): AlertRow["toneClass"] {
  const fromRegistry = resolveRegistryText(
    { kind: "ALERT", key: code },
    ["uiTone"],
    "",
  ).toLowerCase();
  if (fromRegistry === "critical" || fromRegistry === "crit" || fromRegistry === "danger" || fromRegistry === "negative") return "critical";
  if (fromRegistry === "warning" || fromRegistry === "warn" || fromRegistry === "medium") return "warning";
  if (fromRegistry === "inform" || fromRegistry === "info" || fromRegistry === "neutral" || fromRegistry === "default") return "inform";

  const token = String(code || "").toUpperCase();
  if (token.includes("PRICE_ABOVE") || token.includes("PRICE_BELOW") || token.includes("CAPITAL_IMOBILIZADO")) return "critical";
  if (token.includes("OVERSTOCK") || token.includes("ESTOQUE_PARADO") || token.includes("GIRO_CRITICO") || token.includes("MARGEM_APERTADA")) return "warning";
  return "inform";
}

function mapHealthRadar(
  dto: AnalyticsHomeV2Dto | null | undefined,
  pnDescriptionMap: Map<string, string>,
  pnMetaMap: Map<string, {
    description?: string;
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    stockValueNumeric?: number | null;
    stockQtyNumeric?: number | null;
  }>
): { matrix: number[][]; cells: Record<string, HeatCellDetail> } {
  const radar = dto?.blocks.health_radar;
  if (!radar || radar.status !== "OK" || !radar.data) return { matrix: [], cells: {} };
  const impacts = radar.data.impact_levels.length ? radar.data.impact_levels : ["I1", "I2", "I3", "I4", "I5"];
  const urgencies = radar.data.urgency_levels.length ? radar.data.urgency_levels : ["U1", "U2", "U3", "U4", "U5"];
  const maxCount = Math.max(1, ...radar.data.cells.map((cell) => Number(cell.count || 0)));
  const matrix = Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => 0.2));
  const cells: Record<string, HeatCellDetail> = {};
  const impactIndex = new Map(impacts.map((label, idx) => [label, idx]));
  const urgencyIndex = new Map(urgencies.map((label, idx) => [label, idx]));
  for (const cell of radar.data.cells) {
    const i = impactIndex.get(cell.impact);
    const u = urgencyIndex.get(cell.urgency);
    if (i == null || u == null || i > 4 || u > 4) continue;
    const y = 4 - i;
    const x = u;
    const key = `heat-${y}-${x}`;
    const normalized = Math.max(0.15, Math.min(1, Number(cell.count || 0) / maxCount));
    matrix[y][x] = normalized;
    const topDrivers = Array.isArray(cell.top_drivers) ? cell.top_drivers.map(String).filter(Boolean) : [];
    const driverText = topDrivers.length ? topDrivers.join(" | ") : "Sem driver destacado";
    const cellRecord = cell as Record<string, unknown>;
    const pnDetailsRaw = Array.isArray(cellRecord.pn_details)
      ? (cellRecord.pn_details as Array<Record<string, unknown>>)
      : [];
    const pnDetailsByPn = new Map<string, { descricao: string; details: string }>();
    for (const row of pnDetailsRaw) {
      const pn = String(row.pn || "").trim();
      if (!pn) continue;
      const stockValueNumericRaw = Number(row.stock_value_brl);
      const stockQtyNumericRaw = Number(row.stock_qty);
      pnDetailsByPn.set(pn, {
        descricao: String(row.descricao || "").trim(),
        details: String(row.details || "").trim(),
      });
      const known = pnMetaMap.get(pn) || {};
      pnMetaMap.set(pn, {
        description: String(row.descricao || "").trim() || known.description,
        brand: String(row.marca || row.brand || "").trim() || known.brand,
        taxonomyLeafName: String(
          row.taxonomy_root_name ??
          row.taxonomy_level_0_name ??
          row.taxonomy_leaf_0_name ??
          row.taxonomy_leaf0_name ??
          row.taxonomy_l0_name ??
          row.taxonomy_leaf_name ??
          row.taxonomyLeafName ??
          ""
        ).trim() || known.taxonomyLeafName,
        stockType: String(row.tipo_estoque || row.stock_type || row.stockType || "").trim() || known.stockType,
        stockValueNumeric: Number.isFinite(stockValueNumericRaw) ? stockValueNumericRaw : (known.stockValueNumeric ?? null),
        stockQtyNumeric: Number.isFinite(stockQtyNumericRaw) ? stockQtyNumericRaw : (known.stockQtyNumeric ?? null),
      });
    }
    const topPnsRaw = pnDetailsRaw.length
      ? pnDetailsRaw.map((row) => String(row.pn || "").trim())
      : (
          Array.isArray(cellRecord.pns)
            ? (cellRecord.pns as unknown[]).map((pn) => String(pn || "").trim())
            : (Array.isArray(cell.top_pns) ? cell.top_pns.map(String) : [])
        );
    const topPns = topPnsRaw.filter(Boolean);
    cells[key] = {
      key,
      impact: String(cell.impact || ""),
      urgency: String(cell.urgency || ""),
      count: Number(cell.count || 0),
      topDrivers,
      skuDetails: topPns.map((pn) => ({
        ...(pnMetaMap.get(pn) || {}),
        pn,
        description: pnDetailsByPn.get(pn)?.descricao || pnMetaMap.get(pn)?.description || pnDescriptionMap.get(pn) || "-",
        details: pnDetailsByPn.get(pn)?.details || driverText || "Sem detalhes",
        stockValue:
          pnMetaMap.get(pn)?.stockValueNumeric != null && Number.isFinite(Number(pnMetaMap.get(pn)?.stockValueNumeric))
            ? asCurrency(Number(pnMetaMap.get(pn)?.stockValueNumeric || 0))
            : undefined,
        stockQty:
          pnMetaMap.get(pn)?.stockQtyNumeric != null && Number.isFinite(Number(pnMetaMap.get(pn)?.stockQtyNumeric))
            ? String(Number(pnMetaMap.get(pn)?.stockQtyNumeric || 0))
            : undefined,
      })),
    };
  }
  return { matrix, cells };
}

export function mapAnalyticsHomeViewModel(dto: AnalyticsHomeV2Dto | null | undefined): AnalyticsHomeViewModel {
  const actionsBlock = dto?.blocks.actions_today;
  const alertsBlock = dto?.blocks.alerts_prioritarios;
  const portfolioBlock = dto?.blocks.portfolio_distribution;
  const topMetalBlock = dto?.blocks.top_metal;
  const timelineBlock = dto?.blocks.timeline;
  const kpisOp = dto?.blocks.kpis_operational;
  const kpisProducts = dto?.blocks.kpis_products;
  const kpisSeries = dto?.blocks.kpis_series;

  const actionItems =
    actionsBlock?.status === "OK" && actionsBlock.data
      ? asArray(actionsBlock.data.items)
      : [];
  const actionBuckets =
    actionsBlock?.status === "OK" && actionsBlock.data
      ? asArrayOrObjectValues(actionsBlock.data.buckets)
      : [];
  const pnDescriptionMap = new Map<string, string>();
  const pnMetaMap = new Map<string, {
    description?: string;
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    stockValueNumeric?: number | null;
    stockQtyNumeric?: number | null;
  }>();
  const upsertPnMeta = (
    pn: string,
    patch: {
      description?: string;
      brand?: string;
      taxonomyLeafName?: string;
      stockType?: string;
      stockValueNumeric?: number | null;
      stockQtyNumeric?: number | null;
    }
  ) => {
    const token = String(pn || "").trim();
    if (!token) return;
    const prev = pnMetaMap.get(token) || {};
    pnMetaMap.set(token, {
      description: patch.description || prev.description,
      brand: patch.brand || prev.brand,
      taxonomyLeafName: patch.taxonomyLeafName || prev.taxonomyLeafName,
      stockType: patch.stockType || prev.stockType,
      stockValueNumeric:
        patch.stockValueNumeric != null && Number.isFinite(Number(patch.stockValueNumeric))
          ? Number(patch.stockValueNumeric)
          : (prev.stockValueNumeric ?? null),
      stockQtyNumeric:
        patch.stockQtyNumeric != null && Number.isFinite(Number(patch.stockQtyNumeric))
          ? Number(patch.stockQtyNumeric)
          : (prev.stockQtyNumeric ?? null),
    });
  };
  for (const item of actionItems) {
    const pn = String(item.pn || "").trim();
    if (!pn) continue;
    const desc = String(item.descricao || item.description || item.name || "").trim();
    if (desc && !pnDescriptionMap.has(pn)) pnDescriptionMap.set(pn, desc);
    const stockValueRaw = Number(item.stock_value_brl ?? item.capital_exposure_rs ?? item.capital_brl);
    const stockQtyRaw = Number(item.stock_qty ?? item.stock_on_hand_qty);
    upsertPnMeta(pn, {
      description: desc || undefined,
      brand: String(item.marca || item.brand || "").trim() || undefined,
      taxonomyLeafName: String(
        item.taxonomy_root_name ??
        item.taxonomy_level_0_name ??
        item.taxonomy_leaf_0_name ??
        item.taxonomy_leaf0_name ??
        item.taxonomy_l0_name ??
        item.taxonomy_leaf_name ??
        item.taxonomyLeafName ??
        ""
      ).trim() || undefined,
      stockType: String(item.tipo_estoque ?? item.stock_type ?? item.stockType ?? "").trim() || undefined,
      stockValueNumeric: Number.isFinite(stockValueRaw) ? stockValueRaw : null,
      stockQtyNumeric: Number.isFinite(stockQtyRaw) ? stockQtyRaw : null,
    });
  }
  for (const bucket of actionBuckets) {
    for (const row of asArray(bucket.top_skus)) {
      const pn = String(row.pn || "").trim();
      if (!pn) continue;
      const desc = String(row.descricao || row.description || "").trim();
      if (desc && !pnDescriptionMap.has(pn)) pnDescriptionMap.set(pn, desc);
      const stockValueRaw = Number(row.stock_value_brl ?? row.capital_brl ?? row.capital_exposure_rs);
      const stockQtyRaw = Number(row.stock_qty ?? row.stock_on_hand_qty);
      upsertPnMeta(pn, {
        description: desc || undefined,
        brand: String(row.marca || row.brand || "").trim() || undefined,
        taxonomyLeafName: String(
          row.taxonomy_root_name ??
          row.taxonomy_level_0_name ??
          row.taxonomy_leaf_0_name ??
          row.taxonomy_leaf0_name ??
          row.taxonomy_l0_name ??
          row.taxonomy_leaf_name ??
          row.taxonomyLeafName ??
          ""
        ).trim() || undefined,
        stockType: String(row.tipo_estoque ?? row.stock_type ?? row.stockType ?? "").trim() || undefined,
        stockValueNumeric: Number.isFinite(stockValueRaw) ? stockValueRaw : null,
        stockQtyNumeric: Number.isFinite(stockQtyRaw) ? stockQtyRaw : null,
      });
    }
  }
  if (alertsBlock?.status === "OK" && alertsBlock.data && alertsBlock.data.top_skus) {
    const topSkus = alertsBlock.data.top_skus as Record<string, unknown>;
    for (const rows of Object.values(topSkus)) {
      for (const row of asArray(rows)) {
        const pn = String(row.pn || "").trim();
        if (!pn) continue;
        const desc = String(row.descricao || row.description || "").trim();
        if (desc && !pnDescriptionMap.has(pn)) pnDescriptionMap.set(pn, desc);
        const stockValueRaw = Number(row.stock_value_brl ?? row.capital_brl ?? row.capital_exposure_rs);
        const stockQtyRaw = Number(row.stock_qty ?? row.stock_on_hand_qty);
        upsertPnMeta(pn, {
          description: desc || undefined,
          brand: String(row.marca || row.brand || "").trim() || undefined,
          taxonomyLeafName: String(
            row.taxonomy_root_name ??
            row.taxonomy_level_0_name ??
            row.taxonomy_leaf_0_name ??
            row.taxonomy_leaf0_name ??
            row.taxonomy_l0_name ??
            row.taxonomy_leaf_name ??
            row.taxonomyLeafName ??
            ""
          ).trim() || undefined,
          stockType: String(row.tipo_estoque ?? row.stock_type ?? row.stockType ?? "").trim() || undefined,
          stockValueNumeric: Number.isFinite(stockValueRaw) ? stockValueRaw : null,
          stockQtyNumeric: Number.isFinite(stockQtyRaw) ? stockQtyRaw : null,
        });
      }
    }
  }

  const actionsTotal = actionBuckets.length
    ? actionBuckets.reduce(
        (sum, bucket) => sum + asNumber(bucket.count ?? bucket.sku_count ?? bucket.count_skus ?? 0, 0),
        0
      )
    : 0;

  const allActionRows: ActionRow[] = (() => {
    const bucketsSource = actionBuckets;
    const skuByAction = new Map<string, string[]>();
    const detailByAction = new Map<
      string,
      Map<string, {
        pn: string;
        description: string;
        urgencyScore: number;
        urgencyLabel: string;
        brand?: string;
        taxonomyLeafName?: string;
        stockType?: string;
        stockValue?: string;
        stockValueNumeric?: number | null;
        stockQty?: string;
        stockQtyNumeric?: number | null;
        decisionState?: string;
        valuePriorityTier?: string;
        riskAdjustedExposureRs?: number | null;
        dominantObjective?: string;
        arbitrationReasonCodes?: string[];
      }>
    >();
    const arbitrationByAction = new Map<
      string,
      {
        decisionState?: string;
        valuePriorityTier?: string;
        riskAdjustedExposureRs?: number | null;
        dominantObjective?: string;
        arbitrationReasonCodes?: string[];
        rankScore: number;
      }
    >();
    const tierRank = (tier: string | undefined): number => {
      if (tier === "P0") return 4;
      if (tier === "P1") return 3;
      if (tier === "P2") return 2;
      if (tier === "P3") return 1;
      return 0;
    };
    const stateRank = (state: string | undefined): number => {
      if (state === "READY") return 4;
      if (state === "CAUTION") return 3;
      if (state === "REVIEW") return 2;
      if (state === "BLOCKED") return 1;
      return 0;
    };
    for (const item of actionItems) {
      const actionCode = String(item.action_code || item.main_action || item.action || item.code || "MONITOR");
      const skuList = uniqueSkuList([
        item.pn,
        ...toPnList(item.top_pns ?? item.top_skus ?? item.pns),
      ]);
      const current = skuByAction.get(actionCode) || [];
      skuByAction.set(actionCode, uniqueSkuList([...current, ...skuList]));
      const pn = String(item.pn || "").trim();
      const vaRaw = asRecord(item.value_arbitration);
      const decisionState = String((vaRaw.decision_state ?? item.decision_state ?? "") || "").trim().toUpperCase() || undefined;
      const valuePriorityTier = String((vaRaw.value_priority_tier ?? item.value_priority_tier ?? "") || "").trim().toUpperCase() || undefined;
      const dominantObjective = String((vaRaw.dominant_objective ?? item.dominant_objective ?? "") || "").trim().toUpperCase() || undefined;
      const riskAdjustedExposureRs = asNumber(
        vaRaw.risk_adjusted_exposure_rs ?? item.risk_adjusted_exposure_rs,
        Number.NaN,
      );
      const reasonCodes = toList(vaRaw.arbitration_reason_codes ?? item.arbitration_reason_codes).map((r) => String(r || "").trim()).filter(Boolean);
      const resolvedRiskAdjustedExposure = Number.isFinite(riskAdjustedExposureRs) ? riskAdjustedExposureRs : null;
      const readNumber = (...values: unknown[]): number | null => {
        for (const value of values) {
          if (value == null || value === "") continue;
          const parsed = Number(value);
          if (Number.isFinite(parsed)) return parsed;
        }
        return null;
      };
      const brand = String(item.marca ?? item.brand ?? "").trim() || undefined;
      const taxonomyLeafName =
        String(
          item.taxonomy_root_name ??
          item.taxonomy_level_0_name ??
          item.taxonomy_leaf_0_name ??
          item.taxonomy_leaf0_name ??
          item.taxonomy_l0_name ??
          item.taxonomy_leaf_name ??
          item.taxonomyLeafName ??
          ""
        ).trim() || undefined;
      const stockType = String(item.tipo_estoque ?? item.stock_type ?? item.stockType ?? "").trim() || undefined;
      const stockValueNumeric = readNumber(
        item.stock_value_brl,
        item.capital_brl,
        item.capital_exposure_rs,
        vaRaw.capital_exposure_rs,
        vaRaw.risk_adjusted_exposure_rs,
        item.risk_adjusted_exposure_rs,
      );
      const stockQtyNumeric = readNumber(
        item.stock_qty,
        item.stock_on_hand_qty,
        item.estoque_atual,
        item.estoque,
      );
      const rankScore = (tierRank(valuePriorityTier) * 1_000_000)
        + (stateRank(decisionState) * 100_000)
        + Math.max(0, resolvedRiskAdjustedExposure ?? 0);
      const prevActionRank = arbitrationByAction.get(actionCode);
      if (!prevActionRank || rankScore > prevActionRank.rankScore) {
        arbitrationByAction.set(actionCode, {
          decisionState,
          valuePriorityTier,
          riskAdjustedExposureRs: resolvedRiskAdjustedExposure,
          dominantObjective,
          arbitrationReasonCodes: reasonCodes,
          rankScore,
        });
      }
      if (pn) {
        const urgencyScore = asNumber(item.urgency_score ?? item.score ?? 0, 0);
        const urgencyLabel = normalizeUrgencyLabel(item.urgency_label, urgencyScore);
        const desc = String(item.descricao || item.description || item.name || "").trim();
        const byPn = detailByAction.get(actionCode) || new Map<string, {
          pn: string;
          description: string;
          urgencyScore: number;
          urgencyLabel: string;
          brand?: string;
          taxonomyLeafName?: string;
          stockType?: string;
          stockValue?: string;
          stockValueNumeric?: number | null;
          stockQty?: string;
          stockQtyNumeric?: number | null;
          decisionState?: string;
          valuePriorityTier?: string;
          riskAdjustedExposureRs?: number | null;
          dominantObjective?: string;
          arbitrationReasonCodes?: string[];
        }>();
        const prev = byPn.get(pn);
        if (!prev || urgencyScore > prev.urgencyScore) {
          byPn.set(pn, {
            pn,
            description: desc,
            urgencyScore,
            urgencyLabel,
            brand,
            taxonomyLeafName,
            stockType,
            stockValue: stockValueNumeric != null ? asCurrency(stockValueNumeric) : "-",
            stockValueNumeric,
            stockQty: stockQtyNumeric != null ? asQty(stockQtyNumeric) : "-",
            stockQtyNumeric,
            decisionState,
            valuePriorityTier,
            riskAdjustedExposureRs: resolvedRiskAdjustedExposure,
            dominantObjective,
            arbitrationReasonCodes: reasonCodes,
          });
        }
        detailByAction.set(actionCode, byPn);
      }
    }
    const bucketByAction = new Map<string, Record<string, unknown>>();
    for (const bucket of bucketsSource) {
      const actionCode = String(bucket.action_code || bucket.main_action || bucket.action || bucket.code || "MONITOR");
      if (!actionCode) continue;
      if (!bucketByAction.has(actionCode)) bucketByAction.set(actionCode, bucket);
    }

    return HOME_ACTION_CATALOG.map((item, index) => {
      const bucket = bucketByAction.get(item.actionCode) || {};
      const actionCode = item.actionCode;
      const skuCount = asNumber(bucket.count ?? bucket.sku_count ?? bucket.count_skus ?? 0, 0);
      const bucketTopRows = asArray(bucket.top_skus ?? bucket.top_examples);
      const skus = uniqueSkuList([
        ...toPnList(bucket.top_pns ?? bucket.top_skus ?? bucket.pns ?? bucket.top_examples),
        ...(skuByAction.get(actionCode) || []),
      ]);
      const explicitDetails = Array.from((detailByAction.get(actionCode) || new Map()).values());
      const fallbackUrgencyScore = asNumber(
        bucket.urgency_score ?? bucket.score ?? bucket.urgency_max ?? 0,
        0,
      );
      const fallbackUrgencyLabel = normalizeUrgencyLabel(
        bucket.urgency_label,
        fallbackUrgencyScore,
      );
      const bucketTopDetails = bucketTopRows.map((row) => {
        const pn = String(row.pn || "").trim();
        if (!pn) return null;
        const urgencyScore = asNumber(row.urgency_score ?? row.score ?? fallbackUrgencyScore, fallbackUrgencyScore);
        const urgencyLabel = normalizeUrgencyLabel(row.urgency_label, urgencyScore);
        const stockValueNumeric = asNumberOrNull(
          row.stock_value_brl ??
          row.capital_brl ??
          row.capital_exposure_rs ??
          row.valor_estoque ??
          row.estoque_valor
        );
        const stockQtyNumeric = asNumberOrNull(
          row.stock_qty ??
          row.stock_on_hand_qty ??
          row.estoque_atual ??
          row.estoque
        );
        return {
          pn,
          description: String(row.descricao || row.description || pnDescriptionMap.get(pn) || "-").trim() || "-",
          urgencyScore,
          urgencyLabel,
          brand: String(row.marca || row.brand || "").trim() || undefined,
          taxonomyLeafName: String(row.taxonomy_leaf_name || row.taxonomyLeafName || "").trim() || undefined,
          stockType: String(row.tipo_estoque || row.stock_type || row.stockType || "").trim() || undefined,
          stockValue: stockValueNumeric != null ? asCurrency(stockValueNumeric) : "-",
          stockValueNumeric,
          stockQty: stockQtyNumeric != null ? asQty(stockQtyNumeric) : "-",
          stockQtyNumeric,
        };
      }).filter((row): row is NonNullable<typeof row> => row != null);
      const skuDetails = (explicitDetails.length
        ? explicitDetails
        : bucketTopDetails.length
          ? bucketTopDetails
          : skus.map((pn) => ({
            pn,
            description: pnDescriptionMap.get(pn) || "-",
            urgencyScore: fallbackUrgencyScore,
            urgencyLabel: fallbackUrgencyLabel,
            brand: "-",
            taxonomyLeafName: "-",
            stockType: undefined,
            stockValue: "-",
            stockValueNumeric: null,
            stockQty: "-",
            stockQtyNumeric: null,
          }))
      ).sort((a, b) => (b.urgencyScore - a.urgencyScore) || a.pn.localeCompare(b.pn));
      const stockTotalRs = (() => {
        const bucketTotal = asNumberOrNull(
          bucket.stock_total_rs ??
          bucket.stock_value_total_rs ??
          bucket.capital_total_rs ??
          bucket.capital_exposure_rs ??
          bucket.valor_estoque_total ??
          bucket.estoque_valor_total
        );
        if (bucketTotal != null) return bucketTotal;
        const detailsTotal = skuDetails.reduce((sum, detail) => sum + (detail.stockValueNumeric ?? 0), 0);
        return detailsTotal > 0 ? detailsTotal : null;
      })();
      const resolvedSkuCount = skuCount > 0 ? skuCount : skuDetails.length;
      const arbitrationCtx = arbitrationByAction.get(actionCode);
      return {
        key: `action-bucket:${String(actionCode || index).toLowerCase()}`,
        actionCode,
        name: String(bucket.label || bucket.title || bucket.name || item.name || actionCode || "Acao"),
        desc: String(bucket.headline || bucket.description || bucket.desc || item.defaultDesc),
        skuCount: `${resolvedSkuCount} SKU${resolvedSkuCount === 1 ? "" : "s"}`,
        skus,
        stockTotalRs,
        stockTotalLabel: stockTotalRs != null ? asCurrency(stockTotalRs) : "-",
        skuDetails,
        signalLabel: actionCode.replace(/_/g, " "),
        signalClass: item.signalClass || mapSignalClass(actionCode),
        decisionState: arbitrationCtx?.decisionState,
        valuePriorityTier: arbitrationCtx?.valuePriorityTier,
        riskAdjustedExposureRs: arbitrationCtx?.riskAdjustedExposureRs ?? null,
        dominantObjective: arbitrationCtx?.dominantObjective,
        arbitrationReasonCodes: arbitrationCtx?.arbitrationReasonCodes || [],
      };
    });
  })();
  const rankTier = (tier: string | undefined): number => {
    if (tier === "P0") return 4;
    if (tier === "P1") return 3;
    if (tier === "P2") return 2;
    if (tier === "P3") return 1;
    return 0;
  };
  const rankState = (state: string | undefined): number => {
    if (state === "READY") return 4;
    if (state === "CAUTION") return 3;
    if (state === "REVIEW") return 2;
    if (state === "BLOCKED") return 1;
    return 0;
  };
  const actionRows: ActionRow[] = allActionRows
    .filter((row) => {
      const count = asNumber(String(row.skuCount).split(" ")[0], 0);
      return count > 0 || (row.skuDetails?.length ?? 0) > 0;
    })
    .sort((a, b) => {
      const tierCmp = rankTier(b.valuePriorityTier) - rankTier(a.valuePriorityTier);
      if (tierCmp !== 0) return tierCmp;
      const stateCmp = rankState(b.decisionState) - rankState(a.decisionState);
      if (stateCmp !== 0) return stateCmp;
      const exposureCmp = asNumber(b.riskAdjustedExposureRs, 0) - asNumber(a.riskAdjustedExposureRs, 0);
      if (exposureCmp !== 0) return exposureCmp;
      const qtyCmp = asNumber(String(b.skuCount).split(" ")[0], 0) - asNumber(String(a.skuCount).split(" ")[0], 0);
      if (qtyCmp !== 0) return qtyCmp;
      return a.name.localeCompare(b.name);
    })
    .slice(0, 4);

  type AlertSkuContext = {
    brand?: string;
    taxonomyLeafName?: string;
    stockType?: string;
    financialPriorityTier?: string;
    financialPriorityScore?: number | null;
    stockValueNumeric?: number | null;
    stockQtyNumeric?: number | null;
    rank: number;
  };

  const alertSkuContextByPn = new Map<string, AlertSkuContext>();

  const tierRankForAlert = (tier: string | undefined): number => {
    if (tier === "P0") return 4;
    if (tier === "P1") return 3;
    if (tier === "P2") return 2;
    if (tier === "P3") return 1;
    return 0;
  };

  const tierLabelForAlert = (tier: string | undefined): "Alta" | "Media" | "Baixa" | "" => {
    if (tier === "P0" || tier === "P1" || tier === "HIGH" || tier === "ALTA") return "Alta";
    if (tier === "P2" || tier === "MEDIUM" || tier === "MEDIA") return "Media";
    if (tier === "P3" || tier === "LOW" || tier === "BAIXA") return "Baixa";
    return "";
  };

  const firstText = (...values: unknown[]): string => {
    for (const value of values) {
      const token = String(value || "").trim();
      if (token) return token;
    }
    return "";
  };

  const firstNumber = (...values: unknown[]): number | null => {
    for (const value of values) {
      const parsed = asNumberOrNull(value);
      if (parsed != null) return parsed;
    }
    return null;
  };

  const taxonomyLeaf0Name = (...values: unknown[]): string => {
    for (const value of values) {
      const record = asRecord(value);
      const taxonomy = asRecord(record.taxonomy);
      const taxonomyLeaf = asRecord(taxonomy.leaf);
      const taxonomyLeaf0 = asRecord(taxonomy.leaf0);

      const levels = Array.isArray(taxonomy.levels) ? taxonomy.levels : [];
      const levelZero = levels.find((level) => {
        const levelRecord = asRecord(level);
        return asNumberOrNull(levelRecord.level) === 0;
      });
      const levelZeroRecord = asRecord(levelZero);

      const levelFromLeaf = asNumberOrNull(taxonomy.leaf_level);
      const rootName = firstText(
        record.taxonomy_root_name,
        record.taxonomy_level_0_name,
        record.taxonomy_leaf_0_name,
        record.taxonomy_leaf0_name,
        record.taxonomy_l0_name,
        taxonomy.root_name,
        taxonomy.level_0_name,
        taxonomy.level0_name,
        taxonomy.leaf_0_name,
        taxonomy.leaf0_name,
        taxonomyLeaf0.leaf_name,
        taxonomyLeaf0.name,
        levelZeroRecord.leaf_name,
        levelZeroRecord.name,
        taxonomyLeaf.leaf_name,
        taxonomyLeaf.name,
      );
      if (rootName) return rootName;

      const fallbackLeaf = firstText(
        record.taxonomy_leaf_name,
        record.taxonomyLeafName,
        taxonomy.leaf_name,
      );
      if (fallbackLeaf && (levelFromLeaf == null || levelFromLeaf === 0)) return fallbackLeaf;
      if (fallbackLeaf) return fallbackLeaf;
    }
    return "";
  };

  for (const item of actionItems) {
    const pn = String(item.pn || "").trim();
    if (!pn) continue;
    const valueArbitration = asRecord(item.value_arbitration);
    const priorityTier = firstText(
      valueArbitration.value_priority_tier,
      item.value_priority_tier,
    ).toUpperCase() || undefined;
    const priorityScore = firstNumber(
      valueArbitration.risk_adjusted_exposure_rs,
      item.risk_adjusted_exposure_rs,
      valueArbitration.capital_exposure_rs,
      item.capital_exposure_rs,
      valueArbitration.margin_exposure_rs,
      item.margin_exposure_rs,
      valueArbitration.priority_score,
      item.priority_score,
    );
    const stockValueNumeric = firstNumber(
      item.stock_value_brl,
      item.capital_brl,
      item.capital_exposure_rs,
      valueArbitration.capital_exposure_rs,
      valueArbitration.risk_adjusted_exposure_rs,
      item.risk_adjusted_exposure_rs,
    );
    const stockQtyNumeric = firstNumber(
      item.stock_qty,
      item.stock_on_hand_qty,
      item.estoque_atual,
      item.estoque,
    );
    const rank = (tierRankForAlert(priorityTier) * 1_000_000) + Math.max(0, priorityScore || 0);
    const current = alertSkuContextByPn.get(pn);
    if (!current || rank > current.rank) {
      alertSkuContextByPn.set(pn, {
        brand: firstText(item.marca, item.brand) || undefined,
        taxonomyLeafName: taxonomyLeaf0Name(item) || firstText(item.taxonomy_leaf_name, item.taxonomyLeafName) || undefined,
        stockType: firstText(item.tipo_estoque, item.stock_type, item.stockType) || undefined,
        financialPriorityTier: priorityTier,
        financialPriorityScore: priorityScore,
        stockValueNumeric,
        stockQtyNumeric,
        rank,
      });
    }
  }

  function mapAlertSkuDetails(
    topSkusRaw: Record<string, unknown>[],
    fallbackPns: string[],
    detailText: string,
  ): NonNullable<AlertRow["skuDetails"]> {
    const rowsSource: Record<string, unknown>[] = topSkusRaw.length
      ? topSkusRaw
      : fallbackPns.map((pn) => ({ pn }));

    const mapped = rowsSource
      .map((row) => {
        const record = asRecord(row);
        const details = asRecord(record.details);
        const pn = firstText(record.pn);
        if (!pn) return null;
        const ctx = alertSkuContextByPn.get(pn);
        const rawPriorityTier = firstText(
          record.financial_priority_tier,
          record.value_priority_tier,
          details.financial_priority_tier,
          ctx?.financialPriorityTier,
        ).toUpperCase();
        const rawPriorityScore = firstNumber(
          record.financial_priority_score,
          record.priority_score,
          record.value_priority_score,
          details.financial_priority_score,
          record.severity,
          ctx?.financialPriorityScore,
        );
        const stockValueNumeric = firstNumber(
          record.stock_value_brl,
          record.capital_brl,
          record.capital_exposure_rs,
          record.estoque_valor,
          record.valor_estoque,
          details.stock_value_brl,
          details.capital_brl,
          details.capital_exposure_rs,
          details.estoque_valor,
          details.valor_estoque,
          ctx?.stockValueNumeric,
        );
        const stockQtyNumeric = firstNumber(
          record.stock_qty,
          record.stock_on_hand_qty,
          record.estoque_atual,
          record.estoque,
          details.stock_qty,
          details.stock_on_hand_qty,
          details.estoque_atual,
          details.estoque,
          ctx?.stockQtyNumeric,
        );
        const detailsTextRaw = firstText(record.reason_pt, record.reason, detailText);
        const detailsText = detailsTextRaw
          ? detailsTextRaw.replace(/^motivo\s*:\s*/i, "").trim() || "-"
          : "-";
        const resolvedStockValueNumeric = stockValueNumeric ?? 0;
        return {
          pn,
          description: firstText(record.descricao, record.description, pnDescriptionMap.get(pn), "-") || "-",
          details: detailsText,
          brand: firstText(record.marca, record.brand, ctx?.brand) || "-",
          taxonomyLeafName: taxonomyLeaf0Name(record, ctx) || firstText(record.taxonomy_leaf_name, record.taxonomyLeafName, ctx?.taxonomyLeafName) || "-",
          stockType: firstText(record.tipo_estoque, record.stock_type, record.stockType, details.tipo_estoque, details.stock_type, ctx?.stockType) || "-",
          stockValue: asCurrency(resolvedStockValueNumeric),
          stockValueNumeric: resolvedStockValueNumeric,
          stockQty: stockQtyNumeric != null ? asQty(stockQtyNumeric) : "-",
          stockQtyNumeric,
          rawPriorityTier,
          rawPriorityScore,
        };
      })
      .filter((row): row is {
        pn: string;
        description: string;
        details: string;
        brand: string;
        taxonomyLeafName: string;
        stockType: string;
        stockValue: string;
        stockValueNumeric: number;
        stockQty: string;
        stockQtyNumeric: number | null;
        rawPriorityTier: string;
        rawPriorityScore: number | null;
      } => Boolean(row));

    const rankedScores = mapped
      .map((row) => row.rawPriorityScore)
      .filter((value): value is number => typeof value === "number" && Number.isFinite(value) && value > 0)
      .sort((a, b) => b - a);
    const highCut = rankedScores.length
      ? rankedScores[Math.max(0, Math.floor((rankedScores.length - 1) * 0.2))]
      : 0;
    const mediumCut = rankedScores.length
      ? rankedScores[Math.max(0, Math.floor((rankedScores.length - 1) * 0.5))]
      : 0;

    return mapped
      .map((row) => {
        const explicitLabel = tierLabelForAlert(row.rawPriorityTier);
        const fallbackLabel =
          row.rawPriorityScore != null
            ? row.rawPriorityScore >= highCut
              ? "Alta"
              : row.rawPriorityScore >= mediumCut
                ? "Media"
                : "Baixa"
            : "-";
        const priorityLabel = explicitLabel || fallbackLabel;
        const tierScore = tierRankForAlert(row.rawPriorityTier);
        const priorityScore = row.rawPriorityScore != null
          ? row.rawPriorityScore
          : (tierScore > 0 ? tierScore * 100 : null);
        return {
          pn: row.pn,
          description: row.description,
          details: row.details,
          brand: row.brand,
          taxonomyLeafName: row.taxonomyLeafName,
          stockType: row.stockType,
          stockValue: row.stockValue,
          stockValueNumeric: row.stockValueNumeric,
          stockQty: row.stockQty,
          stockQtyNumeric: row.stockQtyNumeric,
          financialPriority: priorityLabel,
          financialPriorityScore: priorityScore,
        };
      })
      .sort((a, b) => {
        const scoreA = asNumber(a.financialPriorityScore, -1);
        const scoreB = asNumber(b.financialPriorityScore, -1);
        if (scoreA !== scoreB) return scoreB - scoreA;
        return a.pn.localeCompare(b.pn, "pt-BR", { numeric: true });
      });
  }

  function resolveAlertStockTotal(
    bucket: Record<string, unknown>,
    topSkusRaw: Record<string, unknown>[],
    skuDetails: NonNullable<AlertRow["skuDetails"]>,
  ): { value: number | null; label: string } {
    const bucketTotal = asNumberOrNull(
      bucket.stock_total_rs ??
      bucket.stock_value_total_rs ??
      bucket.estoque_total_rs ??
      bucket.total_stock_value_rs ??
      bucket.capital_total_rs
    );
    if (bucketTotal != null) {
      return { value: bucketTotal, label: asCurrency(bucketTotal) };
    }

    const topSkusTotal = topSkusRaw.reduce((sum, row) => {
      const value = asNumberOrNull(
        row.stock_value_brl ??
        row.capital_brl ??
        row.capital_exposure_rs ??
        row.estoque_valor ??
        row.valor_estoque
      );
      return sum + (value ?? 0);
    }, 0);
    if (topSkusTotal > 0) {
      return { value: topSkusTotal, label: asCurrency(topSkusTotal) };
    }

    const skuDetailsTotal = skuDetails.reduce((sum, item) => sum + (item.stockValueNumeric ?? 0), 0);
    if (skuDetailsTotal > 0) {
      return { value: skuDetailsTotal, label: asCurrency(skuDetailsTotal) };
    }
    return { value: null, label: "-" };
  }

  const alertBucketMap = new Map<string, Record<string, unknown>>();
  const alertTopSkusByCode: Record<string, Record<string, unknown>[]> =
    alertsBlock?.status === "OK" && alertsBlock.data && alertsBlock.data.top_skus
      ? (alertsBlock.data.top_skus as Record<string, Record<string, unknown>[]>)
      : {};
  if (alertsBlock?.status === "OK" && alertsBlock.data) {
    for (const bucket of asArrayOrObjectValues(alertsBlock.data.buckets)) {
      const code = String(bucket.code || bucket.alert || bucket.name || "").toLowerCase();
      if (!code) continue;
      if (!alertBucketMap.has(code)) alertBucketMap.set(code, bucket);
    }
  }

  const alertRows: AlertRow[] =
    alertsBlock?.status === "OK" && alertsBlock.data
      ? asArrayOrObjectValues(alertsBlock.data.buckets)
          .sort(
            (a, b) =>
              asNumber(b.count ?? b.sku_count ?? b.count_skus ?? 0, 0) -
              asNumber(a.count ?? a.sku_count ?? a.count_skus ?? 0, 0)
          )
          .slice(0, 5)
          .map((bucket, index) => {
            const rawCode = String(bucket.code || bucket.alert || bucket.name || `ALERTA_${index}`);
            const code = rawCode.toLowerCase();
            const registryLabel =
              resolveRegistryText({ kind: "ALERT", key: rawCode }, ["uiLabel", "uiChipLabel", "uiTitle"], "") ||
              resolveRegistryText({ kind: "ALERT_ALIAS", key: rawCode }, ["uiLabel", "uiChipLabel", "uiTitle"], "");
            const count = asNumber(bucket.count ?? bucket.sku_count ?? bucket.count_skus ?? 0, 0);
            const bucketDetails = String(bucket.headline || bucket.description || bucket.desc || "Sem detalhe");
            const topDrivers = toList(bucket.top_drivers);
            const detailText = topDrivers.length ? topDrivers.join(" | ") : bucketDetails;
            const topSkusRaw = asArray(alertTopSkusByCode[rawCode] ?? alertTopSkusByCode[code]);
            const fallbackPns = uniqueSkuList([...toPnList(bucket.top_pns ?? bucket.top_skus ?? bucket.pns)]);
            const skuDetails = mapAlertSkuDetails(topSkusRaw, fallbackPns, detailText);
            const stockTotal = resolveAlertStockTotal(bucket, topSkusRaw, skuDetails);
            return {
              key: `alert:${code}`,
              code: rawCode,
              name: String(bucket.label || bucket.title || registryLabel || bucket.name || rawCode),
              desc: String(bucket.headline || bucket.description || bucket.desc || "Sem descricao"),
              count: String(count),
              stockTotalRs: stockTotal.value,
              stockTotalLabel: stockTotal.label,
              toneClass: mapAlertTone(rawCode),
              skuDetails,
            };
          })
      : [];

  const allAlerts: AlertRow[] = ALERT_CATALOG.map((alertCode) => {
    const bucket = alertBucketMap.get(alertCode) || {};
    const rawCode = String((bucket.code || bucket.alert || bucket.name || alertCode) ?? alertCode);
    const registryLabel =
      resolveRegistryText({ kind: "ALERT", key: rawCode }, ["uiLabel", "uiChipLabel", "uiTitle"], "") ||
      resolveRegistryText({ kind: "ALERT_ALIAS", key: rawCode }, ["uiLabel", "uiChipLabel", "uiTitle"], "");
    const registryDesc =
      resolveRegistryText({ kind: "ALERT", key: rawCode }, ["uiSummary", "uiPrimaryText", "uiSecondaryText"], "") ||
      resolveRegistryText({ kind: "ALERT_ALIAS", key: rawCode }, ["uiSummary", "uiPrimaryText", "uiSecondaryText"], "");
    const count = asNumber(bucket.count ?? bucket.sku_count ?? bucket.count_skus ?? 0, 0);
    const bucketDetails = String(bucket.headline || bucket.description || bucket.desc || registryDesc || "Sem detalhe");
    const topDrivers = toList(bucket.top_drivers);
    const detailText = topDrivers.length ? topDrivers.join(" | ") : bucketDetails;
    const topSkusRaw = asArray(alertTopSkusByCode[rawCode] ?? alertTopSkusByCode[alertCode]);
    const fallbackPns = uniqueSkuList([...toPnList(bucket.top_pns ?? bucket.top_skus ?? bucket.pns)]);
    const skuDetails = mapAlertSkuDetails(topSkusRaw, fallbackPns, detailText);
    const stockTotal = resolveAlertStockTotal(bucket, topSkusRaw, skuDetails);
    return {
      key: `alert:${alertCode}`,
      code: rawCode,
      name: String(bucket.label || bucket.title || registryLabel || rawCode),
      desc: String(bucket.headline || bucket.description || bucket.desc || registryDesc || "Sem descricao"),
      count: String(count),
      stockTotalRs: stockTotal.value,
      stockTotalLabel: stockTotal.label,
      toneClass: mapAlertTone(rawCode),
      skuDetails,
    };
  });

  const portfolioRows: PortfolioRow[] =
    portfolioBlock?.status === "OK" && portfolioBlock.data
      ? portfolioBlock.data.buckets.map((bucket, index) => {
          const keyToken = String(bucket.key || "").toUpperCase();
          const iconStyle: PortfolioRow["iconStyle"] = keyToken === "PRUNE" ? "err" : keyToken === "EXPAND" ? "ok" : "info";
          const bucketRecord = bucket as Record<string, unknown>;
          const drivers = toList(bucket.top_drivers);
          const pnDetailsRaw = Array.isArray(bucketRecord.pn_details)
            ? (bucketRecord.pn_details as Array<Record<string, unknown>>)
            : [];
          const pns = uniqueSkuList(
            pnDetailsRaw.length
              ? pnDetailsRaw.map((row) => String(row.pn || "").trim())
              : toList((bucketRecord.pns as unknown) ?? bucket.top_pns)
          );
          const detailsByPn = new Map<string, { descricao: string; details: string }>();
          for (const row of pnDetailsRaw) {
            const pn = String(row.pn || "").trim();
            if (!pn) continue;
            detailsByPn.set(pn, {
              descricao: String(row.descricao || "").trim(),
              details: String(row.details || "").trim(),
            });
          }
          return {
            key: `pf-${String(bucket.key || index).toLowerCase()}`,
            icon: iconStyle === "err" ? "\u{1F4C9}" : iconStyle === "ok" ? "\u{1F680}" : "\u{1F441}",
            iconStyle,
            label: String(bucket.label || bucket.key || "N/D"),
            value: `${asNumber(bucket.count_skus, 0)} SKUs`,
            countSkus: asNumber(bucket.count_skus, 0),
            pct: Math.max(0, Math.min(100, Math.round(asNumber(bucket.share_skus, 0) * 100))),
            fillStyle: iconStyle,
            drivers,
            skus: pns,
            skuDetails: pns.map((pn) => ({
              pn,
              description: detailsByPn.get(pn)?.descricao || pnDescriptionMap.get(pn) || "-",
              details: detailsByPn.get(pn)?.details || drivers.join(" | ") || "Top SKU da classe",
            })),
          };
        })
      : [];

  const topMetalRows: TopMetalRow[] =
    topMetalBlock?.status === "OK" && topMetalBlock.data
      ? [
          {
            key: "tm-sku",
            tone: "sku",
            k: "Melhor SKU",
            name: topMetalBlock.data.best_sku
              ? `${topMetalBlock.data.best_sku.pn} \u2022 ${topMetalBlock.data.best_sku.descricao || "SKU"}`
              : "N/D",
            val: asCurrency(topMetalBlock.data.best_sku?.lucro_mes),
            subVal: `Receita: ${asCurrency(topMetalBlock.data.best_sku?.receita_liq_mes)}`,
          },
          {
            key: "tm-brand",
            tone: "brand",
            k: "Melhor Marca",
            name: topMetalBlock.data.best_brand?.brand || "N/D",
            val: asCurrency(topMetalBlock.data.best_brand?.lucro_mes),
            subVal: `Receita: ${asCurrency(topMetalBlock.data.best_brand?.receita_liq_mes)}`,
          },
          {
            key: "tm-taxonomy",
            tone: "taxonomy",
            k: "Melhor Grupo",
            name: (() => {
              const bestTaxonomyLeaf = asRecord(topMetalBlock.data?.best_taxonomy_leaf);
              const taxonomy = asRecord(bestTaxonomyLeaf.taxonomy);
              const leaf0 = asRecord(taxonomy.leaf0);
              return (
                firstText(
                  leaf0.leaf_name,
                  leaf0.name,
                  bestTaxonomyLeaf.taxonomy_leaf_0_name,
                  bestTaxonomyLeaf.taxonomy_level_0_name,
                  bestTaxonomyLeaf.taxonomy_leaf_name,
                  bestTaxonomyLeaf.taxonomy_leaf_id,
                ) || "N/D"
              );
            })(),
            val: asCurrency(topMetalBlock.data.best_taxonomy_leaf?.lucro_mes),
            subVal: `Receita: ${asCurrency(topMetalBlock.data.best_taxonomy_leaf?.receita_liq_mes)}`,
          },
        ]
      : [];

  const timelineRows: TimelineRow[] =
    timelineBlock?.status === "OK" && timelineBlock.data
      ? timelineBlock.data.rows.slice(0, 3).map((row, index) => {
          const rec = asRecord(row);
          const status = String(rec.status || "").toUpperCase();
          const pin: TimelineRow["pin"] = status.includes("OK") ? "green" : status.includes("WARN") ? "blue" : "wine";
          const timeRaw = String(rec.occurred_at || rec.timestamp || "");
          const date = timeRaw ? new Date(timeRaw) : null;
          const time = date && !Number.isNaN(date.getTime())
            ? new Intl.DateTimeFormat("pt-BR", { hour: "2-digit", minute: "2-digit" }).format(date)
            : "--:--";
          return {
            key: `ev-${String(rec.id || index)}`,
            pin,
            name: String(rec.title || rec.type || "Evento"),
            desc: String(rec.description || ""),
            time,
          };
        })
      : [];

  const kpiRows: KpiRow[] = (() => {
    if (!kpisOp || !kpisOp.data) return [];
    const salesSeries =
      kpisSeries?.status === "OK" && kpisSeries.data && Array.isArray((kpisSeries.data as Record<string, unknown>).sales_6m)
        ? ((kpisSeries.data as Record<string, unknown>).sales_6m as Array<Record<string, unknown>>)
        : [];
    const marginSeries =
      kpisSeries?.status === "OK" && kpisSeries.data && Array.isArray((kpisSeries.data as Record<string, unknown>).margin_6m)
        ? ((kpisSeries.data as Record<string, unknown>).margin_6m as Array<Record<string, unknown>>)
        : [];
    const runsSeries =
      kpisSeries?.status === "OK" && kpisSeries.data && Array.isArray((kpisSeries.data as Record<string, unknown>).runs_7d)
        ? ((kpisSeries.data as Record<string, unknown>).runs_7d as Array<Record<string, unknown>>)
        : [];
    const sales = salesSeries.length ? Number(salesSeries[salesSeries.length - 1].receita_liq || 0) : null;
    const margin = marginSeries.length ? Number(marginSeries[marginSeries.length - 1].margem_pct || 0) : null;
    const runs = runsSeries.length ? Number(runsSeries[runsSeries.length - 1].skus_processed || 0) : null;
    const runsHeights = normalizeBarHeights(runsSeries.map((row) => Number(row.skus_processed || 0)));
    const salesHeights = normalizeBarHeights(salesSeries.map((row) => Number(row.receita_liq || 0)));
    const marginHeights = normalizeBarHeights(marginSeries.map((row) => Number(row.margem_pct || 0)));
    return [
      {
        key: "kpi-runs",
        label: "RUNs de precos",
        badge: "7D",
        value: runs != null ? String(Math.max(0, Math.round(runs))) : "N/D",
        note: "SKUs processados por dia",
        bars: runsSeries.map((row, idx) => ({
          key: String(row.date || idx),
          heightPct: runsHeights[idx] ?? 16,
          tipLabel: dateLabel(String(row.date || "")),
          tipValue: `${Math.max(0, Math.round(Number(row.skus_processed || 0)))} SKUs`,
        })),
        tone: "default",
      },
      {
        key: "kpi-sales",
        label: "Vendas",
        badge: "6M",
        value: sales != null ? asCurrency(sales) : "N/D",
        note: "Receita liquida mensal",
        bars: salesSeries.map((row, idx) => ({
          key: String(row.month || idx),
          heightPct: salesHeights[idx] ?? 16,
          tipLabel: monthLabel(String(row.month || "")),
          tipValue: asCurrency(row.receita_liq),
        })),
        tone: "blue",
      },
      {
        key: "kpi-margin",
        label: "Margem media",
        badge: "6M",
        value: margin != null ? asPct(margin) : "N/D",
        note: "Margem percentual mensal",
        bars: marginSeries.map((row, idx) => ({
          key: String(row.month || idx),
          heightPct: marginHeights[idx] ?? 16,
          tipLabel: monthLabel(String(row.month || "")),
          tipValue: asPct(row.margem_pct),
        })),
        tone: "default",
      },
    ];
  })();

  const miniStats: AnalyticsHomeViewModel["miniStats"] = [
    {
      key: "mini-actions",
      label: "Acoes prioritarias",
      value: String(actionsTotal),
      sub: "itens mapeados por regra",
      badge: "Hoje",
    },
    {
      key: "mini-alerts",
      label: "Alertas ativos",
      value: String(alertRows.reduce((sum, row) => sum + asNumber(row.count, 0), 0)),
      sub: "sinais com impacto real",
      badge: "Live",
    },
  ];
  if (kpisProducts?.status === "OK" && kpisProducts.data) {
    miniStats.push(
      {
        key: "mini-abc",
        label: "MIX ABC",
        value: "ABC",
        sub: "curva de pareto por grupos",
        badge: "6M",
      },
      {
        key: "mini-capital",
        label: "Capital imobilizado",
        value: asCurrency(kpisProducts.data.capital_brl_total),
        sub: `${Math.max(0, Math.round(asNumber(kpisProducts.data.products_active_count, 0)))} ativos`,
        badge: "6M",
      },
      {
        key: "mini-potential",
        label: "Receita potencial",
        value: asCurrency(kpisProducts.data.potential_revenue_brl_total_market),
        sub: "mercado (portfolio)",
        badge: "6M",
      },
      {
        key: "mini-margin",
        label: "Margem ponderada",
        value: kpisProducts.data.weighted_margin_pct_total == null ? "N/D" : asPct(kpisProducts.data.weighted_margin_pct_total),
        sub: "portfolio produtos",
        badge: "Can",
      },
    );
  }

  const healthRadar = mapHealthRadar(dto, pnDescriptionMap, pnMetaMap);

  return {
    actions: actionRows,
    allActions: allActionRows,
    alerts: alertRows,
    allAlerts,
    heatmap: healthRadar.matrix,
    heatCells: healthRadar.cells,
    portfolio: portfolioRows,
    topMetal: topMetalRows,
    timeline: timelineRows,
    kpis: kpiRows,
    miniStats,
  };
}

