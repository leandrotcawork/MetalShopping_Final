---
name: metalshopping-architecture-direction-review
description: Review MetalShopping implementation, structure, and roadmap alignment against the repository's frozen architecture, principles, ADRs, and future product goals. Use when checking whether current code or plans are drifting toward shortcuts, weak boundaries, poor scalability, unsafe coupling, or under-modeled foundations instead of the professional long-term target for analytics, CRM, procurement/buying, adaptive integrations, governance, and multi-tenant platform growth.
---

# MetalShopping Architecture Direction Review

## Overview

Use this skill to perform a senior-level architectural review against the MetalShopping target state. Treat it as a direction-and-quality review, not a syntax or style pass. Favor truth over convenience.

## Workflow

1. Read only the minimum frozen context:
   `ARCHITECTURE.md`
   `docs/PROJECT_SOT.md`
   `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`
   `docs/IMPLEMENTATION_PLAN.md`
   `docs/DATA_CONTRACT_MAP.md`
   `docs/SYSTEM_PRINCIPLES.md`
   `docs/ENGINEERING_PRINCIPLES.md`
   `docs/PLATFORM_BOUNDARIES.md`
   `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
   `docs/CONTRACT_EVOLUTION_RULES.md`
2. Load only the ADRs and canonical-model documents relevant to the area under review.
3. If legacy reuse is involved, read `docs/METALDOCS_REUSE_MATRIX.md` before trusting legacy patterns.
4. Review the target through the lenses in `references/review-lenses.md`.
5. Produce findings first, ordered by severity, with concrete file references and direct consequences.
6. End with a short verdict:
   `aligned`
   `partially aligned`
   `misaligned`
7. Recommend the next best move toward the long-term platform, not the fastest patch.

## Review rules

- review for future-fit, not just current correctness
- reject shortcuts that weaken modular boundaries, tenancy, governance, contracts, or data ownership
- prefer explicit canonical models over hidden convenience fields
- prefer contract-first and platform-first moves over local tactical wins
- call out when a change helps today but harms future modules such as analytics, CRM, procurement, market intelligence, or adaptive integrations
- do not rubber-stamp code that is merely functional
- make tradeoffs explicit when the repo is intentionally taking a phased approach

## Output expectations

When using this skill:

- findings come first
- cite the frozen repo rules being violated or satisfied
- distinguish temporary bootstrap choices from structural debt
- state whether the reviewed slice improves or weakens the target architecture
- recommend the next best structural step, not just the next coding task

## References

- For the core review checklist, read `references/review-lenses.md`.
- For a fast repo navigation path, read `references/repo-review-flow.md`.
