# ADR-0039: Shopping Manual URL Panel v1 (Operational Table)

- Status: accepted
- Date: 2026-03-20

## Context

Legacy Shopping provides a dense operational panel for manual URL management:

- filters
- pagination
- table density
- inline editing and "open URL" behavior

The target Shopping page currently offers a small "single row" form, which is not sufficient for the real workflow and does not exploit the backend-owned `shopping_supplier_product_signals` model introduced by ADR-0024 and evolved by ADR-0035.

## Decision

Manual URL management becomes an explicit operational panel with a table-first UX:

- A dedicated "Manual URLs" panel exists inside Step 2 (Configurar).
- The default representation is a paginated table.
- Row fields: product identity (at least `productId` with a clear affordance to copy), supplier code, product URL (editable), URL status, manual override state.
- Row fields: URL lifecycle visibility (`nextDiscoveryAt`, `notFoundCount`) for operational clarity.
- Data source: viewing and filtering signals via `GET /api/v1/shopping/supplier-signals`.
- Persistence: applying changes via `PUT /api/v1/shopping/supplier-signals`.

Contract behavior remains unchanged in v1. Bulk edits, if needed, will be introduced only by a dedicated follow-up ADR with an explicit batch contract.

## Contracts (touchpoints)

- `contracts/api/openapi/shopping_v1.openapi.yaml`
- `contracts/api/jsonschema/shopping_supplier_signal_v1.schema.json`
- `contracts/api/jsonschema/shopping_upsert_supplier_signal_request_v1.schema.json`

## Implementation Checklist

- Implement a table-based panel consistent with the legacy density and layout.
- Ensure the panel is SDK-driven only; no manual DTOs.
- Hide advanced lifecycle fields behind a compact disclosure if needed, but keep them available.
- Keep editing deterministic and auditable (manual override always set on save).

## Acceptance Evidence (for Status: accepted)

- Manual URL panel matches legacy intent with a table-first operational UX.
- Table can filter by supplier and product id and can persist manual URL changes via SDK.
- The panel displays lifecycle fields from ADR-0035 without leaking internal DB details.
- `npm.cmd run web:typecheck` passes.
- `npm.cmd run web:build` passes.

## Consequences

- Manual URL management becomes scalable as the catalog grows.
- Operational debugging becomes easier because cooldown state is visible in the UI.
