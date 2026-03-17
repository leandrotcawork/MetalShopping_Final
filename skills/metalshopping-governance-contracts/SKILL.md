---
name: metalshopping-governance-contracts
description: Create or review MetalShopping governance contracts under `contracts/governance` using the repo templates, runtime hierarchy rules, metadata conventions, and shared semantics for policies, thresholds, and feature flags. Use when implementing or revising governance artifacts, checking override scope, or standardizing runtime-governance contracts against the repo SoT and ADRs.
---

# MetalShopping Governance Contracts

## Overview

Use this skill to create or review MetalShopping governance contracts with the repository's runtime-governance standards. Keep the work anchored to the repo templates and the accepted resolution hierarchy instead of inventing ad hoc configuration models.

## Workflow

1. Read only the minimum repo context:
   `docs/PROJECT_SOT.md`
   `docs/CONTRACT_CONVENTIONS.md`
   `docs/adrs/ADR-0004-runtime-governance.md`
2. Confirm the governance artifact type:
   `contracts/governance/policies/`
   `contracts/governance/thresholds/`
   `contracts/governance/feature_flags/`
3. Start from the matching repository template instead of drafting governance JSON ad hoc.
4. Keep scope hierarchy explicit and aligned with runtime semantics.
5. Reference shared schemas in `contracts/api/jsonschema/` whenever validation or reuse is needed.
6. Finish with the review checklist in `references/governance-checklist.md`.

## Contract rules

- Governance contracts belong only under `contracts/governance/`.
- Keep artifact type explicit: policy, threshold, or feature flag.
- Preserve the accepted resolution hierarchy:
  `global`, `environment`, `tenant`, `module`, `entity/profile`, `feature-target`
- Do not encode hardcoded business thresholds in app code when they should be governed artifacts.
- Keep semantics consistent for both Go and Python consumers.
- Treat governance artifacts as auditable platform contracts.

## Output expectations

When creating or updating a governance contract:

- preserve the repository naming and metadata conventions
- preserve explicit versioning
- keep scope and resolution behavior explicit
- note any required companion schema files in `contracts/api/jsonschema/`
- keep the artifact understandable for both core and worker runtime consumers

## References

- For the exact repo workflow and file touchpoints, read `references/repo-governance-flow.md`.
- For the final review pass, read `references/governance-checklist.md`.

