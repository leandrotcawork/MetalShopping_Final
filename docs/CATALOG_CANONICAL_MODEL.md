# Catalog Canonical Model

## Purpose

Define the canonical product model for `catalog` before opening `pricing` and other dependent modules.

This document turns the current minimal `catalog_products` slice into an explicit target model informed by selective legacy reuse, without copying the legacy schema wholesale.

## Why this exists now

The current `catalog` implementation is intentionally minimal. It proves:

- canonical product ownership in the core
- tenant-aware writes
- RLS-backed isolation
- auth and authz integration
- runtime governance affecting real behavior

It does **not** yet represent the full MetalShopping product model.

Before `pricing`, `inventory`, `market_intelligence`, and analytics grow on top of `catalog`, the product model must be frozen at a more professional level.

## Legacy signals worth preserving

The legacy stack contains several useful signals that should inform the new model:

- `products_master` is treated as a canonical integration dataset
- `inventory_position` is treated as a separate canonical dataset
- product classification is moving from legacy `products.grupo` to canonical taxonomy
- taxonomy is modeled as nodes plus level definitions
- `marca` is treated as a transversal dimension, not a taxonomy level
- import flows preserve existing populated values instead of overwriting them with null or empty values

Relevant legacy references:

- `contexts/integration/contracts/defaults.py`
- `db/schema.sql`
- `db/products_write_repo_pg.py`
- `docs/analytics/TAXONOMY_GRUPO_TRANSITION_COMPAT_BRIDGE_V1.md`
- `docs/analytics/ANALYTICS_PRODUCT_HIERARCHY_TAXONOMY_N_LEVEL_DECISION_V1.md`

## Frozen ownership boundaries

### `catalog` owns

- canonical product identity
- canonical SKU identity inside the tenant
- product display name and business-facing naming
- brand as a transversal product dimension
- taxonomy ownership and classification assignment
- stable master-data attributes that define what the product is
- active/inactive commercial state of the product record
- canonical identifiers and compatibility bridges from upstream systems

### `catalog` does not own

- internal prices, costs, and pricing history
- stock quantities and inventory position
- open sales orders or purchase orders
- supplier lead time
- market-observed seller and competitor data
- transient ERP operational snapshots that are not product identity

These belong to:

- `pricing`
- `inventory`
- `procurement`
- `market_intelligence`
- `sales`

## Canonical product model

## 1. Product identity

The core product aggregate should represent the product as a canonical business entity inside one tenant.

Minimum target fields for the aggregate:

- `product_id`
- `tenant_id`
- `sku`
- `name`
- `brand_name`
- `stock_profile_code`
- `is_active`
- `primary_taxonomy_node_id`
- `created_at`
- `updated_at`

Notes:

- `sku` is the canonical business key inside the tenant
- `brand_name` belongs to `catalog` because it is product identity, not a taxonomy level
- `stock_profile_code` is acceptable in `catalog` only as stable product master metadata, not as live inventory state
- `primary_taxonomy_node_id` is the canonical classification pointer for the current phase

## 2. Product identifiers

Legacy shows that product identity arrives from multiple upstream identifiers:

- internal product number (`pn_interno`)
- `reference`
- `ean`

Those should not all be flattened forever into the main product table.

Target supporting structure:

- `catalog_product_identifiers`

Suggested fields:

- `product_id`
- `tenant_id`
- `identifier_type`
- `identifier_value`
- `source_system`
- `is_primary`
- `created_at`
- `updated_at`

Why:

- supports ERP identity, commercial reference, and barcode without overloading one table
- avoids hardcoding only one upstream identity strategy
- helps future integrations and matching workflows

## 3. Taxonomy

Taxonomy should be a first-class structure of `catalog`, not a string field scattered across modules.

Target structures:

- `catalog_taxonomy_nodes`
- `catalog_taxonomy_level_defs`

### `catalog_taxonomy_nodes`

Owns:

- tree structure
- parent-child relationship
- level depth
- display name
- normalized name or code when useful
- active state

Core fields:

- `taxonomy_node_id`
- `tenant_id`
- `name`
- `name_norm`
- `code`
- `parent_node_id`
- `level`
- `path`
- `is_active`
- `created_at`
- `updated_at`

### `catalog_taxonomy_level_defs`

Owns:

- semantic labels for levels
- per-tenant hierarchy naming

Core fields:

- `tenant_id`
- `level`
- `label`
- `short_label`
- `is_enabled`
- `created_at`
- `updated_at`

Frozen rule:

- new work must not recreate fixed `categoria/grupo/subgrupo` columns as canonical structure
- additional hierarchy levels enter only through taxonomy tables

## 4. Compatibility bridge from legacy classification

Legacy `products.grupo` should not become the new canonical structure.

Accepted transition rule:

- canonical classification is taxonomy
- legacy group labels may exist only as compatibility input or derived display fallback during migration
- no new fixed structural columns should be introduced for category/group/subgroup

## 5. Master data ingest semantics

`catalog` should absorb product master data from integration flows with explicit ingest semantics.

Frozen rules:

- integration publishes canonical `products_master`
- `catalog` is the write owner of canonical product identity in the core
- ingestion should preserve useful existing values when incoming values are null or empty
- enrichment and imports must follow no-null-overwrite semantics unless a field explicitly allows hard replacement

This rule is important because the legacy flow already treated product master updates carefully and that behavior is worth preserving.

## What stays out of `catalog`

### Pricing data

Do not place these in `catalog_products`:

- `preco_interno`
- `custo_variavel`
- `custo_medio`
- internal pricing history

Those belong to `pricing`.

### Inventory state

Do not place these in `catalog_products`:

- `estoque_disponivel`
- on-hand quantity
- capital in stock
- stock aging
- rupture or excess state

Those belong to `inventory`.

### ERP operational snapshot fields

Legacy fields such as these should not be treated as canonical product identity:

- `dt_compra`
- `dt_venda`
- `dias_sem_venda`
- `st_imposto`
- `competitivo`
- `classificacao`

If needed later, they should be rehomed deliberately into the correct module or read model, not copied into the product core by inertia.

## Near-term target schema evolution

The current `catalog_products` table should evolve toward this shape in phases:

### Phase A

- keep current `catalog_products`
- add `brand_name`
- add `stock_profile_code`
- add `primary_taxonomy_node_id`
- preserve current tenancy and RLS rules

### Phase B

- add `catalog_product_identifiers`
- move upstream identifiers into explicit supporting structure
- keep `sku` as the canonical business key

### Phase C

- add `catalog_taxonomy_nodes`
- add `catalog_taxonomy_level_defs`
- migrate canonical classification to taxonomy tables

### Phase D

- introduce compatibility import rules from `products_master`
- document no-null-overwrite ingest behavior in implementation

## Recommended next implementation order

1. Expand `catalog` model documents first
2. Introduce taxonomy tables and contracts
3. Introduce product identifiers table
4. Expand `catalog_products` with brand and stock profile metadata
5. Only then open `pricing` on top of the stabilized product model

## What the current `catalog_products` table already validates

The current table is still useful and correct for this phase because it already validates:

- tenant-aware product ownership
- canonical SKU uniqueness per tenant
- product active status
- RLS enforcement
- basic product write and read flow through the core

It should be read as a foundation slice, not as the final product schema.
