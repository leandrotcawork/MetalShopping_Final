---
name: metalshopping-legacy-migration
description: Migrate MetalShopping legacy frontend modules with a visual-parity-first workflow: literal copy, runnable mocks/shims, parity validation, and only then backend/SDK adaptation.
---

# MetalShopping Legacy Migration

Read first: `tasks/lessons.md`, `tasks/todo.md`, `AGENTS.md`.

## Core principle

When request is “copy legacy”, “visual first”, or “deixar igual”, freeze visual baseline first and avoid premature redesign/integration.

## Required sequence

1. **Inventory**
   - identify legacy entrypoint, CSS, shared widgets, payload keys
2. **Freeze must-match**
   - shell/header, tabs, first fold, card hierarchy, spacing rhythm, states
3. **Literal copy**
   - preserve DOM hierarchy and class boundaries
4. **Compatibility shims**
   - local mocks/adapters/context helpers to make page compile/run
5. **Styling prerequisites**
   - token fallbacks for critical surfaces and borders
   - same shell wrapper structure as legacy
6. **Parity validation**
   - structure → data shape → CSS/tokens → interactions/animations
7. **Only after sign-off**
   - contracts/backend/SDK integration and gradual mock replacement

## Non-negotiables

- do not start backend integration before visual sign-off on visual-first requests
- do not infer parity from screenshots only when source HTML/CSS/TSX exists
- do not “clean rewrite” before a runnable matching baseline exists
- do not leave copied pages with missing payload keys
- do not hide essential UI behind animation default states

## Blank/flat screen diagnostics (ordered)

1. DTO exists and matches expected keys?
2. Shell route renders the intended page (not fallback/MVP block)?
3. Critical CSS tokens have fallback values?
4. Overlay/drawer/backdrop state is reset on tab switch?
5. Chart/canvas lifecycle guarded in StrictMode?
6. Runtime exception swallowed by shell error boundary?

## Deliverable format

For complex migrations, keep explicit phases:
- T5-A inventory + baseline
- T5-B literal copy
- T5-C shims/adapters
- T5-D visual parity pass
- T5-E remaining visual sections with mocks
- T1/T2/T4 only after sign-off

## Lessons policy for migration work

- `tasks/lessons.md` receives only structural/global lessons
- page-specific spacing/color tweaks stay in feature notes or PR description
- if a migration issue repeats across pages, then promote it to a lesson
