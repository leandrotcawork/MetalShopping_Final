# SKU Canonical Data Model

## Purpose

Freeze the professional target database model for SKU-related data in MetalShopping Final.

This document is based on:

- the approved modular monolith architecture
- the accepted multi-tenant Postgres model
- the real business semantics present in legacy `products` and `product_erp`
- the long-term need to support pricing, inventory, procurement, analytics, CRM, and adaptive integrations without collapsing everything into one table

## Core rule

The target system must not create one giant SKU table.

The correct model is a canonical split model:

- `catalog` owns what the SKU is
- `pricing` owns what the SKU costs and is sold for
- `inventory` owns how much of the SKU exists right now
- `procurement` owns how the SKU is replenished
- analytics and read models own derived operational meaning

## Legacy reference

Legacy already implies this split, even if it is not yet cleanly formalized.

### Legacy `products`

Semantically behaves as product master data.

Relevant fields:

- `pn_interno`
- `reference`
- `ean`
- `descricao`
- `marca`
- `ativo`
- `tipo_estoque`
- `taxonomy_node_id`

References:

- [products_write_repo_pg.py](C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/Nova%20pasta/MetalShopping/db/products_write_repo_pg.py#L45)
- [products_repo.py](C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/Nova%20pasta/MetalShopping/db/products_repo.py#L234)

### Legacy `product_erp`

Semantically behaves as operational ERP enrichment and current commercial state.

Relevant fields:

- `preco_interno`
- `custo_variavel`
- `custo_medio`
- `dt_compra`
- `dt_venda`
- `dias_sem_venda`
- `estoque_disponivel`
- `st_imposto`
- `competitivo`
- `classificacao`

References:

- [products_write_repo_pg.py](C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/Nova%20pasta/MetalShopping/db/products_write_repo_pg.py#L241)
- [products_write_repo_pg.py](C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/Nova%20pasta/MetalShopping/db/products_write_repo_pg.py#L360)

## Target database model

## 1. Canonical identity in `catalog`

### Table: `catalog_products`

This is the stable anchor of the SKU inside one tenant.

Required target fields:

- `product_id`
- `tenant_id`
- `sku`
- `name`
- `description`
- `brand_name`
- `stock_profile_code`
- `primary_taxonomy_node_id`
- `status`
- `created_at`
- `updated_at`

Notes:

- `sku` is the canonical business key inside the tenant
- `stock_profile_code` is metadata, not live stock state
- `tipo_estoque` from the legacy system should map here only if it remains stable product master semantics
- product identity must stay small, stable, and tenancy-safe

## 2. Alternate identifiers in `catalog`

### Table: `catalog_product_identifiers`

This table exists because one SKU is recognized by multiple business and integration keys.

Required target fields:

- `product_identifier_id`
- `product_id`
- `tenant_id`
- `identifier_type`
- `identifier_value`
- `source_system`
- `is_primary`
- `created_at`
- `updated_at`

Expected identifier types:

- `pn_interno`
- `reference`
- `ean`
- future ERP codes
- future marketplace or supplier aliases

Frozen rule:

- these identifiers must not all be flattened into `catalog_products`
- `pn_interno` remains extremely important, but it is treated as a canonical identifier, not as the only permanent target primary key

## 3. Classification in `catalog`

### Tables

- `catalog_taxonomy_nodes`
- `catalog_taxonomy_level_defs`

Taxonomy remains the canonical classification model.

Frozen rules:

- no canonical `grupo` field
- no recreation of fixed `categoria/grupo/subgrupo` columns
- `taxonomy_node_id` is the proper long-term classification pointer

## 4. Pricing semantics in `pricing`

### Table family

- `pricing_product_prices`
- future supporting tables for cost snapshots or price list variants if needed

The first pricing revision should align with legacy semantics instead of generic naming.

Recommended target fields for the write model:

- `price_id`
- `tenant_id`
- `product_id`
- `currency_code`
- `internal_price_amount`
- `replacement_cost_amount`
- optional `average_cost_amount`
- `pricing_status`
- `effective_from`
- `effective_to`
- `origin_type`
- `origin_ref`
- `reason_code`
- `updated_by`
- `created_at`
- `updated_at`

Important decision:

- legacy `custo_variavel` should be treated as `replacement_cost_amount`
- legacy `custo_medio` should not be silently erased; it should either become `average_cost_amount` or be explicitly postponed
- generic `cost_basis_amount` is not the best business term for MetalShopping
- generic `margin_floor_value` is not part of the canonical business language today and should not remain mandatory if the real product calculates margins instead of governing through explicit floor persistence

## 5. Live stock in `inventory`

### Table family

- `inventory_positions`
- future stock movement or stock aging structures when the module opens

Required ownership:

- `estoque_disponivel`
- on-hand quantity
- stock aging
- coverage and turnover support state
- rupture and overstock operational state

Frozen rule:

- `estoque_disponivel` must not be copied into `catalog_products`
- `estoque_disponivel` must not be treated as a core `pricing` field
- `pricing` may consume inventory signals later, but it does not own them

## 6. Procurement semantics in `procurement`

These fields should not be prematurely collapsed into `pricing`:

- supplier replenishment assumptions
- acquisition-side lead time
- purchase-side restrictions
- replenishment windows

Those belong to `procurement`.

## 7. Read models and analytics

These belong outside canonical write tables:

- realized margin
- contribution margin percent
- capital tied in stock
- days without sale
- demand coverage
- turnover
- competitive positioning summaries
- advisory classifications and recommendations

These are read-model or analytics concerns unless a later domain explicitly becomes the canonical owner.

## Legacy-to-target field map

| Legacy field | Target owner | Target representation | Decision |
| --- | --- | --- | --- |
| `pn_interno` | `catalog` | `catalog_product_identifiers(identifier_type='pn_interno')` | keep as important canonical identifier |
| `reference` | `catalog` | `catalog_product_identifiers(identifier_type='reference')` | keep as identifier |
| `ean` | `catalog` | `catalog_product_identifiers(identifier_type='ean')` | keep as identifier |
| `descricao` | `catalog` | `catalog_products.description` | keep in canonical product |
| `marca` | `catalog` | `catalog_products.brand_name` or future brand dimension | keep in canonical product domain |
| `ativo` | `catalog` | `catalog_products.status` | keep |
| `tipo_estoque` | `catalog` | `catalog_products.stock_profile_code` | keep only as stable product metadata |
| `taxonomy_node_id` | `catalog` | `catalog_products.primary_taxonomy_node_id` | keep |
| `preco_interno` | `pricing` | `pricing_product_prices.internal_price_amount` | keep |
| `custo_variavel` | `pricing` | `pricing_product_prices.replacement_cost_amount` | rename semantically |
| `custo_medio` | `pricing` | `pricing_product_prices.average_cost_amount` or later dedicated cost structure | preserve explicitly |
| `estoque_disponivel` | `inventory` | `inventory_positions.on_hand_quantity` | do not keep in pricing/catalog |
| `dt_compra` | `inventory` or read model | later explicit owner | do not keep in catalog |
| `dt_venda` | `inventory` or read model | later explicit owner | do not keep in catalog |
| `dias_sem_venda` | analytics/read model | derived or imported read model | do not keep in core identity |
| `st_imposto` | future explicit owner | undecided domain | do not place by inertia |
| `competitivo` | analytics/read model | advisory or scoring read model | do not place by inertia |
| `classificacao` | analytics/read model | advisory/read model unless canonized later | do not place by inertia |

## Database design consequences

- prefer smaller domain-owned tables over one universal product table
- prefer canonical identity keys plus module-owned foreign keys
- prefer explicit evolution per module over nullable sprawl
- keep multitenancy explicit with `tenant_id` and RLS in each owned table
- do not use analytics convenience as a reason to collapse write ownership

## Immediate consequence for current work

Before continuing to expand `pricing`, the current first slice should be revised to match this accepted target:

- replace generic `cost_basis_amount` with explicit pricing cost semantics
- stop treating `margin_floor_value` as mandatory canonical persistence if the real business model does not use it
- keep stock outside `pricing`
- keep SKU identity split across `catalog_products` and `catalog_product_identifiers`
