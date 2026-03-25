---
id: lesson-0007
title: Generated artifacts are read-only outputs
region: lessons/cross-domain
tags: [process, build-system, baseline]
links:
  - hippocampus/conventions
  - cortex/backend/index
severity: critical
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T21:00:00Z
---

# Lesson 0007 — Generated artifacts are read-only outputs

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Build system

## Wrong
Manually editing code artifacts that are auto-generated from source contracts (API specs, event schemas, SDK runtime).

## Correct
Edit contracts in the source directory, run contract generation scripts, and allow regeneration to produce downstream artifacts.

## Rule
Auto-generated code is read-only. Never edit generated files manually. Edits will be overwritten on next generation cycle.

## Impact
**Critical:** Prevents breaking changes that get overwritten, breaking the build system and contract alignment. Maintains single source of truth.

See [[hippocampus/conventions]] for source-of-truth hierarchy.
