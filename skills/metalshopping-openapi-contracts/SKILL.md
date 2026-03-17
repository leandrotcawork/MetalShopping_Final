---
name: metalshopping-openapi-contracts
description: Create or review MetalShopping OpenAPI contracts under `contracts/api/openapi` using the repo template, naming rules, metadata, schema reuse, versioning policy, and thin-client contract discipline. Use when implementing or revising a bounded-context HTTP contract such as `iam_v1.openapi.yaml`, checking OpenAPI consistency against repo SoT, or standardizing a new public API surface for MetalShopping.
---

# MetalShopping OpenAPI Contracts

## Overview

Use this skill to create or review OpenAPI files for MetalShopping with the repository's contract-first standards. Keep the work anchored to the repo template and contract conventions instead of inventing per-file structure.

## Workflow

1. Read only the minimum repo context:
   `docs/PROJECT_SOT.md`
   `docs/CONTRACT_CONVENTIONS.md`
   `contracts/api/openapi/_template.openapi.yaml`
2. Confirm the bounded context and target filename under `contracts/api/openapi/`.
3. Start from the template structure instead of writing a spec from scratch.
4. Reuse shared JSON Schemas from `contracts/api/jsonschema/` when possible.
5. Keep versioning explicit in path and contract metadata.
6. Finish with the review checklist in `references/openapi-checklist.md`.

## Contract rules

- Create one bounded-context OpenAPI file per public API surface unless a shared root spec is clearly justified.
- Use lowercase snake_case filenames such as `iam_v1.openapi.yaml`.
- Keep `x-metalshopping` metadata present and meaningful.
- Prefer additive evolution over breaking change.
- Do not create manual parallel contract definitions in app code.
- Keep frontend consumption aligned with generated SDK flow.

## Output expectations

When creating or updating a contract:

- preserve the repository naming convention
- preserve explicit versioning
- use shared schemas where possible
- keep paths, operations, and payload meaning stable and reviewable
- note any required companion schema files in `contracts/api/jsonschema/`

## References

- For the exact repo workflow and file touchpoints, read `references/repo-openapi-flow.md`.
- For the final review pass, read `references/openapi-checklist.md`.

