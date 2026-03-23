// @ts-nocheck
import type { AnalyticsHomeViewModel } from "./analyticsHomeViewModel";

type SignalTone = "info" | "warn" | "crit" | "ok";

type SpotlightBody = {
  title: string;
  meta: string;
  signals: Array<{ label: string; tone: SignalTone }>;
  why: string;
  nextSteps: string[];
  skus: string[];
};

type KnownSpotlightKey =
  | "hero-command"
  | "cmd-run"
  | "cmd-review"
  | "cmd-changes"
  | "cmd-alerts"
  | "inbox-actions"
  | "mini-actions"
  | "mini-alerts"
  | "mini-heat"
  | "mini-portfolio"
  | "mini-abc"
  | "stat-alerts";

export type SpotlightKey =
  | KnownSpotlightKey
  | `sku:${string}`
  | `alert:${string}`
  | `action:${string}`
  | `queue:${string}`
  | `heat-${number}-${number}`
  | string;

export type SpotlightExtra = {
  picked?: string;
  source?: string;
};

export type SpotlightState = {
  open: boolean;
  key: SpotlightKey | null;
  extra?: SpotlightExtra;
};

export type SpotlightContent = SpotlightBody & {
  tags: string[];
};

const defaultContent: SpotlightContent = {
  title: "",
  meta: "",
  signals: [],
  why: "",
  nextSteps: [],
  skus: [],
  tags: []
};

function mk(
  partial: Partial<SpotlightBody> & {
    tags?: string[];
  }
): SpotlightContent {
  return {
    title: partial.title || "",
    meta: partial.meta || "",
    signals: partial.signals || [],
    why: partial.why || "",
    nextSteps: partial.nextSteps || [],
    skus: partial.skus || [],
    tags: partial.tags || []
  };
}

function normalizeSignalTone(code: string): SignalTone {
  if (code === "critical") return "crit";
  if (code === "warning") return "warn";
  return "info";
}

function arbitrationDecisionLabel(state?: string): string {
  const token = String(state || "").toUpperCase();
  if (token === "READY") return "Pronto";
  if (token === "CAUTION") return "Cautela";
  if (token === "REVIEW") return "Revisao";
  if (token === "BLOCKED") return "Bloqueado";
  return token || "Sem estado";
}

function arbitrationDecisionTone(state?: string): SignalTone {
  const token = String(state || "").toUpperCase();
  if (token === "READY") return "ok";
  if (token === "CAUTION" || token === "REVIEW") return "warn";
  return "crit";
}

function dominantObjectiveLabel(objective?: string): string {
  const token = String(objective || "").toUpperCase();
  if (token === "CAPITAL_RELIEF") return "Alivio de capital";
  if (token === "DEMAND_GENERATION") return "Geracao de demanda";
  if (token === "COMPETITIVE_POSITIONING") return "Posicionamento competitivo";
  if (token === "MIX_OPTIMIZATION") return "Otimizacao de mix";
  if (token === "NONE") return "Sem objetivo dominante";
  return token || "Nao definido";
}

function arbitrationReasonLabel(code: string): string {
  const token = String(code || "").toUpperCase();
  if (token === "CAPITAL_DOMINANCE_TRIGGERED") return "Dominancia de capital";
  if (token === "DOMAIN_CONFLICT_STOCK_CAMPAIGN") return "Conflito estoque x campanha";
  if (token === "DOMAIN_CONFLICT_PRICE_CAMPAIGN") return "Conflito preco x campanha";
  if (token === "RISK_OVERRIDE_OPPORTUNITY") return "Risco sobrepoe oportunidade";
  if (token === "DATA_CONFIDENCE_BLOCK") return "Bloqueio por confianca de dados";
  if (token === "MARKET_SIGNAL_WEAK") return "Sinal de mercado fraco";
  if (token === "POLICY_OBJECTIVE_OVERRIDE") return "Override de objetivo por policy";
  return token.replace(/_/g, " ").trim();
}

function asCurrencyInteger(value: number): string {
  return `R$ ${Math.round(value).toLocaleString("pt-BR")}`;
}

export function toSpotlightKey(rawKey: string): SpotlightKey {
  return rawKey as SpotlightKey;
}

export function resolveSpotlightContent(
  key: SpotlightKey | null,
  extra?: SpotlightExtra,
  model?: AnalyticsHomeViewModel
): SpotlightContent {
  if (!key) {
    return defaultContent;
  }

  if (key === "mini-actions" && model) {
    return mk({
      title: "Acoes prioritarias",
      meta: `${model.allActions.length} acoes mapeadas`,
      signals: [],
      why: "",
      nextSteps: [
        "Selecione uma acao para ver os SKUs impactados.",
        "Defina owner e prioridade da acao.",
        "Execute o ajuste no fluxo operacional."
      ],
      skus: [],
      tags: []
    });
  }

  if ((key === "mini-alerts" || key === "stat-alerts") && model) {
    return mk({
      title: "Alertas ativos",
      meta: `${model.allAlerts.length} alertas mapeados`,
      signals: [],
      why: "",
      nextSteps: [
        "Selecione um alerta para ver os SKUs impactados.",
        "Defina prioridade e responsavel por alerta.",
        "Execute a mitigacao no fluxo operacional."
      ],
      skus: [],
      tags: []
    });
  }
  if (key === "mini-portfolio" && model) {
    const total = (model.portfolio || []).reduce((sum, row) => sum + Number(row.countSkus || 0), 0);
    return mk({
      title: "Portfolio Distribution",
      meta: `${model.portfolio.length} classes • ${total} SKUs`,
      signals: [],
      why: "",
      nextSteps: [
        "Selecione uma classe para ver os SKUs do snapshot.",
        "Valide os drivers dominantes por classe.",
        "Converta a classe em plano operacional."
      ],
      skus: [],
      tags: []
    });
  }

  if (key === "mini-heat" && model) {
    const populated = Object.values(model.heatCells || {}).filter((cell) => Number(cell.count || 0) > 0);
    const totalSkus = populated.reduce((sum, cell) => sum + Number(cell.count || 0), 0);
    return mk({
      title: "Radar de saude",
      meta: `${populated.length} celulas com sinal • ${totalSkus} SKUs`,
      signals: [],
      why: "",
      nextSteps: [
        "Selecione uma celula para ver os SKUs impactados.",
        "Priorize as celulas com maior urgencia e impacto.",
        "Execute as acoes e monitore o deslocamento no proximo snapshot."
      ],
      skus: [],
      tags: []
    });
  }
  if (key === "mini-capital" && model) {
    return mk({
      title: "Eficiencia de Capital",
      meta: "Alocacao de capital por grupos e retorno",
      signals: [{ label: "Capital", tone: "warn" }],
      why: "Analise de capital imobilizado, risco e retorno para orientar realocacao e priorizacao.",
      nextSteps: [
        "Filtrar grupos com maior capital e risco.",
        "Validar retorno (GMROI) vs capital alocado.",
        "Priorizar acoes para liberar caixa com menor impacto comercial."
      ],
      skus: [],
      tags: ["Capital", "Eficiência", "Alocacao"]
    });
  }
  if (key === "mini-abc" && model) {
    return mk({
      title: "MIX ABC",
      meta: "Curva de Pareto por grupos e faixa A/B/C",
      signals: [{ label: "Pareto", tone: "info" }],
      why: "Classifica grupos por contribuicao de receita acumulada para destacar concentracao e alocacao de foco comercial.",
      nextSteps: [
        "Defender grupos A com maior contribuicao.",
        "Promover grupos B para ganho de share.",
        "Tratar grupos C com foco em capital e giro."
      ],
      skus: [],
      tags: ["ABC", "Pareto", "Mix"]
    });
  }

  const action = (model?.actions || []).find((row) => row.key === key)
    || (model?.allActions || []).find((row) => row.key === key);
  if (action) {
    const signals: Array<{ label: string; tone: SignalTone }> = [];
    const priorityLabel = action.valuePriorityTier === "P0"
      ? "Prioridade Financeira: Critica"
      : action.valuePriorityTier === "P1"
        ? "Prioridade Financeira: Alta"
        : action.valuePriorityTier === "P2"
          ? "Prioridade Financeira: Media"
          : action.valuePriorityTier === "P3"
            ? "Prioridade Financeira: Baixa"
            : "";
    if (priorityLabel) signals.push({ label: priorityLabel, tone: "crit" });
    if (action.decisionState) {
      signals.push({
        label: arbitrationDecisionLabel(action.decisionState),
        tone: arbitrationDecisionTone(action.decisionState)
      });
    }
    if (action.riskAdjustedExposureRs != null && Number.isFinite(Number(action.riskAdjustedExposureRs))) {
      const exposure = Number(action.riskAdjustedExposureRs || 0);
      signals.push({ label: `Exposicao ${asCurrencyInteger(exposure)}`, tone: "info" });
    }
    const hasStockTotal = action.stockTotalRs != null && Number.isFinite(Number(action.stockTotalRs));
    const stockTotalLabel = hasStockTotal
      ? asCurrencyInteger(Number(action.stockTotalRs || 0))
      : String(action.stockTotalLabel || "-");
    signals.push({ label: `Estoque total ${stockTotalLabel}`, tone: "info" });
    const objectiveText = action.dominantObjective
      ? `Objetivo dominante: ${dominantObjectiveLabel(action.dominantObjective)}. `
      : "";
    const reasonText = action.arbitrationReasonCodes?.length
      ? `Motivos: ${action.arbitrationReasonCodes.slice(0, 3).map(arbitrationReasonLabel).join(" | ")}.`
      : "";
    return mk({
      title: action.name,
      meta: `${action.skuCount} • ${action.signalLabel}`,
      signals,
      why: `${objectiveText}${action.desc}${reasonText ? ` ${reasonText}` : ""}`,
      nextSteps: [
        "Revise os SKUs impactados por esta acao.",
        "Valide a estrategia com comercial e estoque.",
        "Execute o ajuste em lote e acompanhe resultado."
      ],
      skus: action.skus || [],
      tags: ["Acoes", "Priorizacao", action.actionCode]
    });
  }

  const alert = (model?.alerts || []).find((row) => row.key === key)
    || (model?.allAlerts || []).find((row) => row.key === key);
  if (alert) {
    return mk({
      title: alert.name,
      meta: `${alert.count} SKUs`,
      signals: [{ label: "Alerta", tone: normalizeSignalTone(alert.toneClass) }],
      why: alert.desc,
      nextSteps: ["Abrir lista de SKUs", "Revisar regra aplicada", "Acionar owner comercial"],
      skus: [],
      tags: ["Alertas", "Risco"]
    });
  }

  const kpi = model?.kpis.find((row) => row.key === key);
  if (kpi) {
    return mk({
      title: kpi.label,
      meta: `${kpi.badge} • ${kpi.value}`,
      signals: [{ label: kpi.badge, tone: kpi.tone === "blue" ? "info" : "ok" }],
      why: kpi.note,
      nextSteps: ["Comparar com ultima semana", "Abrir detalhe por classificacoes", "Registrar decisao no board"],
      skus: [],
      tags: ["KPI", "Executivo"]
    });
  }

  const tile = model?.topMetal.find((row) => row.key === key);
  if (tile) {
    return mk({
      title: tile.k,
      meta: tile.name,
      signals: [{ label: "Top Metal", tone: "ok" }],
      why: "Ranking por maior margem no ultimo mes fechado.",
      nextSteps: ["Abrir detalhe do ranking", "Comparar com snapshot anterior", "Validar concentracao por classificacoes"],
      skus: [],
      tags: ["Top Metal", "Margem"]
    });
  }

  const evt = model?.timeline.find((row) => row.key === key);
  if (evt) {
    return mk({
      title: evt.name,
      meta: evt.time,
      signals: [{ label: "Timeline", tone: "info" }],
      why: evt.desc,
      nextSteps: ["Abrir log da run", "Correlacionar com alertas", "Gerar checkpoint"],
      skus: [],
      tags: ["Eventos", "Operacao"]
    });
  }

  const dist = model?.portfolio.find((row) => row.key === key);
  if (dist) {
    const driversText = dist.drivers.length ? dist.drivers.join(" | ") : "Sem driver dominante";
    return mk({
      title: dist.label,
      meta: `${dist.value} • ${dist.pct}% do portfolio`,
      signals: [{ label: "Portfolio", tone: "info" }],
      why: `Drivers da classe: ${driversText}.`,
      nextSteps: ["Revisar SKUs da classe no snapshot atual", "Priorizar execução por urgência operacional", "Monitorar migração de classe na próxima run"],
      skus: dist.skus || [],
      tags: ["Portfolio", "Distribuicao", ...(dist.drivers.slice(0, 1))]
    });
  }

  const stat = model?.miniStats.find((row) => row.key === key);
  if (stat) {
    return mk({
      title: stat.label,
      meta: `${stat.value} • ${stat.badge}`,
      signals: [{ label: stat.badge, tone: "warn" }],
      why: stat.sub,
      nextSteps: ["Abrir inbox da fila", "Validar urgencia por regra", "Executar sprint de acoes"],
      skus: [],
      tags: ["Command Bar", "Resumo rapido"]
    });
  }

  if (key.startsWith("sku:")) {
    const sku = key.split(":")[1] || "N/D";
    return mk({
      title: `SKU ${sku}`,
      meta: extra?.source || "SKU Lite",
      signals: [{ label: "SKU", tone: "info" }],
      why: "Visao simplificada para decisao rapida de preco/estoque.",
      nextSteps: ["Abrir historico de preco", "Comparar competitividade", "Aplicar acao recomendada"],
      skus: [sku],
      tags: ["SKU", "Drill-down"]
    });
  }

  if (key.startsWith("alert:")) {
    const code = key.split(":")[1] || "N/D";
    return mk({
      title: "Detalhe do alerta",
      meta: code.toUpperCase(),
      signals: [{ label: "Critico", tone: "crit" }],
      why: "Alerta agregado por regra com impacto relevante.",
      nextSteps: ["Ver SKUs desta regra", "Avaliar severidade", "Priorizar no board"],
      skus: [],
      tags: ["Alert Inbox", "Drill-down"]
    });
  }

  if (key.startsWith("heat-")) {
    const cell = model?.heatCells?.[key];
    return mk({
      title: cell ? `Radar de saude • ${cell.impact} x ${cell.urgency}` : "Radar de saude",
      meta: cell ? `${cell.count} SKUs na celula` : (extra?.picked || "Matriz impacto x urgencia"),
      signals: [{ label: "Heatmap", tone: "warn" }],
      why: cell?.topDrivers?.length
        ? `Drivers dominantes: ${cell.topDrivers.slice(0, 3).join(" | ")}.`
        : "A celula selecionada concentra SKUs com combinacao relevante de impacto e urgencia.",
      nextSteps: [
        "Validar os drivers dominantes da celula.",
        "Priorizar os SKUs com maior urgencia.",
        "Executar a acao recomendada e monitorar a proxima run."
      ],
      skus: cell?.skuDetails?.map((row) => row.pn) || [],
      tags: ["Radar", "Priorizacao"]
    });
  }

  return defaultContent;
}

