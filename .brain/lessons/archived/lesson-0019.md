---
id: lesson-0019
title: Portaled UI must redeclare local CSS tokens
region: lessons
tags: [frontend, css, portals, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0019 — Portaled UI must redeclare local CSS tokens

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Using CSS variables scoped to page containers in components rendered via portal (`document.body`), causing unresolved hover colors.

## Correct
Redeclare required tokens (`--wine`, `--muted`, `--surface`, etc.) on the portal root (`.drawer`) with dark-theme overrides.

## Rule
Any portaled surface must be self-sufficient for token resolution.

## Impact
**Medium:** Prevents CSS variable resolution failures in portaled components.
