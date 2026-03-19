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

  return (
    <AppFrame
      eyebrow="MetalShopping"
      title="Home"
      subtitle="Resumo operacional com dados reais do backend."
      aside={
        <p className={styles.meta}>
          {loading ? "Atualizando..." : `Ultima atualizacao: ${lastUpdatedLabel}`}
        </p>
      }
    >
      <div className={styles.stack}>
        {error ? <p className={styles.error}>{error}</p> : null}

        <div className={styles.metrics}>
          <MetricCard
            label="Produtos"
            value={summary?.productCount ?? 0}
            hint="Total de produtos cadastrados"
          />
          <MetricCard
            label="Produtos ativos"
            value={summary?.activeProductCount ?? 0}
            hint="Itens ativos no catalogo"
          />
          <MetricCard
            label="Com preco atual"
            value={summary?.pricedProductCount ?? 0}
            hint="Itens com preco vigente"
          />
          <MetricCard
            label="Com estoque atual"
            value={summary?.inventoryTrackedCount ?? 0}
            hint="Itens com posicao de estoque vigente"
          />
        </div>

        <SurfaceCard title="Status do modulo Home">
          <p className={styles.cardText}>
            {loading
              ? "Carregando indicadores..."
              : "Resumo operacional conectado em dados reais do Postgres."}
          </p>
        </SurfaceCard>
      </div>
    </AppFrame>
  );
}
