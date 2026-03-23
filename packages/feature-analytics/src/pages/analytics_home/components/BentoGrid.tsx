// @ts-nocheck
import { AlertsList } from "./AlertsList";
import { CardShell } from "./CardShell";
import { HeatmapMatrix } from "./HeatmapMatrix";
import { KpisPanel } from "./KpisPanel";
import { PortfolioDistribution } from "./PortfolioDistribution";
import { Timeline } from "./Timeline";
import { TopActionsList } from "./TopActionsList";
import { TopMetalTiles } from "./TopMetalTiles";
import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import styles from "../analytics_home.module.css";

type FilterKey = "all" | "critical" | "pricing" | "stock" | "data";

type Props = {
  onOpenSpotlight: (key: string) => void;
  activeFilter: FilterKey;
  model: AnalyticsHomeViewModel;
};

function cardClass(activeFilter: FilterKey, filterSet: FilterKey[]) {
  if (activeFilter === "all") return "";
  return filterSet.includes(activeFilter) ? "" : styles.cardMuted;
}

export function BentoGrid({ onOpenSpotlight, activeFilter, model }: Props) {
  const defaultHeatSpotlightKey = (() => {
    const candidate = Object.values(model.heatCells)
      .filter((cell) => Number(cell.count || 0) > 0)
      .sort((a, b) => Number(b.count || 0) - Number(a.count || 0))[0];
    return candidate?.key || "mini-heat";
  })();

  return (
    <section className={styles.grid}>
      <div className={styles.c_actions}>
        <CardShell
          className={`${cardClass(activeFilter, ["stock", "critical"])} ${styles.cardActions}`.trim()}
          emoji="⚡"
          title="Ações Prioritárias Hoje"
          subtitle="Clique para abrir o Spotlight com o por que + próximos passos"
          actionLabel="Abrir fila ->"
          onAction={() => onOpenSpotlight("mini-actions")}
        >
          <TopActionsList onOpenSpotlight={onOpenSpotlight} rows={model.actions} />
        </CardShell>
      </div>

      <div className={styles.c_heat}>
        <CardShell
          className={cardClass(activeFilter, ["critical", "pricing"])}
          emoji="🧭"
          title="Radar de Saude (amostra)"
          subtitle="Matriz: impacto x urgencia (quanto mais quente, mais acao)"
          actionLabel="Como ler ->"
          onAction={() => onOpenSpotlight(defaultHeatSpotlightKey)}
        >
          <HeatmapMatrix onOpenSpotlight={onOpenSpotlight} values={model.heatmap} />
        </CardShell>
      </div>

      <div className={styles.c_alerts}>
        <CardShell
          className={cardClass(activeFilter, ["critical", "data"])}
          emoji="🔔"
          title="Alertas Ativos"
          subtitle="Apenas alertas com regra definida (sem ruido)"
          actionLabel="Ver todos ->"
          onAction={() => onOpenSpotlight("stat-alerts")}
        >
          <AlertsList onOpenSpotlight={onOpenSpotlight} rows={model.alerts} />
        </CardShell>
      </div>

      <div className={styles.c_portstack}>
        <CardShell
          className={`${cardClass(activeFilter, ["pricing"])} ${styles.cardTopMetal}`.trim()}
          emoji="🏆"
          title="Top Metal"
          subtitle="Maior lucro no ultimo mes fechado (sales_monthly)"

        >
          <TopMetalTiles rows={model.topMetal} />
        </CardShell>
        <CardShell
          className={`${cardClass(activeFilter, ["stock", "pricing"])} ${styles.cardPortfolio}`.trim()}
          emoji="📦"
          title="Portfolio Distribution"
          subtitle="Decisao por SKU (regra deterministica + explicavel)"
          actionLabel="Detalhar ->"
          onAction={() => onOpenSpotlight("mini-portfolio")}
        >
          <PortfolioDistribution onOpenSpotlight={onOpenSpotlight} rows={model.portfolio} />
        </CardShell>
      </div>

      <div className={styles.c_kpis}>
        <CardShell
          className={cardClass(activeFilter, ["all"])}
          emoji="📊"
          title="KPIs Executivos"
          subtitle="Mini-tendencias (7D/semana) • clique para ver contexto"
        >
          <KpisPanel rows={model.kpis} />
        </CardShell>
      </div>

      <div className={styles.c_timeline}>
        <CardShell
          className={cardClass(activeFilter, ["data"])}
          emoji="🕒"
          title="Timeline"
          subtitle="O que aconteceu desde a ultima run"
        >
          <Timeline rows={model.timeline} />
        </CardShell>
      </div>
    </section>
  );
}

