---
id: lesson-0003
title: Outbox must be atomic with writes
region: lessons
tags: [outbox, events, atomicity, go-adapter, baseline]
links:
  - cortex/backend/index
  - hippocampus/conventions
severity: critical
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0003 — Outbox must be atomic with writes

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Go adapter + events

## Wrong
Appending events after `tx.Commit`.

## Correct
Use `outbox.AppendInTx` in the same transaction before commit.

## Rule
Write + event publish intent must be atomic or not happen.

## Impact
**Critical:** Prevents lost events and ensures eventual consistency across services.

See [[sinapses/outbox-event-flow]] for complete atomicity pattern.
