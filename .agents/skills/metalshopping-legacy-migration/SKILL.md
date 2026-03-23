---
name: metalshopping-legacy-migration
description: Migrate MetalShopping legacy frontend modules into the new app with a legacy-first visual parity workflow. Use when the goal is to copy a legacy page or flow first, make it runnable with mocks and adapters, reach 1:1 visual parity, and only then replace mocks with contracts, SDK, and real backend integration.
---

# MetalShopping Legacy Migration

Read first: `tasks/lessons.md`, `tasks/todo.md`, `AGENTS.md`.

## Core rule

Freeze the visual baseline before redesigning anything.

If the user says "copy legacy", "leave identical", "visual first", or provides a legacy HTML/CSS/TSX source, use this skill before normal frontend adaptation.

## Migration sequence

1. Inventory the legacy source:
   - page entrypoints
   - CSS modules/global CSS
   - shared widgets
   - app/session dependencies
   - payload shape expected by the viewmodel
2. Freeze the must-match surface:
   - shell/header/tab rail
   - first fold
   - card hierarchy
   - spacing rhythm
   - states/chips/badges
3. Copy literal structure first:
   - preserve markup hierarchy
   - preserve class boundaries
   - preserve tab labels and IA ordering
   - do not simplify layout before parity
4. Add compatibility shims so the copy runs:
   - local `AppSessionProvider` or equivalent
   - mock DTOs shaped like legacy payloads
   - adapters for registry/text helpers
   - temporary UI wrappers only when existing components block compile/runtime
5. Restore styling prerequisites:
   - local token defaults for legacy CSS variables
   - shell container structure identical to legacy (`header` vs inner wrapper, rails, pills)
   - animation fallbacks must not hide content by default
6. Validate parity in this order:
   - structure
   - data shape
   - CSS/tokens
   - animation/state polish
7. Only after visual sign-off:
   - write/read contracts
   - implement backend
   - regenerate SDK
   - replace mocks incrementally without changing the visual shell

## Non-negotiable rules

- Do not start with backend integration if the task is explicitly visual-first.
- Do not judge parity from screenshots alone if legacy source files exist; inspect the actual legacy HTML/CSS/TSX.
- Do not replace the copied page with a clean rewrite before a matching baseline exists.
- Do not leave copied components starved of data; if real data is not ready, mock the exact payload keys the legacy viewmodel expects.
- Do not rely on animation to reveal essential UI; default state must be visible.

## Architecture check

State only:
- module type: `frontend-only` now, `read-only` later
- exact files/folders to change
- dependencies that require shims
- what is parity-critical now vs deferred integration later

## Required diagnostics when something looks blank or flat

Check in this order:

1. Is the copied page receiving the DTO it expects?
2. Are payload keys aligned with what the viewmodel reads?
3. Are token variables like `--surface`, `--surface-border`, `--radius-lg`, `--grid-gap` defined in scope?
4. Is shell markup structure identical to legacy (`header > inner`, tab rail wrapper, right controls)?
5. Are any elements hidden by default with `opacity: 0` outside keyframes?
6. Is a runtime exception being swallowed by the route shell?

## Deliverable pattern

For a complex migration, keep phases explicit:

- T5-A inventory + freeze baseline
- T5-B literal copy
- T5-C compatibility adapters
- T5-D visual parity pass
- T5-E remaining visual sections with mocks
- T1/T2/T4 only after visual sign-off

## References

- `references/migration-checklist.md` - field checklist and common failure modes
- `../metalshopping-frontend/references/migration-rules.md` - frontend adaptation rules after parity
