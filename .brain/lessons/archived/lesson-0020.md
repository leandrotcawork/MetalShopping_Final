---
id: lesson-0020
title: Header hover must be bound to interactive target only
region: lessons
tags: [frontend, css, ux, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0020 — Header hover must be bound to interactive target only

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Applying hover color on `th:hover`, which changes label color even outside the clickable text area.

## Correct
Bind hover/focus color only to the sortable button label (`.spotlightSkuSortBtn:hover .spotlightSkuHeadLabel`).

## Rule
Visual feedback must match the real interactive hit area.

## Impact
**Medium:** Improves UX accuracy and visual feedback.
