---
name: metalshopping-sdk-generation
description: Plan, create, or review MetalShopping generated SDK and type workflows using `contracts/` as the single source of truth and `packages/generated/*` as downstream outputs. Use when defining or revising generation flows for `sdk_ts`, `types_ts`, or `sdk_py`, reviewing generation boundaries, or checking that frontend and worker consumers stay aligned with the repo SDK generation strategy.
---

# MetalShopping SDK Generation

## Overview

Use this skill to define or review how MetalShopping generates SDKs and shared types from contracts. Keep the work anchored to the repository generation strategy and output boundaries instead of inventing ad hoc generation flows.

## Workflow

1. Read only the minimum repo context:
   `docs/SDK_GENERATION_STRATEGY.md`
   `docs/CONTRACT_CONVENTIONS.md`
   `packages/generated/README.md`
   `scripts/generate_contract_artifacts.ps1`
2. Confirm the generation target:
   `packages/generated/sdk_ts/`
   `packages/generated/types_ts/`
   `packages/generated/sdk_py/`
3. Confirm the source contract set in `contracts/`.
4. Keep generated outputs downstream only; do not treat them as authoring sources.
5. Keep generation reproducible, scriptable, and all-or-nothing per target where possible.
6. Finish with the review checklist in `references/sdk-generation-checklist.md`.

## Generation rules

- `contracts/` is the only source of truth.
- `packages/generated/*` contains generated outputs only.
- frontend consumes generated TypeScript SDKs and types.
- workers consume generated Python models or SDKs where contract-driven interaction exists.
- do not create manual fallback type systems that compete with generated outputs.
- keep generation workflows in `scripts/`.

## Output expectations

When defining or reviewing a generation workflow:

- preserve source-versus-output boundaries
- make target ownership explicit
- keep the workflow reproducible from scripts
- keep versioning behavior explicit
- ensure the result supports frontend and worker consumers without parallel contract drift

## References

- For the exact repo workflow and file touchpoints, read `references/repo-sdk-flow.md`.
- For the final review pass, read `references/sdk-generation-checklist.md`.

