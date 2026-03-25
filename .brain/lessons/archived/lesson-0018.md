---
id: lesson-0018
title: Table header hover must bind to explicit label node
region: lessons
tags: [frontend, css, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0018 — Table header hover must bind to explicit label node

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Depending on inherited text color from sortable header button for hover state.

## Correct
Wrap header text in a dedicated label element and style hover/focus directly on that label.

## Rule
Interactive table headers should expose a stable label selector for deterministic hover parity.

## Impact
**Medium:** Ensures precise hover state targeting.
