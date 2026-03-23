// @ts-nocheck
import { makeAnalyticsHomeV2Dto, type AnalyticsHomeV2Dto } from "../../legacy_dto";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useAppSession } from "../../app/providers/AppProviders";
import { AnalyticsHomePage } from "../analytics_home/AnalyticsHomePage";
import { AnalyticsProductsPage } from "./AnalyticsProductsPage";
import { TaxonomyHomePage } from "./TaxonomyHomePage";
import "./analytics.css";

type TabKey = "home" | "products" | "taxonomy" | "brands" | "actions";
type Domain = "all" | "critical" | "pricing" | "stock" | "capital" | "data";
type Urgency = "HIGH" | "MED" | "LOW";
type Lane = "TRIAGE" | "DOING" | "BLOCKED" | "DONE";
type Drawer = { type: "action"; id: string } | { type: "alert"; id: string } | { type: "sku"; sku: string } | null;

type ActionItem = {
  id: string;
  title: string;
  reason: string;
  domain: Domain;
  urgency: Urgency;
  score: number;
  skuCount: number;
  topSkus: string[];
  topTaxonomyLeaf: string;
  topBrand: string;
};

type AlertItem = {
  id: string;
  title: string;
  reason: string;
  domain: Domain;
  urgency: Urgency;
  total: number;
  topSkus: string[];
};

const LANES: Array<{ key: Lane; label: string }> = [
  { key: "TRIAGE", label: "Triage" },
  { key: "DOING", label: "Doing" },
  { key: "BLOCKED", label: "Blocked" },
  { key: "DONE", label: "Done" }
];
const NAV_TABS: Array<{ key: TabKey; label: string }> = [
  { key: "home", label: "Analytics | Home" },
  { key: "brands", label: "Analytics | Marca" },
  { key: "actions", label: "Analytics | Execucao" },
  { key: "taxonomy", label: "Analytics | Classificacoes" },
  { key: "products", label: "Analytics | Produtos" },
];

function rec(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}
function txt(value: unknown): string {
  return String(value ?? "").trim();
}
function num(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  const parsed = Number(txt(value));
  return Number.isFinite(parsed) ? parsed : null;
}
function pickT(row: Record<string, unknown>, keys: string[]): string {
  for (const key of keys) {
    const value = txt(row[key]);
    if (value) return value;
  }
  return "";
}
function pickN(row: Record<string, unknown>, keys: string[]): number | null {
  for (const key of keys) {
    const value = num(row[key]);
    if (value !== null) return value;
  }
  return null;
}
function toList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map((it) => (typeof it === "string" || typeof it === "number" ? txt(it) : pickT(rec(it), ["pn", "name", "label", "code"]))).filter(Boolean);
}
function fmtInt(value: number | null): string {
  if (value === null) return "N/D";
  return new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 0 }).format(value);
}
function fmtPct(value: number | null): string {
  if (value === null) return "N/D";
  return `${new Intl.NumberFormat("pt-BR", { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(value)}%`;
}
function fmtMoney(value: number | null): string {
  if (value === null) return "N/D";
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL", maximumFractionDigits: 0 }).format(value);
}
function fmtDate(value: string | null | undefined): string {
  const token = txt(value);
  if (!token) return "N/D";
  const parsed = new Date(token);
  if (Number.isNaN(parsed.getTime())) return token;
  return new Intl.DateTimeFormat("pt-BR", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" }).format(parsed);
}
function boolish(value: unknown, fallback = false): boolean {
  if (typeof value === "boolean") return value;
  const token = txt(value).toLowerCase();
  if (!token) return fallback;
  return ["1", "true", "t", "yes", "y", "on"].includes(token);
}
function domainOf(raw: string): Domain {
  const token = raw.toLowerCase();
  if (/preco|price|margem|compet/.test(token)) return "pricing";
  if (/estoque|stock|ruptura|reposicao|cobertura/.test(token)) return "stock";
  if (/capital|receita|lucro|caixa|giro/.test(token)) return "capital";
  if (/dado|cadastro|ean|descricao|qualidade/.test(token)) return "data";
  return "critical";
}
function urgency(score: number): Urgency {
  if (score >= 80) return "HIGH";
  if (score >= 55) return "MED";
  return "LOW";
}
function defaultLane(level: Urgency): Lane {
  if (level === "HIGH") return "TRIAGE";
  if (level === "MED") return "DOING";
  return "BLOCKED";
}
function tabFromParam(token: string | undefined): TabKey {
  const tab = txt(token).toLowerCase();
  return ["home", "products", "taxonomy", "brands", "actions"].includes(tab)
    ? (tab as TabKey)
    : "home";
}

function parseActions(dto: AnalyticsHomeV2Dto | null): ActionItem[] {
  if (!dto || dto.blocks.actions_today.status !== "OK" || !dto.blocks.actions_today.data) return [];
  return dto.blocks.actions_today.data.items.map((item, idx) => {
    const row = rec(item);
    const presentation = rec(row.presentation);
    const id = pickT(row, ["action_code", "id", "code", "pn"]) || `act-${idx + 1}`;
    const title = pickT(row, ["title", "headline", "name"]) || pickT(presentation, ["title", "label", "headline"]) || id;
    const score = pickN(row, ["urgency_score", "score", "priority_score"]) ?? 0;
    const reason = pickT(row, ["reason_short", "reason", "why", "headline"]) || pickT(presentation, ["headline", "reason"]) || "Sem contexto adicional.";
    const topSkusRaw = toList(row.top_pns ?? row.top_skus ?? row.pns);
    const topSkus = topSkusRaw.length ? topSkusRaw : (txt(row.pn) ? [txt(row.pn)] : []);
    const topTaxonomyLeaf = toList(row.top_taxonomy_leafs ?? row.top_taxonomy_leaf_names)[0] || "-";
    const topBrand = toList(row.top_brands ?? row.brands)[0] || "-";
    const skuCount = Math.max(0, Math.trunc(pickN(row, ["sku_count", "count_skus", "count"]) ?? Math.max(1, topSkus.length)));
    return { id, title, reason, domain: domainOf(`${id} ${title} ${reason}`), urgency: urgency(score), score, skuCount, topSkus, topTaxonomyLeaf, topBrand };
  });
}

function parseAlerts(dto: AnalyticsHomeV2Dto | null): AlertItem[] {
  if (!dto || dto.blocks.alerts_prioritarios.status !== "OK" || !dto.blocks.alerts_prioritarios.data) return [];
  return dto.blocks.alerts_prioritarios.data.buckets.map((item, idx) => {
    const row = rec(item);
    const id = pickT(row, ["alert_code", "id", "code"]) || `al-${idx + 1}`;
    const title = pickT(row, ["title", "headline", "name"]) || id;
    const reason = pickT(row, ["reason_short", "reason", "why"]) || "Sem contexto adicional.";
    const topSkus = toList(row.top_pns ?? row.top_skus ?? row.pns);
    const total = Math.max(0, Math.trunc(pickN(row, ["count", "total", "sku_count"]) ?? topSkus.length));
    const score = pickN(row, ["urgency_score", "score", "priority_score"]) ?? 0;
    return { id, title, reason, domain: domainOf(`${id} ${title} ${reason}`), urgency: urgency(score), total, topSkus };
  });
}

export function AnalyticsPage() {
  const { tab } = useParams();
  const navigate = useNavigate();
  const activeTab = tabFromParam(tab);
  const { api, analyticsHomeSnapshot, setAnalyticsHomeSnapshot } = useAppSession();
  const snapshotRef = useRef(analyticsHomeSnapshot);
  const searchRef = useRef<HTMLInputElement | null>(null);

  const [loadState, setLoadState] = useState<"loading" | "ready" | "error">(analyticsHomeSnapshot?.data ? "ready" : "loading");
  const [pageError, setPageError] = useState("");
  const [query, setQuery] = useState("");
  const [chip, setChip] = useState<Domain>("all");
  const [drawer, setDrawer] = useState<Drawer>(null);
  const [laneByAction, setLaneByAction] = useState<Record<string, Lane>>({});
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    try {
      const stored = window.localStorage.getItem("ms:theme");
      if (stored === "dark" || stored === "light") return stored;
      const attr = document.documentElement.getAttribute("data-theme");
      return attr === "dark" ? "dark" : "light";
    } catch {
      return "light";
    }
  });
  const [isOnline, setIsOnline] = useState<boolean>(() => {
    try {
      return Boolean(navigator.onLine);
    } catch {
      return true;
    }
  });

  useEffect(() => {
    // Defensive: tab switches must never leave any modal/backdrop mounted.
    setDrawer(null);

    if (typeof window === "undefined") return undefined;
    const frameId = window.requestAnimationFrame(() => {
      window.scrollTo({ top: 0, left: 0, behavior: "auto" });
    });
    return () => window.cancelAnimationFrame(frameId);
  }, [activeTab]);

  useEffect(() => {
    snapshotRef.current = analyticsHomeSnapshot;
    if (analyticsHomeSnapshot?.data) setLoadState("ready");
  }, [analyticsHomeSnapshot]);

  useEffect(() => {
    try {
      document.documentElement.setAttribute("data-theme", theme);
      window.localStorage.setItem("ms:theme", theme);
    } catch {
      // ignore
    }
  }, [theme]);

  useEffect(() => {
    function onOnline() {
      setIsOnline(true);
    }
    function onOffline() {
      setIsOnline(false);
    }
    window.addEventListener("online", onOnline);
    window.addEventListener("offline", onOffline);
    return () => {
      window.removeEventListener("online", onOnline);
      window.removeEventListener("offline", onOffline);
    };
  }, []);

  const loadAnalytics = useCallback(async () => {
    const hasSnapshot = Boolean(snapshotRef.current?.data);
    if (!hasSnapshot) setLoadState("loading");
    setPageError("");
    try {
      const envelope = await api.home.workspace(undefined, { includeOperational: true });
      const dto = makeAnalyticsHomeV2Dto(envelope as { data: Record<string, unknown>; meta?: Record<string, unknown> }, "current");
      const current = snapshotRef.current;
      const nextKey = `${dto.snapshot.resolved_id || ""}|${dto.snapshot.as_of || ""}`;
      const currKey = current ? `${current.data.snapshot.resolved_id || ""}|${current.data.snapshot.as_of || ""}` : "";
      if (!current || nextKey !== currKey) setAnalyticsHomeSnapshot({ data: dto, asOf: String(dto.snapshot.as_of || ""), updatedAt: Date.now() });
      setLoadState("ready");
    } catch (err) {
      if (!hasSnapshot) setLoadState("error");
      setPageError(err instanceof Error ? err.message : String(err));
    }
  }, [api, setAnalyticsHomeSnapshot]);

  useEffect(() => {
    void loadAnalytics();
  }, [loadAnalytics]);

  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") setDrawer(null);
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        searchRef.current?.focus();
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const dto = analyticsHomeSnapshot?.data ?? null;
  const actions = useMemo(() => parseActions(dto), [dto]);
  const alerts = useMemo(() => parseAlerts(dto), [dto]);
  useEffect(() => {
    if (!actions.length) return;
    setLaneByAction((prev) => {
      const next = { ...prev };
      for (const item of actions) if (!next[item.id]) next[item.id] = defaultLane(item.urgency);
      return next;
    });
  }, [actions]);

  const filteredActions = useMemo(() => {
    const token = query.toLowerCase().trim();
    return actions.filter((it) => {
      if (chip !== "all") {
        if (chip === "critical" && it.urgency !== "HIGH") return false;
        if (chip !== "critical" && it.domain !== chip) return false;
      }
      if (!token) return true;
      return `${it.id} ${it.title} ${it.reason}`.toLowerCase().includes(token);
    });
  }, [actions, chip, query]);
  const filteredAlerts = useMemo(() => {
    const token = query.toLowerCase().trim();
    return alerts.filter((it) => {
      if (chip !== "all") {
        if (chip === "critical" && it.urgency !== "HIGH") return false;
        if (chip !== "critical" && it.domain !== chip) return false;
      }
      if (!token) return true;
      return `${it.id} ${it.title} ${it.reason}`.toLowerCase().includes(token);
    });
  }, [alerts, chip, query]);

  const selectedAction = drawer?.type === "action" ? actions.find((it) => it.id === drawer.id) || null : null;
  const selectedAlert = drawer?.type === "alert" ? alerts.find((it) => it.id === drawer.id) || null : null;

  return (
    <section className="ah-page">
      <header className="ah-topline">
        <div className="ah-topline-inner">
          <div className="ah-topline-brand">
            <h1>Metal <span>Analytics</span></h1>
            <small>Visao Estrategica</small>
          </div>
          <nav className="ah-topline-tabs" aria-label="Analytics navigation">
            {NAV_TABS.map((item) => (
              <button
                key={item.key}
                type="button"
                className={activeTab === item.key ? "active" : ""}
                onClick={() => {
                  setDrawer(null);
                  navigate(`/analytics/${item.key}`);
                }}
              >
                {item.label}
              </button>
            ))}
          </nav>
          <div className="ah-topline-right" aria-label="Connection and theme">
            <div className={`ah-status ${isOnline ? "on" : "off"}`} title={isOnline ? "Online" : "Offline"}>
              <span className="ah-dot" aria-hidden="true" />
              <span>{isOnline ? "Online" : "Offline"}</span>
            </div>
            <button
              type="button"
              className="ah-themeBtn"
              aria-label={theme === "dark" ? "Ativar tema claro" : "Ativar tema escuro"}
              title={theme === "dark" ? "Tema claro" : "Tema escuro"}
              onClick={() => setTheme((cur) => (cur === "dark" ? "light" : "dark"))}
            >
              {theme === "dark" ? "☀" : "☾"}
            </button>
          </div>
        </div>
      </header>

      {loadState === "loading" && !dto ? <div className="ah-banner">Carregando analytics...</div> : null}
      {loadState === "error" && !dto ? <div className="ah-banner error">Falha: {pageError}</div> : null}
      {pageError && dto ? <div className="ah-banner warning">Falha no refresh: {pageError}</div> : null}

      {activeTab === "home" ? (
        <AnalyticsHomePage
          dto={dto}
          updatedAtLabel={fmtDate(dto?.snapshot?.as_of)}
          onRefresh={loadAnalytics}
          isRefreshing={loadState === "loading"}
        />
      ) : null}

      {activeTab === "products" ? <AnalyticsProductsPage /> : null}
      {activeTab === "taxonomy" ? <TaxonomyHomePage /> : null}

      {dto && activeTab !== "home" && activeTab !== "products" && activeTab !== "taxonomy" ? (
        <>
          <section className="ah-hero">
            <div className="ah-hero-main">
              <h2>Visao Inteligente do Portfolio</h2>
              <p>Janela 6M principal + 3M sensibilidade. Atualizado {fmtDate(dto.snapshot.as_of)}.</p>
              <div className="ah-shortcuts">
                <button type="button" onClick={() => navigate("/analytics/actions")}><b>Ver Acoes</b><small>Fila por urgencia</small></button>
                <button type="button" onClick={() => setChip("critical")}><b>Criticos</b><small>Somente prioridade alta</small></button>
                <button type="button" onClick={() => setDrawer({ type: "alert", id: filteredAlerts[0]?.id || alerts[0]?.id || "" })}><b>Alertas</b><small>Abrir spotlight</small></button>
              </div>
            </div>
            <aside className="ah-command">
              <div className="ah-command-head">
                <h3>Barra de Comando</h3>
                <small>Buscar SKU, hierarquia, marca ou regra</small>
              </div>
              <div className="ah-chips">
                <button type="button" className={chip === "all" ? "active" : ""} onClick={() => setChip("all")}>Tudo</button>
                <button type="button" className={chip === "critical" ? "active" : ""} onClick={() => setChip("critical")}>Critico</button>
                <button type="button" className={chip === "pricing" ? "active" : ""} onClick={() => setChip("pricing")}>Preco</button>
                <button type="button" className={chip === "stock" ? "active" : ""} onClick={() => setChip("stock")}>Estoque</button>
                <button type="button" className={chip === "capital" ? "active" : ""} onClick={() => setChip("capital")}>Capital</button>
                <button type="button" className={chip === "data" ? "active" : ""} onClick={() => setChip("data")}>Dados</button>
              </div>
              <div className="ah-search">
                <span>⌕</span>
                <input ref={searchRef} value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Ex.: 18578 | porcelanato | preco..." />
                <small>Ctrl K</small>
              </div>
              <div className="ah-mini">
                <button type="button" onClick={() => setDrawer({ type: "action", id: filteredActions[0]?.id || actions[0]?.id || "" })}>
                  <span>Acoes prioritarias</span><b>{fmtInt(actions.length)}</b><small>Hoje</small>
                </button>
                <button type="button" onClick={() => setDrawer({ type: "alert", id: filteredAlerts[0]?.id || alerts[0]?.id || "" })}>
                  <span>Alertas ativos</span><b>{fmtInt(alerts.length)}</b><small>Live</small>
                </button>
              </div>
            </aside>
          </section>

          {activeTab === "actions" && (
            <section className="ah-kanban">
              {LANES.map((lane) => (
                <article key={lane.key} className="ah-lane">
                  <header><h3>{lane.label}</h3></header>
                  <div>
                    {filteredActions.filter((item) => (laneByAction[item.id] || defaultLane(item.urgency)) === lane.key).map((item) => (
                      <div key={item.id} className="ah-kanban-card">
                        <b>{item.title}</b>
                        <small>{item.reason}</small>
                        <p>{fmtInt(item.skuCount)} SKUs | {item.domain.toUpperCase()}</p>
                        <div>
                          <button type="button" onClick={() => setDrawer({ type: "action", id: item.id })}>Detalhar</button>
                          <select value={laneByAction[item.id] || defaultLane(item.urgency)} onChange={(event) => setLaneByAction((prev) => ({ ...prev, [item.id]: event.target.value as Lane }))}>
                            {LANES.map((opt) => <option key={opt.key} value={opt.key}>{opt.label}</option>)}
                          </select>
                        </div>
                      </div>
                    ))}
                  </div>
                </article>
              ))}
            </section>
          )}

          {activeTab === "brands" && (
            <section className="ah-simple">
              <header><h3>Marcas</h3><p>Agrupamento por impacto de acoes.</p></header>
              <div className="ah-chip-grid">
                {Object.entries(actions.reduce<Record<string, number>>((acc, item) => {
                  const key = item.topBrand || "Sem marca";
                  acc[key] = (acc[key] || 0) + item.skuCount;
                  return acc;
                }, {})).sort((a, b) => b[1] - a[1]).map(([brand, count]) => (
                  <div key={brand} className="ah-chip-card"><b>{brand}</b><small>{fmtInt(count)} SKUs impactados</small></div>
                ))}
              </div>
            </section>
          )}
        </>
      ) : null}

      {drawer ? (
        <>
          <div className="ah-backdrop open" onClick={() => setDrawer(null)} />
          <aside className="ah-drawer open">
            <header>
              <h3>
                {selectedAction?.title ||
                  selectedAlert?.title ||
                  (drawer?.type === "sku" ? `SKU ${drawer.sku}` : "Spotlight")}
              </h3>
              <button type="button" onClick={() => setDrawer(null)}>
                x
              </button>
            </header>
            <div>
              {selectedAction ? (
                <>
                  <h4>Why</h4>
                  <p>{selectedAction.reason}</p>
                  <h4>Next steps</h4>
                  <div className="ah-drawer-actions">
                    <button type="button" onClick={() => navigate("/analytics/actions")}>
                      Abrir no Kanban
                    </button>
                    <button type="button">Exportar</button>
                  </div>
                  <h4>Ver SKUs</h4>
                  <div className="ah-drawer-skus">
                    {selectedAction.topSkus.map((sku) => (
                      <button key={sku} type="button" onClick={() => setDrawer({ type: "sku", sku })}>
                        {sku}
                      </button>
                    ))}
                  </div>
                </>
              ) : null}
              {selectedAlert ? (
                <>
                  <h4>Why</h4>
                  <p>{selectedAlert.reason}</p>
                  <h4>SKUs relacionados</h4>
                  <div className="ah-drawer-skus">
                    {selectedAlert.topSkus.map((sku) => (
                      <button key={sku} type="button" onClick={() => setDrawer({ type: "sku", sku })}>
                        {sku}
                      </button>
                    ))}
                  </div>
                </>
              ) : null}
              {drawer?.type === "sku" ? <p>SKU Drawer Lite para {drawer.sku}. Workspace de produto ainda em construcao.</p> : null}
            </div>
          </aside>
        </>
      ) : null}
    </section>
  );
}



