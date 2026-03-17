---
name: metalshopping-contract-authoring
description: Coordinate MetalShopping contract work that spans API, event, governance, and generated artifact concerns using the repo contract conventions and evolution rules. Use when a feature requires more than one contract type, when checking contract completeness across folders, or when deciding which contract artifacts must be added before implementation.
---

# MetalShopping Contract Authoring

## Overview

Use this skill when a feature or capability touches multiple contract surfaces and you need a single disciplined workflow instead of treating API, events, and governance separately.

## Workflow

1. Read only the minimum repo context:
   `docs/CONTRACT_CONVENTIONS.md`
   `docs/CONTRACT_EVOLUTION_RULES.md`
   `docs/SDK_GENERATION_STRATEGY.md`
2. Identify which contract types are required:
   `api/openapi`
   `api/jsonschema`
   `events`
   `governance`
3. Determine whether generation targets are affected.
4. Ensure ownership, naming, and compatibility are explicit.
5. Finish with the review checklist in `references/contract-authoring-checklist.md`.

## Contract rules

- `contracts/` remains the only source of truth
- multi-surface changes should be coordinated, not discovered late
- generated outputs are downstream consequences, not design inputs

## References

- For the repo workflow and file touchpoints, read `references/repo-contract-flow.md`.
- For the final review pass, read `references/contract-authoring-checklist.md`.

