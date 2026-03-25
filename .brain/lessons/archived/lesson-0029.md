---
id: lesson-0029
title: Legacy copy must be normalized to local DTO shapes
region: lessons
tags: [frontend, legacy-migration, types, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0029 — Legacy copy must be normalized to local DTO shapes

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Keeping legacy field names (`x`, `our`, `competitors`) and raw `Record<string, unknown>` values in typed code, causing runtime-safe but compile-broken paths.

## Correct
Map legacy fields to current DTO contract (`date`, `our_price`, `suppliers`) and normalize unknown payload values with explicit converters before rendering.

## Rule
In parity migrations, literal copy is allowed, but every boundary to local contracts/types must be normalized explicitly.

## Impact
**Medium:** Prevents type safety violations in migrated code.
