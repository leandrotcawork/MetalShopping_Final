---
id: lesson-0026
title: Workspace root must provide token fallbacks
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

# Lesson 0026 — Workspace root must provide token fallbacks

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Relying on parent route tokens for workspace surfaces (`--surface`, `--muted`, `--wine`, `--success`), causing inconsistent rendering.

## Correct
Define token fallbacks on workspace root `.page` with explicit dark-mode overrides.

## Rule
Self-contained route modules must declare required visual tokens at their own root.

## Impact
**Medium:** Ensures workspace styling is self-contained and independent.
