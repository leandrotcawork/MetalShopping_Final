---
name: analytics-intelligence
description: STUB — Analytics intelligence sub-agent. Owns the 8-layer intelligence model (metrics, rules, calculations, classifications, recommendations, formulas, data quality). Full design pending dedicated analytics brainstorm session.
---

# $analytics-intelligence — Analytics Intelligence Sub-Agent (STUB)

**Status: STUB — full design pending dedicated analytics brainstorm session**
See: docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md

## Owns
8-layer intelligence model: metrics, rules, calculations, classifications,
recommendations, formulas, data quality scoring.

## Input
Task description + data model context from $analytics-orchestrator

## Output
Intelligence layer changes (Go modules, Python worker logic, read model updates)

## When this stub is invoked
Route back to $analytics-orchestrator with a note that full intelligence
implementation requires the analytics brainstorm session to be completed first.
