---
id: lesson-0032
title: Simulator hero metrics need tolerant alias mapping
region: lessons
tags: [frontend, analytics, correction]
links:
  - cortex/frontend/index
severity: medium
occurrence_count: 1
escalated: false
created_at: 2026-03-23T00:00:00Z
updated_at: 2026-03-24T10:00:00Z
---

# Lesson 0032 — Simulator hero metrics need tolerant alias mapping

**Date:** 2026-03-23 | **Trigger:** correction | **Layer:** Frontend

## Wrong
Depending on a single metric label in workspace payload causes business KPIs to disappear when label names vary.

## Correct
Parse simulator baseline metrics using alias keys + deterministic fallback values (`preco real efetivo`, `gasto var`, gap baseline).

## Rule
UI-facing KPI extraction from loose payloads must be resilient to label variation.

## Impact
**Medium:** Prevents KPI rendering failures due to label changes.
