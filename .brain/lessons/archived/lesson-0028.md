---
id: lesson-0028
title: Do not downgrade legacy charts in parity phase
region: lessons
tags: [frontend, legacy-migration, components, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0028 — Do not downgrade legacy charts in parity phase

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Replacing legacy chart components with simplified SVG placeholders during visual migration.

## Correct
Preserve legacy chart implementation in parity phase and adapt only integration imports when required.

## Rule
Visual parity includes chart behavior and interaction fidelity, not only static layout.

## Impact
**Medium:** Prevents functionality loss during legacy migration.
