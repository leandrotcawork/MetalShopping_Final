---
name: analytics-ai
description: STUB — Analytics AI copilot sub-agent. Owns natural language explanations, scenario simulations, decision drafting, operator-facing conversational interface. Full design pending dedicated analytics brainstorm session.
---

# $analytics-ai — Analytics AI Copilot Sub-Agent (STUB)

**Status: STUB — full design pending dedicated analytics brainstorm session**
See: docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md

## Owns
AI copilot layer: natural language explanations of recommendations,
scenario simulations, decision drafting, operator-facing conversational interface.

## Input
Task description + intelligence layer outputs from $analytics-orchestrator
(metrics, classifications, recommendations produced by $analytics-intelligence)

## Output
Explanation/simulation/drafting features: Go handlers for copilot endpoints,
frontend AI panel components, prompt templates and inference orchestration.

## When this stub is invoked
Route back to $analytics-orchestrator with a note that full AI copilot
implementation requires the analytics brainstorm session to be completed first.
