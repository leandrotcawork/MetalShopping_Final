---
name: analytics-surfaces
description: STUB — Analytics surfaces sub-agent. Owns all 11 analytics read surfaces, charts, workspaces, and frontend component library for analytics. Full design pending dedicated analytics brainstorm session.
---

# $analytics-surfaces — Analytics Surfaces Sub-Agent (STUB)

**Status: STUB — full design pending dedicated analytics brainstorm session**
See: docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md

## Owns
All 11 analytics read surfaces: executive home, products overview, product workspace,
brands overview, brand workspace, taxonomy overview, taxonomy workspace, alerts center,
campaigns center, buying recommendations board, AI copilot panel.
Also owns: charts, workspace layouts, frontend component library for analytics.

## Input
Task description + design system context from $analytics-orchestrator

## Output
Frontend components (React/TypeScript), CSS modules, chart configurations,
workspace page layouts matching the $metalshopping-design-system conventions.

## When this stub is invoked
Route back to $analytics-orchestrator with a note that full surface
implementation requires the analytics brainstorm session to be completed first.
