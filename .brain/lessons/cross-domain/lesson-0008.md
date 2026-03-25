---
id: lesson-0008
title: Completion requires validation + commit
region: lessons/cross-domain
tags: [process, completion, baseline]
links:
  - hippocampus/conventions
severity: high
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T21:00:00Z
---

# Lesson 0008 — Completion requires validation + commit

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Process

## Wrong
Considering a task "done" after implementing code changes but before running build/tests and committing.

## Correct
A task is done only when: (1) build passes, (2) tests pass, (3) real data verified (not just happy path), (4) code committed with one commit per task.

## Rule
Never leave work uncommitted at session end. Unvalidated work is work in limbo.

## Impact
**High:** Prevents half-finished work from stalling, ensures continuity across sessions, maintains clean commit history.

See [[hippocampus/conventions]] for commit format rules.
