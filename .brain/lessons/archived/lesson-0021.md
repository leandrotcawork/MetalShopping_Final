---
id: lesson-0021
title: Feature CSS modules need local token baselines
region: lessons
tags: [frontend, css, design-tokens, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0021 — Feature CSS modules need local token baselines

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Using `var(--surface)` / `var(--surface-border)` in feature styles without defining a local token baseline, causing transparent controls.

## Correct
Define token defaults on the page root class and dark overrides in the same module.

## Rule
Any feature stylesheet that depends on design tokens must be self-sufficient for surface/border readability.

## Impact
**Medium:** Prevents missing CSS variables from breaking styling.
