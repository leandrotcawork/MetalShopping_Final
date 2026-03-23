import { useEffect, useMemo, useState } from "react";

import type { AnalyticsHomeApi, AnalyticsHomeResult } from "./api";
import "./analytics_legacy.css";

type TabKey = "home" | "products" | "taxonomy" | "brands" | "actions";

const TABS: Array<{ key: TabKey; label: string }> = [
  { key: "home", label: "Home" },
  { key: "products", label: "Produtos" },
  { key: "taxonomy", label: "Taxonomia" },
  { key: "brands", label: "Marcas" },
  { key: "actions", label: "Ações" },
];

function n(value: unknown, fallback = 0): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function fmtInt(value: number): string {
  return new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 0 }).format(value);
}

function fmtPct(value: number): string {
  return `${new Intl.NumberFormat("pt-BR", { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(value)}%`;
}

function fmtDate(value: string | null | undefined): string {
  if (!value) return "N/D";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(parsed);
}

function blockData(result: AnalyticsHomeResult | null, key: keyof AnalyticsHomeResult["blocks"]): Record<string, unknown> {
  if (!result) return {};
  const block = result.blocks[key];
  if (!block || !block.data || typeof block.data !== "object") return {};
  return block.data as Record<string, unknown>;
}

export function AnalyticsHomePage(props: { api: AnalyticsHomeApi }) {
  const [activeTab, setActiveTab] = useState<TabKey>("home");
  const [result, setResult] = useState<AnalyticsHomeResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    let disposed = false;
    async function load() {
      setLoading(true);
      setError("");
      try {
        const response = await props.api.getHome();
        if (!disposed) setResult(response);
      } catch (err) {
        if (!disposed) setError(err instanceof Error ? err.message : "Falha ao carregar analytics.");
      } finally {
        if (!disposed) setLoading(false);
      }
    }
    void load();
    return () => {
      disposed = true;
    };
  }, [props.api]);

  const operational = useMemo(() => blockData(result, "kpisOperational"), [result]);
  const products = useMemo(() => blockData(result, "kpisProducts"), [result]);

  const mini = useMemo(
    () => ({
      productsRegistered: n(operational.productsRegistered, 12840),
      suppliersAvailable: n(operational.suppliersAvailable, 32),
      successRateLastRun: n(operational.successRateLastRun, 0.81),
      skusLastRun: n(operational.skusLastRun, 6410),
      productsActiveCount: n(products.productsActiveCount, 10980),
      weightedMarginPctTotal: n(products.weightedMarginPctTotal, 27.4),
    }),
    [operational, products],
  );

  return (
    <section className="ah-page">
      <section className="ah-tabs">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            type="button"
            className={activeTab === tab.key ? "active" : ""}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </section>

      {loading ? <div className="ah-banner">Carregando analytics...</div> : null}
      {error ? <div className="ah-banner warning">Falha no refresh: {error}</div> : null}

      <section className="ah-hero">
        <div className="ah-hero-main">
          <h2>Visao Inteligente do Portfolio</h2>
          <p>Janela 6M principal + 3M sensibilidade. Atualizado {fmtDate(result?.snapshot.asOf ?? null)}.</p>
          <div className="ah-shortcuts">
            <button type="button" onClick={() => setActiveTab("actions")}>
              <b>Ver Acoes</b>
              <small>Fila por urgencia</small>
            </button>
            <button type="button" onClick={() => setActiveTab("products")}>
              <b>Produtos Criticos</b>
              <small>Top SKUs prioritarios</small>
            </button>
            <button type="button" onClick={() => setActiveTab("taxonomy")}>
              <b>Radar de Taxonomia</b>
              <small>Concentracao e risco</small>
            </button>
          </div>
        </div>

        <aside className="ah-command">
          <div className="ah-command-head">
            <h3>Barra de Comando</h3>
            <small>Visual legado replicado para adaptação incremental</small>
          </div>
          <div className="ah-mini">
            <button type="button">
              <span>Produtos cadastrados</span>
              <b>{fmtInt(mini.productsRegistered)}</b>
              <small>Base ativa</small>
            </button>
            <button type="button">
              <span>Fornecedores ativos</span>
              <b>{fmtInt(mini.suppliersAvailable)}</b>
              <small>Diretório</small>
            </button>
          </div>
          <div className="ah-mini">
            <button type="button">
              <span>Taxa de sucesso</span>
              <b>{fmtPct(mini.successRateLastRun * 100 > 1 ? mini.successRateLastRun / 100 : mini.successRateLastRun)}</b>
              <small>Última run</small>
            </button>
            <button type="button">
              <span>SKUs na run</span>
              <b>{fmtInt(mini.skusLastRun)}</b>
              <small>Execução mais recente</small>
            </button>
          </div>
        </aside>
      </section>

      {activeTab === "home" ? (
        <section className="ah-simple">
          <header>
            <h3>Home</h3>
            <p>Painel visual legado com cards de operação e produto.</p>
          </header>
          <div className="ah-chip-grid">
            <div className="ah-chip-card">
              <b>Produtos Ativos</b>
              <small>{fmtInt(mini.productsActiveCount)} SKUs</small>
            </div>
            <div className="ah-chip-card">
              <b>Margem Ponderada</b>
              <small>{fmtPct(mini.weightedMarginPctTotal)}</small>
            </div>
            <div className="ah-chip-card">
              <b>Snapshot servido</b>
              <small>{fmtDate(result?.snapshot.servedAt ?? null)}</small>
            </div>
            <div className="ah-chip-card">
              <b>Status da tranche</b>
              <small>Visual legado ativo</small>
            </div>
          </div>
        </section>
      ) : null}

      {activeTab === "products" ? (
        <section className="ah-simple">
          <header>
            <h3>Produtos</h3>
            <p>Visão inicial copiada para manter paridade visual com o legacy.</p>
          </header>
          <div className="ah-chip-grid">
            {["PN-18578", "PN-90210", "PN-43122", "PN-88401", "PN-77090", "PN-65012", "PN-32011", "PN-55003"].map((sku) => (
              <button key={sku} type="button" className="ah-chip-card">
                <b>{sku}</b>
                <small>Abrir SKU Drawer Lite</small>
              </button>
            ))}
          </div>
        </section>
      ) : null}

      {activeTab === "taxonomy" ? (
        <section className="ah-simple">
          <header>
            <h3>Taxonomia</h3>
            <p>Cards por hierarquia para espelhar o layout legado.</p>
          </header>
          <div className="ah-chip-grid">
            {["Revestimentos", "Tintas", "Ferragens", "Elétrica", "Hidráulica", "Louças", "Coberturas", "Pisos"].map((name) => (
              <div key={name} className="ah-chip-card">
                <b>{name}</b>
                <small>Impacto + urgência</small>
              </div>
            ))}
          </div>
        </section>
      ) : null}

      {activeTab === "brands" ? (
        <section className="ah-simple">
          <header>
            <h3>Marcas</h3>
            <p>Agrupamento visual por impacto estimado.</p>
          </header>
          <div className="ah-chip-grid">
            {["Deca", "Portobello", "Suvinil", "Vedacit", "Quartzolit", "Tramontina", "Tigre", "Lorenzetti"].map((name, index) => (
              <div key={name} className="ah-chip-card">
                <b>{name}</b>
                <small>{fmtInt(1200 - index * 90)} SKUs impactados</small>
              </div>
            ))}
          </div>
        </section>
      ) : null}

      {activeTab === "actions" ? (
        <section className="ah-kanban">
          {["Triage", "Doing", "Blocked", "Done"].map((lane) => (
            <article key={lane} className="ah-lane">
              <header>
                <h3>{lane}</h3>
              </header>
              <div>
                {["Reprecificar top 20", "Ajustar estoque segurança", "Revisar alertas preço"].map((title) => (
                  <div key={`${lane}-${title}`} className="ah-kanban-card">
                    <b>{title}</b>
                    <small>Contexto visual legado</small>
                    <p>48 SKUs | PRICING</p>
                    <div>
                      <button type="button">Detalhar</button>
                      <select defaultValue={lane}>
                        <option value="Triage">Triage</option>
                        <option value="Doing">Doing</option>
                        <option value="Blocked">Blocked</option>
                        <option value="Done">Done</option>
                      </select>
                    </div>
                  </div>
                ))}
              </div>
            </article>
          ))}
        </section>
      ) : null}
    </section>
  );
}
