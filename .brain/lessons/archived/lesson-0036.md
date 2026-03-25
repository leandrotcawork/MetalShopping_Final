---
id: lesson-0036
title: Shell bars inside padded app mains must break out at the route root
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

# Lesson 0036 — Shell bars inside padded app mains must break out at the route root

**Date:** 2026-03-24 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Recreating a shell top bar under a parent `main` with fixed padding but keeping it inset inside that padding box.

## Correct
Match the legacy route shape (`section > header` + content container sibling) and let the outer header wrapper escape parent padding when needed.

## Rule
When app shell padding exists above the route, shell-level bars must break out at the route root or they will never visually match legacy.

## Impact
**Medium:** Ensures shell bars break out of constrained content areas correctly.
