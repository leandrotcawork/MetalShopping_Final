import actionNeededIcon from "../../../assets/insights/action_needed.svg";
import campaignIcon from "../../../assets/insights/campaign.svg";
import campaignActionIcon from "../../../assets/insights/campaign_action.svg";
import classificationIcon from "../../../assets/insights/classification.svg";
import classificationActionIcon from "../../../assets/insights/classification_action.svg";
import diagnosticIcon from "../../../assets/insights/diagnostic.svg";
import diagnosticTraceIcon from "../../../assets/insights/diagnostic_trace.svg";
import guardrailDiscountIcon from "../../../assets/insights/guardrail_discount.svg";
import guardrailFloorIcon from "../../../assets/insights/guardrail_floor.svg";
import highConfidenceIcon from "../../../assets/insights/high_confidence.svg";
import lowConfidenceIcon from "../../../assets/insights/low_confidence.svg";
import marketMarginIcon from "../../../assets/insights/market_margin.svg";
import marketPriceAvgIcon from "../../../assets/insights/market_price_avg.svg";
import marketTurnoverIcon from "../../../assets/insights/market_turnover.svg";
import mediumConfidenceIcon from "../../../assets/insights/medium_confidence.svg";
import portfolioIcon from "../../../assets/insights/portfolio.svg";
import pricingIcon from "../../../assets/insights/pricing.svg";
import pricingActionIcon from "../../../assets/insights/pricing_action.svg";
import pruneActionIcon from "../../../assets/insights/prune_action.svg";
import recommendationIcon from "../../../assets/insights/recommendation.svg";
import riskIcon from "../../../assets/insights/risk.svg";
import riskActionIcon from "../../../assets/insights/risk_action.svg";
import circleIcon from "../../../assets/insights/circle.svg";
import csvRaw from "./analytics_registry.csv?raw";

export type AnalyticsRegistryKind =
  | "DOMAIN"
  | "ACTION"
  | "ALERT"
  | "ALERT_ALIAS"
  | "ACTION_TAG"
  | "TAG"
  | "IMPACT"
  | "ICON";

export type AnalyticsRegistryRow = {
  kind: AnalyticsRegistryKind;
  domain: string;
  key: string;
  backendMeaning: string;
  uiLabel: string;
  uiTitle: string;
  uiSummary: string;
  uiPrimaryText: string;
  uiSecondaryText: string;
  uiCtaLabel: string;
  uiChipLabel: string;
  uiIconKey: string;
  uiTone: string;
  notes: string;
  dataType: string;
  unit: string;
  assetPath: string;
};

type LookupInput = {
  kind: AnalyticsRegistryKind;
  key: string;
  domain?: string;
};

function normalizeToken(value: string): string {
  return String(value || "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toUpperCase()
    .replace(/[^A-Z0-9]+/g, "_")
    .replace(/_+/g, "_")
    .replace(/^_+|_+$/g, "");
}

function parseCsvLine(line: string): string[] {
  const output: string[] = [];
  let value = "";
  let inQuotes = false;
  for (let i = 0; i < line.length; i += 1) {
    const ch = line[i];
    if (ch === '"') {
      const next = line[i + 1];
      if (inQuotes && next === '"') {
        value += '"';
        i += 1;
      } else {
        inQuotes = !inQuotes;
      }
      continue;
    }
    if (ch === ";" && !inQuotes) {
      output.push(value);
      value = "";
      continue;
    }
    value += ch;
  }
  output.push(value);
  return output.map((part) => part.trim());
}

function parseRegistryCsv(raw: string): AnalyticsRegistryRow[] {
  const lines = raw.split(/\r?\n/).filter((line) => line.trim().length > 0);
  if (lines.length <= 1) return [];
  const headers = parseCsvLine(lines[0]);
  const idx = (name: string) => headers.findIndex((item) => item === name);
  const read = (cols: string[], name: string) => {
    const pos = idx(name);
    if (pos < 0 || pos >= cols.length) return "";
    return cols[pos] || "";
  };
  const rows: AnalyticsRegistryRow[] = [];
  for (const line of lines.slice(1)) {
    const cols = parseCsvLine(line);
    const kindRaw = read(cols, "kind") as AnalyticsRegistryKind;
    if (!kindRaw) continue;
    rows.push({
      kind: kindRaw,
      domain: read(cols, "domain"),
      key: read(cols, "key"),
      backendMeaning: read(cols, "backend_meaning"),
      uiLabel: read(cols, "ui_label"),
      uiTitle: read(cols, "ui_title"),
      uiSummary: read(cols, "ui_summary"),
      uiPrimaryText: read(cols, "ui_primary_text"),
      uiSecondaryText: read(cols, "ui_secondary_text"),
      uiCtaLabel: read(cols, "ui_cta_label"),
      uiChipLabel: read(cols, "ui_chip_label"),
      uiIconKey: read(cols, "ui_icon_key"),
      uiTone: read(cols, "ui_tone"),
      notes: read(cols, "notes"),
      dataType: read(cols, "data_type"),
      unit: read(cols, "unit"),
      assetPath: read(cols, "asset_path"),
    });
  }
  return rows;
}

function makeLookupKey(kind: string, domain: string, key: string): string {
  return `${normalizeToken(kind)}|${normalizeToken(domain)}|${normalizeToken(key)}`;
}

const rows = parseRegistryCsv(csvRaw);
const byExact = new Map<string, AnalyticsRegistryRow>();
const byAnyDomain = new Map<string, AnalyticsRegistryRow>();

for (const row of rows) {
  byExact.set(makeLookupKey(row.kind, row.domain, row.key), row);
  const anyDomainKey = makeLookupKey(row.kind, "", row.key);
  if (!byAnyDomain.has(anyDomainKey)) byAnyDomain.set(anyDomainKey, row);
}

export const analyticsRegistryRows = rows;

export function findRegistryRow(input: LookupInput): AnalyticsRegistryRow | null {
  const key = input.key || "";
  if (!key) return null;
  const domain = input.domain || "";
  const exact = byExact.get(makeLookupKey(input.kind, domain, key));
  if (exact) return exact;
  return byAnyDomain.get(makeLookupKey(input.kind, "", key)) || null;
}

function firstNonEmpty(...values: Array<string | null | undefined>): string {
  for (const value of values) {
    if (value && String(value).trim().length > 0) return String(value).trim();
  }
  return "";
}

export function resolveRegistryText(
  input: LookupInput,
  fieldOrder: Array<keyof AnalyticsRegistryRow>,
  fallback = "",
): string {
  const row = findRegistryRow(input);
  if (!row) return fallback;
  const values = fieldOrder.map((field) => row[field]);
  return firstNonEmpty(...values, fallback);
}

export function resolveChipText(token: string, domain?: string): string {
  const fromTag =
    resolveRegistryText({ kind: "TAG", key: token, domain }, ["uiChipLabel", "uiLabel", "uiTitle"]) ||
    resolveRegistryText({ kind: "ACTION_TAG", key: token, domain }, ["uiChipLabel", "uiLabel", "uiTitle"]) ||
    resolveRegistryText({ kind: "ALERT", key: token, domain }, ["uiChipLabel", "uiLabel", "uiTitle"]) ||
    resolveRegistryText({ kind: "ACTION", key: token, domain }, ["uiChipLabel", "uiLabel", "uiTitle"]);
  return fromTag || token;
}

const iconByKey: Record<string, string> = {
  action_needed: actionNeededIcon,
  campaign: campaignIcon,
  campaign_action: campaignActionIcon,
  circle: circleIcon,
  classification: classificationIcon,
  classification_action: classificationActionIcon,
  diagnostic: diagnosticIcon,
  diagnostic_trace: diagnosticTraceIcon,
  guardrail_discount: guardrailDiscountIcon,
  guardrail_floor: guardrailFloorIcon,
  high_confidence: highConfidenceIcon,
  low_confidence: lowConfidenceIcon,
  market_margin: marketMarginIcon,
  market_price_avg: marketPriceAvgIcon,
  market_turnover: marketTurnoverIcon,
  medium_confidence: mediumConfidenceIcon,
  portfolio: portfolioIcon,
  pricing: pricingIcon,
  pricing_action: pricingActionIcon,
  prune_action: pruneActionIcon,
  recommendation: recommendationIcon,
  risk: riskIcon,
  risk_action: riskActionIcon,
};

export function getInsightIconAsset(iconKey: string): string | null {
  const normalized = normalizeToken(iconKey).toLowerCase();
  return iconByKey[normalized] || null;
}

export function resolveIconKey(token: string, domain?: string): string {
  return (
    resolveRegistryText({ kind: "TAG", key: token, domain }, ["uiIconKey"]) ||
    resolveRegistryText({ kind: "ACTION_TAG", key: token, domain }, ["uiIconKey"]) ||
    resolveRegistryText({ kind: "ALERT", key: token, domain }, ["uiIconKey"]) ||
    resolveRegistryText({ kind: "ACTION", key: token, domain }, ["uiIconKey"]) ||
    ""
  );
}
