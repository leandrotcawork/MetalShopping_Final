import type { AnalyticsProductWorkspaceV1Dto } from "@metalshopping/feature-analytics";
import type { WorkspaceInsightsRecommendationItemV2, WorkspaceInsightsV2 } from "@metalshopping/api-client";
import { ApiClientError } from "@metalshopping/api-client";
import { useEffect, useMemo, useState, type CSSProperties } from "react";
import { useNavigate, useOutletContext } from "react-router-dom";

import actionNeededIcon from "../../../../assets/insights/action_needed.svg";
import diagnosticIcon from "../../../../assets/insights/diagnostic.svg";
import recommendationIcon from "../../../../assets/insights/recommendation.svg";
import pricingIcon from "../../../../assets/insights/pricing.svg";
import campaignIcon from "../../../../assets/insights/campaign.svg";
import portfolioIcon from "../../../../assets/insights/portfolio.svg";
import riskIcon from "../../../../assets/insights/risk.svg";
import classificationIcon from "../../../../assets/insights/classification.svg";
import marketPriceAvgIcon from "../../../../assets/insights/market_price_avg.svg";
import marketMarginIcon from "../../../../assets/insights/market_margin.svg";
import marketTurnoverIcon from "../../../../assets/insights/market_turnover.svg";
import guardrailFloorIcon from "../../../../assets/insights/guardrail_floor.svg";
import guardrailDiscountIcon from "../../../../assets/insights/guardrail_discount.svg";
import diagnosticTraceIcon from "../../../../assets/insights/diagnostic_trace.svg";
import pricingActionIcon from "../../../../assets/insights/pricing_action.svg";
import campaignActionIcon from "../../../../assets/insights/campaign_action.svg";
import pruneActionIcon from "../../../../assets/insights/prune_action.svg";
import riskActionIcon from "../../../../assets/insights/risk_action.svg";
import classificationActionIcon from "../../../../assets/insights/classification_action.svg";
import highConfidenceIcon from "../../../../assets/insights/high_confidence.svg";
import mediumConfidenceIcon from "../../../../assets/insights/medium_confidence.svg";
import lowConfidenceIcon from "../../../../assets/insights/low_confidence.svg";

import type { ProductWorkspaceOutletContext } from "../ProductWorkspaceLayout";
import { useAppSession } from "../../../../app/providers/AppProviders";
import styles from "./insights.module.css";
import { InsightCard } from "./insights/InsightCard";
import { InsightBadge } from "./insights/InsightBadge";
import { buildCampaignChips, buildPortfolioChips, buildPricingChips, buildRiskChips, buildStockChips } from "./insights/chips_mapper";
import {
  toFriendlyActionPlanPrimaryText,
  toFriendlyActionPlanSecondaryText,
  toFriendlyRecommendationSummary,
  toFriendlyRecommendationTitle,
  toFriendlyStockStatus,
} from "./insights/presentation_mapper";
import { resolveRegistryText } from "../../registry/analyticsRegistry";
import type { Insight, InsightSection, InsightSeverity } from "./insights/types";

type FilterKey = "all" | Insight["category"];

const FILTERS: Array<{ key: FilterKey; label: string }> = [
  { key: "all", label: "Todos" },
  { key: "preco", label: "Preço" },
  { key: "campanha", label: "Campanha" },
  { key: "estoque", label: "Estoque" },
  { key: "portfolio", label: "Portfólio" },
  { key: "risco", label: "Risco" },
];

const DIAGNOSTIC_ICON_BY_DOMAIN: Record<Insight["domain"], string> = {
  COMPETITIVIDADE: pricingIcon,
  ESTOQUE: portfolioIcon,
  RENTABILIDADE: portfolioIcon,
  RISCO: riskIcon,
  DADOS: classificationIcon,
  MERCADO: campaignIcon,
};

type DiagnosticTheme = "pricing" | "campaign" | "portfolio" | "risk" | "classification";

const DIAGNOSTIC_ACTION_ICON_BY_DOMAIN: Record<Insight["domain"], string> = {
  COMPETITIVIDADE: pricingActionIcon,
  ESTOQUE: pruneActionIcon,
  RENTABILIDADE: classificationActionIcon,
  RISCO: riskActionIcon,
  DADOS: classificationActionIcon,
  MERCADO: campaignActionIcon,
};

function diagnosticThemeByDomain(domain: Insight["domain"]): DiagnosticTheme {
  if (domain === "COMPETITIVIDADE") return "pricing";
  if (domain === "MERCADO") return "campaign";
  if (domain === "ESTOQUE") return "portfolio";
  if (domain === "RENTABILIDADE") return "classification";
  if (domain === "RISCO") return "risk";
  return "classification";
}

function confidenceLevelByPct(confidence: number | undefined): "high" | "medium" | "low" {
  if (confidence == null || !Number.isFinite(confidence)) return "low";
  if (confidence >= 67) return "high";
  if (confidence >= 33) return "medium";
  return "low";
}

function confidenceLabelByLevel(level: "high" | "medium" | "low"): string {
  if (level === "high") return "Confianca Alta";
  if (level === "medium") return "Confianca Media";
  return "Confianca Baixa";
}

function stockStatusLabel(stockRecoRaw?: WorkspaceInsightsRecommendationItemV2 | null): string {
  return toFriendlyStockStatus(stockRecoRaw);
}

function pricingActionLabelFromRaw(raw?: WorkspaceInsightsRecommendationItemV2 | null): string {
  const action = norm(String(raw?.action || ""));
  if (action.includes("baixar_preco")) return "Abaixar Preco";
  if (action.includes("subir_preco")) return "Aumentar Preco";
  return "Monitorar";
}

function confidenceIconByLevel(level: "high" | "medium" | "low"): string {
  if (level === "high") return highConfidenceIcon;
  if (level === "medium") return mediumConfidenceIcon;
  return lowConfidenceIcon;
}

function diagnosticActionLabel(item: Insight): string {
  if (item.domain === "COMPETITIVIDADE") {
    const action = item.evidence?.find((row) => norm(row.label).includes("acao"))?.value;
    if (action) return action;
    const gapValue = extractNumber(item.evidence?.find((row) => norm(row.label).includes("gap"))?.value);
    if (gapValue != null) {
      if (gapValue > 0) return "Baixar Preco";
      if (gapValue < 0) return "Subir Preco";
    }
    return "Monitorar";
  }
  if (item.domain === "MERCADO") {
    const action = item.evidence?.find((row) => norm(row.label).includes("acao"))?.value;
    if (action) return action;
    return "Monitorar";
  }
  if (item.domain === "ESTOQUE") {
    const status = item.evidence?.find((row) => {
      const label = norm(row.label);
      return label.includes("status") || label.includes("acao");
    })?.value;
    if (status) return status;
    const fallback = item.title?.trim();
    return fallback || "Monitorar";
  }
  if (item.domain === "RENTABILIDADE") {
    const action = item.evidence?.find((row) => norm(row.label).includes("acao"))?.value;
    return action || "Monitorar";
  }
  if (item.domain === "RISCO") {
    if (item.severity === "CRITICAL") return "Alto";
    if (item.severity === "WARN") return "Medio";
    return "Baixo";
  }
  if (item.domain === "DADOS") {
    const role = item.evidence?.find((row) => {
      const label = norm(row.label);
      return label.includes("classificacao") || label.includes("papel") || label.includes("role") || label.includes("segmento");
    })?.value;
    return role || "Sem Classificacao";
  }
  return "Monitorar";
}

function norm(value: string): string {
  return value
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim();
}

function extractNumber(text: string | null | undefined): number | null {
  if (!text) return null;
  const raw = String(text).replace(/[^\d,.-]/g, "").trim();
  if (!raw) return null;
  const hasComma = raw.includes(",");
  const hasDot = raw.includes(".");
  const normalized = hasComma && hasDot
    ? raw.replace(/\./g, "").replace(",", ".")
    : hasComma
      ? raw.replace(",", ".")
      : raw;
  const parsed = Number(normalized);
  return Number.isFinite(parsed) ? parsed : null;
}

function fmtCurrency(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(value)) return "-";
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL", minimumFractionDigits: 2 });
}

function fmtPct(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(value)) return "-";
  return `${value.toFixed(1)}%`;
}

function fmtNumber(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(value)) return "-";
  return value.toLocaleString("pt-BR", { minimumFractionDigits: 0, maximumFractionDigits: 2 });
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

function asNumber(value: unknown): number | null {
  if (value == null || value === "") return null;
  const num = Number(value);
  return Number.isFinite(num) ? num : null;
}

function asTextList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map((item) => String(item || "")).filter((item) => item.trim().length > 0);
}

type ExecutiveLayerSelection = {
  governance: Record<string, unknown>;
  executionRecommendation: Record<string, unknown>;
  decisionState: string;
  actionReadiness: string;
  summary: string[];
};

function selectExecutiveLayer(payload: WorkspaceInsightsV2): ExecutiveLayerSelection {
  const governance = asRecord(payload.governance);
  const executionRecommendation = asRecord(payload.execution_recommendation);
  return {
    governance,
    executionRecommendation,
    decisionState: String(governance.decision_state || "").trim().toUpperCase(),
    actionReadiness: String(governance.action_readiness || "").trim().toUpperCase(),
    summary: asTextList(governance.summary_pt),
  };
}

function impactLabelValue(impact: Record<string, unknown>, key: string, formatter: (value: number | null) => string): string {
  return formatter(asNumber(impact[key]));
}

function getPricingImpactMetrics(impact: Record<string, unknown>) {
  return [
    { label: "Preco atual", value: impactLabelValue(impact, "price_current", fmtCurrency) },
    { label: "Media mercado", value: impactLabelValue(impact, "price_market_mean", fmtCurrency) },
    { label: "Preco alvo", value: impactLabelValue(impact, "price_target", fmtCurrency) },
    { label: "Δ preco", value: impactLabelValue(impact, "delta_price", fmtCurrency) },
    { label: "Δ preco %", value: impactLabelValue(impact, "delta_price_pct", fmtPct) },
    { label: "Contrib atual", value: impactLabelValue(impact, "contrib_unit_current", fmtCurrency) },
    { label: "Contrib alvo", value: impactLabelValue(impact, "contrib_unit_target", fmtCurrency) },
    { label: "Δ contrib", value: impactLabelValue(impact, "delta_contrib_unit", fmtCurrency) },
    { label: "Δ contrib %", value: impactLabelValue(impact, "delta_contrib_unit_pct", fmtPct) },
  ];
}

function getCampaignImpactMetrics(impact: Record<string, unknown>) {
  return [
    { label: "Preco atual", value: impactLabelValue(impact, "price_current", fmtCurrency) },
    { label: "Media mercado", value: impactLabelValue(impact, "price_market_mean", fmtCurrency) },
    { label: "Piso contrib 0", value: impactLabelValue(impact, "price_floor_contrib0", fmtCurrency) },
    { label: "Max desconto", value: impactLabelValue(impact, "max_discount_rs", fmtCurrency) },
    { label: "Max desconto %", value: impactLabelValue(impact, "max_discount_pct", fmtPct) },
    { label: "Estoque", value: impactLabelValue(impact, "estoque_un", fmtNumber) },
    { label: "Capital imob.", value: impactLabelValue(impact, "capital_tied_rs", fmtCurrency) },
  ];
}

function bandScenario(band: Record<string, unknown>, share: number) {
  const scenarios = Array.isArray(band.scenarios) ? band.scenarios : [];
  return scenarios.find((scenario) => Number((scenario as Record<string, unknown>).sell_through_share) === share) as
    | Record<string, unknown>
    | undefined;
}

type SpotlightCTA = {
  label: string;
  action?: "overview" | "history" | "simulator";
  kind: "primary" | "secondary";
  disabled?: boolean;
};

type SpotlightChip = { label: string; value: string };

type SpotlightWhyCard = {
  icon: string;
  title: string;
  text: string;
};

type SpotlightStructuredCard = {
  icon: string;
  title: string;
  textHtml: string;
};

type SpotlightStructuredBullet = {
  icon: string;
  textHtml: string;
};

type SpotlightStructuredChecklist = {
  icon: string;
  title: string;
  textHtml: string;
};

type SpotlightStructuredPhase = {
  key: string;
  icon: string;
  label: string;
  priceRs: number | null;
  deltaPriceRs: number | null;
  discountPct: number | null;
  contribMarginRs: number | null;
  contribMarginPct: number | null;
};

type SpotlightStructuredEvidence = {
  key: string;
  label: string;
  value: number | null;
  fmt: string;
  suffix: string;
  benchmark?: {
    op?: string;
    value?: number | null;
    fmt?: string;
    label?: string;
  };
  criterion?: {
    label?: string;
    role?: string;
    passed?: boolean;
  };
  verdict?: {
    label?: string;
    tone?: string;
  };
  explain?: string;
};

function normalizeConfidence(value: number | null | undefined): number | undefined {
  if (value == null || !Number.isFinite(value)) return undefined;
  if (value <= 1) return Math.max(0, Math.min(100, value * 100));
  return Math.max(0, Math.min(100, value));
}

function formatStructuredValue(row: SpotlightStructuredEvidence): string {
  if (row.value == null || !Number.isFinite(row.value)) return "-";
  if (row.fmt === "currency") return fmtCurrency(row.value);
  if (row.fmt === "pct") return fmtPct(row.value);
  if (row.fmt === "days") return `${Math.round(row.value)} d`;
  if (row.fmt === "score") return `${Math.round(row.value)}`;
  if (row.fmt === "text") return String(row.value);
  return fmtNumber(row.value);
}

function formatEvidenceBenchmark(row: SpotlightStructuredEvidence): string {
  const benchmark = row.benchmark;
  if (!benchmark || benchmark.value == null || !Number.isFinite(benchmark.value)) return "";
  const fmt = String(benchmark.fmt || row.fmt || "").trim();
  const value = formatStructuredValue({ ...row, value: benchmark.value, fmt });
  const opRaw = String(benchmark.op || ">=").trim() || ">=";
  const op = opRaw === ">=" ? "≥" : opRaw === "<=" ? "≤" : opRaw;
  const label = String(benchmark.label || "").trim();
  const actual = formatStructuredValue(row);
  if (label) return `${actual} ${op} ${value} (${label})`;
  return `${actual} ${op} ${value}`;
}


function mapDomain(domainRaw: string): Insight["domain"] {
  const token = norm(domainRaw).replace(/\s+/g, "_");
  if (token.includes("preco") || token.includes("pricing") || token.includes("compet")) return "COMPETITIVIDADE";
  if (token.includes("campanha") || token.includes("campaign") || token.includes("mercado")) return "MERCADO";
  if (token.includes("estoque") || token.includes("stock")) return "ESTOQUE";
  if (token.includes("repo") || token.includes("reposi")) return "ESTOQUE";
  if (token.includes("rentab") || token.includes("portfolio")) return "RENTABILIDADE";
  if (token.includes("risco") || token.includes("risk")) return "RISCO";
  return "DADOS";
}

function categoryFromDomain(domain: Insight["domain"]): Insight["category"] {
  if (domain === "COMPETITIVIDADE") return "preco";
  if (domain === "MERCADO") return "campanha";
  if (domain === "ESTOQUE") return "estoque";
  if (domain === "RENTABILIDADE") return "portfolio";
  return "risco";
}

function isPricingActionCode(actionCodeRaw: string | null | undefined): boolean {
  const token = norm(String(actionCodeRaw || "")).replace(/\s+/g, "_");
  if (!token) return false;
  if (token.includes("preco") || token.includes("price")) return true;
  if (token === "baixar_preco" || token === "subir_preco") return true;
  return false;
}

function severityFromPriority(priority: number | null | undefined): InsightSeverity {
  const value = asNumber(priority);
  if (value == null) return "INFO";
  if (value >= 80) return "CRITICAL";
  if (value >= 60) return "WARN";
  if (value >= 40) return "INFO";
  return "GOOD";
}

function severityFromRiskScore(score: number | null | undefined): InsightSeverity {
  const value = asNumber(score);
  if (value == null) return "INFO";
  if (value >= 80) return "CRITICAL";
  if (value >= 50) return "WARN";
  return "GOOD";
}

function spotlightTheme(domain: Insight["domain"]): "pricing" | "campaign" | "portfolio" | "risk" | "classification" {
  if (domain === "COMPETITIVIDADE") return "pricing";
  if (domain === "MERCADO") return "campaign";
  if (domain === "ESTOQUE") return "portfolio";
  if (domain === "RENTABILIDADE") return "classification";
  if (domain === "RISCO") return "risk";
  return "classification";
}

function recommendationActions(domain: Insight["domain"]): Insight["actions"] {
  if (domain === "COMPETITIVIDADE" || domain === "MERCADO") {
    return [{ label: "Simular", goTo: "simulator", kind: "primary" }];
  }
  if (domain === "RISCO" || domain === "ESTOQUE") {
    return [{ label: "Ver historico", goTo: "history", kind: "secondary" }];
  }
  return [{ label: "Visao geral", goTo: "overview", kind: "secondary" }];
}

function mapRecommendationToInsight(
  item: WorkspaceInsightsRecommendationItemV2,
  idx: number,
  invEvidence: Record<string, unknown>,
  tsEvidence: Record<string, unknown>,
): Insight {
  const domain = mapDomain(String(item.domain || ""));
  const impact = asRecord(item.impact);
  const baseFacts = asRecord(item.facts);
  const tags = Array.isArray(item.tags) ? item.tags : [];
  const alerts = Array.isArray(item.alerts) ? item.alerts : [];
  const actions = Array.isArray(item.actions) ? item.actions : [];
  const actionCode = String(item.action || "");

  const mergedFacts =
    domain === "RISCO"
      ? {
          ...baseFacts,
          dos_days: asNumber(invEvidence.dos),
          pme_days: asNumber(invEvidence.pme_days),
          giro: asNumber(invEvidence.giro),
          estoque_un: asNumber(invEvidence.stock_un),
          trend_un_per_month: asNumber(tsEvidence.slope),
        }
      : baseFacts;
  const chipsInput = {
    tags,
    alerts,
    actions,
    action: actionCode,
    impact,
    facts: mergedFacts,
  };
  let chips: { tags: string[]; alerts: string[]; actions: string[] } = { tags, alerts, actions };
  if (domain === "COMPETITIVIDADE") chips = buildPricingChips(chipsInput);
  else if (domain === "MERCADO") chips = buildCampaignChips(chipsInput);
  else if (domain === "ESTOQUE") chips = buildStockChips(chipsInput);
  else if (domain === "RISCO") chips = buildRiskChips(chipsInput);
  else chips = buildPortfolioChips(chipsInput);

  const evidence: Insight["evidence"] = [];
  if (asNumber(baseFacts.target_price_rs) != null || asNumber(impact.price_target) != null) {
    const targetPrice = asNumber(baseFacts.target_price_rs) != null ? asNumber(baseFacts.target_price_rs) : asNumber(impact.price_target);
    evidence.push({ label: "Preco alvo", value: fmtCurrency(targetPrice) });
  }
  if (asNumber(baseFacts.discount_pct) != null || asNumber(impact.discount_pct) != null) {
    const discountPct = asNumber(baseFacts.discount_pct) != null ? asNumber(baseFacts.discount_pct) : asNumber(impact.discount_pct);
    evidence.push({ label: "Desconto", value: fmtPct(discountPct) });
  }
  if (asNumber(impact.max_discount_pct) != null) evidence.push({ label: "Desconto max", value: fmtPct(asNumber(impact.max_discount_pct)) });
  if (asNumber(impact.price_floor_contrib0) != null) evidence.push({ label: "Piso", value: fmtCurrency(asNumber(impact.price_floor_contrib0)) });

  return {
    id: `reco-${idx}-${domain.toLowerCase()}`,
    title: toFriendlyRecommendationTitle(item),
    summary: toFriendlyRecommendationSummary(item),
    domain,
    category: categoryFromDomain(domain),
    severity: severityFromPriority(item.priority_score),
    confidence: normalizeConfidence(item.confidence_pct),
    evidence,
    tags: chips.tags,
    alerts: chips.alerts,
    actionTags: chips.actions,
    actions: recommendationActions(domain),
  };
}

function buildInsightsFromPayload(
  model: AnalyticsProductWorkspaceV1Dto["model"],
  payload: WorkspaceInsightsV2,
): {
  hero: Insight;
  heroPriorityBadge: string;
  heroPrimaryText: string;
  heroSecondaryText: string;
  diagnostics: InsightSection;
  recommendations: InsightSection;
  heroSimulator: { priceTarget: string; discount: string; priceTargetValue: number | null };
  recommendationsRaw: WorkspaceInsightsRecommendationItemV2[];
  sidebarMetrics: Array<{ label: string; value: string; tone?: "positive" | "default" }>;
  limitsMetrics: Array<{ label: string; value: string }>;
} {
  const stack = asRecord(payload.recommendation_stack);
  const recommendationItemsRaw = Array.isArray(stack.items) ? (stack.items as WorkspaceInsightsRecommendationItemV2[]) : [];
  const evidence = asRecord(payload.evidence);
  const marketEvidence = asRecord(evidence.market);
  const invEvidence = asRecord(evidence.inventory);
  const profitEvidence = asRecord(evidence.profitability);
  const tsEvidence = asRecord(evidence.time_series);
  const executiveLayer = selectExecutiveLayer(payload);
  const governance = executiveLayer.governance;
  const executionRecommendation = executiveLayer.executionRecommendation;
  const recommendations = recommendationItemsRaw.map((item, idx) => mapRecommendationToInsight(item, idx, invEvidence, tsEvidence));
  const rawByDomain = new Map<Insight["domain"], WorkspaceInsightsRecommendationItemV2>();
  for (const item of recommendationItemsRaw) {
    const domain = mapDomain(String(item.domain || ""));
    if (!rawByDomain.has(domain)) rawByDomain.set(domain, item);
  }

  const topByDomain = new Map<Insight["domain"], Insight>();
  for (const item of recommendations) {
    if (!topByDomain.has(item.domain)) topByDomain.set(item.domain, item);
  }

  const priceReco = topByDomain.get("COMPETITIVIDADE");
  const priceRecoRaw =
    rawByDomain.get("COMPETITIVIDADE") ||
    recommendationItemsRaw.find((item) => mapDomain(String(item.domain || "")) === "COMPETITIVIDADE") ||
    null;
  const campaignReco = topByDomain.get("MERCADO");
  const portfolioReco = topByDomain.get("RENTABILIDADE");
  const stockReco = topByDomain.get("ESTOQUE");
  const riskReco = topByDomain.get("RISCO");
  const stockRecoRaw =
    rawByDomain.get("ESTOQUE") ||
    recommendationItemsRaw.find((item) => mapDomain(String(item.domain || "")) === "ESTOQUE") ||
    null;
  const diagnosis = asRecord(payload.diagnosis);
  const riskDiagnosis = asRecord(diagnosis.risk_report);
  const portfolioClassification = asRecord(diagnosis.portfolio_classification);
  const classificationRoleRaw = String(portfolioClassification.role || "").trim();
  const classificationRole =
    classificationRoleRaw &&
    classificationRoleRaw !== "-" &&
    classificationRoleRaw.toUpperCase() !== "NAO_CLASSIFICAR"
      ? classificationRoleRaw.toUpperCase()
      : "";
  const classificationActionCode = classificationRole ? `PORTFOLIO::${classificationRole}` : "";
  const classificationTitle =
    resolveRegistryText({ kind: "ACTION", key: classificationActionCode }, ["uiTitle", "uiLabel"], "") ||
    (classificationRole ? `Papel: ${classificationRole}` : "Sem classificacao");
  const classificationSummary =
    resolveRegistryText({ kind: "ACTION", key: classificationActionCode }, ["uiPrimaryText", "uiSummary", "uiSecondaryText"], "") ||
    (classificationRole ? "Classificacao de portfolio." : "Sem classificacao de portfolio.");
  const classificationConfidence = normalizeConfidence(asNumber(portfolioClassification.confidence));

  const riskRaw = asRecord(rawByDomain.get("RISCO"));
  const riskFacts = asRecord(riskRaw.facts);
  const riskCode = String(riskFacts.risk_flag_code || "-");
  const riskHorizon = asNumber(riskFacts.risk_horizon_days);
  const riskConfidenceRaw = asNumber(riskFacts.risk_confidence_pct);
  const riskConfidence = normalizeConfidence(riskConfidenceRaw != null && riskConfidenceRaw <= 1 ? riskConfidenceRaw * 100 : riskConfidenceRaw);
  const riskSeverityScore = asNumber(riskDiagnosis.severity_score);
  const primaryAction = asRecord(executionRecommendation.primary_action);
  const secondaryAction = asRecord(executionRecommendation.secondary_action);
  const primaryActionCode = String(primaryAction.code || "");
  const executiveDecisionState = executiveLayer.decisionState;
  const executiveActionReadiness = executiveLayer.actionReadiness;
  const executiveSummary = executiveLayer.summary;
  const heroRawByAction = primaryActionCode
    ? recommendationItemsRaw.find((item) => String(item.action || "").trim().toUpperCase() === primaryActionCode.toUpperCase()) || null
    : null;
  const heroCandidate =
    (heroRawByAction
      ? recommendations.find((item, idx) => recommendationItemsRaw[idx] === heroRawByAction)
      : null) ||
    (isPricingActionCode(primaryActionCode)
      ? (recommendations.find((item) => item.domain === "COMPETITIVIDADE") || recommendations[0])
      : recommendations[0]);

  const hero = heroCandidate || {
    id: "hero-empty",
    title: "Sem recomendacoes",
    summary: executiveSummary[0] || "Nenhum insight disponivel para este produto no snapshot atual.",
    domain: "DADOS" as const,
    category: "risco" as const,
    severity: "INFO" as const,
    confidence: undefined,
  };
  const heroRawItem =
    heroRawByAction ||
    (isPricingActionCode(primaryActionCode)
      ? recommendationItemsRaw.find((item) => mapDomain(String(item.domain || "")) === "COMPETITIVIDADE")
      : recommendationItemsRaw[0]);
  const heroRaw = heroRawItem ? asRecord(heroRawItem) : {};
  const heroFacts = asRecord(heroRaw.facts);
  const heroImpact = asRecord(heroRaw.impact);
  const heroPriceTarget =
    asNumber(heroFacts.target_price_rs) != null ? asNumber(heroFacts.target_price_rs) : asNumber(heroImpact.price_target);
  const heroDiscount = asNumber(heroFacts.discount_pct);
  const heroPrimaryText = toFriendlyActionPlanPrimaryText(primaryAction);
  const heroSecondaryText = toFriendlyActionPlanSecondaryText(secondaryAction);
  const heroPriorityBadge =
    executiveDecisionState === "BLOCKED"
      ? "Bloqueado para automacao"
      : executiveDecisionState === "REVIEW"
        ? "Revisao obrigatoria"
        : executiveDecisionState === "READY"
          ? "Pronto para executar"
          : executiveActionReadiness === "CAUTION"
            ? "Execucao com cautela"
            : "Prioridade moderada";

  const hasPriceReco = Boolean(priceReco);
  const hasCampaignReco = Boolean(campaignReco);

  const diagnostics: Insight[] = [
    {
      id: "diag-price",
      title: "Estrategia de Preco",
      summary: priceReco?.summary || "",
      domain: "COMPETITIVIDADE",
      category: "preco",
      severity: priceReco?.severity || "INFO",
      confidence: hasPriceReco ? priceReco?.confidence : undefined,
      evidence: [
        { label: "Acao", value: hasPriceReco ? pricingActionLabelFromRaw(priceRecoRaw) : "Monitorar" },
        { label: "Gap vs mercado", value: fmtPct(asNumber(marketEvidence.gap_pct)) },
        { label: "Mercado medio", value: fmtCurrency(asNumber(marketEvidence.market_mean)) },
      ],
    },
    {
      id: "diag-campaign",
      title: hasCampaignReco ? "Campanha" : "Campanha (sem recomendacao)",
      summary: campaignReco?.summary || "",
      domain: "MERCADO",
      category: "campanha",
      severity: campaignReco?.severity || "INFO",
      confidence: hasCampaignReco ? campaignReco?.confidence : undefined,
      evidence: [
        { label: "Acao", value: campaignReco?.title || "Monitorar" },
        { label: "DOS", value: asNumber(invEvidence.dos) != null ? `${Math.round(asNumber(invEvidence.dos) || 0)} d` : "-" },
        { label: "Estoque", value: asNumber(invEvidence.stock_un) != null ? `${Math.round(asNumber(invEvidence.stock_un) || 0)}` : "-" },
      ],
    },
    {
      id: "diag-stock",
      title: "Status de Estoque",
      summary: stockReco?.summary || "Sem sinais de estoque.",
      domain: "ESTOQUE",
      category: "estoque",
      severity: stockReco?.severity || "INFO",
      confidence: stockReco?.confidence,
      evidence: [{ label: "Status", value: stockStatusLabel(stockRecoRaw) }],
    },
    {
      id: "diag-class",
      title: "Classificacao de Portfolio",
      summary: classificationSummary,
      domain: "DADOS",
      category: "portfolio",
      severity: "INFO",
      confidence: classificationConfidence,
      evidence: [{ label: "Classificacao", value: classificationRole || classificationTitle || "-" }],
    },
    {
      id: "diag-risk",
      title: "Nivel de Risco",
      summary: riskReco?.summary || "Sem flags de risco.",
      domain: "RISCO",
      category: "risco",
      severity: severityFromRiskScore(riskSeverityScore),
      confidence: riskReco?.confidence ? riskConfidence : undefined,
      evidence: [
        { label: "Flag", value: riskCode },
        { label: "Horizonte", value: riskHorizon != null ? `${Math.round(riskHorizon)} d` : "-" },
      ],
    },
  ];

  const margin = asNumber(profitEvidence.margin_pct);
  return {
    hero,
    heroPriorityBadge,
    heroPrimaryText,
    heroSecondaryText,
    diagnostics: {
      title: executiveSummary.length ? "Diagnostico executivo" : "Diagnóstico",
      items: diagnostics,
    },
    recommendations: {
      title: "Recomendacoes",
      items: recommendations,
    },
    recommendationsRaw: recommendationItemsRaw,
    heroSimulator: {
      priceTarget: fmtCurrency(heroPriceTarget),
      discount: fmtPct(heroDiscount),
      priceTargetValue: heroPriceTarget != null ? heroPriceTarget : null,
    },
    sidebarMetrics: [
      { label: "Nosso preco", value: fmtCurrency(asNumber(marketEvidence.our_price)) },
      { label: "Preco medio concorr.", value: fmtCurrency(asNumber(marketEvidence.market_mean)) },
      { label: "Margem", value: fmtPct(margin), tone: margin != null && margin >= 20 ? "positive" : "default" },
      { label: "Giro", value: asNumber(invEvidence.giro) != null ? (asNumber(invEvidence.giro) || 0).toFixed(2) : "-" },
    ],
    limitsMetrics: [
      { label: "Preco piso", value: fmtCurrency(asNumber(asRecord(payload.guardrails).price_floor_contrib0)) },
      { label: "Desconto max.", value: fmtPct(asNumber(asRecord(payload.guardrails).max_discount_pct)) },
      { label: "Margem minima", value: fmtPct(asNumber(asRecord(payload.guardrails).min_margin_safe_for_discount_pct)) },
    ],
  };
}

function buildEmptyInsightsView(model: AnalyticsProductWorkspaceV1Dto["model"]) {
  return {
    hero: {
      id: "hero-empty",
      title: "Sem insights disponiveis",
      summary: "Nao ha payload de insights para este produto no snapshot atual.",
      domain: "DADOS" as const,
      category: "risco" as const,
      severity: "INFO" as const,
      confidence: undefined,
      evidence: [{ label: "SKU", value: model.pn }],
    },
    heroPriorityBadge: "Sem leitura executiva",
    heroPrimaryText: "-",
    heroSecondaryText: "-",
    diagnostics: {
      title: "Diagnóstico",
      items: [] as Insight[],
    },
    recommendations: {
      title: "Recomendacoes",
      items: [] as Insight[],
    },
    recommendationsRaw: [] as WorkspaceInsightsRecommendationItemV2[],
    heroSimulator: {
      priceTarget: "-",
      discount: "-",
      priceTargetValue: null,
    },
    sidebarMetrics: [
      { label: "Nosso preco", value: "-" },
      { label: "Preco medio concorr.", value: "-" },
      { label: "Margem", value: "-", tone: "default" as const },
      { label: "Giro", value: "-" },
    ],
    limitsMetrics: [
      { label: "Preco piso", value: "-" },
      { label: "Desconto max.", value: "-" },
      { label: "Margem minima", value: "-" },
    ],
  };
}

export function InsightsTab() {
  const { api } = useAppSession();
  const navigate = useNavigate();
  const { model } = useOutletContext<ProductWorkspaceOutletContext>();
  const [filter, setFilter] = useState<FilterKey>("all");
  const [insightsPayload, setInsightsPayload] = useState<WorkspaceInsightsV2 | null>(null);
  const [insightsError, setInsightsError] = useState("");
  const [spotlightIndex, setSpotlightIndex] = useState<number | null>(null);
  const [spotlightSelectedPhaseKey, setSpotlightSelectedPhaseKey] = useState<string | null>(null);

  useEffect(() => {
    let disposed = false;
    async function loadInsights() {
      setInsightsError("");
      try {
        const env = await api.products.workspaceInsights(model.pn);
        if (disposed) return;
        const payload = env.data?.insights;
        setInsightsPayload(payload || null);
      } catch (err) {
        if (disposed) return;
        const apiErr = err instanceof ApiClientError ? err : null;
        setInsightsError(apiErr?.message || (err instanceof Error ? err.message : String(err)));
        setInsightsPayload(null);
      }
    }
    void loadInsights();
    return () => {
      disposed = true;
    };
  }, [api, model.pn]);

  const insightView = useMemo(
    () => (insightsPayload ? buildInsightsFromPayload(model, insightsPayload) : buildEmptyInsightsView(model)),
    [insightsPayload, model],
  );

  const recommendationPairs = useMemo(() => {
    const order: Record<Insight["category"], number> = {
      preco: 1,
      campanha: 2,
      estoque: 3,
      portfolio: 3,
      risco: 4,
    };
    return insightView.recommendations.items
      .map((insight, index) => ({
        insight,
        raw: insightView.recommendationsRaw[index],
        index,
      }))
      .sort((a, b) => {
        const aOrder = order[a.insight.category] ?? 9;
        const bOrder = order[b.insight.category] ?? 9;
        if (aOrder !== bOrder) return aOrder - bOrder;
        return a.index - b.index;
      });
  }, [insightView.recommendations.items, insightView.recommendationsRaw]);

  const filteredRecommendations = useMemo(() => {
    if (filter === "all") return recommendationPairs;
    return recommendationPairs.filter((item) => item.insight.category === filter);
  }, [filter, recommendationPairs]);

  const spotlightNavIndexes = useMemo(
    () => filteredRecommendations.map((entry) => entry.index),
    [filteredRecommendations],
  );
  const spotlightNavPos = spotlightIndex == null ? -1 : spotlightNavIndexes.indexOf(spotlightIndex);
  const spotlightHasPrev = spotlightNavPos > 0;
  const spotlightHasNext = spotlightNavPos >= 0 && spotlightNavPos < spotlightNavIndexes.length - 1;

  const spotlightRaw = spotlightIndex != null ? insightView.recommendationsRaw[spotlightIndex] : null;
  const spotlightInsight = spotlightIndex != null ? insightView.recommendations.items[spotlightIndex] : null;
  const spotlightImpact = asRecord(spotlightRaw?.impact);
  const spotlightStructured = asRecord((spotlightRaw as { spotlight?: unknown } | null)?.spotlight);
  const spotlightStructuredSections = asRecord(spotlightStructured.sections);
  const spotlightReasons = asTextList(spotlightRaw?.reasons);
  const spotlightHints = asTextList(spotlightRaw?.hints);
  const spotlightNotes = asTextList(spotlightImpact.notes);
  const spotlightNarrative = asTextList(spotlightImpact.why_narrative);
  const spotlightSummary = spotlightInsight?.summary || spotlightRaw?.summary || "-";
  const spotlightTitle = spotlightInsight?.title || spotlightRaw?.title || "Recomendacao";
  const spotlightConfidence = normalizeConfidence(asNumber(spotlightRaw?.confidence_pct));
  const spotlightDomain = spotlightInsight?.domain || "DADOS";
  const spotlightPriority = spotlightInsight?.severity || "INFO";
  const spotlightExecSummary = asTextList(spotlightImpact.executive_summary);
  const spotlightBands = asRecord(spotlightImpact.bands);
  const spotlightBandsRationale = String(spotlightImpact.bands_rationale || "").trim();
  const spotlightHasStructuredSections = Object.keys(spotlightStructuredSections).length > 0;
  const spotlightHasStructuredCampaign = spotlightDomain === "MERCADO" && spotlightHasStructuredSections;
  const spotlightHasStructuredPricing = spotlightDomain === "COMPETITIVIDADE" && spotlightHasStructuredSections;
  const spotlightHasStructuredRisk = spotlightDomain === "RISCO" && spotlightHasStructuredSections;
  const spotlightHasStructured = spotlightHasStructuredSections;
  const spotlightCampaignMissing = spotlightDomain === "MERCADO" && !spotlightHasStructuredCampaign;
  const spotlightActionCode = String(spotlightRaw?.action || "");
  const spotlightIsLiquidacaoProgressiva =
    spotlightDomain === "MERCADO" && spotlightActionCode.toUpperCase().includes("LIQUIDACAO_PROGRESSIVA");

  const spotlightImportanceSection = asRecord(spotlightStructuredSections.importance);
  const spotlightEvidenceSection = asRecord(spotlightStructuredSections.evidence);
  const spotlightHowPracticeSection = asRecord(spotlightStructuredSections.how_practice);
  const spotlightHowShowBands = Boolean(spotlightHowPracticeSection.show_bands);
  const spotlightExecuteSection = asRecord(spotlightStructuredSections.execute);
  const spotlightPricingMetrics = getPricingImpactMetrics(spotlightImpact);
  const spotlightCampaignMetrics = getCampaignImpactMetrics(spotlightImpact);
  const spotlightInventoryMetrics = [
    { label: "Estoque", value: impactLabelValue(spotlightImpact, "estoque_un", fmtNumber) },
    { label: "Demanda dia", value: impactLabelValue(spotlightImpact, "demanda_diaria", fmtNumber) },
    { label: "DOS", value: impactLabelValue(spotlightImpact, "dos", fmtNumber) },
    { label: "PME", value: impactLabelValue(spotlightImpact, "pme_dias", fmtNumber) },
    { label: "Giro", value: impactLabelValue(spotlightImpact, "giro", fmtNumber) },
    { label: "GMROI", value: impactLabelValue(spotlightImpact, "gmroi", fmtNumber) },
    { label: "Capital", value: impactLabelValue(spotlightImpact, "capital_tied_rs", fmtCurrency) },
  ];

  const spotlightWhyCards: SpotlightWhyCard[] = useMemo(() => {
    const raw = spotlightImpact.why_cards;
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const obj = asRecord(item);
        const icon = String(obj.icon || "").trim();
        const title = String(obj.title || "").trim();
        const text = String(obj.text || "").trim();
        if (!icon || !title || !text) return null;
        return { icon, title, text };
      })
      .filter((item): item is SpotlightWhyCard => item != null);
  }, [spotlightImpact]);

  const spotlightStructuredIntro = useMemo(
    () => asTextList(spotlightImportanceSection.intro_html),
    [spotlightImportanceSection],
  );

  const spotlightStructuredCards: SpotlightStructuredCard[] = useMemo(() => {
    const raw = spotlightImportanceSection.cards;
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const obj = asRecord(item);
        const icon = String(obj.icon || "").trim();
        const title = String(obj.title || "").trim();
        const textHtml = String(obj.text_html || "").trim();
        if (!icon || !title || !textHtml) return null;
        return { icon, title, textHtml };
      })
      .filter((item): item is SpotlightStructuredCard => item != null);
  }, [spotlightImportanceSection]);

  const spotlightStructuredEvidence: SpotlightStructuredEvidence[] = useMemo(() => {
    const raw = spotlightEvidenceSection.items;
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const obj = asRecord(item);
        const key = String(obj.key || "").trim();
        const label = String(obj.label || "").trim();
        const value = asNumber(obj.value);
        const fmt = String(obj.fmt || "number").trim().toLowerCase();
        const suffix = String(obj.suffix || "").trim();
        const benchmarkRaw = asRecord(obj.benchmark);
        const verdictRaw = asRecord(obj.verdict);
        const benchmark =
          Object.keys(benchmarkRaw).length > 0
            ? {
                op: String(benchmarkRaw.op || "").trim() || undefined,
                value: asNumber(benchmarkRaw.value),
                fmt: String(benchmarkRaw.fmt || "").trim().toLowerCase() || undefined,
                label: String(benchmarkRaw.label || "").trim() || undefined,
              }
            : undefined;
        const criterionRaw = asRecord(obj.criterion);
        const criterion =
          Object.keys(criterionRaw).length > 0
            ? {
                label: String(criterionRaw.label || "").trim(),
                role: String(criterionRaw.role || "").trim().toLowerCase(),
                passed: Boolean(criterionRaw.passed),
              }
            : undefined;
        const verdict =
          Object.keys(verdictRaw).length > 0
            ? {
                label: String(verdictRaw.label || "").trim(),
                tone: String(verdictRaw.tone || "").trim().toLowerCase(),
              }
            : undefined;
        const explain = String(obj.explain || "").trim();
        if (!key || !label) return null;
        return { key, label, value, fmt, suffix, benchmark, criterion, verdict, explain };
      })
      .filter((item) => item != null) as SpotlightStructuredEvidence[];
  }, [spotlightEvidenceSection]);

  const spotlightStructuredPhases: SpotlightStructuredPhase[] = useMemo(() => {
    const raw = spotlightHowPracticeSection.phases;
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const obj = asRecord(item);
        const key = String(obj.key || "").trim();
        const icon = String(obj.icon || "").trim();
        const label = String(obj.label || "").trim();
        if (!key || !icon || !label) return null;
        return {
          key,
          icon,
          label,
          priceRs: asNumber(obj.price_rs),
          deltaPriceRs: asNumber(obj.delta_price_rs),
          discountPct: asNumber(obj.discount_pct),
          contribMarginRs: asNumber(obj.contrib_margin_rs),
          contribMarginPct: asNumber(obj.contrib_margin_pct),
        };
      })
      .filter((item): item is SpotlightStructuredPhase => item != null);
  }, [spotlightHowPracticeSection]);

  const spotlightStructuredGuidance: SpotlightStructuredBullet[] = useMemo(() => {
    const raw = spotlightHowPracticeSection.guidance;
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const obj = asRecord(item);
        const icon = String(obj.icon || "").trim();
        const textHtml = String(obj.text_html || "").trim();
        if (!icon || !textHtml) return null;
        return { icon, textHtml };
      })
      .filter((item): item is SpotlightStructuredBullet => item != null);
  }, [spotlightHowPracticeSection]);

  const spotlightStructuredChecklist: SpotlightStructuredChecklist[] = useMemo(() => {
    const raw = spotlightExecuteSection.checklist;
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const obj = asRecord(item);
        const icon = String(obj.icon || "").trim();
        const title = String(obj.title || "").trim();
        const textHtml = String(obj.text_html || "").trim();
        if (!icon || !title || !textHtml) return null;
        return { icon, title, textHtml };
      })
      .filter((item): item is SpotlightStructuredChecklist => item != null);
  }, [spotlightExecuteSection]);

  const spotlightWhyCardsView: SpotlightWhyCard[] = useMemo(() => {
    if (spotlightCampaignMissing) return [];
    if (spotlightHasStructured && spotlightStructuredCards.length > 0) {
      return spotlightStructuredCards.map((card) => ({
        icon: card.icon,
        title: card.title,
        text: card.textHtml.replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim(),
      }));
    }
    return spotlightWhyCards;
  }, [spotlightCampaignMissing, spotlightHasStructured, spotlightStructuredCards, spotlightWhyCards]);

  const spotlightHasWhyCards = spotlightWhyCardsView.length > 0;

  const spotlightEvidence: SpotlightChip[] = useMemo(() => {
    if (spotlightCampaignMissing) return [];
    if (spotlightHasStructured) {
      return spotlightStructuredEvidence
        .map((row) => ({
          label: row.label,
          value: formatStructuredValue(row),
        }))
        .filter((row) => row.value !== "-");
    }
    if (spotlightDomain === "MERCADO") {
      return [
        { label: "Estoque", value: impactLabelValue(spotlightImpact, "estoque_un", fmtNumber) },
        { label: "Capital", value: impactLabelValue(spotlightImpact, "capital_tied_rs", fmtCurrency) },
        { label: "Max desc.", value: impactLabelValue(spotlightImpact, "max_discount_pct", fmtPct) },
        { label: "Piso contrib.", value: impactLabelValue(spotlightImpact, "price_floor_contrib0", fmtCurrency) },
      ].filter((row) => row.value !== "-");
    }
    if (spotlightDomain === "COMPETITIVIDADE") {
      return [
        { label: "Preco atual", value: impactLabelValue(spotlightImpact, "price_current", fmtCurrency) },
        { label: "Media mercado", value: impactLabelValue(spotlightImpact, "price_market_mean", fmtCurrency) },
        { label: "Preco alvo", value: impactLabelValue(spotlightImpact, "price_target", fmtCurrency) },
        { label: "Δ preco", value: impactLabelValue(spotlightImpact, "delta_price_pct", fmtPct) },
      ].filter((row) => row.value !== "-");
    }
    if (spotlightDomain === "ESTOQUE") {
      return spotlightInventoryMetrics.filter((row) => row.value !== "-").slice(0, 5);
    }
    if (spotlightDomain === "RISCO") {
      return [
        {
          label: "Severidade",
          value:
            asNumber(spotlightImpact.severity_score) != null
              ? String(asNumber(spotlightImpact.severity_score))
              : "-",
        },
        { label: "Horizonte", value: asNumber(spotlightImpact.horizon_days) != null ? `${Math.round(asNumber(spotlightImpact.horizon_days) || 0)} d` : "-" },
        { label: "Capital", value: impactLabelValue(spotlightImpact, "capital_parado", fmtCurrency) },
      ].filter((row) => row.value !== "-");
    }
    if (spotlightDomain === "RENTABILIDADE") {
      return [
        { label: "Giro", value: impactLabelValue(spotlightImpact, "giro", fmtNumber) },
        { label: "DOS", value: impactLabelValue(spotlightImpact, "dos", fmtNumber) },
        { label: "PME", value: impactLabelValue(spotlightImpact, "pme", fmtNumber) },
        { label: "Capital", value: impactLabelValue(spotlightImpact, "capital_exposto", fmtCurrency) },
      ].filter((row) => row.value !== "-");
    }
    return [];
  }, [spotlightCampaignMissing, spotlightHasStructured, spotlightStructuredEvidence, spotlightDomain, spotlightImpact]);

  const spotlightPlan: string[] = useMemo(() => {
    if (spotlightCampaignMissing) return [];
    if (spotlightHasStructured) {
      return spotlightStructuredChecklist.map((item) =>
        `${item.icon} ${item.title}: ${item.textHtml.replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim()}`,
      );
    }
    return [];
  }, [spotlightCampaignMissing, spotlightHasStructured, spotlightStructuredChecklist]);

  const spotlightGuardrails: SpotlightChip[] = useMemo(() => {
    if (spotlightDomain === "COMPETITIVIDADE" || spotlightDomain === "MERCADO") {
      return [
        { label: "Piso contrib.", value: impactLabelValue(spotlightImpact, "price_floor_contrib0", fmtCurrency) },
        { label: "Desc. max", value: impactLabelValue(spotlightImpact, "max_discount_pct", fmtPct) },
      ].filter((row) => row.value !== "-");
    }
    if (spotlightDomain === "ESTOQUE") {
      return [
        { label: "Target 90d", value: impactLabelValue(spotlightImpact, "target_stock_90d_un", fmtNumber) },
        { label: "Target 30d", value: impactLabelValue(spotlightImpact, "target_stock_30d_un", fmtNumber) },
      ].filter((row) => row.value !== "-");
    }
    return [];
  }, [spotlightDomain, spotlightImpact]);

  const spotlightWhyTitle = useMemo(() => {
    if (spotlightHasStructured) return String(spotlightImportanceSection.title || "Por que isso e importante?");
    if (spotlightDomain === "RENTABILIDADE") return "O que significa";
    return "Por que importa";
  }, [spotlightHasStructured, spotlightImportanceSection, spotlightDomain]);

  const spotlightEvidenceTitle = useMemo(() => {
    if (spotlightHasStructured) return String(spotlightEvidenceSection.title || "Evidencias do caso");
    if (spotlightDomain === "COMPETITIVIDADE") return "Diagnóstico de preço";
    if (spotlightDomain === "MERCADO") return "Evidencias do caso";
    if (spotlightDomain === "ESTOQUE") return "Diagnostico operacional";
    if (spotlightDomain === "RISCO") return "Evidencias do risco";
    if (spotlightDomain === "RENTABILIDADE") return "Por que este SKU esta aqui";
    return "Evidencias";
  }, [spotlightHasStructured, spotlightEvidenceSection, spotlightDomain]);

  const spotlightPlanTitle = useMemo(() => {
    if (spotlightHasStructured) return String(spotlightExecuteSection.title || "Como executar?");
    if (spotlightDomain === "COMPETITIVIDADE") return "Plano de ajuste";
    if (spotlightDomain === "MERCADO") return "Plano de execucao";
    if (spotlightDomain === "ESTOQUE") return "Acoes recomendadas";
    if (spotlightDomain === "RISCO") return "Mitigacao recomendada";
    if (spotlightDomain === "RENTABILIDADE") return "Estrategia recomendada";
    return "Plano";
  }, [spotlightHasStructured, spotlightExecuteSection, spotlightDomain]);

  const spotlightHowTitle = useMemo(() => {
    if (spotlightHasStructured) return String(spotlightHowPracticeSection.title || "Como funciona na pratica?");
    if (spotlightDomain === "MERCADO") return "Como a campanha funciona";
    return "Como fazer";
  }, [spotlightHasStructured, spotlightHowPracticeSection, spotlightDomain]);

  const spotlightIsLiberacaoCapital =
    spotlightDomain === "MERCADO" && spotlightActionCode.toUpperCase().includes("LIBERACAO_CAPITAL");
  const spotlightPhasePriceLabel = spotlightDomain === "COMPETITIVIDADE" ? "Preço" : "Preco";
  const spotlightPhaseDeltaLabel = spotlightDomain === "COMPETITIVIDADE" ? "Δ preço" : "Delta preco";
  const spotlightPhaseDiscountLabel = spotlightDomain === "COMPETITIVIDADE" ? "Ajuste" : "Desconto";

  const spotlightShowHow =
    (spotlightDomain === "MERCADO" && !spotlightCampaignMissing && (
      spotlightHasStructuredCampaign
        ? ((spotlightHowShowBands && spotlightStructuredPhases.length > 0) || spotlightStructuredGuidance.length > 0)
        : ((!spotlightIsLiberacaoCapital && Object.keys(spotlightBands).length > 0) || spotlightHints.length > 0 || spotlightNotes.length > 0)
    )) ||
    (spotlightHasStructuredPricing && ((spotlightHowShowBands && spotlightStructuredPhases.length > 0) || spotlightStructuredGuidance.length > 0));

  const spotlightCampaignPhases = useMemo(() => {
    if (spotlightCampaignMissing) return [];
    if (spotlightHasStructured && !spotlightHowShowBands) return [];
    if (spotlightHasStructured) {
      return spotlightStructuredPhases.map((phase) => ({
        key: phase.key,
        icon: phase.icon,
        label: phase.label,
        price: fmtCurrency(phase.priceRs),
        deltaPrice: fmtCurrency(phase.deltaPriceRs),
        discountPct: phase.discountPct != null ? `${Math.abs(phase.discountPct).toFixed(1)}%` : "-",
        contrib: fmtCurrency(phase.contribMarginRs),
        contribPct: phase.contribMarginPct != null ? fmtPct(phase.contribMarginPct) : "-",
      }));
    }
    if (spotlightDomain !== "MERCADO") return [];
    if (spotlightIsLiberacaoCapital) return [];
    const order = spotlightIsLiquidacaoProgressiva
      ? ([
          { key: "conservative", label: "Fase 1 - Conservador", icon: "*" },
          { key: "base", label: "Fase 2 - Base", icon: "*" },
          { key: "aggressive", label: "Fase 3 - Agressivo", icon: "*" },
        ] as const)
      : ([
          { key: "conservative", label: "Sugestao - Conservador", icon: "*" },
          { key: "base", label: "Sugestao - Base", icon: "*" },
          { key: "aggressive", label: "Sugestao - Agressivo", icon: "*" },
        ] as const);
    const out: Array<{ key: string; icon: string; label: string; price: string; deltaPrice: string; discountPct: string; contrib: string; contribPct: string }> = [];
    for (const item of order) {
      const band = asRecord(spotlightBands[item.key]);
      if (Object.keys(band).length === 0) continue;
      const price = fmtCurrency(asNumber(band.price));
      const deltaPrice = fmtCurrency(asNumber(band.delta_price));
      const deltaPct = asNumber(band.delta_price_pct);
      const discountPct = deltaPct != null ? `${Math.abs(deltaPct).toFixed(1)}%` : "-";
      const contrib = fmtCurrency(asNumber(band.contrib_margin_rs));
      const contribPct = asNumber(band.contrib_margin_pct);
      out.push({
        key: item.key,
        icon: item.icon,
        label: item.label,
        price,
        deltaPrice,
        discountPct,
        contrib,
        contribPct: contribPct != null ? fmtPct(contribPct) : "-",
      });
    }
    return out;
  }, [spotlightCampaignMissing, spotlightHasStructured, spotlightStructuredPhases, spotlightDomain, spotlightBands, spotlightIsLiquidacaoProgressiva, spotlightHowShowBands, spotlightIsLiberacaoCapital]);

  const spotlightNarrativeView = spotlightCampaignMissing
    ? []
    : spotlightHasStructured && spotlightStructuredIntro.length > 0
      ? spotlightStructuredIntro
      : spotlightNarrative;
  const spotlightHintsView = spotlightCampaignMissing
    ? []
    : spotlightHasStructured && spotlightStructuredGuidance.length > 0
      ? spotlightStructuredGuidance.map((item) => `${item.icon} ${item.textHtml.replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim()}`)
      : spotlightHints;

  const spotlightCtas: SpotlightCTA[] = useMemo(() => {
    if (spotlightDomain === "MERCADO") {
      if (spotlightHasStructuredCampaign) {
        const raw = spotlightStructured.ctas;
        if (Array.isArray(raw) && raw.length > 0) {
          const mapped: SpotlightCTA[] = [];
          for (const item of raw) {
            const obj = asRecord(item);
            const id = String(obj.id || "").trim();
            const label = String(obj.label || "").trim();
            const kind: SpotlightCTA["kind"] =
              String(obj.kind || "secondary").trim().toLowerCase() === "primary" ? "primary" : "secondary";
            const enabled = Boolean(obj.enabled);
            if (!id || !label) continue;
            if (id === "open_simulator") {
              mapped.push({ label, kind, action: "simulator", disabled: !enabled });
              continue;
            }
            mapped.push({ label, kind, disabled: !enabled });
          }
          if (mapped.length > 0) return mapped;
        }
      }
      return [
        { label: "Criar Campanha", kind: "primary", disabled: true },
        { label: "Abrir Simulador", kind: "secondary", action: "simulator" },
      ];
    }
    if (spotlightDomain === "COMPETITIVIDADE") {
      return [
        { label: "Abrir Simulador", kind: "primary", action: "simulator" },
        { label: "Aplicar Preco", kind: "secondary", disabled: true },
      ];
    }
    if (spotlightDomain === "ESTOQUE") {
      return [];
    }
    return [];
  }, [spotlightDomain, spotlightHasStructuredCampaign, spotlightStructured]);

  const selectedPricingTarget = useMemo(() => {
    if (spotlightDomain !== "COMPETITIVIDADE") return null;
    if (!spotlightSelectedPhaseKey) return null;
    const match = spotlightStructuredPhases.find((phase) => phase.key === spotlightSelectedPhaseKey);
    return asNumber(match?.priceRs) ?? null;
  }, [spotlightDomain, spotlightSelectedPhaseKey, spotlightStructuredPhases]);

  function onActionNavigate(target: "overview" | "history" | "simulator" | undefined) {
    if (!target) return;
    if (target === "simulator") {
      const priceTarget =
        spotlightDomain === "COMPETITIVIDADE" && selectedPricingTarget != null
          ? selectedPricingTarget
          : insightView.heroSimulator.priceTargetValue;
      navigate(`/analytics/products/${model.pn}/${target}`, {
        state: { priceTarget },
      });
      return;
    }
    navigate(`/analytics/products/${model.pn}/${target}`);
  }

  function openSpotlight(index: number) {
    setSpotlightIndex(index);
    setSpotlightSelectedPhaseKey(null);
  }

  function closeSpotlight() {
    setSpotlightIndex(null);
    setSpotlightSelectedPhaseKey(null);
  }

  function moveSpotlightBy(delta: number) {
    if (spotlightNavPos < 0) return;
    const nextPos = spotlightNavPos + delta;
    if (nextPos < 0 || nextPos >= spotlightNavIndexes.length) return;
    setSpotlightIndex(spotlightNavIndexes[nextPos]);
  }

  useEffect(() => {
    if (spotlightIndex == null) return;
    const prevOverflow = document.body.style.overflow;
    document.body.classList.add("spotlight-open");
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = prevOverflow;
      document.body.classList.remove("spotlight-open");
    };
  }, [spotlightIndex]);

  useEffect(() => {
    if (spotlightIndex == null) return;
    if (spotlightDomain !== "COMPETITIVIDADE") return;
    if (spotlightSelectedPhaseKey) return;
    if (spotlightCampaignPhases.length === 0) return;
    setSpotlightSelectedPhaseKey(spotlightCampaignPhases[0].key);
  }, [spotlightIndex, spotlightDomain, spotlightSelectedPhaseKey, spotlightCampaignPhases]);

  useEffect(() => {
    if (spotlightIndex == null) return;
    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        closeSpotlight();
      }
    }
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [spotlightIndex]);

  return (
    <section className={styles.insightsRoot}>
      <section className={styles.heroSection}>
        <div className={styles.heroHeader}>
          <div className={styles.heroHeaderLeft}>
            <span className={styles.heroIcon} aria-hidden>
              <img src={actionNeededIcon} alt="" />
            </span>
            <h2 className={styles.heroTitle}>Foco Imediato</h2>
          </div>
        </div>
        <div className={styles.heroDivider} />
        <div className={styles.heroContent}>
          <div className={styles.heroLeft}>
            <div className={styles.heroSubHeader}>
              <span className={styles.heroPriority}>{insightView.heroPriorityBadge}</span>
              <h3 className={styles.heroActionTitle}>{insightView.hero.title}</h3>
            </div>
            <div className={styles.heroActionDivider} />
            <div className={styles.heroActions}>
              <div className={styles.heroActionItem}>
                <span className={styles.heroActionLabel}>Principal:</span>
                <span className={styles.heroActionValue}>{insightView.heroPrimaryText}</span>
              </div>
              <div className={styles.heroActionItem}>
                <span className={styles.heroActionLabel}>Secundaria:</span>
                <span className={styles.heroActionValue}>{insightView.heroSecondaryText}</span>
              </div>
            </div>
          </div>

          <div className={styles.heroRight}>
            <div className={`${styles.heroSimulatorBlock} ${styles.heroSimulatorBlockRight}`}>
              <div className={styles.heroSimulatorMeta}>
                <span className={styles.heroSimulatorMetric}>
                  <span className={styles.heroSimulatorMetricLabel}>Preco-Alvo:</span>{" "}
                  <span className={styles.heroSimulatorMetricValue}>{insightView.heroSimulator.priceTarget}</span>
                </span>
                <span className={styles.heroSimulatorSep}>|</span>
                <span className={styles.heroSimulatorMetric}>
                  <span className={styles.heroSimulatorMetricLabel}>Desconto:</span>{" "}
                  <span className={styles.heroSimulatorMetricValue}>{insightView.heroSimulator.discount}</span>
                </span>
              </div>
              <button
                type="button"
                className={styles.heroCta}
                onClick={() => onActionNavigate("simulator")}
              >
                Abrir Simulador
              </button>
            </div>
          </div>
        </div>
      </section>

      <section className={styles.mainGrid}>
        <div>
          <section className={styles.diagnosticsSection}>
            <h3 className={styles.sectionHeader}>
              <span className={styles.sectionHeaderIcon} aria-hidden>
                <img src={diagnosticIcon} alt="" />
              </span>
              {insightView.diagnostics.title}
            </h3>
            <div className={styles.diagnosticsRow}>
              {insightView.diagnostics.items.map((item) => {
                const theme = diagnosticThemeByDomain(item.domain);
                const confLevel = item.confidence == null ? "high" : confidenceLevelByPct(item.confidence);
                return (
                  <article
                    key={item.id}
                    className={`${styles.diagnosticCard} ${styles[`diagnosticCardTheme_${theme}`]}`}
                    role="button"
                    tabIndex={0}
                    onClick={() => setFilter((prev) => (prev === item.category ? "all" : item.category))}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        setFilter((prev) => (prev === item.category ? "all" : item.category));
                      }
                    }}
                  >
                    <div className={styles.diagHeader}>
                      <span className={styles.diagIcon} aria-hidden>
                        <img src={DIAGNOSTIC_ICON_BY_DOMAIN[item.domain]} alt="" />
                      </span>
                      <span className={styles.diagTitle}>{item.title}</span>
                    </div>
                    <div className={styles.diagDivider} />
                    <div className={styles.diagActionRow}>
                      <span
                        className={`${styles.diagActionIcon} ${styles[`diagActionIconTheme_${theme}`]}`}
                        style={{ "--diag-action-icon": `url(${DIAGNOSTIC_ACTION_ICON_BY_DOMAIN[item.domain]})` } as CSSProperties}
                        aria-hidden
                      />
                      <span className={styles.diagActionValue}>{diagnosticActionLabel(item)}</span>
                    </div>
                    <div className={styles.diagDivider} />
                    <div className={styles.diagConfidenceRow}>
                      <span
                        className={styles.diagConfidenceIcon}
                        style={{ "--diag-confidence-icon": `url(${confidenceIconByLevel(confLevel)})` } as CSSProperties}
                        aria-hidden
                      />
                      <span className={`${styles.diagConfidenceText} ${styles[`diagConfidenceText_${confLevel}`]}`}>
                        {item.confidence == null
                          ? item.domain === "COMPETITIVIDADE"
                            ? "Sem Recomendar"
                            : ""
                          : confidenceLabelByLevel(confLevel)}
                      </span>
                    </div>
                  </article>
                );
              })}
            </div>
          </section>

          <section className={styles.recommendationsSection}>
            <div className={styles.recoHeader}>
              <h3 className={styles.recoTitleRow}>
                <span className={styles.sectionHeaderIcon} aria-hidden>
                  <img src={recommendationIcon} alt="" />
                </span>
                {insightView.recommendations.title}
              </h3>
              <div className={styles.recoHeaderMid}>
                <span className={`${styles.recoHeaderLine} ${styles.recoHeaderLineLeft}`} aria-hidden />
                <div className={styles.filterTabs}>
                  {FILTERS.map((entry) => (
                    <button
                      key={entry.key}
                      type="button"
                      className={`${styles.filterTab}${filter === entry.key ? ` ${styles.filterTabActive}` : ""}`}
                      onClick={() => setFilter(entry.key)}
                    >
                      {entry.label}
                    </button>
                  ))}
                </div>
                <span className={`${styles.recoHeaderLine} ${styles.recoHeaderLineRight}`} aria-hidden />
              </div>
            </div>

            {filteredRecommendations.length > 0 ? (
              <div className={styles.recoList}>
                {filteredRecommendations.map((entry) => (
                  <InsightCard
                    key={entry.insight.id}
                    insight={entry.insight}
                    onClick={() => openSpotlight(entry.index)}
                  />
                ))}
              </div>
            ) : (
              <div className={styles.emptyState}>
                <h4>Nenhum insight neste filtro</h4>
                <p>{insightsError || "Ajuste o filtro para visualizar outras recomendacoes."}</p>
              </div>
            )}
          </section>
        </div>

        <aside className={styles.sidebar}>
          <section className={styles.sidebarCard}>
            <h3 className={styles.sidebarTitle}>Mercado e Estoque</h3>
            {insightView.sidebarMetrics.map((row) => (
              <div key={row.label} className={styles.metricRow}>
                <span className={styles.metricLabel}>
                  <span className={styles.metricLabelIcon} aria-hidden>
                    <img
                      src={
                        row.label === "Nosso preco" || row.label === "Preco medio concorr."
                          ? marketPriceAvgIcon
                          : row.label === "Margem"
                            ? marketMarginIcon
                            : marketTurnoverIcon
                      }
                      alt=""
                    />
                  </span>
                  {row.label}
                </span>
                <span className={`${styles.metricValue}${row.tone === "positive" ? ` ${styles.metricValuePositive}` : ""}`}>{row.value}</span>
              </div>
            ))}
          </section>

          <section className={styles.sidebarCard}>
            <h3 className={styles.sidebarTitle}>Limites</h3>
            {insightView.limitsMetrics.map((row) => (
              <div key={row.label} className={styles.metricRow}>
                <span className={styles.metricLabel}>
                  <span className={styles.metricLabelIcon} aria-hidden>
                    <img
                      src={
                        row.label === "Preco piso"
                          ? guardrailFloorIcon
                          : row.label === "Desconto max."
                            ? guardrailDiscountIcon
                            : marketMarginIcon
                      }
                      alt=""
                    />
                  </span>
                  {row.label}
                </span>
                <span className={styles.metricValue}>{row.value}</span>
              </div>
            ))}
          </section>

          <section className={styles.sidebarCard}>
            <h3 className={styles.sidebarTitle}>
              <span className={styles.sidebarTitleIcon} aria-hidden>
                <img src={diagnosticTraceIcon} alt="" />
              </span>
              Rastro de diagnostico
            </h3>
            <button type="button" className={styles.expandBtn} onClick={() => onActionNavigate("history")}>
              Ver historico completo
            </button>
          </section>
        </aside>
      </section>

      <div
        className={`${styles.spotlightOverlay}${spotlightIndex != null ? ` ${styles.spotlightOverlayActive}` : ""}`}
        onClick={closeSpotlight}
        aria-hidden={spotlightIndex == null}
      />
      <aside
        className={`${styles.spotlightPanel}${spotlightIndex != null ? ` ${styles.spotlightPanelActive}` : ""}`}
        role="dialog"
        aria-modal="true"
        aria-label="Detalhes da recomendacao"
      >
        <div className={styles.spotlightHeader}>
          <span className={`${styles.spotlightBadge} ${styles[`spotlightBadge_${spotlightTheme(spotlightDomain)}`]}`}>
            {spotlightDomain === "COMPETITIVIDADE"
              ? "Preco"
              : spotlightDomain === "MERCADO"
                ? "Campanha"
                : spotlightDomain === "ESTOQUE"
                  ? "Estoque"
                  : spotlightDomain === "RENTABILIDADE"
                    ? "Portfolio"
                    : spotlightDomain === "RISCO"
                ? "Risco"
                : "Dados"}
          </span>
          <div className={styles.spotlightNav} aria-hidden={spotlightIndex == null}>
            <button
              type="button"
              className={styles.spotlightNavBtn}
              onClick={() => moveSpotlightBy(-1)}
              disabled={!spotlightHasPrev}
              aria-label="Insight anterior"
            >
              <span aria-hidden dangerouslySetInnerHTML={{ __html: "&lsaquo;" }} />
            </button>
            <button
              type="button"
              className={styles.spotlightNavBtn}
              onClick={() => moveSpotlightBy(1)}
              disabled={!spotlightHasNext}
              aria-label="Proximo insight"
            >
              <span aria-hidden dangerouslySetInnerHTML={{ __html: "&rsaquo;" }} />
            </button>
          </div>
          <button type="button" className={styles.spotlightClose} onClick={closeSpotlight} aria-label="Fechar spotlight">
            <span aria-hidden dangerouslySetInnerHTML={{ __html: "&times;" }} />
          </button>
          <h3 className={styles.spotlightTitle}>{spotlightTitle}</h3>
          <p className={styles.spotlightSubtitle}>{spotlightSummary}</p>
          <div className={styles.spotlightMeta}>
            <span className={styles.spotlightMetaItem}>
              Prioridade: <strong>{spotlightPriority === "CRITICAL" ? "Alta" : spotlightPriority === "WARN" ? "Media" : "Baixa"}</strong>
            </span>
            <span className={styles.spotlightMetaItem}>
              Confianca: <strong>{spotlightConfidence != null ? `${Math.round(spotlightConfidence)}%` : "-"}</strong>
            </span>
            {spotlightRaw?.action ? (
              <span className={styles.spotlightMetaItem}>
                Acao: <strong>{String(spotlightRaw.action)}</strong>
              </span>
            ) : null}
          </div>
        </div>

        <div className={styles.spotlightBody}>
          {!spotlightCampaignMissing ? (
            <section className={styles.spotlightSection}>
              <div className={styles.spotlightSectionHeader}>
                <span className={`${styles.spotlightSectionIcon} ${styles.spotlightSectionIconWhy}`} aria-hidden>
                  💡
                </span>
                <h4 className={styles.spotlightSectionTitle}>{spotlightWhyTitle}</h4>
              </div>
              <div className={styles.spotlightSectionContent}>
                {spotlightNarrativeView.length > 0 ? (
                  <div className={styles.spotlightNarrative}>
                    {(spotlightHasStructuredPricing && spotlightHasWhyCards
                      ? spotlightNarrativeView.slice(0, 1)
                      : spotlightNarrativeView
                    ).map((line, idx) => (
                      <p
                        key={`narrative-${idx}`}
                        className={styles.spotlightNarrativeText}
                        dangerouslySetInnerHTML={{ __html: line }}
                      />
                    ))}
                  </div>
                ) : spotlightNarrativeView.length === 0 && !spotlightHasWhyCards ? (
                  <p className={styles.spotlightMuted}>Sem justificativas detalhadas neste item.</p>
                ) : null}
                {(spotlightDomain === "MERCADO" ||
                  spotlightDomain === "COMPETITIVIDADE" ||
                  spotlightDomain === "RISCO" ||
                  spotlightDomain === "RENTABILIDADE" ||
                  spotlightDomain === "ESTOQUE") &&
                spotlightHasWhyCards ? (
                  <div className={styles.spotlightWhyBulletList}>
                    {spotlightWhyCardsView.slice(0, 6).map((card, idx) => (
                      <div key={`why-card-${idx}`} className={styles.spotlightWhyBulletItem}>
                        <span className={styles.spotlightWhyBulletIcon} aria-hidden>
                          {card.icon}
                        </span>
                        <span className={styles.spotlightWhyBulletText}>
                          <strong>{card.title}:</strong> {card.text}
                        </span>
                      </div>
                    ))}
                  </div>
                ) : null}
                {spotlightNarrativeView.length === 0 && !spotlightHasWhyCards && spotlightReasons.length > 0 ? (
                  <div className={styles.spotlightBulletList}>
                    {spotlightReasons.map((reason, idx) => (
                      <div key={`reason-${idx}`} className={styles.spotlightBulletItem}>
                        <span className={styles.spotlightBulletIcon} dangerouslySetInnerHTML={{ __html: "&bull;" }} />
                        <span className={styles.spotlightBulletText}>{reason}</span>
                      </div>
                    ))}
                  </div>
                ) : null}
              </div>
            </section>
          ) : null}

          {!spotlightCampaignMissing ? (
            <section className={styles.spotlightSection}>
              <div className={styles.spotlightSectionHeader}>
                <span className={`${styles.spotlightSectionIcon} ${styles.spotlightSectionIconWhat}`} aria-hidden>
                  🧾
                </span>
                <h4 className={styles.spotlightSectionTitle}>{spotlightEvidenceTitle}</h4>
              </div>
            <div className={styles.spotlightSectionContent}>
              {spotlightHasStructured ? (
                spotlightStructuredEvidence.length > 0 ? (
                  <div className={styles.spotlightChipGrid}>
                    {spotlightStructuredEvidence
                      .map((row) => {
                        const value = formatStructuredValue(row);
                        if (value === "-") return null;
                        const benchmarkText = formatEvidenceBenchmark(row);
                        const hasBenchmark = Boolean(benchmarkText);
                        return (
                          <div key={row.key} className={styles.spotlightChip}>
                            <div className={styles.spotlightChipHeader}>
                              <span className={styles.spotlightChipLabel}>{row.label}</span>
                            </div>
                            <span className={styles.spotlightChipValue}>{value}</span>
                            {hasBenchmark ? <span className={styles.spotlightChipMeta}>{benchmarkText}</span> : null}
                          </div>
                        );
                      })
                      .filter(Boolean)}
                  </div>
                ) : (
                  <p className={styles.spotlightMuted}>Sem evidencias numericas para este item.</p>
                )
              ) : spotlightEvidence.length > 0 ? (
                <div className={styles.spotlightChipGrid}>
                  {spotlightEvidence.map((row) => (
                    <div key={row.label} className={styles.spotlightChip}>
                      <span className={styles.spotlightChipLabel}>{row.label}</span>
                      <span className={styles.spotlightChipValue}>{row.value}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className={styles.spotlightMuted}>Sem evidencias numericas para este item.</p>
              )}
            </div>
          </section>
          ) : null}

          {spotlightShowHow ? (
            <section className={styles.spotlightSection}>
              <div className={styles.spotlightSectionHeader}>
                <span className={`${styles.spotlightSectionIcon} ${styles.spotlightSectionIconHow}`} aria-hidden>
                  🧭
                </span>
                <h4 className={styles.spotlightSectionTitle}>{spotlightHowTitle}</h4>
              </div>
              <div className={styles.spotlightSectionContent}>
                {(spotlightDomain === "MERCADO" || spotlightDomain === "COMPETITIVIDADE") && spotlightCampaignPhases.length > 0 ? (
                  <div className={styles.spotlightPhaseGrid}>
                    {spotlightCampaignPhases.map((phase) => (
                      <div
                        key={phase.key}
                        className={`${styles.spotlightPhaseCard}${
                          spotlightDomain === "COMPETITIVIDADE" ? ` ${styles.spotlightPhaseCardSelectable}` : ""
                        }${
                          spotlightDomain === "COMPETITIVIDADE" && spotlightSelectedPhaseKey === phase.key
                            ? ` ${styles.spotlightPhaseCardSelected}`
                            : ""
                        }`}
                        role={spotlightDomain === "COMPETITIVIDADE" ? "button" : undefined}
                        tabIndex={spotlightDomain === "COMPETITIVIDADE" ? 0 : undefined}
                        onClick={() => {
                          if (spotlightDomain === "COMPETITIVIDADE") {
                            setSpotlightSelectedPhaseKey(phase.key);
                          }
                        }}
                        onKeyDown={(event) => {
                          if (spotlightDomain !== "COMPETITIVIDADE") return;
                          if (event.key === "Enter" || event.key === " ") {
                            event.preventDefault();
                            setSpotlightSelectedPhaseKey(phase.key);
                          }
                        }}
                      >
                        <div className={styles.spotlightPhaseHeader}>
                          <span className={styles.spotlightPhaseIcon} aria-hidden>
                            {phase.icon}
                          </span>
                          <span className={styles.spotlightPhaseTitle}>{phase.label}</span>
                        </div>
                        <div className={styles.spotlightPhaseRow}>
                          <span>{spotlightPhasePriceLabel}</span>
                          <strong>{phase.price}</strong>
                        </div>
                        <div className={styles.spotlightPhaseRow}>
                          <span>{spotlightPhaseDeltaLabel}</span>
                          <strong>{phase.deltaPrice}</strong>
                        </div>
                        <div className={styles.spotlightPhaseRow}>
                          <span>{spotlightPhaseDiscountLabel}</span>
                          <strong>{phase.discountPct}</strong>
                        </div>
                        <div className={styles.spotlightPhaseRow}>
                          <span>Margem contrib. R$</span>
                          <strong>{phase.contrib}</strong>
                        </div>
                        <div className={styles.spotlightPhaseRow}>
                          <span>Margem contrib. %</span>
                          <strong>{phase.contribPct}</strong>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : null}
                {spotlightHintsView.length > 0 ? (
                  <div className={styles.spotlightBulletList}>
                    {spotlightHintsView.map((hint, idx) => (
                      <div key={`hint-${idx}`} className={styles.spotlightBulletItem}>
                        <span className={styles.spotlightBulletIcon} dangerouslySetInnerHTML={{ __html: "&bull;" }} />
                        <span className={styles.spotlightBulletText}>{hint}</span>
                      </div>
                    ))}
                  </div>
                ) : null}
                {spotlightNotes.length > 0 && !spotlightHasStructured ? (
                  <div className={styles.spotlightBulletList}>
                    {spotlightNotes.map((note, idx) => (
                      <div key={`note-${idx}`} className={styles.spotlightBulletItem}>
                        <span className={styles.spotlightBulletIcon} dangerouslySetInnerHTML={{ __html: "&bull;" }} />
                        <span className={styles.spotlightBulletText}>{note}</span>
                      </div>
                    ))}
                  </div>
                ) : null}
              </div>
            </section>
          ) : null}

          {!spotlightCampaignMissing && spotlightDomain !== "COMPETITIVIDADE" ? (
            <section className={styles.spotlightSection}>
              <div className={styles.spotlightSectionHeader}>
                <span className={`${styles.spotlightSectionIcon} ${styles.spotlightSectionIconWhat}`} aria-hidden>
                  ✅
                </span>
                <h4 className={styles.spotlightSectionTitle}>{spotlightPlanTitle}</h4>
              </div>
              <div className={styles.spotlightSectionContent}>
                {spotlightHasStructured ? (
                  spotlightStructuredChecklist.length > 0 ? (
                    <div className={styles.spotlightPlanList}>
                      {spotlightStructuredChecklist.map((step, idx) => (
                        <div key={`structured-plan-${idx}`} className={styles.spotlightPlanItem}>
                          <span className={`${styles.spotlightPlanIndex} ${styles.spotlightPlanIndexEmoji}`} aria-hidden>
                            {step.icon}
                          </span>
                          <span className={styles.spotlightPlanText}>
                            <strong>{step.title}:</strong> <span dangerouslySetInnerHTML={{ __html: step.textHtml }} />
                          </span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className={styles.spotlightMuted}>Sem plano sugerido para este item.</p>
                  )
                ) : spotlightPlan.length > 0 ? (
                  <div className={styles.spotlightPlanList}>
                    {spotlightPlan.map((step, idx) => (
                      <div key={`plan-${idx}`} className={styles.spotlightPlanItem}>
                        <span
                          className={`${styles.spotlightPlanIndex}${spotlightDomain === "MERCADO" ? ` ${styles.spotlightPlanIndexEmoji}` : ""}`}
                          aria-hidden
                        >
                          {String(idx + 1)}
                        </span>
                        <span className={styles.spotlightPlanText}>{step}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className={styles.spotlightMuted}>Sem plano sugerido para este item.</p>
                )}
              </div>
            </section>
          ) : null}

          {spotlightCtas.length > 0 ? (
            <div className={styles.spotlightCtaRowWrap}>
              <div className={styles.spotlightCtaRow}>
                {spotlightCtas.map((cta) => (
                  <button
                    key={cta.label}
                    type="button"
                    className={`${styles.actionBtn} ${cta.kind === "primary" ? styles.actionBtnPrimary : styles.actionBtnSecondary}${cta.disabled ? ` ${styles.spotlightCtaDisabled}` : ""}`}
                    onClick={() => {
                      if (!cta.disabled) onActionNavigate(cta.action);
                    }}
                    disabled={cta.disabled}
                  >
                    {cta.label}
                  </button>
                ))}
              </div>
            </div>
          ) : null}

          {!spotlightHasStructured && spotlightGuardrails.length > 0 ? (
          <section className={styles.spotlightSection}>
              <div className={styles.spotlightSectionHeader}>
                <span className={`${styles.spotlightSectionIcon} ${styles.spotlightSectionIconHow}`} aria-hidden>
                  🛡️
                </span>
                <h4 className={styles.spotlightSectionTitle}>Regras e limites</h4>
              </div>
              <div className={styles.spotlightSectionContent}>
                <div className={styles.spotlightChipGrid}>
                  {spotlightGuardrails.map((row) => (
                    <div key={row.label} className={styles.spotlightChip}>
                      <span className={styles.spotlightChipLabel}>{row.label}</span>
                      <span className={styles.spotlightChipValue}>{row.value}</span>
                    </div>
                  ))}
                </div>
              </div>
          </section>
          ) : null}

        </div>
      </aside>
    </section>
  );
}
