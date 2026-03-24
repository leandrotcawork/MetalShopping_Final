---
name: analytics-orchestrator
description: Entry point for ALL analytics domain tasks. $ms routes here for anything touching packages/feature-analytics/, analytics_serving Go module, analytics_worker, or the 11 analytics read surfaces. Never bypassed by $ms directly to sub-agents.
---

# $analytics-orchestrator — Analytics Domain Orchestrator

Entry point for ALL tasks touching the analytics domain. $ms routes here.
Never bypass this orchestrator to call sub-agents directly from $ms.

## Triggers
Route to this skill when the task involves any of:
- packages/feature-analytics/
- apps/server_core/internal/modules/analytics_serving/
- apps/analytics_worker/
- Any of the 11 analytics read surfaces:
  executive home, products overview, product workspace, brands overview,
  brand workspace, taxonomy overview, taxonomy workspace, alerts center,
  campaigns center, buying recommendations board, AI copilot panel

## Routing decisions

### Which sub-agent handles this task?

| Task type | Sub-agent |
|-----------|-----------|
| Intelligence layer (metrics, rules, calculations, classifications, recommendations, formulas) | $analytics-intelligence |
| Frontend surfaces, charts, workspaces, UI components | $analytics-surfaces |
| Campaigns, actions, alerts, approval flows | $analytics-campaigns |
| AI copilot, explanations, simulations, drafting | $analytics-ai |

### Model
- Sonnet for routing and standard tasks
- Escalate to Opus for read model architecture decisions

### T-stage mapping
This orchestrator applies the same T1→T7 chain from $ms but with analytics-specific context:
- T2/T3 backend → $analytics-intelligence sub-agent
- T5 frontend → $analytics-surfaces sub-agent
- T5.5 tests → Codex (same as standard flow)
- T6 ADR → Claude/Opus (escalate to main context)
- T7 docs → $metalshopping-docs (same as standard flow)

## Log entry
Write agent activity log entry to logs/agent-activity-YYYY-MM.jsonl on every dispatch and on completion.

## Note
Full sub-agent definitions are in docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md.
These stubs define routing only. Full capability design requires a dedicated brainstorm session.
