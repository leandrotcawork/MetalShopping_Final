# ADR-0027: Shopping Driver Manifest Validation And Activation (Phase 2)

- Status: draft
- Date: 2026-03-19

## Context

ADR-0019 introduces driver manifests stored as tenant-scoped data:

- versioned
- includes `family` + `config_json`
- has validation state and errors

Without a frozen validation/activation workflow, manifests will become either:

- "data that is never trusted" (so code fallback persists forever), or
- "data that can break execution silently" (unsafe operationally)

The legacy system implicitly relied on code-level drivers. The target system must make manifest activation safe and auditable.

## Decision

Manifests follow an explicit lifecycle:

- `pending` on creation
- validated deterministically against the executor family's rules
- only `valid` manifests can be activated (`is_active=true`)
- exactly one active manifest per (tenant, supplier_code)

Validation ownership:

- validation runs in `server_core` (deterministic, audit-friendly) using a family registry
- worker uses only `valid` + `active` manifests at runtime

Rules:

- activation is an explicit action (not implicit on write)
- validation errors are persisted (not only logged)
- code fallback is allowed only for bootstrap-local until manifests are seeded and validated

## Contracts (touchpoints)

- Data model: `apps/server_core/migrations/0022_suppliers_directory_and_driver_manifests.sql`
- OpenAPI:
  - if embedded in Shopping: add explicit management endpoints to `shopping_v1` (not preferred long-term)
  - if split (ADR-0026): define endpoints in `suppliers_v1`:
    - create manifest version
    - validate manifest
    - activate manifest
- JSON Schemas:
  - per-family config schema is introduced as versioned JSON schema:
    - `contracts/api/jsonschema/supplier_driver_manifest_<family>_v1.schema.json`
  - management response includes validation errors in a stable shape
- Events: not required for baseline
- Governance: optional later for RBAC/policy gating

## Implementation checklist

- Define family registry and validators: `metalshopping-platform-packages` (registry) + `metalshopping-module-scaffold` (module glue)
- Expose validation/activation endpoints: `metalshopping-openapi-contracts` + `metalshopping-module-scaffold`
- Worker runtime uses only active/valid manifests: `metalshopping-worker-scaffold`
- Observability: log validation and activation actions with correlation ids: `metalshopping-observability-security`
- SDK generation: `metalshopping-sdk-generation`

## Consequences

- Manifests become safe operational configuration, not "dangerous JSON blobs".
- Bootstrap-local can start from code fallback but converges to data-driven config.
- Later procurement/integration modules can trust supplier driver configuration safely.

