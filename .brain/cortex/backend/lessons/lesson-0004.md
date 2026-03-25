---
id: lesson-0004
title: Worker writes require tenant context and idempotency
region: lessons
tags: [worker, python, tenant-context, idempotency, baseline]
links:
  - cortex/backend/index
  - hippocampus/conventions
severity: critical
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0004 — Worker writes require tenant context and idempotency

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Python worker

## Wrong
Writing without `set_config('app.current_tenant_id', ...)` or without conflict-safe upsert semantics.

## Correct
Start write tx with tenant `set_config` and use `ON CONFLICT ... DO UPDATE` where applicable.

## Rule
Worker write paths must be tenant-safe and retry-safe.

## Impact
**Critical:** Prevents cross-tenant data contamination in async workers and data duplication on retries.
