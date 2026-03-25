---
id: lesson-0015
title: Legacy migration must preserve interactive behavior
region: lessons
tags: [frontend, legacy-migration, correction]
links:
  - cortex/frontend/index
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0015 — Legacy migration must preserve interactive behavior

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Replacing legacy interactive table logic with a simplified static version during visual migration.

## Correct
Keep legacy interaction model (filters, sorting, pagination, state callbacks) and only adapt imports/wiring.

## Rule
In parity-first migration, functional UX behavior is part of visual fidelity and cannot be downgraded.

## Impact
**High:** Prevents functionality regression during legacy migration.
