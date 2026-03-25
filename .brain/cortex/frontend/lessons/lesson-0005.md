---
id: lesson-0005
title: Frontend data flow must use platform SDK contracts
region: lessons
tags: [frontend, sdk, contracts, baseline]
links:
  - cortex/frontend/index
  - hippocampus/conventions
severity: critical
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0005 — Frontend data flow must use platform SDK contracts

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Frontend

## Wrong
Fetching data directly via `fetch()` or bypassing contract-generated runtime types.

## Correct
Use `@metalshopping/sdk-runtime` hooks/runtime and keep loading/error/empty states explicit.

## Rule
Frontend data access is SDK-first, contract-aligned, and state-complete.

## Impact
**Critical:** Ensures type safety, contract alignment, and prevents manual HTTP errors.

See [[sinapses/sdk-data-flow]] for complete SDK integration pattern.
