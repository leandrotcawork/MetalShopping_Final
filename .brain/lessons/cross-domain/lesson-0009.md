---
id: lesson-0009
title: tasks/todo.md edits must be block-scoped
region: lessons/cross-domain
tags: [process, collaboration, baseline]
links:
  - hippocampus/conventions
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T21:00:00Z
---

# Lesson 0009 — tasks/todo.md edits must be block-scoped

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Process

## Wrong
Editing scattered lines throughout tasks/todo.md across multiple commits or contributors.

## Correct
Group all edits to a single feature's task block. Edit only within that block's boundaries. One commit per complete task.

## Rule
Keep todo.md edit scope tight to prevent merge conflicts when multiple task branches work in parallel.

## Impact
**Medium:** Prevents merge conflicts that stall task completion. Keeps todo.md a reliable single source of sprint state.

See [[hippocampus/conventions]] for collaboration rules.
