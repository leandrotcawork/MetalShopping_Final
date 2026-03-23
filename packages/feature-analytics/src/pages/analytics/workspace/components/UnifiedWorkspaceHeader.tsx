// @ts-nocheck
import type { AnalyticsProductWorkspaceV1Dto } from "@metalshopping/feature-analytics";
import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAppSession } from "../../../../app/providers/AppProviders";

import { WorkspaceTabs } from "./WorkspaceTabs";
import styles from "../product_workspace.module.css";

type SkuPickerRow = {
  pn: string;
  description: string;
  brand: string;
  taxonomyLeafName: string;
  ourPrice: number | null;
  marketPrice: number | null;
  stockUn: number | null;
  gapPct: number | null;
  marginPct: number | null;
};

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});
const integerFormatter = new Intl.NumberFormat("pt-BR");

function asNumberOrNull(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return null;
}

function formatMoney(value: number | null): string {
  if (value == null) return "-";
  return currencyFormatter.format(value);
}

function formatStock(value: number | null): string {
  if (value == null) return "-";
  return `${integerFormatter.format(Math.max(0, Math.round(value)))} un`;
}

function formatPercent(value: number | null): string {
  if (value == null) return "-";
  return `${value.toFixed(1)}%`;
}

type UnifiedWorkspaceHeaderProps = {
  model: AnalyticsProductWorkspaceV1Dto["model"];
  fromPath: string | null;
  fromScrollY: number | null;
};

export function UnifiedWorkspaceHeader({ model, fromPath, fromScrollY }: UnifiedWorkspaceHeaderProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { api, shellState } = useAppSession();
  const isOnline = shellState !== "sidecar_offline";
  const classBadges = model.badges.filter((badge) => badge.tone === "class");
  const [isSkuPickerOpen, setIsSkuPickerOpen] = useState(false);
  const [skuQuery, setSkuQuery] = useState("");
  const [skuLoading, setSkuLoading] = useState(false);
  const [skuRows, setSkuRows] = useState<SkuPickerRow[]>([]);

  function classBadgeTier(label: string): "A" | "B" | "C" | null {
    const token = String(label || "").trim().toUpperCase();
    if (token.startsWith("A")) return "A";
    if (token.startsWith("B")) return "B";
    if (token.startsWith("C")) return "C";
    return null;
  }

  function handleBack() {
    if (fromPath) {
      navigate(fromPath, {
        state: fromScrollY != null ? { restoreScrollY: fromScrollY } : undefined,
      });
      return;
    }
    if (window.history.length > 1) {
      navigate(-1);
      return;
    }
    navigate("/analytics/products");
  }

  function resolveTabSuffix(pathname: string): string {
    const match = pathname.match(/^\/analytics\/products\/[^/]+(\/.*)?$/);
    return match?.[1] || "/overview";
  }

  function handleSwapSku(targetPn: string) {
    const token = String(targetPn || "").trim();
    if (!token) return;
    const suffix = resolveTabSuffix(location.pathname);
    navigate(`/analytics/products/${encodeURIComponent(token)}${suffix}`, {
      state: {
        from: fromPath || "/analytics/products",
        fromScrollY: fromScrollY ?? 0,
      },
    });
    setIsSkuPickerOpen(false);
  }

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      const isHotkey = (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k";
      if (!isHotkey) return;
      event.preventDefault();
      setIsSkuPickerOpen(true);
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  useEffect(() => {
    if (!isSkuPickerOpen) return;
    function onEsc(event: KeyboardEvent) {
      if (event.key === "Escape") setIsSkuPickerOpen(false);
    }
    window.addEventListener("keydown", onEsc);
    let active = true;
    const timer = window.setTimeout(async () => {
      setSkuLoading(true);
      try {
        const env = await api.products.analyticsIndex({
          search: skuQuery.trim() || undefined,
          limit: 15,
          offset: 0,
        });
        if (!active) return;
        const rows = Array.isArray(env.data?.rows) ? env.data.rows : [];
        setSkuRows(
          rows.map((row) => ({
            pn: String(row.pn || ""),
            description: String(row.description || ""),
            brand: String(row.brand || ""),
            taxonomyLeafName: String(row.taxonomy_leaf_name || ""),
            ourPrice: asNumberOrNull(row.kpis?.our_price),
            marketPrice: asNumberOrNull(row.kpis?.comp_price_mean),
            stockUn: asNumberOrNull(row.kpis?.estoque_un),
            gapPct: asNumberOrNull(row.kpis?.gap_pct),
            marginPct: asNumberOrNull(row.kpis?.margem_contrib_pct),
          })),
        );
      } catch {
        if (active) setSkuRows([]);
      } finally {
        if (active) setSkuLoading(false);
      }
    }, 150);
    return () => {
      window.removeEventListener("keydown", onEsc);
      active = false;
      window.clearTimeout(timer);
    };
  }, [api, isSkuPickerOpen, skuQuery]);

  return (
    <header className={styles.headerUnified}>
      <div className={styles.headerInner}>
        <div className={styles.headerRow}>
          <div className={styles.branding}>
            METAL <span>ANALYTICS</span>
          </div>
          <div className={styles.breadcrumb}>Produtos / Workspace</div>
          <div className={styles.headerActions}>
            <div className={`${styles.connectPill} ${isOnline ? styles.connectPillOnline : styles.connectPillOffline}`}>
              <span className={`${styles.connectDot} ${isOnline ? styles.connectDotOnline : styles.connectDotOffline}`} />
              {isOnline ? "Online" : "Offline"}
            </div>
          </div>
        </div>

        <div className={styles.headerRow}>
          <button type="button" className={styles.backBtn} onClick={handleBack} aria-label="Voltar para lista">
            <svg className={styles.backBtnIcon} viewBox="0 0 20 20" fill="none" aria-hidden>
              <path
                d="M12.5 4.5L7.5 10L12.5 15.5"
                stroke="currentColor"
                strokeWidth="2.4"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </button>
          <div className={styles.skuIdentity}>
            <div className={styles.skuTitle}>{model.pn} • {model.brand} • {model.taxonomyLeafName}</div>
            <div className={styles.skuSub}>{model.subtitle}</div>
          </div>
          <div className={styles.skuBadges}>
            {classBadges.map((badge) => (
              <span
                key={`${badge.tone}-${badge.label}`}
                className={`${styles.badge} ${styles[`badge_${badge.tone}`]} ${
                  styles[`badge_class_${classBadgeTier(badge.label) || "default"}`]
                }`}
              >
                {badge.label}
              </span>
            ))}
          </div>
          <WorkspaceTabs className={styles.tabsInline} />
          <div className={styles.headerRightControls}>
            <button type="button" className={styles.swapBtn} onClick={() => setIsSkuPickerOpen(true)}>
              Trocar SKU (Ctrl+K)
            </button>
            <div className={styles.updateTime}>Atualizado em: {model.updatedAt}</div>
          </div>
        </div>
      </div>
      {isSkuPickerOpen ? (
        <div className={styles.skuPickerOverlay} onClick={() => setIsSkuPickerOpen(false)}>
          <div className={styles.skuPickerPanel} onClick={(event) => event.stopPropagation()}>
            <div className={styles.skuPickerHead}>
              <strong>Trocar SKU</strong>
              <button type="button" className={styles.skuPickerClose} onClick={() => setIsSkuPickerOpen(false)}>Fechar</button>
            </div>
            <input
              value={skuQuery}
              onChange={(event) => setSkuQuery(event.target.value)}
              className={styles.skuPickerInput}
              placeholder="Digite PN, descricao, marca ou hierarquia"
              autoFocus
            />
            <div className={styles.skuPickerList}>
              {skuLoading ? <div className={styles.skuPickerHint}>Buscando...</div> : null}
              {!skuLoading && skuRows.length === 0 ? <div className={styles.skuPickerHint}>Nenhum SKU encontrado.</div> : null}
              {!skuLoading
                ? skuRows.map((row) => (
                    <button key={row.pn} type="button" className={styles.skuPickerItem} onClick={() => handleSwapSku(row.pn)}>
                      <span className={styles.skuPickerMain}>
                        <span className={styles.skuPickerPn}>{row.pn}</span>
                        <span className={styles.skuPickerDesc}>{row.description}</span>
                        <span className={styles.skuPickerMeta}>{row.brand} • {row.taxonomyLeafName}</span>
                      </span>
                      <span className={styles.skuPickerKpis}>
                        <span className={styles.skuPickerKpi}><strong>Nosso:</strong> {formatMoney(row.ourPrice)}</span>
                        <span className={styles.skuPickerKpi}><strong>Mercado:</strong> {formatMoney(row.marketPrice)}</span>
                        <span className={styles.skuPickerKpi}><strong>Estoque:</strong> {formatStock(row.stockUn)}</span>
                        <span className={styles.skuPickerKpi}><strong>Gap:</strong> {formatPercent(row.gapPct)}</span>
                        <span className={styles.skuPickerKpi}><strong>Margem:</strong> {formatPercent(row.marginPct)}</span>
                      </span>
                    </button>
                  ))
                : null}
            </div>
          </div>
        </div>
      ) : null}
    </header>
  );
}

