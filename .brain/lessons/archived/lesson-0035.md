---
id: lesson-0035
title: Workspace top bars that belong to the shell must be rendered as shell strips
region: lessons
tags: [frontend, layout, ux, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-24T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0035 — Workspace top bars that belong to the shell must be rendered as shell strips

**Date:** 2026-03-24 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Styling a workspace top bar as a self-contained route header/card with its own padded surface.

## Correct
Render shell-level top bars in a full-width wrapper that owns background/border/sticky behavior, and keep the header component as inner constrained content only.

## Rule
If the bar is part of the shell, shell chrome lives in the outer wrapper, not in the page component itself.

## Impact
**Medium:** Ensures proper shell layout hierarchy.
