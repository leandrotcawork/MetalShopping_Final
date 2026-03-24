# Analytics Intelligence — Brainstorm Backlog

**Status:** Pending dedicated brainstorm session
**Priority:** High — most complex and strategically important module in MetalShopping
**Trigger:** Start this session after the developer workflow design is fully implemented

---

## Context

The existing `ANALYTICS_INTELLIGENCE_VISION.md` is a strong foundation but is not the final structure. This document captures what needs to be added, redesigned, and debated in the dedicated brainstorm session.

---

## Agent topology (already decided in workflow brainstorm)

The analytics domain is too complex for a single agent. Four specialized agents under one orchestrator:

| Agent | Owns |
|-------|------|
| `$analytics-orchestrator` | Routes to sub-agents, holds the full vision, decides which agent handles what |
| `$analytics-intelligence` | 8-layer brain: metrics, rules, calculations, classifications, recommendations, formulas, data quality |
| `$analytics-surfaces` | All 11 read surfaces, charts, workspaces, frontend patterns |
| `$analytics-campaigns` | Campaigns, actions, alerts, approval flows, routing rules |
| `$analytics-ai` | AI copilot: explanations, simulations, natural language, drafting |

`$ms` talks only to `$analytics-orchestrator`. Everything below is internal to the analytics domain.

---

## What the current vision document is missing

### 1. Feedback loop
When a recommendation is accepted, rejected, or ignored — the system should learn. Currently there is no feedback mechanism. Without it, the recommendation engine never improves and operators lose trust over time.

Questions to answer:
- How is acceptance/rejection captured?
- What adjusts as a result — thresholds, weights, confidence bands?
- How do we prevent gaming the feedback (operator always rejecting to avoid action)?

### 2. Predictive layer
Everything in the current document is **reactive** — what happened, what is wrong now. A supreme intelligence system needs:
- Demand forecasting (where is a SKU heading, not just where it is)
- Seasonal pattern detection and adjustment
- Leading indicators (early signals before metrics confirm a trend)
- Stockout probability before it happens, not after

Questions to answer:
- What forecasting methods fit our data maturity (ARIMA, exponential smoothing, ML)?
- How do we handle low-data SKUs where forecasting is unreliable?
- How does the predictive layer interact with the procurement suggestion engine?

### 3. Cross-entity correlation
The current document analyzes SKU, brand, and taxonomy **in isolation**. Real commercial intelligence requires:
- Cannibalization detection (brand A rising while brand B declines in same taxonomy node)
- Halo effects (premium SKU driving category traffic that converts to adjacent products)
- Substitution signals (when a competitor SKU goes out of stock, do we see demand transfer?)
- Portfolio concentration risk (too dependent on one brand, one taxonomy, one supplier)

Questions to answer:
- What is the minimum data volume needed for correlation to be meaningful?
- How do we surface cross-entity insights without overwhelming operators?

### 4. Supplier intelligence (underdesigned in current doc)
Supplier analytics is mentioned briefly but is commercially as important as product health. Should include:
- Supplier reliability scoring (on-time delivery, fill rate, quality issues)
- Lead-time variability tracking (not just average — the variance matters)
- Price negotiation intelligence (cost trend, market cost benchmarks, leverage points)
- Supplier concentration risk (what % of revenue depends on one supplier?)
- Supplier-brand alignment (which suppliers are gaining/losing strategic importance?)
- Suggested supplier switches based on reliability + cost trajectory

Questions to answer:
- Does supplier intelligence live in `analytics` or `procurement` domain?
- What data do we need from the `suppliers` module to feed this?

### 5. Real-time anomaly detection
The current document assumes **periodic batch computation**. Some signals need near-real-time detection:
- Sudden sales spike on a specific SKU (trending product, competitor stockout, viral moment)
- Abrupt price drop by a competitor (requires market price monitoring feed)
- Unexpected stockout velocity (stock depleting faster than the model predicted)
- Campaign performance deviation from expected trajectory

Questions to answer:
- What is the latency target for anomaly detection? (minutes, hours, next-day?)
- Does this require streaming infrastructure or is high-frequency batch sufficient for v1?
- How does this integrate with the alerts layer?

### 6. Opportunity radar (proactive, not reactive)
Everything in the current document is **health monitoring** — what is wrong. A supreme intelligence system also proactively detects opportunities:
- Whitespace in a growing taxonomy node where we have no SKUs
- A competitor going out of stock on a high-margin SKU we carry
- A brand gaining market momentum before competitors react
- A taxonomy branch with accelerating demand and our assortment is thin
- A supplier with excess stock willing to negotiate better terms

Questions to answer:
- How do we define and score "opportunity" vs. "health issue"?
- What external data feeds does the opportunity radar need (market prices, competitor availability)?
- How do opportunities surface differently from alerts (they are not problems — they are upside)?

---

## Architecture questions to resolve

- Where does the `analytics_worker` boundary end and `analytics_serving` begin for each intelligence layer?
- Which computations are real-time (serving) vs. batch (worker)?
- How does the AI copilot layer access governed data without bypassing domain ownership rules?
- What is the event contract between `analytics_worker` outputs and `analytics_serving` read models?
- How do campaign executions feed back into analytics as observed impact vs. expected impact?

---

## Vision statement to challenge

The current document's definition of success is good. Push it further:

> The analytics module is succeeding when the commercial team can run the entire day's decision-making — pricing, buying, campaigns, supplier negotiations, assortment changes — from a single system, with every recommendation backed by traceable evidence, and the system getting measurably better at predictions every month because it learns from what was accepted and what actually worked.

---

## Session start prompt (use this when opening the dedicated brainstorm)

> "Let's design the MetalShopping analytics intelligence system from scratch. We have an existing vision document at `docs/ANALYTICS_INTELLIGENCE_VISION.md` and a backlog of gaps at `docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md`. Read both before we start. The goal is a supreme commercial intelligence system — not a dashboard."
