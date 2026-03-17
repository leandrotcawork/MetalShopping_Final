---
name: metalshopping-adr-updates
description: Update MetalShopping ADRs, SoT, and planning documents when a structural decision changes. Use when a new architectural rule must be frozen, when an existing rule changes, or when `PROJECT_SOT`, implementation planning, and ADRs must stay synchronized without duplicating or drifting platform decisions.
---

# MetalShopping ADR Updates

## Overview

Use this skill to keep architecture decisions and planning documents synchronized whenever the platform model changes.

## Workflow

1. Read only the minimum repo context:
   `docs/PROJECT_SOT.md`
   `docs/IMPLEMENTATION_PLAN.md`
   `docs/PROGRESS.md`
   relevant files under `docs/adrs/`
2. Confirm whether the change is a new decision or a modification of an accepted rule.
3. Update ADRs first when the rule itself changes.
4. Update SoT and planning docs to reflect the accepted state.
5. Finish with the review checklist in `references/adr-checklist.md`.

## References

- For the repo workflow and file touchpoints, read `references/repo-adr-flow.md`.
- For the final review pass, read `references/adr-checklist.md`.

