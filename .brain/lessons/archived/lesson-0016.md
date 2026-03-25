---
id: lesson-0016
title: Mock semantics must match UI contract keys
region: lessons
tags: [frontend, legacy-migration, testing, correction]
links:
  - cortex/frontend/index
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0016 — Mock semantics must match UI contract keys

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Feeding display labels like `Info` or `OK` into UI code that expects metric keys such as `giro_6m` and `margin_sales_pct`.

## Correct
Mock payloads must preserve the exact semantic keys consumed by the migrated UI, even in visual-first phases.

## Rule
In legacy migration, mock data shape is part of the contract and cannot be approximated with human labels.

## Impact
**High:** Prevents mismatch between mock data and real data contracts.
