---
id: sinapse-analytics-routing
title: Analytics Routing & Orchestration
region: sinapses
tags: [analytics, orchestration, $ms, $analytics-orchestrator, cross-cutting]
links:
  - cortex/backend/index
  - hippocampus/conventions
weight: 0.85
updated_at: 2026-03-24T10:00:00Z
---

# Analytics Routing & Orchestration

How the analytics domain is routed and orchestrated within MetalShopping's task system.

## Routing Decision

All tasks in MetalShopping are routed through `$ms` (master orchestrator). The router checks:

```
Is the task related to analytics (packages/feature-analytics/,
                                   apps/analytics_worker,
                                   cortex/database read models)?

  ↓ YES
  Route to $analytics-orchestrator

  ↓ NO
  Route to standard implementation skill
```

**Rule:** The orchestrator never bypass `$analytics-orchestrator` directly. All analytics work goes through the coordinator.

## Analytics Orchestrator Responsibilities

The analytics orchestrator handles:

1. **Feature-level tasks** — Changes to Analytics surfaces (Home, Products, Taxonomy, Brands)
2. **Worker-level tasks** — Python async compute, event scoring, projections
3. **Read model tasks** — Denormalized view updates, cache invalidation
4. **Contract tasks** — Analytics event schemas, governance definitions
5. **Cross-domain tasks** — How analytics integrates with procurement, CRM, strategy modules

## Task Breakdown by Analytics Subdomain

### 1. Analytics Surfaces (Frontend Features)

All UI/UX work related to analytics dashboards and reports.

```
Example tasks:
- "Add a new chart to the Products surface showing SKU performance"
- "Fix the hover behavior in the Home insights strip"
- "Migrate the legacy Brands surface to the new component library"
```

Routed to: `metalshopping-design-system` (or analytics sub-agent once staffed)

### 2. Analytics Worker (Python Async Compute)

Background processing that calculates metrics, aggregates data, updates read models.

```
Example tasks:
- "Implement daily revenue rollup in the worker"
- "Add tenant isolation to the product scoring algorithm"
- "Write event handlers that populate the products_analytics materialized view"
```

Routed to: `metalshopping-implement` (Go/Python modules)

### 3. Analytics Governance (Feature Flags, Thresholds)

Runtime policies that control feature rollouts and calculation parameters.

```
Example tasks:
- "Add a feature flag for the new SKU performance calculation"
- "Increase the anomaly detection threshold from 2.5σ to 3.0σ"
```

Routed to: `metalshopping-governance-contracts`

### 4. Analytics Events

Domain events that flow from procurement/products → analytics pipeline.

```
Example tasks:
- "Define the ProductSoldEvent schema with unit_price, margin_percent"
- "Create an OrderConfirmedEvent and wire it to the worker"
```

Routed to: `metalshopping-event-contracts`

## Orchestrator Decision Tree

```
/ms "Add new metric to analytics dashboard"
  ↓
Is it a visual/component task?
  → metalshopping-design-system (layout, styling, component hierarchy)

Is it a Go/Python implementation task?
  → Check: Does it need background compute?
    → metalshopping-implement (async worker code)
    → Check: Does it update read models?
      → cortex/database (schema additions)

Is it a contract/governance task?
  → Check: Event schema or policy?
    → metalshopping-event-contracts
    → metalshopping-governance-contracts

Does it span multiple domains (e.g., "Products module should feed analytics with margin data")?
  → Dispatch as multi-phase:
    1. Define event schema (event-contracts)
    2. Wire event emission in products module (implement)
    3. Wire event handler in analytics worker (implement)
    4. Update analytics view (database)
    5. Display in UI (design-system)
```

## Example: Multi-Phase Analytics Task

**Task:** "Add product margin to the Products analytics surface."

```
T1 — Define contract:
  Create ProductSoldEvent schema in contracts/events/ with margin_percent field
  Skill: $metalshopping-event-contracts

T2 — Implement emission:
  Wire ProductSoldEvent emission in products.application when an order is placed
  Skill: $metalshopping-implement

T3 — Implement handler:
  Create OnProductSold event handler in analytics_worker
  Updates products_analytics materialized view with margin data
  Skill: $metalshopping-implement

T4 — Update read model:
  Add margin_percent column to products_analytics materialized view
  Add index on margin_percent for sorting/filtering
  Skill: $metalshopping-implement (with $cortex/database consultation)

T5 — Display in UI:
  Add margin column to Products surface table
  Implement margin sparkline in insights strip
  Skill: $metalshopping-design-system

T6 — Test & verify:
  Integration test: order placed → event → worker processes → analytics updated
  Skill: test verification

T7 — Documentation:
  Update docs/PROGRESS.md and architecture docs
  Skill: $metalshopping-docs
```

## No Bypass Rule

| Scenario | What Happens | Why |
|----------|------|---|
| Task mentions "analytics" | Routes to `$analytics-orchestrator` | Ensures domain expertise |
| Developer asks for a shortcut | Still routes to orchestrator (may be faster, but through proper channels) | Maintains consistency, prevents cross-tenant issues, ensures testing |
| One-off metric calculation | Still goes through orchestrator if it touches read models/events | Prevents temporary code from staying permanent |

## Integration Points

Analytics depends on these domains:

| Domain | What It Provides | Example |
|--------|---------|---------|
| **Procurement** | Order events, supplier cost changes | OnOrderCreated, OnSupplierCostUpdated |
| **Products** | Product metadata, pricing changes | OnProductPriceChanged, OnProductArchived |
| **CRM** | Customer segments, account changes | OnAccountCreated, OnCustomerSegmentUpdated |
| **Strategy** | Pricing rules, promotion flags | OnPricingRuleChanged |

All these integrations are event-driven (outbox pattern). No direct database access between domains.

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.85
