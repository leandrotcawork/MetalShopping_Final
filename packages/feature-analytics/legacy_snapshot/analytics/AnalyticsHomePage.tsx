import { makeAnalyticsHomeV2Dto } from "@metalshopping/feature-analytics";
import { useCallback, useEffect, useMemo, useState } from "react";

import { useAppSession } from "../../app/providers/AppProviders";
import { AnalyticsHomePage as AnalyticsHomeContent } from "../analytics_home/AnalyticsHomePage";
import styles from "./analytics_home.module.css";

function formatAsOf(value: string | null | undefined): string {
  const token = String(value || "").trim();
  if (!token) return "N/D";
  const parsed = new Date(token);
  if (Number.isNaN(parsed.getTime())) return token;
  return new Intl.DateTimeFormat("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  }).format(parsed);
}

export function AnalyticsHomePage() {
  const { api, analyticsHomeSnapshot, setAnalyticsHomeSnapshot } = useAppSession();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const refreshSnapshot = useCallback(async () => {
    setIsRefreshing(true);
    try {
      const [workspaceEnvelope, operationalEnvelope] = await Promise.all([
        api.home.workspace(undefined, { includeOperational: false }),
        api.home.operational(),
      ]);
      const envelope = {
        ...workspaceEnvelope,
        data: {
          ...(workspaceEnvelope?.data || {}),
          operational: (operationalEnvelope as { data?: Record<string, unknown> } | null)?.data?.operational || {},
        },
      };
      const dto = makeAnalyticsHomeV2Dto(
        envelope as { data: Record<string, unknown>; meta?: Record<string, unknown> },
        "current"
      );
      setAnalyticsHomeSnapshot({
        data: dto,
        asOf: String(dto.snapshot.as_of || ""),
        updatedAt: Date.now()
      });
    } finally {
      setIsRefreshing(false);
    }
  }, [api, setAnalyticsHomeSnapshot]);

  useEffect(() => {
    let disposed = false;

    async function hydrateSnapshot() {
      if (analyticsHomeSnapshot?.data) return;
      try {
        const [workspaceEnvelope, operationalEnvelope] = await Promise.all([
          api.home.workspace(undefined, { includeOperational: false }),
          api.home.operational(),
        ]);
        const envelope = {
          ...workspaceEnvelope,
          data: {
            ...(workspaceEnvelope?.data || {}),
            operational: (operationalEnvelope as { data?: Record<string, unknown> } | null)?.data?.operational || {},
          },
        };
        const dto = makeAnalyticsHomeV2Dto(
          envelope as { data: Record<string, unknown>; meta?: Record<string, unknown> },
          "current"
        );
        if (disposed) return;
        setAnalyticsHomeSnapshot({
          data: dto,
          asOf: String(dto.snapshot.as_of || ""),
          updatedAt: Date.now()
        });
      } catch {
        // Keep page render resilient even when snapshot fetch fails.
      }
    }

    void hydrateSnapshot();
    return () => {
      disposed = true;
    };
  }, [api, analyticsHomeSnapshot?.data, setAnalyticsHomeSnapshot]);

  const updatedAtLabel = useMemo(() => {
    return formatAsOf(analyticsHomeSnapshot?.asOf || analyticsHomeSnapshot?.data.snapshot.as_of);
  }, [analyticsHomeSnapshot]);

  return (
    <section className={styles.page}>
      <AnalyticsHomeContent
        updatedAtLabel={updatedAtLabel}
        dto={analyticsHomeSnapshot?.data || null}
        onRefresh={refreshSnapshot}
        isRefreshing={isRefreshing}
      />
    </section>
  );
}
