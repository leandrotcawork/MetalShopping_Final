import { resolveChipText, resolveIconKey } from "../../../registry/analyticsRegistry";

export type ChipTriplet = {
  tags: string[];
  alerts: string[];
  actions: string[];
};

type ChipsInput = {
  tags?: string[];
  alerts?: string[];
  actions?: string[];
  action?: string;
  impact?: Record<string, unknown>;
  facts?: Record<string, unknown>;
};

function asNumber(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const n = Number(value.replace(",", "."));
    return Number.isFinite(n) ? n : null;
  }
  return null;
}

function asText(value: unknown): string {
  return value == null ? "" : String(value).trim();
}

export function fmtPctShort(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(value)) return "";
  if (Math.abs(value - Math.round(value)) < 0.05) return `${Math.round(value)}%`;
  return `${value.toFixed(1).replace(".", ",")}%`;
}

function titleCaseChip(text: string): string {
  const words = String(text || "").split(" ").filter(Boolean);
  if (words.length === 0) return "";
  const lowers = new Set(["de", "da", "do", "das", "dos", "e", "em", "para", "por", "com"]);
  return words
    .map((word, idx) => {
      const lw = word.toLowerCase();
      return idx > 0 && lowers.has(lw) ? lw : lw.charAt(0).toUpperCase() + lw.slice(1);
    })
    .join(" ");
}

export function normalizeChipToken(value: string): string {
  return String(value || "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toUpperCase()
    .replace(/%/g, "")
    .replace(/[^A-Z0-9]+/g, "_")
    .replace(/_+/g, "_")
    .replace(/^_+|_+$/g, "");
}

export function friendlyChipText(value: string): string {
  const token = asText(value);
  if (!token) return "";
  const fromRegistry = resolveChipText(token);
  let text = fromRegistry && fromRegistry !== token ? fromRegistry : token.includes("_") ? titleCaseChip(token.replace(/_/g, " ")) : token;

  const normalized = normalizeChipToken(text);
  const emojiMap: Record<string, string> = {
    PRICE_ABOVE_MARKET: "📈",
    PRICE_BELOW_MARKET: "📉",
    PRECO_ACIMA: "📈",
    PRECO_ABAIXO: "📉",
    PRECO_ALVO: "🎯",
    MARGEM_ALVO: "🧮",
    DELTA_PRECO: "↕️",
    DESCONTO_MAX: "🏷️",
    PISO_DE_PRECO: "🧱",
    COBERTURA_DADOS: "🧪",
    CONFIANCA_DADOS: "✅",
    DOS: "📦",
    PME: "📆",
    GIRO: "🔁",
    GMROI: "💰",
  };
  const emoji = emojiMap[normalized] || "";
  if (emoji && !text.startsWith(emoji)) text = `${emoji} ${text}`;
  return text;
}

function isPriceDownActionToken(value: string): boolean {
  const normalized = normalizeChipToken(value);
  if (
    normalized === "BAIXAR_PRECO" ||
    normalized === "AJUSTAR_PRECO_BAIXAR" ||
    normalized === "AJUSTAR_PRECO_PARA_BAIXAR" ||
    normalized === "REDUZIR_PRECO" ||
    normalized.includes("BAIXAR_PRECO") ||
    normalized.includes("AJUSTAR_PRECO_BAIXAR")
  ) return true;
  const hasDown = normalized.includes("BAIXAR") || normalized.includes("REDUZIR");
  const hasPrice = normalized.includes("PRECO") || normalized.includes("PRICE");
  return hasDown && hasPrice;
}

function dedupePush(target: string[], value: string): void {
  const text = friendlyChipText(value);
  if (text && !target.includes(text)) target.push(titleCaseChip(text));
}

function pricingTargetMarginPct(impact: Record<string, unknown>): number | null {
  const target = asNumber(impact.price_target);
  const contrib = asNumber(impact.contrib_unit_target);
  if (target == null || contrib == null || target <= 0) return null;
  return (contrib / target) * 100;
}

export function buildPricingChips(input: ChipsInput): ChipTriplet {
  const tags = input.tags || [];
  const alerts = input.alerts || [];
  const actions = input.actions || [];
  const actionUpper = asText(input.action).toUpperCase();
  const impact = input.impact || {};
  const outTags: string[] = [];
  const outAlerts: string[] = [];
  const outActions: string[] = [];

  if (alerts.includes("price_above_market")) outAlerts.push("price_above_market");
  if (alerts.includes("price_below_market")) outAlerts.push("price_below_market");

  const targetMargin = pricingTargetMarginPct(impact);
  if (targetMargin != null) outTags.push(`Margem Alvo: ${fmtPctShort(targetMargin)}`);

  const targetPrice = asNumber(impact.price_target);
  if (targetPrice != null) {
    outTags.push(`Preco Alvo: ${targetPrice.toLocaleString("pt-BR", { style: "currency", currency: "BRL", minimumFractionDigits: 2 })}`);
  }

  const delta = asNumber(impact.delta_price_pct);
  if (delta != null) {
    outTags.push(`Delta Preco: ${fmtPctShort(delta)}`);
    if (!alerts.includes("price_above_market") && !alerts.includes("price_below_market")) {
      if (delta < 0) outAlerts.push("price_above_market");
      else if (delta > 0) outAlerts.push("price_below_market");
    }
  }

  if (actionUpper === "SUBIR_PRECO") outActions.push("Ajustar Preco para Subir");

  for (const token of tags) {
    if (actionUpper === "BAIXAR_PRECO" && isPriceDownActionToken(token)) continue;
    dedupePush(outTags, token);
  }
  for (const token of alerts) dedupePush(outAlerts, token);
  for (const token of actions) {
    if (actionUpper === "BAIXAR_PRECO" && isPriceDownActionToken(token)) continue;
    dedupePush(outActions, token);
  }

  return { tags: outTags.slice(0, 4), alerts: outAlerts.slice(0, 4), actions: outActions.slice(0, 4) };
}

export function buildCampaignChips(input: ChipsInput): ChipTriplet {
  const tags = input.tags || [];
  const alerts = input.alerts || [];
  const actions = input.actions || [];
  const actionUpper = asText(input.action).toUpperCase();
  const impact = input.impact || {};
  const outTags: string[] = [];
  const outAlerts: string[] = [];
  const outActions: string[] = [];

  const riskLabel =
    alerts.includes("overstock") ? "Risco de Overstock" :
    alerts.includes("estoque_parado") ? "Estoque Parado" :
    alerts.includes("giro_critico") ? "Giro Critico" :
    alerts.includes("margem_apertada") ? "Margem Apertada" :
    alerts.includes("price_above_market") ? "Preco Acima do Mercado" :
    alerts.includes("price_below_market") ? "Preco Abaixo do Mercado" :
    "";
  if (riskLabel) outTags.push(riskLabel);

  const maxDiscount = asNumber(impact.max_discount_pct);
  if (maxDiscount != null) outTags.push(`Desconto Max: ${fmtPctShort(maxDiscount)}`);

  const floor = asNumber(impact.price_floor_contrib0);
  if (floor != null) outTags.push(`Piso de Preco: ${floor.toLocaleString("pt-BR", { style: "currency", currency: "BRL", minimumFractionDigits: 2 })}`);

  const estoque = asNumber(impact.estoque_un);
  const capital = asNumber(impact.capital_tied_rs);
  const dos = asNumber(impact.dos);
  const pme = asNumber(impact.pme_dias);
  const giro = asNumber(impact.giro);
  const gmroi = asNumber(impact.gmroi);
  const coverage = asNumber(impact.coverage_ratio);
  const tier = asText(impact.confidence_tier).toUpperCase();

  if (actionUpper.includes("LIBERAR CAPITAL") || actionUpper.includes("LIQUIDACAO") || actionUpper.includes("OUTLET")) {
    if (estoque != null) outTags.push(`Estoque: ${Math.round(estoque)} un`);
    if (capital != null) outTags.push(`Capital: ${capital.toLocaleString("pt-BR", { style: "currency", currency: "BRL", minimumFractionDigits: 0 })}`);
    if (dos != null) outTags.push(`DOS: ${Math.round(dos)} d`);
    if (pme != null) outTags.push(`PME: ${Math.round(pme)} d`);
  } else if (actionUpper.includes("TRACAO") || actionUpper.includes("OPORTUNIDADE") || actionUpper.includes("ANCORAGEM")) {
    if (giro != null) outTags.push(`Giro: ${giro.toFixed(2)}`);
    if (gmroi != null) outTags.push(`GMROI: ${gmroi.toFixed(2)}`);
    if (dos != null) outTags.push(`DOS: ${Math.round(dos)} d`);
  } else if (actionUpper.includes("TESTE") || actionUpper.includes("REATIVACAO")) {
    if (dos != null) outTags.push(`DOS: ${Math.round(dos)} d`);
    if (giro != null) outTags.push(`Giro: ${giro.toFixed(2)}`);
    if (coverage != null) outTags.push(`Cobertura: ${fmtPctShort(coverage)}`);
  }

  if (coverage != null) outTags.push(`Cobertura Dados: ${fmtPctShort(coverage)}`);
  if (tier) {
    const tierPt = ({ HIGH: "Alta", MEDIUM: "Media", LOW: "Baixa" } as Record<string, string>)[tier] || titleCaseChip(tier.toLowerCase());
    outTags.push(`Confianca Dados: ${tierPt}`);
  }

  if (actionUpper.includes("LIBERAR CAPITAL")) outActions.push("Liberar Capital");
  else if (actionUpper.includes("OUTLET")) outActions.push("Campanha Outlet");
  else if (actionUpper.includes("TESTE DE DEMANDA") || actionUpper.includes("TESTE_DEMANDA")) outActions.push("Teste de Demanda");
  else if (actionUpper.includes("BUNDLE MIX") || actionUpper.includes("BUNDLE_MIX")) outActions.push("Bundle Mix");
  else if (actionUpper.includes("REATIVACAO SELETIVA") || actionUpper.includes("REATIVACAO_SELETIVA")) outActions.push("Reativacao Seletiva");
  else if (actionUpper.includes("LIQUIDACAO PROGRESSIVA") || actionUpper.includes("LIQUIDACAO_PROGRESSIVA")) outActions.push("Liquidacao Progressiva");
  else if (actionUpper.includes("TRACAO")) outActions.push("Campanha de Tracao");
  else if (actionUpper.includes("ANCORAGEM")) outActions.push("Ancoragem de Preco");
  else if (actionUpper.includes("OPORTUNIDADE")) outActions.push("Oportunidade Competitiva");

  for (const token of tags) dedupePush(outTags, token);
  for (const token of alerts) dedupePush(outAlerts, token);
  for (const token of actions) dedupePush(outActions, token);

  return { tags: outTags.slice(0, 4), alerts: outAlerts.slice(0, 4), actions: outActions.slice(0, 4) };
}

export function buildStockChips(input: ChipsInput): ChipTriplet {
  const impact = input.impact || {};
  const outTags: string[] = [];
  const outAlerts: string[] = [];
  const outActions: string[] = [];

  const dos = asNumber(impact.dos);
  if (dos != null) outTags.push(`DOS: ${Math.round(dos)} d`);
  const pme = asNumber(impact.pme_dias);
  if (pme != null) outTags.push(`PME: ${Math.round(pme)} d`);
  const giro = asNumber(impact.giro);
  if (giro != null) outTags.push(`Giro: ${giro.toFixed(2)}`);
  const gmroi = asNumber(impact.gmroi);
  if (gmroi != null) outTags.push(`GMROI: ${gmroi.toFixed(2)}`);

  const overstockExcess = asNumber(impact.overstock_excess_un);
  if (overstockExcess != null) outTags.push(`Excesso: ${Math.round(overstockExcess)} un`);
  const rupturaLack = asNumber(impact.ruptura_lack_un);
  if (rupturaLack != null) outTags.push(`Falta: ${Math.round(rupturaLack)} un`);

  const actionUpper = asText(input.action).toUpperCase();
  if (actionUpper.includes("REDUZIR") || actionUpper.includes("LIBERAR")) outActions.push("Liberar Capital");
  if (actionUpper.includes("REVER")) outActions.push("Rever Compras");

  for (const token of input.tags || []) dedupePush(outTags, token);
  for (const token of input.alerts || []) dedupePush(outAlerts, token);
  for (const token of input.actions || []) dedupePush(outActions, token);

  return { tags: outTags.slice(0, 4), alerts: outAlerts.slice(0, 4), actions: outActions.slice(0, 4) };
}

export function buildRiskChips(input: ChipsInput): ChipTriplet {
  const facts = input.facts || {};
  const outTags: string[] = [];
  const outAlerts: string[] = [];
  const outActions: string[] = [];

  const actionText = asText(input.action);
  const code = actionText.includes("RISK::") ? actionText.split("RISK::")[1] : asText(facts.risk_flag_code);
  if (code) outTags.push(`Risco: ${titleCaseChip(code.replace(/_/g, " ").toLowerCase())}`);

  const horizon = asNumber(facts.risk_horizon_days);
  if (horizon != null) outTags.push(`Horizonte: ${Math.round(horizon)} d`);

  const dos = asNumber(facts.dos_days);
  const pme = asNumber(facts.pme_days);
  const giro = asNumber(facts.giro);
  const trend = asNumber(facts.trend_un_per_month);

  const codeNorm = normalizeChipToken(code);
  if (codeNorm === "CAPITAL_IMOBILIZADO") {
    if (dos != null) outTags.push(`DOS: ${Math.round(dos)} d`);
    if (pme != null) outTags.push(`PME: ${Math.round(pme)} d`);
  } else if (codeNorm === "GIRO_CRITICO") {
    if (giro != null) outTags.push(`Giro: ${giro.toFixed(2)}`);
    if (trend != null) outTags.push(`Tendencia: ${trend.toFixed(2)} un/mes`);
  }

  return { tags: outTags.slice(0, 4), alerts: outAlerts.slice(0, 4), actions: outActions.slice(0, 4) };
}

export function buildPortfolioChips(input: ChipsInput): ChipTriplet {
  const outTags: string[] = [];
  const outAlerts: string[] = [];
  const outActions: string[] = [];

  const actionText = asText(input.action);
  if (actionText.startsWith("PORTFOLIO::")) {
    const role = actionText.replace("PORTFOLIO::", "");
    outTags.push(`Papel: ${titleCaseChip(role.replace(/_/g, " ").toLowerCase())}`);
  }

  // Mantem portfolio restrito ao papel para evitar chips de estoque/reposicao.

  return { tags: outTags.slice(0, 4), alerts: outAlerts.slice(0, 4), actions: outActions.slice(0, 4) };
}

export function chipIconKey(tokenText: string): string | null {
  const fromRegistry = resolveIconKey(tokenText);
  if (fromRegistry) return fromRegistry;
  const token = normalizeChipToken(tokenText);
  if (token.startsWith("MARGEM")) return "market_margin";
  if (token.startsWith("MARKUP")) return "market_margin";
  if (token.startsWith("CUSTO_MEDIO")) return "market_margin";
  if (token.startsWith("PRECO_ALVO") || token.startsWith("PISO_DE_PRECO")) return "guardrail_floor";
  if (token.includes("PRECO_ACIMA") || token.includes("PRECO_ABAIXO") || token.includes("PRECO_ACIMA_DE_MERCADO")) return "market_price_avg";
  if (token.startsWith("PME") || token.startsWith("GIRO")) return "market_turnover";
  if (token.startsWith("TENDENCIA")) return "market_price_avg";
  if (token.startsWith("RISCO") || token.includes("TENDENCIA")) return "risk_action";
  if (token.startsWith("SEGMENTO") || token.startsWith("CLASSE")) return "classification_action";
  if (token.includes("AJUSTAR_PRECO_PARA_BAIXAR") || token.includes("REDUZIR_PRECO") || token.startsWith("DESCONTO_MAX")) return "guardrail_discount";
  if (
    token.includes("LIBERAR_CAPITAL") ||
    token.includes("REVER_COMPRAS") ||
    token.includes("CAMPANHA_OUTLET") ||
    token.includes("CAMPANHA_DE_TRACAO") ||
    token.includes("ANCORAGEM_DE_PRECO") ||
    token.includes("OPORTUNIDADE_COMPETITIVA")
  ) return "campaign_action";
  if (token.startsWith("COBERTURA_DADOS") || token.startsWith("CONFIANCA_DADOS")) return "diagnostic_trace";
  return null;
}
