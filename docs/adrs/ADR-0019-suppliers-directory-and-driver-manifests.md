# ADR-0019: Suppliers Directory and Driver Manifests as Tenant-Scoped Data

- Status: accepted
- Date: 2026-03-19

## Context

The legacy Shopping core relies on a supplier registry plus an optional manifest-driven runtime:

- suppliers have code/label/kind/lookup policy
- manifests describe the execution family and config (HTML search, VTEX JSON, Playwright PDP-first, etc)
- manifests can be versioned and validated

In the target platform, suppliers are not a Shopping-only concern. Suppliers will also feed procurement, integrations control, and market intelligence over time. Supplier identity and operational configuration must be backend-owned, tenant-aware, and auditable.

## Decision

We will model suppliers and driver configuration as tenant-scoped data owned by `server_core`:

- `apps/server_core/internal/modules/suppliers` owns the directory and driver configuration
- Shopping reads supplier options from Suppliers; Shopping does not hardcode suppliers in the UI

The Suppliers directory must capture, at minimum:

- `supplier_code` (canonical)
- `supplier_label` (UI-facing)
- `execution_kind` (`HTTP` or `PLAYWRIGHT`)
- `lookup_policy` (EAN vs reference precedence)
- `enabled` flag (tenant-specific enablement)

Driver manifests must be stored as data (not hand-edited generated code) and support:

- `family` (executor family)
- `config_json` (family-specific config)
- validation state and errors
- version history

Worker runtime rule:

- prefer manifests stored in Postgres when available and enabled
- allow code registry fallback only for bootstrap-local development until the manifest table is seeded

## Contracts (touchpoints)

- Data model (tenant-scoped + RLS):
  - `apps/server_core/migrations/0022_suppliers_directory_and_driver_manifests.sql`
- OpenAPI (v1 baseline):
  - Suppliers may be returned embedded in `GET /api/v1/shopping/bootstrap` for workflow selection.
  - Dedicated management surfaces (create/update/activate manifest) may later be added as either:
    - an extension of `shopping_v1` (operational-only, minimal), or
    - a separate `contracts/api/openapi/suppliers_v1.openapi.yaml` (recommended when procurement/integrations start consuming it).
- JSON Schemas: only required when exposing suppliers/manifests through an OpenAPI surface.
- Governance: none required initially; enablement is tenant-owned data (`enabled`).
- Events: optional future; not required for v1 execution.

## Implementation checklist

- Keep Shopping UI free of hardcoded supplier lists.
- Seed a minimal directory + a first manifest version for bootstrap-local environments.
- Ensure worker reads the directory/manifests under tenant context (ADR-0022).

## Consequences

- Shopping "Configurar" can be driven by real backend-owned supplier state.
- Supplier enablement and driver configuration become auditable and evolvable.
- The worker can run with stable semantics across environments (no hardcoded supplier lists in UI).

## Follow-up (Skills)

- ADR governance and sync: `metalshopping-adr-updates`
- Suppliers contract authoring (read surface): `metalshopping-openapi-contracts` and `metalshopping-contract-authoring`
- Suppliers module scaffold: `metalshopping-module-scaffold`
- Worker integration with manifests: `metalshopping-worker-patterns` and `metalshopping-worker-scaffold`
- Observability and security review: `metalshopping-observability-security`
