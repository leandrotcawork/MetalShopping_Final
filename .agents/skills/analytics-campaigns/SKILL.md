---
name: analytics-campaigns
description: STUB — Analytics campaigns sub-agent. Owns campaigns, actions, alerts, approval flows, routing rules. Full design pending dedicated analytics brainstorm session.
---

# $analytics-campaigns — Analytics Campaigns Sub-Agent (STUB)

**Status: STUB — full design pending dedicated analytics brainstorm session**
See: docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md

## Owns
Campaigns, actions, alerts, approval flows, routing rules.
Covers the full lifecycle: creation → scheduling → execution → outcome tracking.

## Input
Task description + governance context from $analytics-orchestrator

## Output
Campaign/alert logic: Go domain + application layer changes, frontend campaign
workflow components, governance schema updates (approval rules, thresholds).

## When this stub is invoked
Route back to $analytics-orchestrator with a note that full campaign
implementation requires the analytics brainstorm session to be completed first.
