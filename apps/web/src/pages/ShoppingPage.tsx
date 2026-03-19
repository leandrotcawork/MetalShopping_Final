import { useEffect, useMemo, useState } from "react";

import { AppFrame, MetricCard, SurfaceCard } from "@metalshopping/ui";
import type { ServerCoreSdk, ShoppingRunStatus } from "@metalshopping/sdk-runtime";
import type { ShoppingRunV1 } from "@metalshopping/sdk-types";

import styles from "./ShoppingPage.module.css";

type ShoppingPageProps = {
  api: ServerCoreSdk["shopping"];
};

const statusOptions: Array<{ value: ShoppingRunStatus | "all"; label: string }> = [
  { value: "all", label: "Todos" },
  { value: "queued", label: "Queued" },
  { value: "running", label: "Running" },
  { value: "completed", label: "Completed" },
  { value: "failed", label: "Failed" },
];

function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return "--";
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) {
    return value;
  }
  return parsed.toLocaleString("pt-BR");
}

function statusClass(status: string) {
  switch (status) {
    case "completed":
      return styles.statusCompleted;
    case "failed":
      return styles.statusFailed;
    case "running":
      return styles.statusRunning;
    case "queued":
      return styles.statusQueued;
    default:
      return "";
  }
}

export function ShoppingPage({ api }: ShoppingPageProps) {
  const [selectedStatus, setSelectedStatus] = useState<ShoppingRunStatus | "all">("all");
  const [summary, setSummary] = useState<Awaited<ReturnType<ServerCoreSdk["shopping"]["getSummary"]>> | null>(null);
  const [runs, setRuns] = useState<ShoppingRunV1[]>([]);
  const [selectedRun, setSelectedRun] = useState<ShoppingRunV1 | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const [nextSummary, nextRuns] = await Promise.all([
          api.getSummary(),
          api.listRuns(selectedStatus === "all" ? {} : { status: selectedStatus, limit: 20, offset: 0 }),
        ]);
        if (cancelled) {
          return;
        }
        setSummary(nextSummary);
        setRuns(nextRuns.rows);
        if (nextRuns.rows.length > 0) {
          const firstRun = await api.getRun(nextRuns.rows[0].runId);
          if (!cancelled) {
            setSelectedRun(firstRun);
          }
        } else {
          setSelectedRun(null);
        }
      } catch (loadError) {
        if (!cancelled) {
          const message = loadError instanceof Error ? loadError.message : "Falha ao carregar Shopping.";
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
  }, [api, selectedStatus]);

  const lastRunLabel = useMemo(() => formatDateTime(summary?.lastRunAt), [summary?.lastRunAt]);

  async function handleRunSelect(runId: string) {
    setError(null);
    try {
      const run = await api.getRun(runId);
      setSelectedRun(run);
    } catch (selectError) {
      const message = selectError instanceof Error ? selectError.message : "Falha ao carregar detalhe do run.";
      setError(message);
    }
  }

  return (
    <AppFrame
      eyebrow="MetalShopping"
      title="Shopping de Precos"
      subtitle="Leitura operacional de runs e snapshots ja persistidos no Postgres."
      aside={
        <div className={styles.metaPanel}>
          <p className={styles.metaKicker}>Ultimo run</p>
          <p className={styles.metaValue}>{loading ? "Atualizando..." : lastRunLabel}</p>
        </div>
      }
    >
      <div className={styles.stack}>
        {error ? <p className={styles.error}>{error}</p> : null}

        <div className={styles.metrics}>
          <MetricCard label="Total runs" value={String(summary?.totalRuns ?? 0)} hint="Historico total de execucoes" />
          <MetricCard label="Running" value={String(summary?.runningRuns ?? 0)} hint="Execucoes em andamento" />
          <MetricCard label="Completed" value={String(summary?.completedRuns ?? 0)} hint="Execucoes concluidas" />
          <MetricCard label="Failed" value={String(summary?.failedRuns ?? 0)} hint="Execucoes com falha" />
        </div>

        <div className={styles.grid}>
          <SurfaceCard title="Runs recentes" subtitle="Filtrados por status">
            <div className={styles.filterBar}>
              {statusOptions.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  className={`${styles.filterButton} ${selectedStatus === option.value ? styles.filterActive : ""}`.trim()}
                  onClick={() => setSelectedStatus(option.value)}
                >
                  {option.label}
                </button>
              ))}
            </div>

            {runs.length === 0 ? (
              <p className={styles.empty}>Nenhum run encontrado para o filtro atual.</p>
            ) : (
              <ul className={styles.runList}>
                {runs.map((run) => (
                  <li key={run.runId}>
                    <button type="button" className={styles.runButton} onClick={() => void handleRunSelect(run.runId)}>
                      <span className={styles.runMain}>
                        <strong className={styles.runId}>{run.runId}</strong>
                        <small className={styles.runTime}>{formatDateTime(run.startedAt)}</small>
                      </span>
                      <span className={`${styles.statusPill} ${statusClass(run.status)}`.trim()}>{run.status}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </SurfaceCard>

          <SurfaceCard title="Detalhe do run" subtitle="Snapshot do run selecionado">
            {!selectedRun ? (
              <p className={styles.empty}>Selecione um run para visualizar os detalhes.</p>
            ) : (
              <dl className={styles.detailGrid}>
                <div>
                  <dt>Run ID</dt>
                  <dd>{selectedRun.runId}</dd>
                </div>
                <div>
                  <dt>Status</dt>
                  <dd>{selectedRun.status}</dd>
                </div>
                <div>
                  <dt>Inicio</dt>
                  <dd>{formatDateTime(selectedRun.startedAt)}</dd>
                </div>
                <div>
                  <dt>Fim</dt>
                  <dd>{formatDateTime(selectedRun.finishedAt ?? null)}</dd>
                </div>
                <div>
                  <dt>Itens processados</dt>
                  <dd>{selectedRun.processedItems}</dd>
                </div>
                <div>
                  <dt>Total de itens</dt>
                  <dd>{selectedRun.totalItems}</dd>
                </div>
                <div className={styles.notes}>
                  <dt>Notas</dt>
                  <dd>{selectedRun.notes?.trim() ? selectedRun.notes : "Sem observacoes."}</dd>
                </div>
              </dl>
            )}
          </SurfaceCard>
        </div>
      </div>
    </AppFrame>
  );
}
