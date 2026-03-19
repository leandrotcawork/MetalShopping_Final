# ADR-0024: Persisted Supplier Product URLs and Lookup Mode for Shopping Execution

- Status: accepted
- Date: 2026-03-19

## Context

The legacy engine improves scraping success and speed with two persisted signals:

- supplier product URL per product (PDP-first shortcut and stale tracking)
- lookup mode per (product, supplier) (EAN vs reference) inferred from results

These signals are operationally valuable:

- reduce repeated discovery cost
- stabilize results when supplier search quality varies
- allow a "manual URLs" workflow for high-value SKUs

In the target platform, these must be tenant-scoped, auditable, and owned by backend tables (not page-local state).

## Decision

Shopping execution will support persisted per-tenant signals:

1. Supplier product URL
   - stores the best known PDP URL per `(tenant_id, product_id, supplier_code)`
   - tracks URL status (`ACTIVE`, `STALE`, `INVALID`) and last check signals
   - supports both discovery updates (worker) and manual overrides (admin/UI workflow)

2. Supplier lookup mode
   - stores the preferred lookup identifier per `(tenant_id, product_id, supplier_code)`
   - vocabulary: `EAN` or `REFERENCE`
   - updated by the worker based on inference from successful results

Worker rules:

- may read these signals before execution to choose the best request plan
- must upsert updates idempotently

## Contracts (touchpoints)

- Data model (tenant-scoped + RLS): `apps/server_core/migrations/0024_shopping_supplier_product_signals.sql`
- OpenAPI: `contracts/api/openapi/shopping_v1.openapi.yaml`
  - `GET /api/v1/shopping/supplier-signals`
  - `PUT /api/v1/shopping/supplier-signals` (upsert)
- JSON Schemas (v1): `contracts/api/jsonschema/shopping_supplier_signal_list_v1.schema.json`, `shopping_supplier_signal_v1.schema.json`, `shopping_upsert_supplier_signal_request_v1.schema.json`
- Events: none required in v1
- Governance: none required in v1

## Implementation checklist

- UI edits must write through the Shopping API, not localStorage.
- Worker must treat manual overrides as authoritative unless marked stale/invalid by checks.
- Logs must avoid leaking URLs with sensitive query params.

## Consequences

- The "Configurar URLs" legacy capability has a clean backend-owned home.
- Scraping becomes faster and more reliable without embedding heuristics in the UI.
- Future analytics can use these signals as inputs for supplier quality scoring.

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Data model + RLS review: `metalshopping-platform-packages`
- Contract authoring for manual URL management surfaces: `metalshopping-openapi-contracts`
- Worker persistence pattern: `metalshopping-worker-scaffold`
- Security review (sensitive data in URLs/logs): `metalshopping-observability-security`
