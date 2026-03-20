# ADR-0044: Catalog Import Idempotency and Identifier Conflicts v1

- Status: accepted
- Date: 2026-03-20

## Context

Legacy `products` data is not guaranteed to satisfy the new canonical constraints without an explicit policy.

The new core schema has stronger invariants:

- `catalog_products` requires `sku` and `name` (non-empty).
- `catalog_product_identifiers` enforces uniqueness:
  - `(tenant_id, identifier_type, identifier_value)` is unique.
  - only one `is_primary=true` per `(tenant_id, product_id, identifier_type)`.

Legacy commonly contains:

- missing/blank descriptions
- duplicated EANs or duplicated references across different internal product numbers
- taxonomy links that are null or point to inactive/missing nodes

If we do not freeze a policy, the importer will either:

- fail halfway through (non-deterministic retries), or
- silently “fix” data in inconsistent ways (creating new debt)

## Decision

Define import-time guardrails so the pipeline in ADR-0043 is deterministic, repeatable, and produces explicit conflict evidence.

### 1) Deterministic IDs (import-owned)

The importer must generate stable identifiers for imported rows to make re-imports idempotent.

Rules:

- Destination `catalog_products.product_id` is deterministic from legacy `pn_interno`.
  - Example strategy (v1): `prd_<sha256(lower(trim(pn_interno)))[:24]>`.
- Destination `catalog_taxonomy_nodes.taxonomy_node_id` is deterministic from legacy numeric `taxonomy_nodes.id`.
  - Example strategy (v1): `tx_<id>`.
- Destination `catalog_product_identifiers.product_identifier_id` is deterministic from `(product_id, identifier_type, identifier_value)`.
  - Example strategy (v1): `pid_<sha256(product_id + '|' + type + '|' + value)[:24]>`.

Notes:

- This does not change the `server_core` runtime product creation strategy (which may remain random for user-created products).
- Deterministic IDs are required only for the import pipeline to be repeatable.

### 2) Required-field fallback rules

The importer must never create an invalid product row.

Rules:

- `sku` always comes from `pn_interno` and must be non-empty.
- `name` fallback order:
  1. `descricao` (trimmed)
  2. `reference` (trimmed)
  3. `ean` (trimmed)
  4. `pn_interno` (trimmed)
- `description`:
  - set to legacy `descricao` when present; otherwise null.
- `brand_name`:
  - set when present; otherwise null.
- `stock_profile_code`:
  - set from legacy `tipo_estoque` when present; otherwise null.

### 3) Identifier insertion policy (conflict-safe)

The importer must preserve the strongest signal (`pn_interno`) even when other identifiers conflict.

Rules:

- Always insert identifier:
  - `identifier_type='pn_interno'`, `identifier_value=pn_interno`, `is_primary=true`
- Insert `reference` and `ean` only when:
  - the value is non-empty, and
  - there is no existing row with the same `(tenant_id, identifier_type, identifier_value)` assigned to a different product

Conflict behavior:

- If `reference` conflicts:
  - skip inserting the conflicting `reference` identifier
  - write a row to the conflict report (see next section)
- If `ean` conflicts:
  - skip inserting the conflicting `ean` identifier
  - write a row to the conflict report

This policy keeps the import running and makes conflicts explicit for later human resolution, without weakening canonical constraints.

### 4) Reporting (required)

The importer must emit explicit reports (CSV/JSON) for:

- duplicated EAN values (with the set of `pn_interno` involved)
- duplicated reference values (with the set of `pn_interno` involved)
- products with missing taxonomy assignment
- products with taxonomy_node_id pointing to missing nodes (integrity failure)

Reports are part of acceptance evidence for ADR-0043.

## Implementation Checklist

Frozen execution order for this ADR:

1. Finalize the deterministic ID scheme and conflict policy (this ADR)
   Skill: `metalshopping-adr-updates`
2. Implement importer with:
  - deterministic IDs
  - conflict-safe identifier insertions
  - reporting output
3. Validate in a full import run and attach evidence

## Acceptance Evidence (for Status: accepted)

- Import can be rerun twice with the same inputs and produces:
  - the same counts
  - the same deterministic IDs for the same legacy keys
  - no duplicate rows created
- Conflict reports exist and are non-empty only when legacy data truly conflicts.
- Spot-check:
  - random sample of 20 products contains correct `pn_interno/reference/ean` placement and primary flags.

Evidence captured on 2026-03-20:

- Rerun summary (no reset):
  - `levels=3`, `nodes=100`, `products_imported=3838`
  - `identifier_conflicts=59`, `missing_taxonomy_rows=1`
  - reports: `.tmp/import_catalog_reports/20260320_210414`

## Alternatives considered

- Relax uniqueness constraints in the new schema:
  - rejected; canonical constraints are intentional and protect downstream modules.
- Force-merge duplicates (choose one product to “own” an EAN/reference):
  - rejected for v1; it encodes business decisions implicitly and is not safe without explicit governance.

## Consequences

- We can import the full legacy universe without weakening the new canonical model.
- Data issues become visible and actionable instead of silently corrupting identity resolution.
- The import pipeline becomes idempotent and safe to rerun during development.
