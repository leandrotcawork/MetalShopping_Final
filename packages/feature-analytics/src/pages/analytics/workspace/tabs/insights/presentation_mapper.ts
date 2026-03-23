import type { WorkspaceInsightsRecommendationItemV2 } from "../../../../../app/apiClient";
import { resolveRegistryText } from "../../../registry/analyticsRegistry";

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

function asNumber(value: unknown): number | null {
  if (value == null || value === "") return null;
  const num = Number(value);
  return Number.isFinite(num) ? num : null;
}

function asText(value: unknown): string {
  if (value == null) return "";
  return String(value).trim();
}

function normalizeActionCode(value: string): string {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toUpperCase()
    .replace(/[^A-Z0-9]+/g, "_")
    .replace(/_+/g, "_")
    .replace(/^_+|_+$/g, "");
}

function formatPct(value: number | null): string {
  if (value == null || !Number.isFinite(value)) return "";
  const normalized = Math.abs(value);
  const rounded = Math.round(normalized * 10) / 10;
  if (Math.abs(rounded - Math.round(rounded)) < 0.05) {
    return `${Math.round(rounded)}%`;
  }
  return `${rounded.toFixed(1).replace(".", ",")}%`;
}

function formatArgValue(value: unknown): string {
  if (value == null) return "";
  if (typeof value === "number" && Number.isFinite(value)) {
    return String(value);
  }
  return String(value);
}

function formatTemplateArg(key: string, value: unknown): string {
  if (key === "discount_pct") {
    const num = asNumber(value);
    if (num != null) {
      return `${Math.abs(num).toFixed(1).replace(".", ",")}%`;
    }
  }
  return formatArgValue(value);
}

function renderTemplate(template: string, args: Record<string, unknown>): string {
  if (!template.trim()) return "";
  const withDouble = template.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (_full, key: string) =>
    formatTemplateArg(key, args[key]),
  );
  return withDouble.replace(/\{\s*([a-zA-Z0-9_]+)\s*\}/g, (_full, key: string) => formatTemplateArg(key, args[key]));
}

function presentationNode(
  item: WorkspaceInsightsRecommendationItemV2,
  kind: "title" | "summary",
): { key: string; args: Record<string, unknown> } {
  const presentation = asRecord(item.presentation);
  const node = asRecord(presentation[kind]);
  const keyRaw = asText(node.key) || asText(item.action);
  const argsRaw = asRecord(node.args);
  return { key: keyRaw, args: argsRaw };
}

function resolveFromPresentationTemplate(
  item: WorkspaceInsightsRecommendationItemV2,
  kind: "title" | "summary",
): string {
  const node = presentationNode(item, kind);
  const key = normalizeActionCode(node.key);
  const template = resolveRegistryText(
    { kind: "ACTION", key },
    kind === "title"
      ? ["uiTitle", "uiLabel", "uiPrimaryText", "uiSummary"]
      : ["uiPrimaryText", "uiSummary", "uiSecondaryText", "uiLabel"],
    "",
  );
  if (!template) return "";
  return renderTemplate(template, node.args).trim();
}

type ActionPlanLike = {
  code?: string;
  label?: string;
  presentation?: {
    title?: {
      key?: string;
      args?: Record<string, unknown>;
    };
    summary?: {
      key?: string;
      args?: Record<string, unknown>;
    };
  };
};

function actionPlanPresentationNode(
  actionPlan: ActionPlanLike | null | undefined,
): { key: string; args: Record<string, unknown> } {
  const presentation = asRecord(actionPlan?.presentation);
  const summary = asRecord(presentation.summary);
  const title = asRecord(presentation.title);
  const keyRaw = asText(summary.key) || asText(title.key) || asText(actionPlan?.code);
  const argsRaw = asRecord(summary.args);
  const titleArgs = asRecord(title.args);
  return { key: keyRaw, args: Object.keys(argsRaw).length > 0 ? argsRaw : titleArgs };
}

function resolveActionPlanTemplate(
  actionPlan: ActionPlanLike | null | undefined,
  fields: string[],
): string {
  const node = actionPlanPresentationNode(actionPlan);
  const key = normalizeActionCode(node.key);
  const template = resolveRegistryText(
    { kind: "ACTION", key },
    fields as ("uiTitle" | "uiLabel" | "uiPrimaryText" | "uiSecondaryText" | "uiSummary")[],
    "",
  );
  if (!template) return "";
  return renderTemplate(template, node.args).trim();
}

function deriveDiscountPctFromImpact(item: WorkspaceInsightsRecommendationItemV2): number | null {
  const facts = asRecord(item.facts);
  const fromFacts = asNumber(facts.discount_pct);
  if (fromFacts != null) return Math.abs(fromFacts);

  const impact = asRecord(item.impact);
  const deltaPct = asNumber(impact.delta_price_pct);
  if (deltaPct != null) return Math.abs(deltaPct);

  const current = asNumber(impact.price_current);
  const target = asNumber(impact.price_target);
  if (current != null && target != null && current > 0) {
    return Math.abs(((current - target) / current) * 100);
  }
  return null;
}

export function toFriendlyRecommendationTitle(item: WorkspaceInsightsRecommendationItemV2): string {
  const fromPresentation = resolveFromPresentationTemplate(item, "title");
  if (fromPresentation) return fromPresentation;

  const domain = asText(item.domain);
  const code = normalizeActionCode(asText(item.action));
  const current = asText(item.title);
  const discountPct = deriveDiscountPctFromImpact(item);
  const discountText = formatPct(discountPct);

  if ((code === "AJUSTAR_PRECO_BAIXAR" || code === "BAIXAR_PRECO") && discountText) {
    return `Reduzir Preco em ${discountText}`;
  }
  if ((code === "AJUSTAR_PRECO_SUBIR" || code === "SUBIR_PRECO") && discountText) {
    return `Aumentar Preco em ${discountText}`;
  }

  return (
    resolveRegistryText({ kind: "ACTION", key: code, domain }, ["uiTitle", "uiLabel"], "") ||
    current ||
    "Recomendacao"
  );
}

export function toFriendlyRecommendationSummary(item: WorkspaceInsightsRecommendationItemV2): string {
  const fromPresentation = resolveFromPresentationTemplate(item, "summary");
  if (fromPresentation) return fromPresentation;

  const domain = asText(item.domain);
  const code = normalizeActionCode(asText(item.action));
  const generic = resolveRegistryText(
    { kind: "ACTION", key: code, domain },
    ["uiPrimaryText", "uiSummary", "uiSecondaryText"],
    "",
  );
  if (generic) return generic;

  return "Sem descricao padronizada para esta recomendacao.";
}

export function toFriendlyStockStatus(item: WorkspaceInsightsRecommendationItemV2 | null | undefined): string {
  if (!item) return "Monitorar";
  const domain = asText(item.domain);
  const code = normalizeActionCode(asText(item.action));
  const fromRegistry = resolveRegistryText({ kind: "ACTION", key: code, domain }, ["uiChipLabel", "uiLabel", "uiTitle"], "");
  if (fromRegistry) return fromRegistry;
  const title = asText(item.title);
  if (title) return title;
  const derived = toFriendlyRecommendationTitle(item);
  return derived || "Monitorar";
}

export function toFriendlyActionPlanPrimaryText(actionPlan: ActionPlanLike | null | undefined): string {
  const rendered = resolveActionPlanTemplate(actionPlan, ["uiPrimaryText", "uiSummary", "uiTitle", "uiLabel"]);
  if (rendered) return rendered;
  const label = asText(actionPlan?.label);
  return label || "—";
}

export function toFriendlyActionPlanSecondaryText(actionPlan: ActionPlanLike | null | undefined): string {
  const rendered = resolveActionPlanTemplate(actionPlan, ["uiSecondaryText", "uiSummary", "uiTitle", "uiLabel"]);
  if (rendered) return rendered;
  const label = asText(actionPlan?.label);
  return label || "—";
}
