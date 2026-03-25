---
id: lesson-0034
title: Hero KPI order must be explicit, not payload-driven
region: lessons
tags: [frontend, analytics, ux, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-24T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0034 — Hero KPI order must be explicit, not payload-driven

**Date:** 2026-03-24 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Rendering workspace or simulator hero KPIs in the raw payload order, which lets backend/mock ordering drift break visual parity.

## Correct
Build hero KPI lists from an explicit ordered view-model that matches the approved legacy surface.

## Rule
First-fold KPI strips must be rendered from fixed presentation order, never implicit source order.

## Impact
**Medium:** Prevents KPI reordering from breaking visual parity.
