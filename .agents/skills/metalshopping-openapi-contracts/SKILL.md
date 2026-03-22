---
name: metalshopping-openapi-contracts
description: Create or review MetalShopping OpenAPI contracts under contracts/api/openapi using the repo template, naming rules, schema reuse, and versioning policy. Called by $ms for T1 of every feature.
---

# MetalShopping OpenAPI Contracts

## Workflow
1. Read `docs/CONTRACT_CONVENTIONS.md` + `docs/SDK_GENERATION_STRATEGY.md`
2. Start from `contracts/api/openapi/_template.openapi.yaml`
3. Reuse schemas from `contracts/api/jsonschema/` when they exist
4. Finish with `references/openapi-checklist.md`

## Rules
- One bounded-context file: `contracts/api/openapi/<domain>_v1.openapi.yaml`
- Lowercase snake_case filenames
- `x-metalshopping` metadata present and meaningful
- Path versioning explicit: `/api/v1/<resource>`
- Prefer additive evolution — never break consumers without version bump
- Shared payloads in `contracts/api/jsonschema/` (not duplicated inline)

## References
- Template: `contracts/api/openapi/_template.openapi.yaml`
- Shared schemas: `contracts/api/jsonschema/`
- Checklist: `references/openapi-checklist.md`
