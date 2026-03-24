# Analytics Intelligence Vision

## Status

- Type: product and architecture vision
- Scope: target operating model for the analytics module
- Binding level: directional until frozen by dedicated ADRs and contracts

## Purpose

Define what the MetalShopping analytics module is supposed to become: the commercial intelligence layer that tells the company where to act, what to sell, what to stop pushing, when to reduce or increase price, when to buy more, when to slow purchases, which brands and taxonomies deserve attention, which campaigns to run, which alerts matter, and how AI should help without breaking domain ownership.

This document is intentionally broader than the current frontend parity tranche. It explains the full target model while staying compatible with the current repo boundaries:

- `catalog` owns product identity, brand, and taxonomy
- `pricing` owns internal price and cost semantics
- `inventory` owns live stock position
- `procurement` owns supplier-side replenishment semantics
- analytics owns derived intelligence, scoring, projections, classifications, and recommendations

## What already exists as module language

Even before the full backend intelligence exists, the current analytics frontend already exposes the right vocabulary for the future module. The current product slice already works with concepts such as:

- statuses: `crit`, `warn`, `ok`, `info`
- actions: `PRUNE`, `EXPAND`, `MONITOR`
- trends: `up`, `down`, `flat`
- core metrics: `gapPct`, `marginPct`, `pme6`, `giro6`, `dos6`, `gmroi6`, `slope6`, `cv6`, `xyz6`, `dataQuality`, `maturity`
- workspace metrics and narratives around price, market, stock, demand, risk, data quality, and recommendation context

That is a good base. The target module should professionalize and operationalize those semantics end-to-end.

## Module mission

Analytics is not just a dashboard. It is the decision engine for the commercial operation.

The module must answer, continuously and with explainability:

- what is selling well
- what is selling badly
- what is growing
- what is declining
- which products have price misalignment
- how much we can reduce or increase price safely
- where capital is stuck
- where stock risk is rising
- which brands are gaining or losing strength
- which taxonomy branches deserve assortment expansion or cleanup
- which SKUs should be monitored, expanded, repriced, pruned, or bought
- which actions should become campaigns
- which issues deserve alerts now
- what the company should do next, with evidence

## Architectural position

The full analytics operating model should be split like this:

- `apps/server_core/internal/modules/analytics_serving`
  - tenant-safe read APIs
  - explainable recommendation payloads
  - alert feeds
  - campaign read surfaces
  - brand and taxonomy read models
- `apps/analytics_worker`
  - heavy computation
  - scoring
  - statistical projections
  - simulations
  - derived classifications
  - batch refresh of read models
- `apps/automation_worker`
  - triggers
  - action orchestration
  - campaign execution
  - delayed follow-ups
- `apps/notifications_worker`
  - alert delivery
  - routing by channel, urgency, and audience

The worker split matters:

- analytics computes intelligence
- automation decides how governed actions run
- notifications delivers signals
- core serves the read surface and remains the canonical synchronous interface

## Ownership boundary

The analytics module must stay strong without stealing write ownership from other domains.

Analytics owns:

- KPIs
- ratios
- classifications
- opportunity scores
- risk scores
- trend detection
- anomaly detection
- price corridor suggestions
- campaign suggestions
- action recommendations
- narrative explainability
- derived read models by SKU, brand, taxonomy, and portfolio

Analytics does not own:

- canonical product identity
- canonical brand registry
- canonical taxonomy structure
- internal price writes
- live stock writes
- purchase order writes
- supplier lead-time writes

This rule is non-negotiable. Analytics can recommend. Other domains remain write owners.

## Analytical grain

The module should operate at multiple grains at the same time:

- portfolio
- brand
- taxonomy node
- product or SKU
- supplier-facing replenishment group
- commercial segment
- time window

The same SKU can therefore be analyzed through several lenses:

- SKU health
- brand role
- taxonomy role
- price position
- inventory pressure
- procurement urgency
- campaign readiness

## Core intelligence model

The module should organize intelligence into eight layers.

### 1. Metrics layer

Raw and derived metrics are normalized into a stable analytical vocabulary.

Main metric families:

- demand and sales
- margin and profitability
- price and market position
- stock and capital exposure
- replenishment pressure
- data quality and model maturity
- strategic opportunity
- operational risk

Representative metrics:

- `sales_units_7d`
- `sales_units_30d`
- `sales_units_90d`
- `revenue_brl_30d`
- `gross_margin_brl_30d`
- `gross_margin_pct_30d`
- `replacement_cost_amount`
- `average_cost_amount`
- `our_price_amount`
- `market_avg_price_amount`
- `market_min_price_amount`
- `market_max_price_amount`
- `price_gap_vs_market_pct`
- `on_hand_quantity`
- `on_hand_value_brl`
- `days_of_supply`
- `days_no_sales`
- `inventory_turnover_6m`
- `pme_days`
- `gmroi_6m`
- `demand_slope_30d`
- `demand_slope_90d`
- `demand_cv_6m`
- `data_quality_score`
- `maturity_score`
- `capital_at_risk_brl`
- `opportunity_score`
- `alert_score`

### 2. Rules layer

Rules convert metrics into business meaning.

Examples:

- detect when price is above a safe competitive band and sales are falling
- detect when stock is excessive relative to demand
- detect when a product has healthy demand but poor availability
- detect when margin is below policy floor
- detect when data quality is too weak to automate a recommendation
- detect when a taxonomy has accelerating demand and low assortment depth
- detect when a brand is losing share inside an important category

Rules must be runtime-governed. Thresholds and policies cannot be hardcoded forever in application code.

### 3. Calculation layer

Calculations transform raw state into decision variables.

Examples:

- trend slope
- moving averages
- weighted margin
- price elasticity proxy
- competitive band
- safety stock
- reorder point
- target stock coverage
- demand volatility
- brand share
- taxonomy share
- alert severity
- campaign priority

### 4. Classification layer

The module classifies entities into operational buckets that humans can act on quickly.

Examples:

- ABC by margin or revenue contribution
- XYZ by demand predictability
- A1, A2, B1, B2, C1 style portfolio classes
- health status: `crit`, `warn`, `ok`, `info`
- recommendation class: `repricing`, `expand`, `prune`, `monitor`, `buy`, `fix_data`, `campaign`

### 5. Recommendation layer

Recommendations are not static labels. They are explicit proposals with rationale.

Each recommendation should include:

- subject
- recommended action
- expected objective
- evidence metrics
- triggered rules
- confidence
- risk notes
- human approval requirement
- review horizon

### 6. Campaign layer

Campaigns group many recommendations into a governed execution unit.

Examples:

- reduce price in slow-moving excess-stock SKUs
- expand coverage in high-demand low-stock SKUs
- reactivate dormant products in a strategic brand
- improve assortment quality inside a taxonomy branch
- negotiate supplier terms for fast-growing items

### 7. Alert layer

Alerts are high-signal exceptions, not just notifications.

Examples:

- critical capital trapped in low-rotation inventory
- abrupt sales drop in a top-margin SKU
- stockout risk in a high-opportunity product
- margin collapse after last price move
- market gap outside policy band
- poor data quality blocking automation

### 8. AI layer

AI sits on top of governed data and rules to accelerate understanding and action.

AI should explain, simulate, summarize, prioritize, and draft actions. AI must not become a hidden source of truth for numbers or policies.

## Metric semantics and first formulas

The module needs explicit formula semantics so recommendations are reproducible.

The exact implementation can evolve, but the first formulas should follow stable business meaning.

### Demand and trend

- `avg_daily_sales_30d = sales_units_30d / 30`
- `avg_daily_sales_90d = sales_units_90d / 90`
- `demand_slope_90d = slope(linear_regression(daily_sales_90d))`
- `demand_acceleration = demand_slope_30d - demand_slope_90d`
- `demand_cv_6m = stddev(monthly_sales_6m) / mean(monthly_sales_6m)`

Interpretation:

- positive slope means demand is increasing
- negative slope means demand is deteriorating
- high `cv` means unstable demand

### Stock and capital

- `days_of_supply = on_hand_quantity / max(avg_daily_sales_30d, demand_floor)`
- `inventory_turnover_6m = cogs_6m / max(avg_inventory_value_6m, value_floor)`
- `pme_days = avg_inventory_value_6m / max(cogs_daily_6m, value_floor)`
- `capital_at_risk_brl = on_hand_value_brl * risk_factor`

Where `risk_factor` increases when:

- `days_of_supply` is above target
- `days_no_sales` is high
- margin is weak
- demand trend is negative

### Price and market

- `price_gap_vs_market_pct = ((our_price_amount / market_avg_price_amount) - 1) * 100`
- `price_gap_vs_market_min_pct = ((our_price_amount / market_min_price_amount) - 1) * 100`
- `price_gap_vs_market_max_pct = ((our_price_amount / market_max_price_amount) - 1) * 100`
- `competitive_dispersion_pct = ((market_max_price_amount - market_min_price_amount) / market_avg_price_amount) * 100`

Interpretation:

- positive gap means we are above market average
- negative gap means we are below market average
- high dispersion means pricing can be less aggressive because the market is noisy

### Margin and return

- `gross_margin_brl = revenue_brl - cogs_brl`
- `gross_margin_pct = gross_margin_brl / max(revenue_brl, value_floor)`
- `margin_unit_pct = (our_price_amount - cost_basis_amount) / max(our_price_amount, value_floor)`
- `gmroi_6m = gross_margin_brl_6m / max(avg_inventory_cost_brl_6m, value_floor)`

The module should support multiple cost bases:

- replacement cost
- average cost
- fallback mode when one basis is missing

### Buying and replenishment

- `lead_time_demand = avg_daily_sales_30d * lead_time_days`
- `safety_stock = service_factor * demand_stddev * sqrt(max(lead_time_days, 1))`
- `reorder_point = lead_time_demand + safety_stock`
- `net_available = on_hand_quantity + on_order_quantity - reserved_quantity`
- `suggested_buy_qty = max(0, target_stock_qty - net_available)`

Where `target_stock_qty` should consider:

- service level policy
- demand class
- volatility
- lead-time variability
- planned campaign uplift
- strategic brand or taxonomy priority

## Decision engines

The analytics module should expose four decision engines.

### 1. Product health engine

Purpose:

- decide if a SKU should be expanded, monitored, repriced, pruned, or bought

Inputs:

- sales trend
- margin
- stock coverage
- market gap
- data quality
- strategic weight

Outputs:

- status
- action
- confidence
- explanation

### 2. Pricing recommendation engine

Purpose:

- answer whether price should go down, go up, stay stable, or be tested

Method:

- define a minimum price floor from cost, minimum margin policy, and governance constraints
- define a maximum ceiling from market band, elasticity evidence, and brand positioning policy
- define a recommended corridor between floor and ceiling
- choose the recommended price based on the current objective

Possible objectives:

- release trapped capital
- maximize margin
- increase volume
- defend premium positioning
- recover competitiveness

The output must not be just one number. It should include:

- current price
- floor
- ceiling
- recommended price
- maximum safe discount
- expected margin impact
- expected turnover impact
- confidence band

### 3. Portfolio intelligence engine

Purpose:

- prioritize which products deserve management attention

Representative composite score:

- `priority_score = w1 * capital_risk + w2 * margin_leakage + w3 * demand_opportunity + w4 * strategic_weight + w5 * urgency`

This score should drive:

- spotlight tables
- work queues
- alert ranking
- campaign selection

### 4. Procurement suggestion engine

Purpose:

- suggest what to buy, when to buy, how much to buy, and where to buy

Important boundary:

- the buying engine may consume analytics outputs
- `procurement` remains the write owner of replenishment decisions and workflow
- procurement must not depend on analytics internals for its canonical upstream facts

Analytics should therefore publish advisory outputs such as:

- buy now
- wait
- reduce order
- accelerate supplier follow-up
- switch supplier candidate
- review MOQ or pack size

## Brand intelligence

Brand analytics should answer whether each brand is healthy, strategic, overpriced, underexposed, overstocked, or declining inside the portfolio.

Brand read surfaces should include:

- sales trend by brand
- margin trend by brand
- stock and capital by brand
- competitive position by brand
- campaign performance by brand
- concentration risk by brand
- supplier dependency by brand
- whitespace opportunities by brand and taxonomy

Representative brand questions:

- which brands deserve more inventory and visibility
- which brands are losing competitiveness
- which brands have excess stock without demand support
- which brands are sustaining premium pricing successfully
- which brands should receive campaign investment now

## Taxonomy intelligence

Taxonomy analytics should tell the company where the portfolio is structurally strong or weak.

Taxonomy read surfaces should include:

- node-level demand trend
- margin density by node
- stock pressure by node
- assortment depth and breadth
- whitespace and missing assortment
- brand mix inside the node
- cannibalization signals
- classification quality gaps

Representative taxonomy questions:

- which taxonomy branches are growing faster than the rest of the company
- where we have too much stock in weak demand branches
- where assortment is too shallow for the opportunity
- where brand distribution is unbalanced
- where pricing is systematically above or below market

## Campaign model

Campaigns are the bridge between analytics and execution.

The campaign module should support:

- campaign objective
- target segment
- inclusion and exclusion rules
- action template
- owner
- approval flow
- schedule
- expected KPI impact
- observed KPI impact
- rollback or closure criteria

Example campaign types:

- `price_correction`
- `inventory_release`
- `brand_acceleration`
- `taxonomy_expansion`
- `slow_mover_cleanup`
- `supplier_rebalance`
- `data_quality_repair`

Campaign segmentation should work across:

- SKU lists
- brand
- taxonomy
- supplier group
- status tier
- action class
- opportunity score band

## Action model

Actions are the atomic operations proposed by analytics or executed through campaigns.

Examples:

- propose price decrease
- propose price increase
- create buy suggestion
- mark SKU for close monitoring
- request supplier review
- request catalog correction
- request taxonomy correction
- open commercial follow-up
- open automation task

Each action should store:

- `action_type`
- `subject_type`
- `subject_id`
- `reason_code`
- `evidence_snapshot`
- `recommended_by`
- `approval_state`
- `executed_at`
- `result_state`

## Alerts model

Alerts should be explicit and operationally triaged.

Alert categories:

- sales deterioration
- stock risk
- capital trapped
- margin erosion
- competitive pricing deviation
- procurement urgency
- campaign underperformance
- data quality degradation
- model confidence degradation

Alert severities:

- `info`
- `warn`
- `critical`

Alert scoring should consider:

- financial impact
- urgency
- confidence
- scope size
- recurrence

Recommended first routing:

- commercial leads receive pricing and campaign alerts
- buyers receive procurement and stock pressure alerts
- category managers receive brand and taxonomy alerts
- operations receive data quality and integration degradation alerts

## AI interface

AI should make analytics operationally useful, not cosmetically impressive.

The AI layer should support:

- explain this KPI movement
- explain why this SKU was classified as `critical`
- summarize the biggest changes since last week
- propose the next 10 actions by financial impact
- simulate a price change before approval
- compare two brands or two taxonomy branches
- draft a campaign from current alerts
- explain buy suggestions in plain language
- identify the likely root cause of poor sales

AI output rules:

- AI never invents canonical numbers
- AI always cites the metrics, rules, and comparisons used
- AI respects governance and approval policies
- AI can draft actions but does not auto-write transactional truth unless an approved automation path exists
- AI explanations must be traceable to deterministic evidence

## Explainability contract

Every recommendation must be explainable in a standard structure.

Minimum explainability payload:

- what changed
- why it matters
- evidence metrics
- triggered rules
- confidence level
- expected upside
- expected downside
- recommended next step
- review window

Example:

- "Reduce price by 4.5% because the SKU is 11.2% above market average, sales slope is negative for 90 days, DOS is 126 days, and margin remains above floor after the reduction."

## Read surfaces

The target analytics product should expose these first-class surfaces:

- executive home
- products overview
- product workspace
- brands overview
- brand workspace
- taxonomy overview
- taxonomy workspace
- alerts center
- campaigns center
- buying recommendations board
- AI copilot panel

Each surface should work with the same underlying intelligence vocabulary so the company is not looking at disconnected dashboards.

## Buying and procurement intelligence

The future buying module should feel native to analytics without violating procurement boundaries.

The user experience should answer:

- what should we buy now
- what should we postpone
- what quantity should we suggest
- which supplier looks best
- where demand is accelerating faster than coverage
- where a campaign will require replenishment first
- where price reduction should happen before replenishment

The suggestion engine should combine:

- demand trend
- stock coverage
- open purchase orders
- supplier lead time
- supplier reliability
- margin profile
- campaign forecast
- strategic brand or taxonomy priority

The most important boundary remains:

- analytics can compute advisory buy intelligence
- procurement owns the actual replenishment workflow and decisions

## Data quality and maturity

The module must expose when its own advice is weak.

Every entity should have:

- `data_quality_score`
- `maturity_score`
- missing field counts
- imputation flags
- freshness status
- model confidence

This is mandatory. A system that recommends actions without showing data quality will lose operator trust.

## Governance requirements

The intelligence module must be governed, not improvised.

The following items should be runtime-configurable over time:

- price floor policy
- competitive band policy
- alert sensitivity
- campaign approval thresholds
- service level targets for buy suggestions
- strategic weighting for brands and taxonomies
- automation permissions

## Phased implementation

### Phase 1. Read intelligence

Deliver:

- product, brand, and taxonomy read models
- stable KPIs
- trend detection
- deterministic classifications

### Phase 2. Recommendation engine

Deliver:

- pricing suggestions
- product health statuses
- portfolio prioritization
- explainability payloads

### Phase 3. Alerts and campaigns

Deliver:

- alert center
- campaign center
- action queue
- governed routing and approvals

### Phase 4. Buying intelligence

Deliver:

- advisory procurement suggestions
- reorder and supplier recommendation logic
- integration with procurement surfaces

### Phase 5. AI copilot

Deliver:

- natural-language explanations
- simulations
- comparative analysis
- campaign drafting
- operator copiloting

## Definition of success

The analytics module is succeeding when a user can answer, in one system and with evidence:

- where should we act first
- what should we sell more
- what should we stop forcing
- where should we lower price
- where can we raise price safely
- which brands deserve investment
- which taxonomy branches are weak or promising
- which stock is trapped
- what should we buy and when
- which alerts matter now
- which campaign should start this week

If the module cannot answer those questions clearly, it is still a dashboard, not intelligence.
