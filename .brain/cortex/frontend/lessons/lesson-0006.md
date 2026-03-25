---
id: lesson-0006
title: Reuse design system before adding UI primitives
region: lessons
tags: [frontend, ui, design-system, baseline]
links:
  - cortex/frontend/index
  - hippocampus/conventions
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0006 — Reuse design system before adding UI primitives

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Frontend

## Wrong
Creating local UI components already available in `packages/ui`.

## Correct
Check `packages/ui/src/index.ts` first; extend shared primitives when needed.

## Rule
Prefer shared UI primitives over feature-local duplication.

## Impact
**High:** Prevents UI inconsistency and reduces maintenance burden.
