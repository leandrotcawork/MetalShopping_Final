---
name: workflow-guardian
description: Inspect-and-report agent. Checks skill staleness, coverage gaps, lesson accumulation, unused skills, and unmapped features. Never auto-updates anything. Writes a [workflow-check] report to tasks/todo.md. User decides what to act on.
---

# $workflow-guardian — Workflow Inspect-and-Report Agent

Inspect-only. Never auto-updates skills or workflow structure.
Reports findings to tasks/todo.md as a [workflow-check] entry.
User decides what to act on.

## Triggers (4)
1. ADR committed → run before next task starts (immediate)
2. 3+ lessons referencing same skill/domain → surfaces in next session WATCH briefing
3. Every 10 tasks completed → periodic health check
4. Manual /workflow-check command → on-demand

## What to inspect

### 1. Skill staleness
For each skill in .agents/skills/:
- List file paths and pattern names the skill references
- Check if those paths still exist in the codebase
- Flag: STALE if a referenced path no longer exists

### 2. Coverage gaps
- Scan codebase for directories/modules added since last check
- Check if any new module type, layer, or surface has no skill covering it
- Flag: GAP if new codebase area has no skill owner

### 3. Lesson accumulation
- Read tasks/lessons.md
- Group lessons by domain keyword (Go, frontend, contract, analytics, etc.)
- Flag: REVIEW if 3+ lessons in the same domain → recommend invoking skill-creator

### 4. Unused skills
- Read agent activity log from logs/agent-activity-*.jsonl
- List skills with no log entries in the last 20 tasks
- Flag: UNUSED — may be stale or redundant

### 5. Unmapped features
- Compare codebase module list against T-stage chain in $ms
- Flag: UNMAPPED if a module or surface exists with no T-stage or skill ownership

## Output format
Write a [workflow-check] entry to tasks/todo.md:

```
## [workflow-check] YYYY-MM-DD

### CRITICAL
- [ ] [skill-stale] $metalshopping-implement references adapters/legacy.go (deleted) — update skill via skill-creator

### WARNING
- [ ] [skill-review] 3 lessons about Go adapter patterns — run skill-creator on $metalshopping-implement
- [ ] [coverage-gap] apps/server_core/internal/modules/crm/ exists with no skill — create $metalshopping-crm or update $ms routing

### INFO
- [ ] [skill-unused] $metalshopping-legacy-migration not invoked in last 20 tasks — confirm still needed
```

## What NOT to do
- Do not edit any skill files
- Do not update $ms routing
- Do not invoke skill-creator
- Do not commit anything
- Do not mark findings as resolved — the user does that
