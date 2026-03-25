---
id: lesson-0001
title: Tenant-safe DB access is mandatory
region: lessons
tags: [tenant-isolation, database, go, baseline]
links:
  - cortex/backend/index
  - hippocampus/conventions
severity: critical
occurrence_count: 1
escalated: true
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0001 — Tenant-safe DB access is mandatory

**Date:** 2026-03-23 | **Trigger:** baseline | **Layer:** Go adapter

## Wrong
Running adapter queries without tenant transaction/runtime tenant scope.

## Correct
Every Postgres adapter query uses `pgdb.BeginTenantTx` and tenant-scoped predicates (`current_tenant_id()` where applicable).

## Rule
No tenant-scoped read/write may run outside tenant transaction and tenant filters.

## Impact
**Critical:** This is a data leak risk. Cross-tenant queries expose sensitive customer data.

See [[sinapses/tenant-isolation-flow]] for complete flow documentation.
