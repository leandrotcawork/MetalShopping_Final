---
id: lesson-0024
title: Remove local wrappers after migration to shared UI
region: lessons
tags: [frontend, refactoring, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0024 — Remove local wrappers after migration to shared UI

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Keeping dead wrapper files after switching to shared UI imports.

## Correct
Delete redundant local wrappers to prevent accidental regressions to non-standard components.

## Rule
Migration to shared UI is complete only when obsolete wrapper paths are removed.

## Impact
**Medium:** Prevents code rot and accidental usage of outdated wrappers.
