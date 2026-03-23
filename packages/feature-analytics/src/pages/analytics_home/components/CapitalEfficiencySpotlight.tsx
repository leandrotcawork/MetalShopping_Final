// @ts-nocheck
import { useMemo, useState } from "react";

import { SelectMenu, type SelectMenuOption } from "../../../components/ui/SelectMenu";
import { createSpotlightSelectClassNames } from "../../../components/ui/spotlightSelect";
import styles from "../analytics_home.module.css";

type CapitalEfficiencyRow = {
  nodeId: number;
  nodeName: string;
  capitalBrl: number;
  riskLevel: "low" | "medium" | "high";
  riskPct: number;
  gmroi: number | null;
  revenueBrl: number;
  marginBrl: number;
};

type CapitalEfficiencySpotlightProps = {
  rows: CapitalEfficiencyRow[];
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  brandOptions: string[];
  brandFilter: string[];
  onBrandFilterChange: (next: string[]) => void;
  loading?: boolean;
  error?: string;
};

const FILTER_SELECT_CLASSNAMES = createSpotlightSelectClassNames({
  wrap: styles.capitalSpotlightSelectWrap,
});

function asCurrency(value: number): string {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
}

function asPct(value: number): string {
  return `${value.toFixed(1)}%`;
}

function asNumber(value: number): string {
  return new Intl.NumberFormat("pt-BR", {
    maximumFractionDigits: 2,
  }).format(value);
}

function toggleMultiValue(current: string[], value: string): string[] {
  if (!value || value === "all") return [];
  if (current.includes(value)) return current.filter((item) => item !== value);
  return [...current, value];
}

export function CapitalEfficiencySpotlight({
  rows,
  searchQuery,
  onSearchQueryChange,
  brandOptions,
  brandFilter,
  onBrandFilterChange,
  loading = false,
  error = "",
}: CapitalEfficiencySpotlightProps) {
  const [riskFilter, setRiskFilter] = useState<"all" | "high" | "medium" | "low">("all");
  const [gmroiOp, setGmroiOp] = useState<"gte" | "eq" | "lte">("lte");
  const [gmroiValue, setGmroiValue] = useState("");
  const [capitalOp, setCapitalOp] = useState<"gte" | "eq" | "lte">("gte");
  const [capitalValue, setCapitalValue] = useState("");
  const [sortBy, setSortBy] = useState<"priority" | "gmroi" | "capital">("priority");

  function passNumericFilter(
    op: "gte" | "eq" | "lte",
    rowValue: number,
    rawTarget: string,
  ): boolean {
    const parsed = Number(String(rawTarget || "").replace(",", "."));
    if (!Number.isFinite(parsed)) return true;
    if (op === "eq") return rowValue === parsed;
    if (op === "lte") return rowValue <= parsed;
    return rowValue >= parsed;
  }

  const filteredRows = useMemo(() => {
    const token = String(searchQuery || "").trim().toLocaleLowerCase("pt-BR");
    const filtered = rows
      .filter((row) => (riskFilter === "all" ? true : row.riskLevel === riskFilter))
      .filter((row) => (token ? row.nodeName.toLocaleLowerCase("pt-BR").includes(token) : true))
      .filter((row) => passNumericFilter(capitalOp, Number(row.capitalBrl || 0), capitalValue))
      .filter((row) => passNumericFilter(gmroiOp, Number(row.gmroi ?? Number.NaN), gmroiValue))
      .map((row) => {
        const gmroi = row.gmroi == null || !Number.isFinite(Number(row.gmroi)) ? 0 : Number(row.gmroi);
        const gmroiGap = Math.max(0, 1.5 - gmroi);
        const priorityScore = Number(row.capitalBrl || 0) * (1 + gmroiGap) * (1 + Math.max(0, Number(row.riskPct || 0)) / 100);
        return { ...row, priorityScore };
      });
    if (sortBy === "gmroi") {
      return filtered.sort((a, b) => {
        const gmA = a.gmroi == null ? Number.POSITIVE_INFINITY : Number(a.gmroi);
        const gmB = b.gmroi == null ? Number.POSITIVE_INFINITY : Number(b.gmroi);
        if (gmA !== gmB) return gmA - gmB;
        return b.capitalBrl - a.capitalBrl;
      });
    }
    if (sortBy === "capital") {
      return filtered.sort((a, b) => b.capitalBrl - a.capitalBrl);
    }
    return filtered.sort((a, b) => b.priorityScore - a.priorityScore);
  }, [capitalOp, capitalValue, gmroiOp, gmroiValue, riskFilter, rows, searchQuery, sortBy]);

  const priorityTopRows = useMemo(() => filteredRows.slice(0, 5), [filteredRows]);

  const maxCapital = useMemo(
    () => Math.max(1, ...filteredRows.map((row) => Math.max(0, Number(row.capitalBrl || 0)))),
    [filteredRows],
  );

  return (
    <section className={styles.capitalSpotlightBlock}>
      <div className={styles.capitalSpotlightToolbar}>
        <label className={styles.capitalSpotlightField}>
          <span>Buscar grupo</span>
          <input
            type="text"
            value={searchQuery}
            onChange={(event) => onSearchQueryChange(event.target.value)}
            placeholder="Ex.: Assentos, Metais, Loucas..."
          />
        </label>
        <label className={styles.capitalSpotlightField}>
          <span>Marca</span>
          <SelectMenu
            id="capital-efficiency-brand-filter"
            mode="multi"
            value=""
            values={brandFilter}
            options={[{ label: "Todas Marcas", value: "all" }, ...brandOptions.map((item) => ({ label: item, value: item }))] as SelectMenuOption[]}
            onSelect={(value) => onBrandFilterChange(toggleMultiValue(brandFilter, value))}
            classNames={FILTER_SELECT_CLASSNAMES}
          />
        </label>
        <label className={styles.capitalSpotlightField}>
          <span>Risco</span>
          <select
            value={riskFilter}
            onChange={(event) => setRiskFilter(event.target.value as "all" | "high" | "medium" | "low")}
          >
            <option value="all">Todos</option>
            <option value="high">Alto</option>
            <option value="medium">Medio</option>
            <option value="low">Baixo</option>
          </select>
        </label>
        <label className={styles.capitalSpotlightField}>
          <span>Capital (R$)</span>
          <div className={styles.capitalSpotlightInline}>
            <select value={capitalOp} onChange={(event) => setCapitalOp(event.target.value as "gte" | "eq" | "lte")}>
              <option value="gte">&gt;=</option>
              <option value="eq">=</option>
              <option value="lte">&lt;=</option>
            </select>
            <input
              type="text"
              inputMode="decimal"
              value={capitalValue}
              onChange={(event) => setCapitalValue(event.target.value)}
              placeholder="Ex.: 50000"
            />
          </div>
        </label>
        <label className={styles.capitalSpotlightField}>
          <span>GMROI</span>
          <div className={styles.capitalSpotlightInline}>
            <select value={gmroiOp} onChange={(event) => setGmroiOp(event.target.value as "gte" | "eq" | "lte")}>
              <option value="lte">&lt;=</option>
              <option value="eq">=</option>
              <option value="gte">&gt;=</option>
            </select>
            <input
              type="text"
              inputMode="decimal"
              value={gmroiValue}
              onChange={(event) => setGmroiValue(event.target.value)}
              placeholder="Ex.: 1.5"
            />
          </div>
        </label>
        <label className={styles.capitalSpotlightField}>
          <span>Ordenar</span>
          <select value={sortBy} onChange={(event) => setSortBy(event.target.value as "priority" | "gmroi" | "capital")}>
            <option value="priority">Prioridade</option>
            <option value="gmroi">Pior GMROI</option>
            <option value="capital">Maior Capital</option>
          </select>
        </label>
      </div>

      {loading ? <div className={styles.capitalSpotlightState}>Carregando eficiencia de capital...</div> : null}
      {!loading && error ? <div className={styles.capitalSpotlightState}>{error}</div> : null}
      {!loading && !error && filteredRows.length === 0 ? (
        <div className={styles.capitalSpotlightState}>Sem dados para os filtros selecionados.</div>
      ) : null}

      {!loading && !error && filteredRows.length > 0 ? (
        <>
          <div className={styles.capitalSpotlightChart}>
            {filteredRows.slice(0, 24).map((row) => {
              const widthPct = Math.max(3, Math.min(100, (Math.max(0, row.capitalBrl) / maxCapital) * 100));
              const toneClass =
                row.riskLevel === "high"
                  ? styles.capitalSpotlightBarHigh
                  : row.riskLevel === "medium"
                    ? styles.capitalSpotlightBarMedium
                    : styles.capitalSpotlightBarLow;
              return (
                <div key={`cap-${row.nodeId}`} className={styles.capitalSpotlightBarRow}>
                  <div className={styles.capitalSpotlightBarLabel}>
                    <strong>{row.nodeName}</strong>
                    <span>GMROI {row.gmroi == null ? "-" : asNumber(row.gmroi)} | Risco {asPct(row.riskPct)}</span>
                  </div>
                  <div className={styles.capitalSpotlightBarTrack}>
                    <div className={`${styles.capitalSpotlightBarFill} ${toneClass}`} style={{ width: `${widthPct}%` }} />
                  </div>
                  <div className={styles.capitalSpotlightBarValue}>{asCurrency(row.capitalBrl)}</div>
                </div>
              );
            })}
          </div>

          <div className={styles.capitalSpotlightPriority}>
            <h4>Top prioridades (eficiência)</h4>
            <div className={styles.capitalSpotlightPriorityList}>
              {priorityTopRows.map((row, index) => (
                <div key={`cap-priority-${row.nodeId}`} className={styles.capitalSpotlightPriorityItem}>
                  <span className={styles.capitalSpotlightPriorityRank}>{index + 1}</span>
                  <span className={styles.capitalSpotlightPriorityName}>{row.nodeName}</span>
                  <span className={styles.capitalSpotlightPriorityMeta}>
                    {asCurrency(row.capitalBrl)} | GMROI {row.gmroi == null ? "-" : asNumber(row.gmroi)}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className={styles.capitalSpotlightTableWrap}>
            <table className={styles.capitalSpotlightTable}>
              <thead>
                <tr>
                  <th>Grupo</th>
                  <th>Capital</th>
                  <th>Receita</th>
                  <th>Margem</th>
                  <th>GMROI</th>
                  <th>Risco</th>
                </tr>
              </thead>
              <tbody>
                {filteredRows.slice(0, 60).map((row) => (
                  <tr key={`cap-row-${row.nodeId}`}>
                    <td>{row.nodeName}</td>
                    <td>{asCurrency(row.capitalBrl)}</td>
                    <td>{asCurrency(row.revenueBrl)}</td>
                    <td>{asCurrency(row.marginBrl)}</td>
                    <td>{row.gmroi == null ? "-" : asNumber(row.gmroi)}</td>
                    <td>{asPct(row.riskPct)}</td>
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

