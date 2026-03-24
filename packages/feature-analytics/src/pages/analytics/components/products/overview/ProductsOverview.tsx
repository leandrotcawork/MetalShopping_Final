import { useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import { AnalyticsSpotlightDrawer } from "../../../../analytics_home/components/AnalyticsSpotlightDrawer";
import { SpotlightSelectionWidget } from "../../../../analytics_home/components/SpotlightSelectionWidget";
import type { ProductsOverviewViewModel } from "./products_overview.viewmodel";
import { KpiCard } from "./KpiCard";
import { SparklineMini } from "./SparklineMini";
import styles from "./products_overview.module.css";

type ProductsOverviewProps = {
  model: ProductsOverviewViewModel;
};

type MatrixQuadrantKey = "stars" | "potential" | "attention" | "critical";

export function ProductsOverview({ model }: ProductsOverviewProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [activeQuadrant, setActiveQuadrant] = useState<MatrixQuadrantKey | null>(null);

  function openWorkspace(pn: string): void {
    const token = encodeURIComponent(String(pn || "").trim());
    if (!token || token === "-") return;
    navigate(`/analytics/products/${token}/overview`, {
      state: {
        from: `${location.pathname}${location.search}${location.hash}`,
      },
    });
  }

  const currencyFormatter = useMemo(
    () => new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL", maximumFractionDigits: 0 }),
    []
  );

  const matrixMeta = useMemo(
    () => {
      type SpotlightRow = (typeof model.matrixSpotlight.stars)[number];
      const displayMap = model.matrixSpotlight.display;

      const metricLabelMap: Record<string, string> = {
        giro_6m: "Giro",
        margin_sales_pct: "Margem nas vendas",
        margin_unit_pct: "Margem individual",
        sales_6m_units: "Vendas",
        stock_qty: "Estoque",
        days_no_sales: "Dias sem venda",
        contribution_brl: "Contribuicao",
        gap_vs_market_pct: "Gap vs mercado",
        capital_brl: "Capital imobilizado",
        dos: "DOS",
      };

      function metricValue(row: SpotlightRow, key: string): { text: string; numeric: number | null } {
        switch (key) {
          case "giro_6m":
            return {
              text: `${row.giro6m.toLocaleString("pt-BR", { minimumFractionDigits: 1, maximumFractionDigits: 1 })}x`,
              numeric: row.giro6m,
            };
          case "margin_sales_pct":
            return { text: `${row.marginSalesPct.toFixed(1).replace(".", ",")}%`, numeric: row.marginSalesPct };
          case "margin_unit_pct":
            return { text: `${row.marginUnitPct.toFixed(1).replace(".", ",")}%`, numeric: row.marginUnitPct };
          case "sales_6m_units":
            return { text: `${Math.round(row.sales6mUnits).toLocaleString("pt-BR")}un`, numeric: row.sales6mUnits };
          case "stock_qty":
            return {
              text: `${row.stockQty.toLocaleString("pt-BR", {
                minimumFractionDigits: Number.isInteger(row.stockQty) ? 0 : 1,
                maximumFractionDigits: Number.isInteger(row.stockQty) ? 0 : 2,
              })}un`,
              numeric: row.stockQty,
            };
          case "days_no_sales":
            return { text: `${Math.round(row.daysNoSales)}d`, numeric: row.daysNoSales };
          case "contribution_brl":
            return { text: currencyFormatter.format(row.contributionBrl), numeric: row.contributionBrl };
          case "gap_vs_market_pct":
            return { text: `${row.gapVsMarketPct >= 0 ? "+" : ""}${row.gapVsMarketPct.toFixed(1).replace(".", ",")}%`, numeric: row.gapVsMarketPct };
          case "capital_brl":
            return { text: currencyFormatter.format(row.capitalBrl), numeric: row.capitalBrl };
          case "dos":
            return { text: `${row.dos.toFixed(0)}d`, numeric: row.dos };
          default:
            return { text: "-", numeric: null };
        }
      }

      function buildSpotlightRows(quadrant: MatrixQuadrantKey, rows: SpotlightRow[]) {
        function normalizePriorityTier(raw: string): "Alta" | "Media" | "Baixa" | null {
          const token = String(raw || "").trim().toLowerCase();
          if (!token) return null;
          if (token === "alta" || token === "high" || token === "p1" || token === "critica" || token === "critico") return "Alta";
          if (token === "media" || token === "medium" || token === "p2") return "Media";
          if (token === "baixa" || token === "low" || token === "p3") return "Baixa";
          return null;
        }

        const scores = rows
          .map((row) => Number(row.financialPriorityScore || 0))
          .filter((value) => Number.isFinite(value) && value > 0)
          .sort((a, b) => b - a);

        const highCut = scores.length ? scores[Math.max(0, Math.floor((scores.length - 1) * 0.2))] : 0;
        const mediumCut = scores.length ? scores[Math.max(0, Math.floor((scores.length - 1) * 0.5))] : 0;

        function fallbackPriorityTier(score: number): "Alta" | "Media" | "Baixa" {
          if (!Number.isFinite(score) || score <= 0) return "Baixa";
          if (score >= highCut) return "Alta";
          if (score >= mediumCut) return "Media";
          return "Baixa";
        }

        const display = displayMap[quadrant] || [];
        const thirdKey = display[0] || (quadrant === "critical" ? "capital_brl" : "contribution_brl");
        const detailKeys = display.slice(1);
        return {
          thirdKey,
          rows: rows.map((row) => {
            const third = metricValue(row, thirdKey);
            const priorityScore = Number(row.financialPriorityScore || 0);
            const priorityTier = normalizePriorityTier(row.financialPriorityTier) || fallbackPriorityTier(priorityScore);
            const details = detailKeys.map((key) => {
              const val = metricValue(row, key);
              return {
                label: metricLabelMap[key] || key,
                value: val.text,
              };
            });
            details.unshift({
              label: metricLabelMap[thirdKey] || "Metrica",
              value: third.text,
            });
            return {
              pn: row.pn,
              description: row.product,
              brand: row.brand,
              taxonomyLeafName: row.taxonomyLeafName,
              stockValue: currencyFormatter.format(row.stockValueBrl),
              stockValueNumeric: row.stockValueBrl,
              stockQty: row.stockQty.toLocaleString("pt-BR", {
                minimumFractionDigits: Number.isInteger(row.stockQty) ? 0 : 1,
                maximumFractionDigits: Number.isInteger(row.stockQty) ? 0 : 2,
              }),
              stockQtyNumeric: row.stockQty,
              financialPriority: priorityTier,
              financialPriorityScore: priorityScore,
              third: third.text,
              thirdNumeric: third.numeric,
              details,
            };
          }),
        };
      }

      const starsRows = buildSpotlightRows("stars", model.matrixSpotlight.stars);
      const potentialRows = buildSpotlightRows("potential", model.matrixSpotlight.potential);
      const attentionRows = buildSpotlightRows("attention", model.matrixSpotlight.attention);
      const criticalRows = buildSpotlightRows("critical", model.matrixSpotlight.critical);

      return ({
      stars: {
        label: "Estrelas",
        subtitle: "Alta margem + alto giro",
        tableTitle: "Produtos Estrela",
        thirdHeader: metricLabelMap[starsRows.thirdKey] || "Metrica",
        emptyText: "Nenhum produto estrela encontrado.",
        defaultSort: { key: "third" as const, dir: "desc" as const },
        rows: starsRows.rows,
      },
      potential: {
        label: "Potencial",
        subtitle: "Alta margem + baixo giro",
        tableTitle: "Produtos em Potencial",
        thirdHeader: metricLabelMap[potentialRows.thirdKey] || "Metrica",
        emptyText: "Nenhum produto em potencial encontrado.",
        defaultSort: { key: "third" as const, dir: "desc" as const },
        rows: potentialRows.rows,
      },
      attention: {
        label: "Atencao",
        subtitle: "Baixa margem + alto giro",
        tableTitle: "Produtos em Atencao",
        thirdHeader: metricLabelMap[attentionRows.thirdKey] || "Metrica",
        emptyText: "Nenhum produto em atencao encontrado.",
        defaultSort: { key: "third" as const, dir: "desc" as const },
        rows: attentionRows.rows,
      },
      critical: {
        label: "Encalhe",
        subtitle: "Baixa margem + baixo giro",
        tableTitle: "Produtos em Encalhe",
        thirdHeader: metricLabelMap[criticalRows.thirdKey] || "Metrica",
        emptyText: "Nenhum produto encalhado encontrado.",
        defaultSort: { key: "third" as const, dir: "desc" as const },
        rows: criticalRows.rows,
      },
    })},
    [currencyFormatter, model.matrixSpotlight]
  );

  const activeSpotlight = activeQuadrant ? matrixMeta[activeQuadrant] : null;

  return (
    <div className={styles.overviewRoot}>
      <section className={styles.heroKpis}>
        {model.kpis.map((kpi) => (
          <KpiCard
            key={kpi.label}
            label={kpi.label}
            value={kpi.value}
            change={kpi.change}
            tone={kpi.tone}
            helpItems={kpi.helpItems}
            showChangeArrow={kpi.showChangeArrow}
          />
        ))}
      </section>

      <section className={styles.mainGrid}>
        <article className={`${styles.glassCard} ${styles.matrixCard}`}>
          <h3 className={styles.cardTitle}>Matriz Margem x Giro</h3>
          <div className={styles.matrixGrid}>
            <button
              type="button"
              className={`${styles.matrixQuadrant} ${styles.matrixQuadrant_stars}`}
              onClick={() => setActiveQuadrant("stars")}
            >
              <span className={styles.matrixIcon} aria-hidden>{"\u{1F31F}"}</span>
              <span className={styles.matrixLabel}>Estrelas</span>
              <strong className={styles.matrixCount}>{model.matrix.stars}</strong>
              <span className={styles.matrixSubtitle}>Alta margem + alto giro</span>
            </button>
            <button
              type="button"
              className={`${styles.matrixQuadrant} ${styles.matrixQuadrant_potential}`}
              onClick={() => setActiveQuadrant("potential")}
            >
              <span className={styles.matrixIcon} aria-hidden>{"\u{1F3AF}"}</span>
              <span className={styles.matrixLabel}>Potencial</span>
              <strong className={styles.matrixCount}>{model.matrix.potential}</strong>
              <span className={styles.matrixSubtitle}>Alta margem + baixo giro</span>
            </button>
            <button
              type="button"
              className={`${styles.matrixQuadrant} ${styles.matrixQuadrant_attention}`}
              onClick={() => setActiveQuadrant("attention")}
            >
              <span className={styles.matrixIcon} aria-hidden>{"\u26A0\uFE0F"}</span>
              <span className={styles.matrixLabel}>Atencao</span>
              <strong className={styles.matrixCount}>{model.matrix.attention}</strong>
              <span className={styles.matrixSubtitle}>Baixa margem + alto giro</span>
            </button>
            <button
              type="button"
              className={`${styles.matrixQuadrant} ${styles.matrixQuadrant_critical}`}
              onClick={() => setActiveQuadrant("critical")}
            >
              <span className={styles.matrixIcon} aria-hidden>{"\u{1F534}"}</span>
              <span className={styles.matrixLabel}>Encalhe</span>
              <strong className={styles.matrixCount}>{model.matrix.critical}</strong>
              <span className={styles.matrixSubtitle}>Baixa margem + baixo giro</span>
            </button>
          </div>
        </article>

        <article className={`${styles.glassCard} ${styles.abcSideCard}`}>
          <h3 className={styles.cardTitle}>Distribuicao ABC (Curva de Pareto)</h3>
          <div className={styles.abcBars}>
            {model.abc.map((item) => (
              <div key={item.letter} className={styles.abcItem}>
                <div className={`${styles.abcLabel} ${styles[`abcLabel_${item.letter.toLowerCase()}`]}`}>{item.letter}</div>
                <div className={styles.abcBarContainer}>
                  <div className={styles.abcBarHeader}>
                    <span>{item.countLabel}</span>
                    <span>{item.shareLabel}</span>
                  </div>
                  <div className={styles.abcBar}>
                    <div
                      className={`${styles.abcBarFill} ${styles[`abcBarFill_${item.letter.toLowerCase()}`]}`}
                      style={{ width: `${item.fillPct}%` }}
                    >
                      {item.valueLabel}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </article>
      </section>

      <section className={styles.rankingsGrid}>
        <article className={styles.rankingCard}>
          <h3 className={styles.cardTitle}>Top 10 Margem (R$ que mais gera)</h3>
          <table className={styles.rankingTable}>
            <thead>
              <tr>
                <th>PN</th>
                <th>Produto</th>
                <th>Margem</th>
                <th>R$/mes</th>
              </tr>
            </thead>
            <tbody>
              {model.rankings.topMargin.map((row) => (
                <tr key={`m-${row.pn}`}>
                  <td>{row.pn}</td>
                  <td
                    className={`${styles.productName} ${styles.productNameInteractive}`}
                    title={row.product}
                    onDoubleClick={() => openWorkspace(row.pn)}
                  >
                    {row.product}
                  </td>
                  <td className={`${styles.metricValue} ${styles.metricValue_positive}`}>{row.metric}</td>
                  <td className={styles.metricValue}>{row.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </article>

        <article className={styles.rankingCard}>
          <h3 className={styles.cardTitle}>Top 10 Giro (volume que vende)</h3>
          <table className={styles.rankingTable}>
            <thead>
              <tr>
                <th>PN</th>
                <th>Produto</th>
                <th>Giro</th>
                <th>Vendas</th>
              </tr>
            </thead>
            <tbody>
              {model.rankings.topGiro.map((row) => (
                <tr key={`g-${row.pn}`}>
                  <td>{row.pn}</td>
                  <td
                    className={`${styles.productName} ${styles.productNameInteractive}`}
                    title={row.product}
                    onDoubleClick={() => openWorkspace(row.pn)}
                  >
                    {row.product}
                  </td>
                  <td className={`${styles.metricValue} ${styles.metricValue_positive}`}>{row.metric}</td>
                  <td className={styles.metricValue}>{row.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </article>

        <article className={styles.rankingCard}>
          <h3 className={styles.cardTitle}>Piores Giro (parados ha mais tempo)</h3>
          <table className={styles.rankingTable}>
            <thead>
              <tr>
                <th>PN</th>
                <th>Produto</th>
                <th>Dias</th>
                <th>R$ inv</th>
              </tr>
            </thead>
            <tbody>
              {model.rankings.worstGiro.map((row) => (
                <tr key={`w-${row.pn}`}>
                  <td>{row.pn}</td>
                  <td
                    className={`${styles.productName} ${styles.productNameInteractive}`}
                    title={row.product}
                    onDoubleClick={() => openWorkspace(row.pn)}
                  >
                    {row.product}
                  </td>
                  <td className={`${styles.metricValue} ${styles.metricValue_negative}`}>{row.metric}</td>
                  <td className={styles.metricValue}>{row.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </article>
      </section>

      <section className={styles.trendsGrid}>
        {model.trends.map((trend) => (
          <article key={trend.label} className={styles.trendCard}>
            <header className={styles.trendHeader}>
              <span className={styles.trendLabel}>{trend.label}</span>
              <span className={`${styles.trendChange} ${styles[`trendChange_${trend.tone}`]}`}>
                <span aria-hidden>{trend.tone === "positive" ? "\u2197" : "\u2198"}</span>
                <span>{trend.change}</span>
              </span>
            </header>
            <SparklineMini points={trend.points} tone={trend.tone} />
          </article>
        ))}
      </section>

      <AnalyticsSpotlightDrawer
        open={Boolean(activeSpotlight)}
        title={activeSpotlight ? `Spotlight - ${activeSpotlight.label}` : "Spotlight"}
        meta={activeSpotlight?.subtitle}
        onClose={() => setActiveQuadrant(null)}
      >
        {activeSpotlight ? (
          <SpotlightSelectionWidget
            tableTitle={activeSpotlight.tableTitle}
            tableThirdHeader={activeSpotlight.thirdHeader}
            tableRows={activeSpotlight.rows}
            tableEmptyText={activeSpotlight.emptyText}
            tableDefaultSort={activeSpotlight.defaultSort}
            onOpenSku={openWorkspace}
          />
        ) : null}
      </AnalyticsSpotlightDrawer>
    </div>
  );
}
