import { useEffect, useMemo, useState } from "react";

import { AppFrame, MetricCard, SurfaceCard } from "@metalshopping/ui";
import type { ServerCoreSdk } from "@metalshopping/sdk-runtime";
import type { HomeSummaryV1 } from "@metalshopping/sdk-types";

import styles from "./HomePage.module.css";

type HomePageProps = {
  api: ServerCoreSdk["home"];
};

function formatDateTime(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) {
    return value;
  }
  return parsed.toLocaleString("pt-BR");
}

function toPercent(numerator: number, denominator: number) {
  if (denominator <= 0) {
    return 0;
  }
  return Math.round((numerator / denominator) * 100);
}

export function HomePage({ api }: HomePageProps) {
  const [summary, setSummary] = useState<HomeSummaryV1 | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const next = await api.getSummary();
        if (!cancelled) {
          setSummary(next);
        }
      } catch (loadError) {
        if (!cancelled) {
          const message = loadError instanceof Error ? loadError.message : "Falha ao carregar resumo da Home.";
          setError(message);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [api]);

  const lastUpdatedLabel = useMemo(() => {
    if (!summary) {
      return "--";
    }
    return formatDateTime(summary.lastUpdated);
  }, [summary]);

  const totalProducts = summary?.productCount ?? 0;
  const activeProducts = summary?.activeProductCount ?? 0;
  const pricedProducts = summary?.pricedProductCount ?? 0;
  const trackedProducts = summary?.inventoryTrackedCount ?? 0;

  const activeCoverage = toPercent(activeProducts, totalProducts);
  const pricingCoverage = toPercent(pricedProducts, totalProducts);
  const inventoryCoverage = toPercent(trackedProducts, totalProducts);

  return (
    <AppFrame
      eyebrow="MetalShopping"
      title="Centro Operacional"
      subtitle="Visao executiva do acervo atual para catalogo, preco e estoque com dados reais do backend."
      aside={
        <div className={styles.metaPanel}>
          <p className={styles.metaKicker}>Sync runtime</p>
          <p className={styles.metaValue}>{loading ? "Atualizando..." : lastUpdatedLabel}</p>
        </div>
      }
    >
      <div className={styles.stack}>
        {error ? <p className={styles.error}>{error}</p> : null}

        <div className={styles.metrics}>
          <MetricCard
            label="Produtos"
            value={totalProducts.toLocaleString("pt-BR")}
            hint="Total de produtos cadastrados"
          />
          <MetricCard
            label="Produtos ativos"
            value={`${activeProducts.toLocaleString("pt-BR")} (${activeCoverage}%)`}
            hint="Itens ativos no catalogo"
          />
          <MetricCard
            label="Com preco atual"
            value={`${pricedProducts.toLocaleString("pt-BR")} (${pricingCoverage}%)`}
            hint="Itens com preco vigente"
          />
          <MetricCard
            label="Com estoque atual"
            value={`${trackedProducts.toLocaleString("pt-BR")} (${inventoryCoverage}%)`}
            hint="Itens com posicao de estoque vigente"
          />
        </div>

        <div className={styles.operationalGrid}>
          <SurfaceCard title="Cobertura operacional" subtitle="Leitura rapida por frente de operacao" tone="soft">
            <div className={styles.coverageList}>
              <article className={styles.coverageItem}>
                <header className={styles.coverageHead}>
                  <span>Catalogo ativo</span>
                  <strong>{activeCoverage}%</strong>
                </header>
                <div className={styles.coverageBar}>
                  <span style={{ width: `${activeCoverage}%` }} />
                </div>
              </article>
              <article className={styles.coverageItem}>
                <header className={styles.coverageHead}>
                  <span>Preco vigente</span>
                  <strong>{pricingCoverage}%</strong>
                </header>
                <div className={styles.coverageBar}>
                  <span style={{ width: `${pricingCoverage}%` }} />
                </div>
              </article>
              <article className={styles.coverageItem}>
                <header className={styles.coverageHead}>
                  <span>Estoque rastreado</span>
                  <strong>{inventoryCoverage}%</strong>
                </header>
                <div className={styles.coverageBar}>
                  <span style={{ width: `${inventoryCoverage}%` }} />
                </div>
              </article>
            </div>
          </SurfaceCard>

          <SurfaceCard title="Leitura do momento" subtitle="Snapshot para decisao rapida">
            <ul className={styles.snapshotList}>
              <li>
                <span>Produtos fora do estado ativo</span>
                <strong>{Math.max(totalProducts - activeProducts, 0).toLocaleString("pt-BR")}</strong>
              </li>
              <li>
                <span>Sem preco vigente</span>
                <strong>{Math.max(totalProducts - pricedProducts, 0).toLocaleString("pt-BR")}</strong>
              </li>
              <li>
                <span>Sem estoque rastreado</span>
                <strong>{Math.max(totalProducts - trackedProducts, 0).toLocaleString("pt-BR")}</strong>
              </li>
            </ul>
          </SurfaceCard>
        </div>

        <SurfaceCard title="Status do modulo Home" subtitle="Entrega nivel 1 no padrao make-it-work-first">
          <p className={styles.cardText}>
            {loading
              ? "Carregando indicadores reais do Postgres..."
              : "Home conectada ao endpoint backend-owned `/api/v1/home/summary`, sem mock e sem composicao de dados no frontend."}
          </p>
        </SurfaceCard>
      </div>
    </AppFrame>
  );
}
