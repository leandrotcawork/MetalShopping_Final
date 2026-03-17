# ADR-0007: Canonical SKU Data Ownership

- Status: accepted
- Date: 2026-03-17

## Context

MetalShopping legacy mixes canonical product identity, internal pricing signals, stock snapshots, and operational ERP fields across `products` and `product_erp`.

That legacy model is valuable as a semantic reference, but reproducing it as a single target table would create ownership ambiguity and long-term coupling between `catalog`, `pricing`, `inventory`, `procurement`, and analytics.

The new platform needs a SKU model that is:

- faithful to the real business semantics of the legacy system
- scalable for future modules such as CRM, analytics, procurement, and adaptive integrations
- explicit about domain ownership
- safe for multi-tenant canonical state

## Decision

- `catalog` is the canonical owner of SKU identity and stable master data
- `pricing` owns internal price and cost semantics
- `inventory` owns live stock position and stock-derived operational state
- `procurement` owns replenishment and supplier-side acquisition semantics
- analytics and read models own derived KPIs, ratios, and advisory projections

The canonical SKU model must therefore be persisted as a split model, not a single wide table.

### `catalog` owns

- canonical product identity
- `sku`
- product naming and description
- brand identity
- classification through taxonomy
- stable stock profile metadata
- canonical identifiers such as `pn_interno`, `reference`, `ean`, and future upstream aliases
- active/inactive lifecycle state

### `pricing` owns

- internal selling price
- replacement cost semantics derived from legacy `custo_variavel`
- additional cost semantics such as average cost when intentionally modeled
- price validity windows
- reasoned price changes and pricing event publication

### `inventory` owns

- `estoque_disponivel`
- on-hand quantity
- stock aging and stock coverage
- operational stock risk state

### `procurement` owns

- supplier replenishment state
- lead time
- purchase-side constraints
- buying workflow data

### Analytics and read models own

- realized margins
- stock capital exposure
- market-relative indicators
- operational classifications and advisory outputs that are not canonical write ownership

## Consequences

- no single `products` replacement table should absorb all legacy columns
- `catalog_products` must stay focused on identity and stable master data
- `pricing` must be revised to reflect MetalShopping legacy semantics instead of generic fields that do not belong to the real business language
- future legacy migration work must map every `products` and `product_erp` field to an explicit target owner before implementation
