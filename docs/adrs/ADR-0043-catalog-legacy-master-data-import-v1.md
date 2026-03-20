# ADR-0043: Catalog Legacy Master Data Import v1

- Status: accepted
- Date: 2026-03-20

## Context

We want the new `catalog` canonical state to start from the same product universe that already exists in the legacy Postgres database (`metalshopping_db`, schema `metalshopping`), instead of continuing with small, manually created test datasets.

Today:

- `catalog` is the canonical owner of product master data (`catalog_products`, `catalog_product_identifiers`, taxonomy tables).
- `shopping` runs depend on a realistic product universe and stable identifiers (EAN/reference/pn_interno).
- The team has already concentrated suppliers, manifests, and URLs under `tenant_default`, so the target tenant for data is `tenant_default`.

We need a professional import path that:

- is repeatable and idempotent
- is safe under multitenancy/RLS
- does not introduce runtime coupling to the legacy backend
- produces objective evidence (counts + conflict reports) before we accept it

## Decision

Introduce an explicit, admin-run import pipeline to migrate **legacy product master data** from `metalshopping_db` into the new core canonical tables.

Scope for v1:

- Import taxonomy structure:
  - legacy `metalshopping.taxonomy_level_defs`
  - legacy `metalshopping.taxonomy_nodes`
- Import product master data:
  - legacy `metalshopping.products`
  - including `pn_interno`, `reference`, `ean`, `descricao`, `marca`, `tipo_estoque`, `ativo`, `taxonomy_node_id`
- Target tables (new core):
  - `catalog_taxonomy_level_defs`
  - `catalog_taxonomy_nodes`
  - `catalog_products`
  - `catalog_product_identifiers`

Explicit non-goals for v1:

- importing legacy operational snapshots (`product_erp`, `product_internal_prices`) into canonical modules
  - those map to `pricing` and `inventory` and must be handled by dedicated, module-owned ingest later
- importing `shopping` URL tables or driver state from legacy as part of the catalog import
  - URL import belongs to shopping ADRs, not catalog

### Operational rule

The import is an **admin-run script** (offline operation), not a runtime service path:

- `server_core` must not connect to legacy `metalshopping_db` in normal request handling.
- The script may connect to both databases and write to the new canonical tables using tenant-scoped sessions.

### Tenancy rule

The import must be **tenant-scoped**, defaulting to `tenant_default`:

- The script sets tenant context in the destination database (RLS must apply).
- It must never write rows for multiple tenants in a single execution.

### Reset semantics (v1)

We support a deliberate reset mode for local/dev bootstrap:

- Delete the destination tenant’s catalog state first (products + identifiers + taxonomy), then import.
- This is a controlled operation and must not be run implicitly by the application.

## Data Mapping (high level)

Legacy -> New:

- `products.pn_interno` -> `catalog_products.sku`
- `products.descricao` -> `catalog_products.name` and `catalog_products.description` (see ADR-0044 fallback rules)
- `products.marca` -> `catalog_products.brand_name`
- `products.tipo_estoque` -> `catalog_products.stock_profile_code`
- `products.ativo` -> `catalog_products.status` (`active|inactive`)
- `products.taxonomy_node_id` -> `catalog_products.primary_taxonomy_node_id` (stable mapping defined in ADR-0044)
- `products.pn_interno/reference/ean` -> `catalog_product_identifiers` (`identifier_type` values are frozen: `pn_interno|reference|ean`)

## Contracts (touchpoints)

No new API/event contracts are required for the import itself in v1.

The import must preserve the already-frozen canonical model:

- `docs/CATALOG_CANONICAL_MODEL.md`
- `docs/SKU_CANONICAL_DATA_MODEL.md`

## Implementation Checklist

Frozen execution order for this ADR:

1. Define importer spec and safety rules (this ADR + ADR-0044)
   Skill: `metalshopping-adr-updates`
2. Implement admin-run importer (scripts only)
   Skill: `metalshopping-observability-security` (review: RLS/tenant session usage)
3. Validate against real legacy DB and produce evidence
   Skill: `metalshopping-architecture-direction-review` (review: boundary safety)
4. Update planning docs and mark ADR accepted
   Skill: `metalshopping-adr-updates`

## Acceptance Evidence (for Status: accepted)

- Pre-check report generated from legacy DB:
  - product counts (total/active)
  - EAN conflicts report
  - reference conflicts report
  - taxonomy integrity report (products referencing missing nodes)
- Destination evidence for `tenant_default`:
  - row counts in `catalog_products`, `catalog_product_identifiers`, `catalog_taxonomy_nodes`
  - spot-check: at least 20 SKUs match legacy fields for brand, description, taxonomy assignment
- UI sanity check (thin-client):
  - Products portfolio filters list real brands and taxonomy leaf0 names
  - Shopping run scope from catalog can select a realistic list of products

Evidence captured on 2026-03-20:

- Import run (reset + import) summary:
  - `levels=3`, `nodes=100`, `products_imported=3838`
  - `identifier_conflicts=59`, `missing_taxonomy_rows=1`
  - reports: `.tmp/import_catalog_reports/20260320_210326`

## Alternatives considered

- Import through `server_core` HTTP endpoints:
  - rejected because it creates a chatty, slow ingestion path and adds avoidable runtime coupling.
- Keep only XLSX import as the source of truth:
  - rejected because we already have a populated canonical legacy DB and want a full universe with taxonomy and brand preserved.

## Consequences

- We gain a repeatable path to populate canonical `catalog` state from legacy master data.
- Data quality issues become explicit via reports instead of being silently absorbed into the new system.
- We keep the core runtime clean: no legacy DB dependency in request serving.
