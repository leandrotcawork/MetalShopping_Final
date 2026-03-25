---
id: lesson-0010
title: Legacy migration follows parity-first sequencing
region: lessons
tags: [frontend, legacy-migration, process, baseline]
links:
  - cortex/frontend/index
  - hippocampus/conventions
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0010 — Legacy migration follows parity-first sequencing

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Frontend + Process

## Wrong
Mixing backend integration before visual parity or rewriting layout before baseline match.

## Correct
Freeze baseline → literal copy → shims/mocks → parity pass → integration phases.

## Rule
For legacy migrations, visual parity is completed before backend adaptation.

## Impact
**High:** Ensures incremental validation and reduces integration risk.
