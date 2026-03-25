---
id: lesson-0031
title: Remove ts-nocheck with minimal explicit callback typing
region: lessons
tags: [frontend, typescript, legacy-migration, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0031 — Remove ts-nocheck with minimal explicit callback typing

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Keeping migrated files compile-green by disabling type checks globally (`// @ts-nocheck`).

## Correct
Re-enable type checking and add explicit callback parameter typing at integration boundaries where DTOs are intentionally loose.

## Rule
Prefer localized type annotations over file-level typecheck suppression in migrated frontend modules.

## Impact
**Medium:** Improves type safety without broad suppression.
