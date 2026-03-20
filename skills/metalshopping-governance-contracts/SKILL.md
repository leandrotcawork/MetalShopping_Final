---
name: metalshopping-governance-contracts
description: Create or review MetalShopping governance contracts under contracts/governance using repo templates, runtime hierarchy rules, and shared semantics for policies, thresholds, and feature flags. Use when implementing or revising governance artifacts.
---

# MetalShopping Governance Contracts

## Workflow
1. Read `docs/CONTRACT_CONVENTIONS.md` + `docs/adrs/ADR-0004-runtime-governance.md`
2. Confirm artifact type: policy | threshold | feature_flag
3. Start from matching template in `contracts/governance/`
4. Keep scope hierarchy explicit: global → environment → tenant → module → entity
5. Finish with `references/governance-checklist.md`

## Rules
- Artifacts only under `contracts/governance/policies/`, `thresholds/`, `feature_flags/`
- Do not encode business thresholds in app code — use governed artifacts
- Semantics consistent for both Go and Python consumers
- Treat as auditable platform contracts

## References
- Templates: `contracts/governance/`
- ADR: `docs/adrs/ADR-0004-runtime-governance.md`
- Checklist: `references/governance-checklist.md`
