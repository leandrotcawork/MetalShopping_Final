---
id: lesson-0030
title: Preserve UTF-8 when patching legacy-copied frontend files
region: lessons
tags: [frontend, legacy-migration, encoding, correction]
links:
  - cortex/frontend/index
severity: low
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0030 — Preserve UTF-8 when patching legacy-copied frontend files

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Editing UTF-8 files through a different code page (e.g., CP-1252) and writing back, which corrupts accents/emojis and changes UI text/icons.

## Correct
Read/write migrated frontend files as UTF-8 and prefer patching tools that preserve encoding.

## Rule
Legacy parity files must keep original text encoding to avoid silent UI regressions.

## Impact
**Low:** Prevents encoding-related UI text corruption.
