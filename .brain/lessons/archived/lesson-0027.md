---
id: lesson-0027
title: Second-pass parity must diff against legacy snapshot
region: lessons
tags: [frontend, legacy-migration, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0027 — Second-pass parity must diff against legacy snapshot

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Applying visual tweaks without checking exact delta versus legacy source-of-truth.

## Correct
Run direct file diff against `legacy_snapshot` and keep only intentional structural deviations.

## Rule
In parity mode, every CSS delta from legacy must be explicit and justified.

## Impact
**Medium:** Ensures intentional changes during legacy migration.
