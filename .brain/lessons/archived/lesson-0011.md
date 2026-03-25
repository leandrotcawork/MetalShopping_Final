---
id: lesson-0011
title: Legacy CSS must define safe token fallbacks
region: lessons
tags: [frontend, css, legacy-migration, baseline]
links:
  - cortex/frontend/index
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0011 — Legacy CSS must define safe token fallbacks

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Frontend

## Wrong
Using critical surface tokens without fallback, causing transparent/unstyled cards.

## Correct
Define local fallback variables for surfaces/borders/radius in migrated CSS modules.

## Rule
Visual-critical styles in migrated pages must remain stable even if tokens are missing.

## Impact
**High:** Prevents broken styling when CSS variables are undefined.
