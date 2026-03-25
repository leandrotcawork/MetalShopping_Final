---
id: lesson-0012
title: Runtime behavior changes require operational verification
region: lessons
tags: [process, debugging, baseline]
links:
  - hippocampus/conventions
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0012 — Runtime behavior changes require operational verification

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Process

## Wrong
Blaming code before checking migration/manifest/config/version/runtime restart state.

## Correct
Verify DB migration, active manifest/config, restart status, and test data freshness first.

## Rule
Diagnose runtime/state drift before code-level fixes on worker/config-driven flows.

## Impact
**High:** Prevents wild goose chases investigating code when the issue is operational.
