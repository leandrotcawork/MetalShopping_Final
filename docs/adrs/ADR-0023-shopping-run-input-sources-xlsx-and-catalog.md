# ADR-0023: Shopping Run Input Sources (XLSX Scope and Catalog Selection)

- Status: accepted
- Date: 2026-03-19

## Context

The legacy Shopping workflow offers two input modes:

- XLSX (current operational spreadsheet)
- Catalog selection (products already registered)

In the target platform, canonical product writes belong to `catalog` and the broader integration platform, not to the Shopping workflow UI. Shopping must be able to define run scope without turning into a parallel product-import subsystem.

## Decision

Shopping Level 2 will support two run-scope input sources, with explicit semantics:

1. Catalog selection (primary path)
   - UI selects `product_id` values from the Products portfolio surface
   - run request stores the explicit list of selected products

2. XLSX scope (workflow input, not canonical import)
   - UI provides an XLSX as a run-scope input
   - server_core extracts the minimal scope identifiers (for example `pn_interno`, `ean`, `reference`)
   - server_core resolves identifiers to existing `catalog_products` for the tenant
   - the run request stores the resolved `product_id` list plus an audit of unresolved identifiers

Non-goal (explicitly out of Shopping Level 2):

- Shopping does not import or mutate canonical catalog products from XLSX.
  Canonical import belongs to the integration platform and catalog governance.

## Consequences

- The legacy UX remains available (XLSX path) without reintroducing legacy data coupling.
- Run scope becomes auditable and reproducible as part of the run request ledger.
- Product import remains governed and modular instead of embedded in an operational workflow page.

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Contract authoring (XLSX scope endpoint shape): `metalshopping-openapi-contracts`
- Go module changes (scope extraction and validation): `metalshopping-module-scaffold`
- SDK refresh: `metalshopping-sdk-generation`
- UI behavior preservation: `metalshopping-frontend-migration-guardrails` and `metalshopping-page-delivery`

