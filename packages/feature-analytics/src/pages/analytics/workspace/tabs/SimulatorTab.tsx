import { useEffect, useMemo, useState, type KeyboardEvent } from "react";
import { useLocation, useOutletContext } from "react-router-dom";

import { InfoTooltipLabel } from "../components/InfoTooltipLabel";
import type { ProductWorkspaceOutletContext } from "../ProductWorkspaceLayout";
import styles from "./simulator.module.css";
import { SIMULATOR_HELP } from "./simulator_help_texts";
import {
  calculateSimulation,
  defaultSimulationInputs,
  deriveDiscountFromPrice,
  deriveMarginFromPrice,
  derivePriceBounds,
  derivePriceFromDiscount,
  derivePriceFromMargin,
  type CurrentSnapshot,
  type IndicatorTone,
  type SimulationInputs,
} from "./simulator/math";

type EditedField = "price" | "discount" | "margin";
type CostMode = "cost_avg" | "cost_variable";

function normalizeToken(value: string): string {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function parseLocaleNumber(raw: string): number | null {
  const cleaned = String(raw || "")
    .replace(/[^\d,.\-]/g, "")
    .replace(/\.(?=\d{3}\b)/g, "")
    .replace(",", ".");
  const value = Number(cleaned);
  return Number.isFinite(value) ? value : null;
}

function parseNavNumber(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") return parseLocaleNumber(value);
  return null;
}

function pickMetricValue(metrics: Array<{ label: string; value: string }>, needle: string): string {
  const token = normalizeToken(needle);
  const found = metrics.find((item) => normalizeToken(item.label).includes(token));
  return found?.value ?? "";
}

function pickMetricValueText(metrics: Array<{ label: string; valueText: string }>, needle: string): string {
  const token = normalizeToken(needle);
  const found = metrics.find((item) => normalizeToken(item.label).includes(token));
  return found?.valueText ?? "";
}

function toCurrentSnapshot(model: ProductWorkspaceOutletContext["model"]): CurrentSnapshot {
  const hero = model.heroMetrics;
  const priceCurrent = parseLocaleNumber(pickMetricValue(hero, "preco (nosso)"));
  const marketAvg = parseLocaleNumber(pickMetricValue(hero, "mercado medio"));
  const marketMin = parseLocaleNumber(pickMetricValue(hero, "mercado min"));
  const marketMax = parseLocaleNumber(pickMetricValue(hero, "mercado max"));
  const priceRealEffective =
    parseLocaleNumber(pickMetricValue(hero, "preco real efetivo")) ??
    parseLocaleNumber(pickMetricValue(hero, "preco real")) ??
    priceCurrent;
  const costAvg = parseLocaleNumber(pickMetricValue(hero, "custo medio"));
  const costVariable = parseLocaleNumber(pickMetricValue(hero, "custo variavel"));
  const marginPctCurrent = parseLocaleNumber(model.profitability.metrics[0]?.value || "");
  const markupPctCurrent = parseLocaleNumber(model.profitability.metrics[1]?.value || "");
  const gmroiCurrent = parseLocaleNumber(model.profitability.lower[0]?.value || "");
  const gapPctCurrent = parseLocaleNumber(pickMetricValueText(model.competitiveness.metrics, "gap vs mercado"));
  const simulator = model.simulator;
  const variableCostUnitAuto = simulator?.variable_cost_unit_auto ?? null;
  const variableSpendUnit =
    parseLocaleNumber(pickMetricValue(hero, "gasto var")) ??
    parseLocaleNumber(pickMetricValue(hero, "gasto variavel")) ??
    variableCostUnitAuto;
  const variableCostSource = simulator?.variable_cost_source ?? "NONE_GROSS_FALLBACK";
  const variableCostCoverageMonths = simulator?.variable_cost_coverage_months ?? 0;
  const marginCalcMode = simulator?.margin_calc_mode ?? "GROSS_FALLBACK_NO_GV";
  const marginIsFallback = Boolean(simulator?.margin_is_fallback ?? true);

  return {
    pn: model.pn,
    title: model.title,
    brand: model.brand,
    taxonomyLeafName: model.taxonomyLeafName,
    priceCurrent,
    priceRealEffective,
    costAvg,
    costVariable,
    variableSpendUnit,
    variableCostUnitAuto,
    variableCostSource,
    variableCostCoverageMonths,
    marginCalcMode,
    marginIsFallback,
    marketAvg,
    marketMin,
    marketMax,
    marginPctCurrent,
    markupPctCurrent,
    gmroiCurrent,
    gapPctCurrent,
  };
}

function formatCurrency(value: number | null): string {
  if (value == null || !Number.isFinite(value)) return "-";
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL", maximumFractionDigits: 2 }).format(value);
}

function formatPct(value: number | null, decimals = 1): string {
  if (value == null || !Number.isFinite(value)) return "-";
  return `${value.toFixed(decimals)}%`;
}

function formatNumber(value: number | null, decimals = 2): string {
  if (value == null || !Number.isFinite(value)) return "-";
  return value.toFixed(decimals);
}

function formatDraftDecimal(value: number | null, decimals = 2): string {
  if (value == null || !Number.isFinite(value)) return "";
  return value.toFixed(decimals).replace(".", ",");
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}

function toPercentWithin(value: number | null, min: number | null, max: number | null): number {
  if (value == null || min == null || max == null || max <= min) return 0;
  return clamp(((value - min) / (max - min)) * 100, 0, 100);
}

function toneClass(tone: IndicatorTone): string {
  if (tone === "good") return styles.toneGood;
  if (tone === "warn") return styles.toneWarn;
  if (tone === "bad") return styles.toneBad;
  return styles.toneNeutral;
}

function syncFromEditedField(
  current: CurrentSnapshot,
  prev: SimulationInputs,
  field: EditedField,
  rawValue: number | null,
): SimulationInputs {
  const freightAdjPct = prev.freightAdjPct;
  const bounds = derivePriceBounds(current);
  if (field === "price") {
    const price = rawValue == null ? null : clamp(rawValue, bounds.min, bounds.max);
    return {
      ...prev,
      price,
      discountPct: deriveDiscountFromPrice(current.priceCurrent, price),
      targetMarginPct: deriveMarginFromPrice(
        current.costAvg,
        current.variableCostUnitAuto,
        freightAdjPct,
        price,
      ),
    };
  }

  if (field === "discount") {
    const discountPct = rawValue == null ? null : clamp(rawValue, 0, 90);
    const price = derivePriceFromDiscount(current.priceCurrent, discountPct);
    return {
      ...prev,
      price,
      discountPct,
      targetMarginPct: deriveMarginFromPrice(
        current.costAvg,
        current.variableCostUnitAuto,
        freightAdjPct,
        price,
      ),
    };
  }

  const margin = rawValue == null ? null : clamp(rawValue, 0, 90);
  const price = derivePriceFromMargin(
    current.costAvg,
    current.variableCostUnitAuto,
    freightAdjPct,
    margin,
  );
  return {
    ...prev,
    price,
    discountPct: deriveDiscountFromPrice(current.priceCurrent, price),
    targetMarginPct: margin,
  };
}

export function SimulatorTab() {
  const { model } = useOutletContext<ProductWorkspaceOutletContext>();
  const location = useLocation();
  const current = useMemo(() => toCurrentSnapshot(model), [model]);
  const costAvgDate = String(model.simulator?.cost_avg_date || "").trim() || null;
  const [costMode, setCostMode] = useState<CostMode>("cost_avg");
  const selectedBaseCost = useMemo(() => {
    if (costMode === "cost_variable") {
      return current.costVariable != null ? current.costVariable : current.costAvg;
    }
    return current.costAvg;
  }, [costMode, current.costAvg, current.costVariable]);
  const currentForCalc = useMemo(
    () => ({
      ...current,
      costAvg: selectedBaseCost,
    }),
    [current, selectedBaseCost],
  );
  const navPriceTarget = useMemo(
    () => parseNavNumber((location.state as { priceTarget?: unknown } | null)?.priceTarget),
    [location.state],
  );
  const [inputs, setInputs] = useState<SimulationInputs>(() => defaultSimulationInputs(currentForCalc));
  const [draftPrice, setDraftPrice] = useState("");
  const [draftDiscount, setDraftDiscount] = useState("");
  const [draftMargin, setDraftMargin] = useState("");
  const [draftFreight, setDraftFreight] = useState("");

  useEffect(() => {
    const base = defaultSimulationInputs(currentForCalc);
    if (navPriceTarget != null) {
      setInputs(syncFromEditedField(currentForCalc, base, "price", navPriceTarget));
      return;
    }
    setInputs(base);
  }, [currentForCalc, navPriceTarget]);

  useEffect(() => {
    setDraftPrice(formatDraftDecimal(inputs.price));
    setDraftDiscount(formatDraftDecimal(inputs.discountPct));
    setDraftMargin(formatDraftDecimal(inputs.targetMarginPct));
    setDraftFreight(formatDraftDecimal(inputs.freightAdjPct));
  }, [inputs.price, inputs.discountPct, inputs.targetMarginPct, inputs.freightAdjPct]);

  const result = useMemo(() => calculateSimulation(currentForCalc, inputs), [currentForCalc, inputs]);
  const priceBounds = useMemo(() => derivePriceBounds(currentForCalc), [currentForCalc]);

  const priceFillPct = toPercentWithin(inputs.price, priceBounds.min, priceBounds.max);
  const discountFillPct = clamp(inputs.discountPct ?? 0, 0, 90) / 90 * 100;
  const marginFillPct = clamp(inputs.targetMarginPct ?? 0, 0, 60) / 60 * 100;

  const minMarkerPct = toPercentWithin(current.marketMin, priceBounds.min, priceBounds.max);
  const maxMarkerPct = toPercentWithin(current.marketMax, priceBounds.min, priceBounds.max);
  const avgMarkerPct = toPercentWithin(current.marketAvg, priceBounds.min, priceBounds.max);
  const currentMarkerPct = toPercentWithin(current.priceCurrent, priceBounds.min, priceBounds.max);
  const simMarkerPct = toPercentWithin(result.priceNew, priceBounds.min, priceBounds.max);

  const deltaMarginPp =
    result.marginPct != null && current.marginPctCurrent != null
      ? result.marginPct - current.marginPctCurrent
      : null;
  const deltaGapPp =
    result.gapPct != null && current.gapPctCurrent != null
      ? result.gapPct - current.gapPctCurrent
      : null;
  const currentContrib =
    selectedBaseCost != null && current.priceCurrent != null
      ? current.priceCurrent - selectedBaseCost
      : null;
  const deltaContrib =
    result.marginAmount != null && currentContrib != null
      ? result.marginAmount - currentContrib
      : null;

  function updateFreightAdj(deltaOrAbsolute: number, absolute = false) {
    setInputs((prev) => {
      const nextFreight = absolute ? clamp(deltaOrAbsolute, 0, 60) : clamp(prev.freightAdjPct + deltaOrAbsolute, 0, 60);
      const nextMargin = deriveMarginFromPrice(
        currentForCalc.costAvg,
        currentForCalc.variableCostUnitAuto,
        nextFreight,
        prev.price,
      );
      return {
        ...prev,
        freightAdjPct: nextFreight,
        targetMarginPct: nextMargin,
      };
    });
  }

  function adjustField(field: EditedField, delta: number) {
    const currentValue = field === "price" ? inputs.price : field === "discount" ? inputs.discountPct : inputs.targetMarginPct;
    const next = (currentValue ?? 0) + delta;
    setInputs((prev) => syncFromEditedField(currentForCalc, prev, field, next));
  }

  function resetSimulation() {
    setInputs(defaultSimulationInputs(currentForCalc));
  }

  function commitPriceDraft() {
    const parsed = parseLocaleNumber(draftPrice);
    if (parsed == null) {
      setDraftPrice(formatDraftDecimal(inputs.price));
      return;
    }
    setInputs((prev) => syncFromEditedField(currentForCalc, prev, "price", parsed));
  }

  function commitDiscountDraft() {
    const parsed = parseLocaleNumber(draftDiscount);
    if (parsed == null) {
      setDraftDiscount(formatDraftDecimal(inputs.discountPct));
      return;
    }
    setInputs((prev) => syncFromEditedField(currentForCalc, prev, "discount", parsed));
  }

  function commitMarginDraft() {
    const parsed = parseLocaleNumber(draftMargin);
    if (parsed == null) {
      setDraftMargin(formatDraftDecimal(inputs.targetMarginPct));
      return;
    }
    setInputs((prev) => syncFromEditedField(currentForCalc, prev, "margin", parsed));
  }

  function commitFreightDraft() {
    const parsed = parseLocaleNumber(draftFreight);
    if (parsed == null) {
      setDraftFreight(formatDraftDecimal(inputs.freightAdjPct));
      return;
    }
    updateFreightAdj(parsed, true);
  }

  function handleDraftKeyDown(
    event: KeyboardEvent<HTMLInputElement>,
    commit: () => void,
    reset: () => void,
  ) {
    if (event.key === "Enter") {
      event.preventDefault();
      commit();
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      reset();
    }
  }

  return (
    <section className={styles.simRoot}>
      <header className={styles.hero}>
        <h2 className={styles.heroTitle}>💰 Simulador de Preco</h2>
        <p className={styles.heroMeta}>
          Preco simulado <span>{formatCurrency(result.priceNew)}</span> | Desconto <span>{formatPct(result.discountPct)}</span> | Margem <span>{formatPct(result.marginPct)}</span>
        </p>
        <div className={styles.heroMetrics}>
          <div>
            <span className={styles.heroMetricLabel}>Nosso preco</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.priceCurrent)}</span>
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Mercado med</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.marketAvg)}</span>
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Mercado min</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.marketMin)}</span>
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Mercado max</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.marketMax)}</span>
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Custo medio</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.costAvg)}</span>
            {costAvgDate ? <span className={styles.heroMetricMeta}>DT_COMPRA: {costAvgDate}</span> : null}
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Custo variavel</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.costVariable)}</span>
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Gasto var.</span>
            <span className={styles.heroMetricValue}>{formatCurrency(current.variableSpendUnit)}</span>
          </div>
          <div>
            <span className={styles.heroMetricLabel}>Modo margem</span>
            <span className={styles.heroMetricValue}>
              {current.variableCostSource !== "NONE_GROSS_FALLBACK"
                ? `Real (${current.variableCostCoverageMonths}m)`
                : "Bruta (fallback)"}
            </span>
          </div>
        </div>
        <button type="button" className={styles.resetBtn} onClick={resetSimulation}>↻ Resetar simulacao</button>
      </header>

      <section className={styles.simGrid}>
        <article className={styles.glassCard}>
          <h3 className={styles.cardTitle}>🎛️ Controles</h3>

          <div className={styles.sliderGroup}>
            <div className={styles.sliderHeader}>
              <span className={styles.sliderLabel}>
                <InfoTooltipLabel label="Preco (R$)" help={SIMULATOR_HELP["preco-rs"]} />
              </span>
              <div className={styles.sliderControls}>
                <button type="button" className={styles.sliderBtn} onClick={() => adjustField("price", -10)}>−</button>
                <input
                  className={styles.sliderValueInput}
                  type="text"
                  inputMode="decimal"
                  value={draftPrice}
                  onChange={(event) => setDraftPrice(event.target.value)}
                  onKeyDown={(event) =>
                    handleDraftKeyDown(
                      event,
                      commitPriceDraft,
                      () => setDraftPrice(formatDraftDecimal(inputs.price)),
                    )}
                />
                <button type="button" className={styles.sliderBtn} onClick={() => adjustField("price", 10)}>+</button>
              </div>
            </div>
            <div className={styles.sliderTrack}>
              <div className={styles.sliderFill} style={{ width: `${priceFillPct}%` }} />
              <span className={styles.sliderKnob} style={{ left: `${priceFillPct}%` }} />
              <input
                className={styles.sliderInput}
                type="range"
                min={priceBounds.min}
                max={priceBounds.max}
                step={1}
                value={inputs.price ?? priceBounds.min}
                onChange={(event) => setInputs((prev) => syncFromEditedField(currentForCalc, prev, "price", Number(event.target.value)))}
              />
            </div>
            <div className={styles.sliderMarkers}>
              <span>{formatCurrency(priceBounds.min)}</span>
              <span>{formatCurrency((priceBounds.min + priceBounds.max) / 2)}</span>
              <span>{formatCurrency(priceBounds.max)}</span>
            </div>
          </div>

          <div className={styles.sliderGroup}>
            <div className={styles.sliderHeader}>
              <span className={styles.sliderLabel}>
                <InfoTooltipLabel label="Desconto (%)" help={SIMULATOR_HELP["desconto-pct"]} />
              </span>
              <div className={styles.sliderControls}>
                <button type="button" className={styles.sliderBtn} onClick={() => adjustField("discount", -0.5)}>−</button>
                <input
                  className={styles.sliderValueInput}
                  type="text"
                  inputMode="decimal"
                  value={draftDiscount}
                  onChange={(event) => setDraftDiscount(event.target.value)}
                  onKeyDown={(event) =>
                    handleDraftKeyDown(
                      event,
                      commitDiscountDraft,
                      () => setDraftDiscount(formatDraftDecimal(inputs.discountPct)),
                    )}
                />
                <button type="button" className={styles.sliderBtn} onClick={() => adjustField("discount", 0.5)}>+</button>
              </div>
            </div>
            <div className={styles.sliderTrack}>
              <div className={styles.sliderFill} style={{ width: `${discountFillPct}%` }} />
              <span className={styles.sliderKnob} style={{ left: `${discountFillPct}%` }} />
              <input
                className={styles.sliderInput}
                type="range"
                min={0}
                max={90}
                step={0.5}
                value={inputs.discountPct ?? 0}
                onChange={(event) => setInputs((prev) => syncFromEditedField(currentForCalc, prev, "discount", Number(event.target.value)))}
              />
            </div>
            <div className={styles.sliderMarkers}>
              <span>0%</span>
              <span>45%</span>
              <span>90%</span>
            </div>
          </div>

          <div className={styles.sliderGroup}>
            <div className={styles.sliderHeader}>
              <span className={styles.sliderLabel}>
                <InfoTooltipLabel label="Margem de contribuicao alvo (%)" help={SIMULATOR_HELP["margem-alvo-pct"]} />
              </span>
              <div className={styles.sliderControls}>
                <button type="button" className={styles.sliderBtn} onClick={() => adjustField("margin", -0.5)}>−</button>
                <input
                  className={styles.sliderValueInput}
                  type="text"
                  inputMode="decimal"
                  value={draftMargin}
                  onChange={(event) => setDraftMargin(event.target.value)}
                  onKeyDown={(event) =>
                    handleDraftKeyDown(
                      event,
                      commitMarginDraft,
                      () => setDraftMargin(formatDraftDecimal(inputs.targetMarginPct)),
                    )}
                />
                <button type="button" className={styles.sliderBtn} onClick={() => adjustField("margin", 0.5)}>+</button>
              </div>
            </div>
            <div className={styles.sliderTrack}>
              <div className={styles.sliderFill} style={{ width: `${marginFillPct}%` }} />
              <span className={styles.sliderKnob} style={{ left: `${marginFillPct}%` }} />
              <input
                className={styles.sliderInput}
                type="range"
                min={0}
                max={60}
                step={0.5}
                value={inputs.targetMarginPct ?? 0}
                onChange={(event) => setInputs((prev) => syncFromEditedField(currentForCalc, prev, "margin", Number(event.target.value)))}
              />
            </div>
            <div className={styles.sliderMarkers}>
              <span>0%</span>
              <span>30%</span>
              <span>60%</span>
            </div>
          </div>

          <div className={styles.sliderGroup}>
            <div className={styles.sliderHeader}>
              <span className={styles.sliderLabel}>
                <InfoTooltipLabel label="Fretes/Encargos (%)" help={SIMULATOR_HELP["frete-encargos-pct"]} />
              </span>
              <div className={styles.sliderControls}>
                <button type="button" className={styles.sliderBtn} onClick={() => updateFreightAdj(-0.5)}>−</button>
                <input
                  className={styles.sliderValueInput}
                  type="text"
                  inputMode="decimal"
                  value={draftFreight}
                  onChange={(event) => setDraftFreight(event.target.value)}
                  onKeyDown={(event) =>
                    handleDraftKeyDown(
                      event,
                      commitFreightDraft,
                      () => setDraftFreight(formatDraftDecimal(inputs.freightAdjPct)),
                    )}
                />
                <button type="button" className={styles.sliderBtn} onClick={() => updateFreightAdj(0.5)}>+</button>
              </div>
            </div>
            <div className={styles.sliderTrack}>
              <div className={styles.sliderFill} style={{ width: `${clamp(inputs.freightAdjPct, 0, 60) / 60 * 100}%` }} />
              <span className={styles.sliderKnob} style={{ left: `${clamp(inputs.freightAdjPct, 0, 60) / 60 * 100}%` }} />
              <input
                className={styles.sliderInput}
                type="range"
                min={0}
                max={60}
                step={0.5}
                value={inputs.freightAdjPct}
                onChange={(event) => updateFreightAdj(Number(event.target.value), true)}
              />
            </div>
            <div className={styles.sliderMarkers}>
              <span>0%</span>
              <span>30%</span>
              <span>60%</span>
            </div>
          </div>

          <div className={styles.priceRangeSection}>
            <h4 className={styles.rangeTitle}>
              <InfoTooltipLabel label="Posicao no Range" help={SIMULATOR_HELP["posicao-no-range"]} />
            </h4>
            <div className={styles.rangePinsOutside}>
              <span className={`${styles.rangePinLabel} ${styles.rangePinLabelSim}`} style={{ left: `${simMarkerPct}%` }}>Simulado</span>
              <span className={`${styles.rangePinLabel} ${styles.rangePinLabelCurrent}`} style={{ left: `${currentMarkerPct}%` }}>Atual</span>
            </div>
            <div className={styles.priceRangeTrack}>
              <div className={`${styles.rangePin} ${styles.pinMin}`} style={{ left: `${minMarkerPct}%` }}><span>min</span></div>
              <div className={`${styles.rangePin} ${styles.pinAvg}`} style={{ left: `${avgMarkerPct}%` }}><span>med</span></div>
              <div className={`${styles.rangePin} ${styles.pinMax}`} style={{ left: `${maxMarkerPct}%` }}><span>max</span></div>
              <div className={`${styles.rangePin} ${styles.pinCurrent}`} style={{ left: `${currentMarkerPct}%` }}><span>atual</span></div>
              <div className={`${styles.rangePin} ${styles.pinSim}`} style={{ left: `${simMarkerPct}%` }}><span>simulado</span></div>
            </div>
            <div className={styles.rangeMarkers}>
              <span className={`${styles.rangeMarkerItem} ${styles.rangeMarkerStart}`} style={{ left: `${minMarkerPct}%` }}>min</span>
              <span className={`${styles.rangeMarkerItem} ${styles.rangeMarkerCenter}`} style={{ left: `${avgMarkerPct}%` }}>med</span>
              <span className={`${styles.rangeMarkerItem} ${styles.rangeMarkerEnd}`} style={{ left: `${maxMarkerPct}%` }}>max</span>
            </div>
            <div className={styles.rangeValues}>
              <span className={`${styles.rangeValueItem} ${styles.rangeMarkerStart}`} style={{ left: `${minMarkerPct}%` }}>{formatCurrency(current.marketMin)}</span>
              <span className={`${styles.rangeValueItem} ${styles.rangeMarkerCenter}`} style={{ left: `${avgMarkerPct}%` }}>{formatCurrency(current.marketAvg)}</span>
              <span className={`${styles.rangeValueItem} ${styles.rangeMarkerEnd}`} style={{ left: `${maxMarkerPct}%` }}>{formatCurrency(current.marketMax)}</span>
            </div>
          </div>
        </article>

        <article className={styles.glassCard}>
          <div className={styles.cardTitleRow}>
            <h3 className={styles.cardTitle}>📊 Impacto da Decisao (ao vivo)</h3>
            <div className={styles.costModeSwitch}>
              <button
                type="button"
                className={`${styles.costModeBtn} ${costMode === "cost_avg" ? styles.costModeBtnActive : ""}`}
                onClick={() => setCostMode("cost_avg")}
              >
                Custo medio
              </button>
              <button
                type="button"
                className={`${styles.costModeBtn} ${costMode === "cost_variable" ? styles.costModeBtnActive : ""}`}
                onClick={() => setCostMode("cost_variable")}
                disabled={current.costVariable == null}
              >
                Custo variavel
              </button>
            </div>
          </div>

          <div className={styles.impactGrid}>
            <div className={styles.metricBox}>
              <div className={styles.metricLabel}>
                <InfoTooltipLabel label="Markup simulado (%)" help={SIMULATOR_HELP["markup-simulado-pct"]} />
              </div>
              <div className={`${styles.metricValue} ${toneClass(result.tones.markup)}`}>{formatPct(result.markupPct, 2)}</div>
              <div className={styles.metricBar}><div className={styles.metricFill} style={{ width: `${clamp(result.markupPct ?? 0, 0, 100)}%` }} /></div>
            </div>
            <div className={styles.metricBox}>
              <div className={styles.metricLabel}>
                <InfoTooltipLabel label="Margem simulada (%)" help={SIMULATOR_HELP["margem-simulada-pct"]} />
              </div>
              <div className={`${styles.metricValue} ${toneClass(result.tones.margin)}`}>{formatPct(result.marginPct, 2)}</div>
              <div className={styles.metricBar}><div className={styles.metricFill} style={{ width: `${clamp(result.marginPct ?? 0, 0, 100)}%` }} /></div>
            </div>
            <div className={styles.metricBox}>
              <div className={styles.metricLabel}>
                <InfoTooltipLabel label="Contribuicao por unidade" help={SIMULATOR_HELP["contribuicao-unidade"]} />
              </div>
              <div className={styles.metricValue}>{formatCurrency(result.marginAmount)}</div>
              <div className={styles.metricBar}><div className={styles.metricFillInfo} style={{ width: `${clamp((result.marginAmount ?? 0) / 10, 0, 100)}%` }} /></div>
            </div>
            <div className={styles.metricBox}>
              <div className={styles.metricLabel}>
                <InfoTooltipLabel label="Novo preco" help={SIMULATOR_HELP["novo-preco"]} />
              </div>
              <div className={styles.metricValue}>{formatCurrency(result.priceNew)}</div>
              <div className={styles.metricBar}><div className={styles.metricFillNeutral} style={{ width: `${clamp((result.marketPositionPct ?? 0), 0, 100)}%` }} /></div>
            </div>
            <div className={styles.metricBox}>
              <div className={styles.metricLabel}>
                <InfoTooltipLabel label="Gap vs mercado (%)" help={SIMULATOR_HELP["gap-mercado-pct"]} />
              </div>
              <div className={`${styles.metricValue} ${toneClass(result.tones.gap)}`}>{formatPct(result.gapPct, 2)}</div>
              <div className={styles.metricBar}><div className={styles.metricFillWarn} style={{ width: `${clamp(Math.abs(result.gapPct ?? 0) * 4, 0, 100)}%` }} /></div>
            </div>
            <div className={styles.metricBox}>
              <div className={styles.metricLabel}>
                <InfoTooltipLabel label="Novo custo" help={SIMULATOR_HELP["novo-custo"]} />
              </div>
              <div className={styles.metricValue}>{formatCurrency(result.effectiveCost)}</div>
              <div className={styles.metricBar}>
                <div
                  className={styles.metricFillWarn}
                  style={{
                    width: `${clamp(
                      result.effectiveCost != null && result.priceNew != null && result.priceNew > 0
                        ? (result.effectiveCost / result.priceNew) * 100
                        : 0,
                      0,
                      100,
                    )}%`,
                  }}
                />
              </div>
            </div>
          </div>

          <div className={styles.statusBadges}>
            {result.statusBadges.map((item) => (
              <span key={`${item.label}-${item.tone}`} className={`${styles.statusBadge} ${styles[`badge_${item.tone}`]}`}>{item.label}</span>
            ))}
          </div>

          <div className={styles.deltaSection}>
            <h4 className={styles.deltaTitle}>📈 Comparado ao preco atual</h4>
            <div className={styles.deltaGrid}>
              <article className={styles.deltaCard}>
                <span>Delta margem (pp)</span>
                <strong>{formatPct(deltaMarginPp, 2)}</strong>
              </article>
              <article className={styles.deltaCard}>
                <span>Delta contribuicao (R$)</span>
                <strong>{formatCurrency(deltaContrib)}</strong>
              </article>
              <article className={styles.deltaCard}>
                <span>Preco real efetivo</span>
                <strong>{formatCurrency(current.priceRealEffective)}</strong>
              </article>
              <article className={styles.deltaCard}>
                <span>Gasto var.</span>
                <strong>{formatCurrency(current.variableSpendUnit)}</strong>
              </article>
              <article className={styles.deltaCard}>
                <span>Gap real vs atual (pp)</span>
                <strong>{formatPct(deltaGapPp, 2)}</strong>
              </article>
            </div>
          </div>

          {result.alerts.length ? (
            <div className={styles.alertList}>
              {result.alerts.map((alert) => (
                <div key={`${alert.level}-${alert.title}`} className={`${styles.alertItem} ${styles[`alert_${alert.level}`]}`}>
                  <strong>{alert.title}</strong>
                  <span>{alert.message}</span>
                </div>
              ))}
            </div>
          ) : null}
        </article>
      </section>
    </section>
  );
}
