# ADR-0026: Suppliers Management API Surface Split (Phase 2)

- Status: draft
- Date: 2026-03-19

## Context

ADR-0019 freezes Suppliers Directory and Driver Manifests as backend-owned, tenant-scoped data.

In Level 1/early Level 2, Shopping can consume suppliers through `GET /api/v1/shopping/bootstrap` without exposing a full management API. As soon as Procurement, Integrations, or Admin surfaces need to manage suppliers/manifests, keeping this embedded inside `shopping_v1` becomes an ownership and evolution risk.

## Decision

When supplier/manifests management is required by more than Shopping, we split the management surface:

- introduce a dedicated bounded context API contract `suppliers_v1`
- Shopping continues to consume supplier options, but does not own supplier lifecycle

Rules:

- Supplier identity/config is owned by the Suppliers module (not Shopping UI)
- Shopping bootstrap may still embed a lightweight view of enabled suppliers for workflow UX
- Driver manifest validation/activation follows ADR-0027

## Contracts (touchpoints)

- OpenAPI (new, Phase 2):
  - `contracts/api/openapi/suppliers_v1.openapi.yaml`
  - includes read + management endpoints for:
    - directory list/upsert/enablement
    - driver manifest create/validate/activate
- JSON Schemas:
  - `contracts/api/jsonschema/suppliers_*.schema.json` for the new API payloads
- Events: optional later (supplier enabled/manifest activated), not required for initial split
- Governance: optional (policy gating who can modify manifests), not required for baseline

## Implementation checklist

- Contract authoring: `metalshopping-openapi-contracts` (+ `metalshopping-contract-authoring` when schema reuse is required)
- Go module:
  - new `apps/server_core/internal/modules/suppliers` (or evolve existing if already created)
  - RLS enforced via `app.tenant_id` and tenant-scoped queries
  Skills: `metalshopping-module-scaffold`, `metalshopping-server-core-modules`
- UI:
  - Shopping consumes enabled suppliers from Shopping bootstrap (or via platform-sdk composition), never hardcoded
  Skills: `metalshopping-frontend-migration-guardrails`, `metalshopping-page-delivery`
- SDK generation: `metalshopping-sdk-generation`

## Consequences

- Supplier operations become reusable across Shopping, Procurement, and Integrations.
- Contract evolution becomes cleaner (Shopping contract stays workflow-focused).
- Admin UI can manage suppliers/manifests without leaking Shopping concerns.

