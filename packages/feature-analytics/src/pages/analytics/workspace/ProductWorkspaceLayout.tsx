import { ApiClientError } from "../../../app/apiClient";
import { makeAnalyticsProductWorkspaceV1Dto, type AnalyticsProductWorkspaceV1Dto } from "@metalshopping/feature-analytics";
import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation, useParams } from "react-router-dom";

import { useAppSession } from "../../../app/providers/AppProviders";
import { ProductHero } from "./components/ProductHero";
import { UnifiedWorkspaceHeader } from "./components/UnifiedWorkspaceHeader";
import styles from "./product_workspace.module.css";

export type ProductWorkspaceOutletContext = {
  model: AnalyticsProductWorkspaceV1Dto["model"];
};

export function ProductWorkspaceLayout() {
  const { api } = useAppSession();
  const params = useParams<{ pn: string }>();
  const location = useLocation();
  const [model, setModel] = useState<AnalyticsProductWorkspaceV1Dto["model"] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");
  const pn = String(params.pn || "").trim();
  const stateFromPath = typeof location.state === "object" && location.state && "from" in location.state
    ? String((location.state as { from?: string }).from || "").trim() || null
    : null;
  const stateFromScrollY = typeof location.state === "object" && location.state && "fromScrollY" in location.state
    ? Number((location.state as { fromScrollY?: unknown }).fromScrollY)
    : NaN;
  const sessionFallback = (() => {
    try {
      const raw = window.sessionStorage.getItem("analytics:workspace:return");
      if (!raw) return { from: null as string | null, fromScrollY: null as number | null };
      const payload = JSON.parse(raw) as { from?: unknown; fromScrollY?: unknown };
      const from = String(payload.from || "").trim() || null;
      const fromScrollY = Number(payload.fromScrollY);
      return {
        from,
        fromScrollY: Number.isFinite(fromScrollY) && fromScrollY >= 0 ? fromScrollY : null,
      };
    } catch {
      return { from: null as string | null, fromScrollY: null as number | null };
    }
  })();
  const fromPath = (() => {
    const statePath = stateFromPath || null;
    const sessionPath = sessionFallback.from || null;
    if (!statePath) return sessionPath;
    if (!sessionPath) return statePath;

    // If navigation explicitly came from a non-spotlight route (ex: /analytics/products),
    // never override with stale spotlight return stored in session.
    const stateHasSpotlight = /[?&]spotlight=/.test(statePath);
    if (!stateHasSpotlight) return statePath;

    const sessionHasSpotlight = /[?&]spotlight=/.test(sessionPath);
    if (sessionHasSpotlight && !stateHasSpotlight) return sessionPath;

    const stateHasPageIndex = /[?&]spotlightPageIndex=/.test(statePath);
    const sessionHasPageIndex = /[?&]spotlightPageIndex=/.test(sessionPath);
    if (sessionHasPageIndex && !stateHasPageIndex) return sessionPath;

    return statePath;
  })();
  const fromScrollYSafe = Number.isFinite(stateFromScrollY) && stateFromScrollY >= 0
    ? stateFromScrollY
    : sessionFallback.fromScrollY;

  useEffect(() => {
    if (!pn) {
      setModel(null);
      setIsLoading(false);
      setErrorMessage("");
      return;
    }
    let disposed = false;
    async function loadWorkspace() {
      setIsLoading(true);
      setErrorMessage("");
      try {
        const env = await api.products.workspace(pn);
        if (disposed) return;
        const dto = makeAnalyticsProductWorkspaceV1Dto(env.data, "current");
        setModel(dto.model);
      } catch (err) {
        if (disposed) return;
        const apiErr = err instanceof ApiClientError ? err : null;
        setErrorMessage(apiErr?.message || (err instanceof Error ? err.message : String(err)));
        setModel(null);
      } finally {
        if (!disposed) setIsLoading(false);
      }
    }
    void loadWorkspace();
    return () => {
      disposed = true;
    };
  }, [api, pn]);

  if (!pn) {
    return <Navigate to="/analytics/products" replace />;
  }

  if (isLoading) {
    return (
      <section className={styles.page}>
        <div className={styles.container}>
          <section className={styles.placeholder}>
            <h2>Carregando workspace...</h2>
          </section>
        </div>
      </section>
    );
  }

  if (!model) {
    return (
      <section className={styles.page}>
        <div className={styles.container}>
          <section className={styles.placeholder}>
            <h2>Nao foi possivel carregar o workspace</h2>
            <p>{errorMessage || "Sem dados para este produto."}</p>
          </section>
        </div>
      </section>
    );
  }

  const showHero = location.pathname.endsWith("/overview");

  return (
    <section className={styles.page}>
      <div className={styles.ambient} aria-hidden>
        <span className={`${styles.blob} ${styles.blobA}`} />
        <span className={`${styles.blob} ${styles.blobB}`} />
      </div>
      <header className={styles.workspaceShell}>
        <UnifiedWorkspaceHeader model={model} fromPath={fromPath} fromScrollY={fromScrollYSafe} />
      </header>
      <div className={styles.container}>
        {showHero ? <ProductHero model={model} /> : null}
        <Outlet context={{ model } satisfies ProductWorkspaceOutletContext} />
      </div>
    </section>
  );
}
