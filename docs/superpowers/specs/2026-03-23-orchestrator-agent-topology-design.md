# Orchestrator + Agent Topology Design

**Date:** 2026-03-23
**Status:** Draft
**Scope:** `$ms` redesign as a 4-decision router, extended T1→T7 implementation chain, plan mode and model selection rules, parallel dispatch pattern, context isolation rule, two new agents, and hooks configuration

**Depends on:**
- Sub-problem 1 spec: `2026-03-23-plugin-set-design.md`
- Sub-problem 2 spec: `2026-03-23-session-structure-design.md`

---

## Problem

The existing `$ms` orchestrator routes tasks to skills but makes four implicit decisions on every task: which model to use, which tool to use (Claude vs Codex), whether to enter plan mode, and whether to dispatch work in parallel. Because these decisions are implicit, the current workflow runs everything sequentially in the main context, uses the same model regardless of task complexity, has no Codex integration, and has no mechanism for keeping progress documentation current. The practical result is: token limits hit mid-task on complex features, slow execution on stages that could run in parallel, and documentation that drifts silently after every completed feature.

## Goals

1. Every task that enters `$ms` receives four explicit routing decisions before any work starts
2. Main context is never used for research or exploration — all codebase reads go to subagents (session-start reads are exempt: they are ritual, not research)
3. Codex executes T3a (adapters + transport), T5 boilerplate, and T5.5 (tests) — the three highest-volume implementation stages
4. `docs/PROGRESS.md` and `docs/PROJECT_SOT.md` are updated at T7 without blocking the review gate or the commit
5. All analytics work routes through `$analytics-orchestrator` — `$ms` never calls analytics sub-agents directly
6. Both Claude and Codex load context before writing any code

## Out of Scope

- Post-task ritual, lesson quality filter, handoff document format (sub-problem 2)
- Plugin installation and memory system seeding (sub-problem 1)
- Analytics intelligence system design (`docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md`)
- Personal workflow guide (post-implementation)

---

## Definitions

**Critical task:** Any task where the files being written contain or interact with auth, tenant isolation, or outbox patterns — regardless of which T-stage is executing. Risk-based, not stage-based.

**Sonnet has failed (current task):** Within the current task instance, Sonnet produced output that violated an absolute rule, was identified as architecturally incorrect by Claude review, or failed to produce working code. Escalation to Opus applies to the current task only — not persistently to that task type in future sessions.

**T5 boilerplate (Codex):** File scaffolding, component stubs, CSS module creation, prop wiring to existing SDK methods, and page layout structure that replicates an existing pattern exactly.

**T5 architecture (Claude):** New SDK method selection, new component abstractions not present in `packages/ui/`, state management decisions, routing changes, and any frontend decision with no direct existing pattern to copy.

**Review gate (blocking):** The review gate always blocks commit. The commit does not happen until `$metalshopping-review` completes and passes. T7 docs update is non-blocking — it runs in parallel and its failure does not hold the commit.

---

## Design

### 1. `$ms` as a 4-decision router

When a task enters `$ms`, four decisions are logged before any T-stage begins:

**Decision 1: Plan mode?**
- YES: new module, multi-file change (>2 files), architectural decision, anything touching auth / tenant isolation / outbox logic
- NO: single-file fix, trivial correction, boilerplate with a complete existing pattern

When plan mode is YES, Claude writes the plan to `tasks/todo.md`. The user reviews it and either approves, requests revisions, or rejects. Approval is required before T1 starts. If the user requests revisions, Claude rewrites the plan and presents it again. If rejected, the task is closed without implementation.

**Decision 2: Which model?**
- Opus: new domain architecture (T1 for new module), ADR (T6), critical task at review gate, Sonnet has failed on this task instance
- Sonnet: all other stages (default)

Codex uses its own internal model and is not subject to Claude's model routing rules.

**Decision 3: Claude or Codex?**
- Codex: plan is complete (all file paths named, constraints listed, no TBD), task qualifies for T3a / T5 boilerplate / T5.5, and the specific files being written are not critical (see Definitions)
- Claude: all orchestration, planning, review, architectural decisions, all critical files

**Decision 4: Parallel dispatch?**
- YES if 2+ independent subtasks exist — dispatch subagents simultaneously
- ALWAYS for research: codebase exploration always dispatches a subagent; main context receives only the summary

---

### 2. Context isolation — hard rule

No research runs in the main context. Before any stage requiring codebase understanding, `$ms` dispatches a subagent to read files, identify the closest existing pattern, and return a structured summary. The main context acts only on the summary.

Session-start reads (`tasks/lessons.md`, `tasks/todo.md`, MEMORY.md) are exempt — these are ritual reads with bounded, known scope, not open-ended exploration.

Rationale: File reads in the main context consume tokens permanently. Subagent context is isolated and discarded. This is the highest single-impact token management decision in the workflow.

Enforcement: behavioral rule in the `$ms` skill definition. Violations are caught at the review gate.

---

### 3. Extended T1→T7 chain with tool ownership

| Stage | What | Tool | Model | Parallel opportunity |
|-------|------|------|-------|---------------------|
| T1 | Contract (OpenAPI / event / governance) | Claude | Opus (new module), Sonnet (extending) | Contract validation subagent runs simultaneously |
| T2 | Backend domain + application layers | Claude | Sonnet | — |
| T3a | Adapters + transport handlers | Codex | Codex internal | T4 SDK generation starts after T3a draft |
| T3b | Outbox wiring + readmodel | Claude | Sonnet | — |
| T3.5 | DB migration (.sql) | Claude | Sonnet | — |
| T4 | SDK generation (script + validation) | Claude | Sonnet | Overlaps T3a review |
| T5 | Boilerplate (Codex) + architecture (Claude) | Split — see Definitions | Sonnet (Claude side); Codex internal | YES — Codex boilerplate and Claude architecture run simultaneously on different files |
| T5.5 | Tests | Codex | Codex internal | — |
| T6 | ADR (only when architectural decision made) | Claude | Opus | — |
| T7 | Docs update (PROGRESS.md, PROJECT_SOT.md) | `$metalshopping-docs` subagent | Sonnet | Parallel with review gate (non-blocking) |
| Review | `$metalshopping-review` | Claude | Opus (critical tasks), Sonnet (standard) | Parallel with T7 (blocking — commit waits for this) |

T3 split rationale: T3a follows strict templates — Codex produces correct output with a complete handoff. T3b contains outbox wiring, the highest-risk correctness constraint in the backend; it always stays with Claude.

T5 split rationale: Boilerplate copies existing patterns — safe for Codex. Architecture has no pattern to copy — requires Claude judgment.

**Stage selection per task type:**

| Task type | Stages |
|-----------|--------|
| Pure frontend fix | T5 + T5.5 + Review |
| New full backend module | T1 → T2 → T3a → T3b → T3.5 → T4 → T5 → T5.5 → T6 → T7 → Review |
| Contract extension (existing module) | T1 → T4 → T5.5 → Review |
| Contract extension touching tenant/auth/outbox | Full chain — classified as critical task |
| Bug fix | Affected stage(s) + T5.5 + Review |

Note: "contract change only" always includes T5.5 (tests) and the review gate. A contract change with no tests is not complete.

---

### 4. Lesson evaluation — three trigger points

Filter defined in sub-problem 2. Runs at:
1. **Codex constraint fix** — Claude fixes, filter evaluates, lesson written if passes. Codex loads `tasks/lessons.md` via handoff document on every task.
2. **User correction** — filter evaluates immediately.
3. **Post-task sweep** — before commit, Claude evaluates full task for uncaught lessons.

---

### 5. Context loading — both tools

**Claude (session start):** Reads lessons → todo → MEMORY.md → STATE/NEXT/WATCH briefing. Defined in sub-problem 2.

**Codex (task start):** Loads `tasks/codex-handoff.md` (format defined in sub-problem 2) and `tasks/lessons.md` before writing. Does not read full project context — Claude distills what Codex needs.

---

### 6. Hooks configuration

| Hook | Trigger | Action |
|------|---------|--------|
| `SessionStart` | Session opens | Reads lessons + todo + MEMORY.md → STATE/NEXT/WATCH briefing |
| `PostTaskComplete` | Task marked done | Post-task sequence from sub-problem 2 |
| `PreFileWrite` | Write to `packages/generated/` or `packages/generated-types/` | Blocks write, reports violation |
| `SecurityCheck` | Security-sensitive file edit | `security-guidance` plugin (already installed) |

---

### 7. Two new agents

**`$metalshopping-docs`**
- Stage: T7, dispatched as subagent in parallel with review gate
- Input: task description + committed diff + current PROGRESS.md + current PROJECT_SOT.md
- Output: delta update only — appends completed features, updates status; never rewrites
- Model: Sonnet
- Failure behavior: failure is logged as a warning entry in `tasks/todo.md`, user is notified, retry attempted at next session start. Does not block commit.

**`$analytics-orchestrator`**
- Entry point for all tasks touching `packages/feature-analytics/`, `analytics_serving` Go module, analytics worker, or any of the 11 analytics read surfaces
- `$ms` never calls analytics sub-agents directly
- Interface: receives a task description + applicable T-stage; returns completed output or escalation to Claude for architectural decisions
- Internal routing to four sub-agents:
  - `$analytics-intelligence` — 8-layer brain (metrics, rules, calculations, classifications, recommendations); input: task + data model context; output: intelligence layer changes
  - `$analytics-surfaces` — 11 read surfaces + charts + workspaces; input: task + design system context; output: frontend components
  - `$analytics-campaigns` — campaigns, actions, alerts, approval flows; input: task + governance context; output: campaign/alert logic
  - `$analytics-ai` — copilot layer; input: task + intelligence layer outputs; output: explanation/simulation/drafting features
- Full sub-agent definitions in dedicated analytics brainstorm
- Model: Sonnet (routing), Opus (read model architecture)

---

### 8. Full execution flow

```
Task enters $ms
      │
      ▼
4 decisions logged: plan mode / model / Claude vs Codex / parallel?
      │
      ▼
Research needed? ──YES──▶ subagent explores ──▶ summary to main context
      │
      ▼
Plan mode? ──YES──▶ Claude writes plan to tasks/todo.md
                              │
                    user reviews plan
                    ├── approves ──▶ continue to T1
                    ├── revises ──▶ Claude rewrites ──▶ present again
                    └── rejects ──▶ task closed, no implementation
      │
      ▼
T1 (Claude) ──parallel──▶ contract validation subagent
      │
      ▼
T2 (Claude/Sonnet)
      │
      ├──▶ T3a (Codex) ──parallel──▶ T4 SDK generation begins
      ├──▶ T3b (Claude — outbox + readmodel)
      │
      ▼
T3.5 migration (Claude)
      │
      ▼
T4 validates (Claude)
      │
      ▼
T5: Codex boilerplate + Claude architecture simultaneously (different files)
      │
      ▼
T5.5 tests (Codex)
      │
      ▼
Codex output → Claude reviews → violation? ──YES──▶ Claude fixes ──▶ lesson eval (trigger 1)
      │
      ▼
T6 ADR if needed (Claude/Opus)
      │
      ▼
Lesson sweep — trigger 3 (post-task, before commit)
      │
      ▼
T7 (non-blocking) ──parallel──▶ $metalshopping-docs runs, logs failure if it occurs
Review (blocking) ──parallel──▶ $metalshopping-review runs, commit waits for pass
      │
      ▼
Commit (after review passes; T7 outcome does not affect this)
      │
      ▼
PostTaskComplete hook fires
```

---

## Acceptance Criteria

1. Every task has all four routing decisions logged before any T-stage executes
2. No Read tool call for research occurs in the main context during any T-stage; session-start reads are exempt
3. T3a and T5.5 execute via Codex; T3b always executes via Claude; handoff document is complete before Codex starts any stage
4. T3b is never assigned to Codex regardless of task size, plan completeness, or line count
5. Commit waits for review gate to pass; T7 failure does not block commit but is logged in `tasks/todo.md`
6. Lesson filter runs at all three trigger points: Codex fix, user correction, post-task sweep
7. Codex loads both handoff document and `tasks/lessons.md` before writing on every task
8. All analytics tasks route through `$analytics-orchestrator`; no direct `$ms` → analytics sub-agent call exists in the `$ms` skill file
9. `PreFileWrite` hook blocks writes to `packages/generated/` and `packages/generated-types/`
10. Opus is used only in cases matching the Definitions section; Codex model column is intentionally absent — Codex manages its own model internally
11. Plan mode produces a written plan, waits for explicit user approval, and handles revision and rejection paths before T1 begins
12. All task types in Section 3 include T5.5 and the review gate — no task type completes without tests and review
