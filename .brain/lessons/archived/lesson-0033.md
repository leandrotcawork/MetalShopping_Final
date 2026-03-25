---
id: lesson-0033
title: First-fold workspace KPIs must be rendered in ProductHero
region: lessons
tags: [frontend, analytics, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-24T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0033 — First-fold workspace KPIs must be rendered in ProductHero

**Date:** 2026-03-24 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Adding requested KPI fields only in tab-level content (e.g., Simulator) when the expected location is the workspace hero block.

## Correct
Inject first-fold KPI metrics in `ProductHero`/`HeroMetrics` with resilient fallback mapping from available model fields.

## Rule
When a KPI is requested for the workspace header area, render it in the hero metrics source, not only inside tabs.

## Impact
**Medium:** Ensures KPIs are visible above the fold where users expect them.
