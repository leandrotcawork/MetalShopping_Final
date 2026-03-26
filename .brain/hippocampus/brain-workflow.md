---
id: hippocampus-brain-workflow
title: Brain Workflow Rules
type: hippocampus
tags: [workflow, pipeline, brain, rules]
updated_at: 2026-03-26
---

# Brain Workflow Rules

## Pipeline

```
brain-decision → brain-map → brain-task → [brain-codex-review] → brain-document → brain-consolidate
```

## File Location Guardrails

### During Task Execution (brain-task Steps 1-5)

| Step | File Location | Purpose |
|------|---------------|---------|
| 1 | `.brain/working-memory/context-packet-[id].md` | Assembled sinapses (context) |
| 2 | `.brain/working-memory/sonnet-context-[id].md` OR `codex-context-[id].md` OR `opus-debug-context-[id].md` | Execution context |
| 4 | `.brain/working-memory/task-completion-[id].md` | Outcome + files + tests + lessons |
| 5 | `.brain/working-memory/sinapse-updates-[id].md` | Proposed sinapse updates (awaiting approval) |

All artifacts during execution go to `.brain/working-memory/` — NOT to `tasks/`.

### After Task Completion (brain-task Step 6)

| Location | Purpose |
|----------|---------|
| `.brain/progress/completed-contexts/[id]-completion-record.md` | Archived completion record |
| `.brain/progress/activity.md` | Activity log (append-only) |

### Sprint Backlog (manual)

| Location | Purpose |
|----------|---------|
| `tasks/todo.md` | Current sprint backlog |
| `tasks/lessons.md` | Permanent lessons learned |

## Model Routing

| Score | Model | Use Case |
|-------|-------|---------|
| < 20 | Haiku | Simple, single-domain, no risk |
| 20-39 | Sonnet | Standard single-domain tasks |
| 40-74 | Codex | Complex, cross-domain, high risk |
| 75+ | Codex + Plan | Architectural, multi-phase |
| debugging (any) | Opus | All debugging regardless of score |

## Complexity Scoring

- Base: 15
- Cross-domain: +30, Backend: +10
- Risk critical: +35, high: +20, medium: +5
- Type debug: +15, architectural: +20, unknown_pattern: +10
