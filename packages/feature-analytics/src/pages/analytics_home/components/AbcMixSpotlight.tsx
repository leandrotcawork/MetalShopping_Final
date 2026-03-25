import { useMemo } from "react";

import { FilterDropdown, type SelectMenuOption } from "@metalshopping/ui";
import styles from "../analytics_home.module.css";

type AbcMixRow = {
  nodeId: number;
  nodeName: string;
  revenueBrl: number;
  sharePct: number;
  cumSharePct: number;
  band: "A" | "B" | "C";
  marginPct: number | null;
};

type AbcMixSpotlightProps = {
  rows: AbcMixRow[];
  aMaxCumPct: number;
  bMaxCumPct: number;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  brandOptions: string[];
  brandFilter: string[];
  onBrandFilterChange: (next: string[]) => void;
  loading?: boolean;
  error?: string;
};

function asCurrency(value: number): string {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
}

function asPct(value: number): string {
  return `${Number(value || 0).toFixed(1)}%`;
}

function toggleMultiValue(current: string[], value: string): string[] {
  if (!value || value === "all") return [];
  if (current.includes(value)) return current.filter((item) => item !== value);
  return [...current, value];
}

function toneClassFromBand(band: "A" | "B" | "C"): string {
  if (band === "A") return styles.abcSpotlightToneA;
  if (band === "B") return styles.abcSpotlightToneB;
  return styles.abcSpotlightToneC;
}

const SPOTLIGHT_FILTER_CLASSNAMES = {
  trigger: styles.spotlightDsTrigger,
  menu: styles.spotlightDsMenu,
  option: styles.spotlightDsOption,
  optionActive: styles.spotlightDsOptionActive,
  searchInput: styles.spotlightDsSearchInput,
};

export function AbcMixSpotlight({
  rows,
  aMaxCumPct,
  bMaxCumPct,
  searchQuery,
  onSearchQueryChange,
  brandOptions,
  brandFilter,
  onBrandFilterChange,
  loading = false,
  error = "",
}: AbcMixSpotlightProps) {
  const topA = useMemo(() => rows.filter((row) => row.band === "A").slice(0, 5), [rows]);
  const topB = useMemo(() => rows.filter((row) => row.band === "B").slice(0, 5), [rows]);
  const topC = useMemo(() => rows.filter((row) => row.band === "C").slice(0, 5), [rows]);
  const previewRows = useMemo(() => rows.slice(0, 30), [rows]);

  const marcaOptions = useMemo<SelectMenuOption[]>(
    () => [{ label: "Todas Marcas", value: "all" }, ...brandOptions.map((item) => ({ label: item, value: item }))],
    [brandOptions]
  );

  return (
    <section className={styles.abcSpotlightBlock}>
      <div className={styles.abcSpotlightToolbar}>
        <label className={styles.abcSpotlightField}>
          <span>Buscar grupo</span>
          <input
            type="text"
            value={searchQuery}
            onChange={(event) => onSearchQueryChange(event.target.value)}
            placeholder="Ex.: Assentos, Metais, Loucas..."
          />
        </label>
        <label className={styles.abcSpotlightField}>
          <span>Marca</span>
          <FilterDropdown
            id="abc-mix-brand-filter"
            selectionMode="duo"
            value=""
            values={brandFilter}
            options={marcaOptions}
            onSelect={(value) => onBrandFilterChange(toggleMultiValue(brandFilter, value))}
            classNamesOverrides={{ wrap: styles.capitalSpotlightSelectWrap, ...SPOTLIGHT_FILTER_CLASSNAMES }}
          />
        </label>
      </div>

      {loading ? <div className={styles.abcSpotlightState}>Carregando mix ABC...</div> : null}
      {!loading && error ? <div className={styles.abcSpotlightState}>{error}</div> : null}
      {!loading && !error && rows.length === 0 ? (
        <div className={styles.abcSpotlightState}>Sem dados para os filtros selecionados.</div>
      ) : null}

      {!loading && !error && rows.length > 0 ? (
        <>
          <div className={styles.abcSpotlightThresholds}>
            <span>Faixa A ate {asPct(aMaxCumPct)}</span>
            <span>Faixa B ate {asPct(bMaxCumPct)}</span>
            <span>Faixa C acima de {asPct(bMaxCumPct)}</span>
          </div>

          <div className={styles.abcSpotlightPareto}>
            {previewRows.map((row) => {
              const widthPct = Math.max(2, Math.min(100, row.cumSharePct));
              return (
                <div key={`abc-${row.nodeId}`} className={styles.abcSpotlightParetoRow}>
                  <div className={styles.abcSpotlightParetoLabel}>
                    <strong>{row.nodeName}</strong>
                    <span>{asCurrency(row.revenueBrl)} | {asPct(row.sharePct)}</span>
                  </div>
                  <div className={styles.abcSpotlightParetoTrack}>
                    <div
                      className={`${styles.abcSpotlightParetoFill} ${toneClassFromBand(row.band)}`}
                      style={{ width: `${widthPct}%` }}
                    />
                  </div>
                  <div className={styles.abcSpotlightParetoMeta}>
                    <span className={styles.abcSpotlightBandPill}>{row.band}</span>
                    <span>{asPct(row.cumSharePct)}</span>
                  </div>
                </div>
              );
            })}
          </div>

          <div className={styles.abcSpotlightTopGrid}>
            <article className={styles.abcSpotlightTopCard}>
              <h4>Top A</h4>
              {topA.map((row, idx) => (
                <div key={`top-a-${row.nodeId}`} className={styles.abcSpotlightTopItem}>
                  <span>{idx + 1}. {row.nodeName}</span>
                  <strong>{asCurrency(row.revenueBrl)}</strong>
                </div>
              ))}
            </article>
            <article className={styles.abcSpotlightTopCard}>
              <h4>Top B</h4>
              {topB.map((row, idx) => (
                <div key={`top-b-${row.nodeId}`} className={styles.abcSpotlightTopItem}>
                  <span>{idx + 1}. {row.nodeName}</span>
                  <strong>{asCurrency(row.revenueBrl)}</strong>
                </div>
              ))}
            </article>
            <article className={styles.abcSpotlightTopCard}>
              <h4>Top C</h4>
              {topC.map((row, idx) => (
                <div key={`top-c-${row.nodeId}`} className={styles.abcSpotlightTopItem}>
                  <span>{idx + 1}. {row.nodeName}</span>
                  <strong>{asCurrency(row.revenueBrl)}</strong>
                </div>
              ))}
            </article>
          </div>

          <div className={styles.abcSpotlightTableWrap}>
            <table className={styles.abcSpotlightTable}>
              <thead>
                <tr>
                  <th>Grupo</th>
                  <th>Faixa</th>
                  <th>Receita</th>
                  <th>Share</th>
                  <th>Acumulado</th>
                  <th>Margem %</th>
                </tr>
              </thead>
              <tbody>
                {rows.slice(0, 80).map((row) => (
                  <tr key={`abc-row-${row.nodeId}`}>
                    <td>{row.nodeName}</td>
                    <td><span className={styles.abcSpotlightBandPill}>{row.band}</span></td>
                    <td>{asCurrency(row.revenueBrl)}</td>
                    <td>{asPct(row.sharePct)}</td>
                    <td>{asPct(row.cumSharePct)}</td>
                    <td>{row.marginPct == null ? "-" : asPct(row.marginPct)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : null}
    </section>
  );
}


