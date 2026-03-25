---
id: lesson-0013
title: Observability is part of the contract
region: lessons
tags: [backend, logging, observability, baseline]
links:
  - cortex/backend/index
  - hippocampus/conventions
severity: high
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0013 — Observability is part of the contract

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Backend + Process

## Wrong
Returning generic errors/logs without actionable context.

## Correct
Use structured error codes and request logs with `trace_id`, `action`, `result`, `duration_ms`.

## Rule
Debuggability requirements are mandatory, not optional.

## Impact
**High:** Enables efficient debugging and production monitoring.
