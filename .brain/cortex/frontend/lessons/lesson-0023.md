---
id: lesson-0023
title: Feature code must import shared UI from package entrypoint
region: lessons
tags: [frontend, ui-reuse, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0023 — Feature code must import shared UI from package entrypoint

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Importing duplicated/local UI wrappers instead of the registered shared component.

## Correct
Import shared controls from `@metalshopping/ui` (`packages/ui/src/index.ts`) for consistent behavior and styling.

## Rule
If a component exists in `packages/ui`, feature modules consume it via package entrypoint.

## Impact
**Medium:** Prevents UI duplication and ensures consistency.
