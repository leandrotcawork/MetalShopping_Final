---
id: lesson-0025
title: Delete orphan facade files when usage hits zero
region: lessons
tags: [frontend, refactoring, correction]
links:
  - cortex/frontend/index
severity: low
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0025 — Delete orphan facade files when usage hits zero

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Leaving unused pass-through facade modules in the tree after import migration.

## Correct
Remove zero-reference facades immediately once consumers are moved.

## Rule
Dead facade files create drift risk and must not remain after cleanup.

## Impact
**Low:** Reduces cognitive load and prevents accidental imports.
