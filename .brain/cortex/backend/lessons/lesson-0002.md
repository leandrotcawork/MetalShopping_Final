---
id: lesson-0002
title: Handlers must fail fast on auth and tenancy
region: lessons
tags: [auth, tenancy, go-handlers, baseline]
links:
  - cortex/backend/index
  - hippocampus/conventions
severity: critical
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0002 — Handlers must fail fast on auth and tenancy

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Go handler

## Wrong
Calling service logic before validating principal/tenant context.

## Correct
`PrincipalFromContext` (401) and `TenantFromContext` (403) execute before any handler operation.

## Rule
Auth and tenancy checks are always first in protected handlers.

## Impact
**Critical:** Prevents unauthorized requests from accessing business logic.

See [[sinapses/tenant-isolation-flow]] for complete validation flow.
