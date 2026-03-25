---
id: lesson-0017
title: Hover parity requires selector specificity on label elements
region: lessons
tags: [frontend, css, legacy-migration, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0017 — Hover parity requires selector specificity on label elements

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Relying on generic hover color rules that are overridden in composed table header styles.

## Correct
Scope hover color to the exact interactive label selector in the spotlight header (`th .spotlightSkuSortBtn`) with sufficient specificity.

## Rule
For legacy visual parity, hover behavior must target the final rendered label node, not only parent containers.

## Impact
**Medium:** Ensures consistent hover behavior in migrated tables.
