# Pricing Legacy Signal Boundaries

## Purpose

Freeze how legacy `product_erp` signals should be split between `pricing`, `inventory`, `procurement`, and analytics-oriented read models.

This document exists to prevent the new `pricing` module from gradually becoming a copy of legacy ERP state.

## Decision

The target platform keeps `pricing` semantically narrow:

- `pricing` owns internal price and explicit cost semantics
- `inventory` owns live stock and stock-timing facts
- `procurement` owns supplier-side replenishment semantics
- analytics and read models own derived commercial pressure and advisory classifications

## Legacy signal split

| Legacy field | Primary target owner | Target representation | Why |
| --- | --- | --- | --- |
| `preco_interno` | `pricing` | `pricing_product_prices.price_amount` | canonical internal commercial price |
| `custo_variavel` | `pricing` | `pricing_product_prices.replacement_cost_amount` | replacement-cost semantics used for internal price decisions |
| `custo_medio` | `pricing` | `pricing_product_prices.average_cost_amount` | preserve explicit average-cost semantics for margin and analytics usage |
| `estoque_disponivel` | `inventory` | `inventory_positions.on_hand_quantity` | live stock state is not price ownership |
| `dt_compra` | `inventory` or inventory-serving read model | explicit inventory timing field later | purchase timing is operational stock context, not pricing state |
| `dt_venda` | `inventory` or inventory-serving read model | explicit inventory timing field later | sale timing belongs to stock/turnover context, not pricing ownership |
| `dias_sem_venda` | analytics or inventory-serving read model | derived stale-stock signal | pressure signal for decisions, not canonical pricing write state |
| `st_imposto` | future explicit owner | undecided domain | do not absorb into pricing by inertia |
| `competitivo` | analytics / market-intelligence-serving read model | advisory competitive state | output of analysis, not write ownership |
| `classificacao` | analytics / portfolio read model | advisory classification | derived operational label, not canonical price state |

## What `pricing` may consume later

`pricing` will likely consume, but not own:

- stock pressure signals
- stale inventory signals
- market competitiveness inputs
- tax-aware commercial rules
- replenishment cost context from procurement

The rule is:

- consuming another module's signal does not justify copying that signal into `pricing`'s canonical write table

## Practical follow-on sequence

After the current `pricing` slice, the next boundary-safe follow-ons should be:

1. `inventory` opening with `on_hand_quantity` and operational timing ownership
2. inventory-serving or analytics-serving read models for `dias_sem_venda` and stale pressure
3. `procurement` cost and replenishment ownership when supplier-side workflows become explicit
4. pricing policy evolution that consumes those downstream signals without absorbing their write ownership

## Non-goals

This document does not authorize:

- adding `estoque_disponivel` to `pricing_product_prices`
- adding `dias_sem_venda` to `pricing_product_prices`
- adding `competitivo` or `classificacao` as canonical pricing columns
- treating `st_imposto` as a pricing field just because it came from `product_erp`
