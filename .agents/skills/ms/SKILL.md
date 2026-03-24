---
name: ms
description: MetalShopping master orchestrator. Routes every task through 4 explicit decisions (plan mode / model / Claude vs Codex / parallel dispatch) before any T-stage begins. Dispatches subagents for all research. Never calls analytics sub-agents directly — routes to $analytics-orchestrator.
---

# $ms — MetalShopping Master Orchestrator

## Session start (mandatory — fires automatically via SessionStart hook)
1. Read `tasks/lessons.md` — apply every rule before touching code
2. Read `tasks/todo.md` — identify STATE (clean / in-progress / blocked)
3. Read MEMORY.md — recover prior-session context
4. Output STATE/NEXT/WATCH briefing. Wait for user direction.

## On every task — make 4 decisions before any T-stage begins

Log all four decisions to tasks/todo.md under the task entry.

### Decision 1: Plan mode?
- YES: new module, >2 files changed, architectural decision, anything touching auth/tenant/outbox
- NO: single-file fix, trivial correction, boilerplate with complete existing pattern

**Both paths:** always write the task intent and 4-decision log to tasks/todo.md first.

**If YES (complex):**
1. Write full plan to tasks/todo.md (all applicable stages, all files, constraints, no TBD)
2. Call `EnterPlanMode` — no file changes are possible until user explicitly approves
3. Present the plan and wait
4. On approve: exit plan mode, start T1 (first applicable T-stage)
5. On revise: update plan in tasks/todo.md, call `EnterPlanMode` again with the revised plan
6. On reject: exit plan mode, close task, no implementation

**If NO (simple):**
1. Write brief intent entry to tasks/todo.md (stage(s) expected + files expected)
2. Proceed directly to T-stages — no plan mode, no approval gate

### Decision 2: Which model?
- Opus: T1 for new domain, T6 ADR, critical task at review gate, Sonnet failed on this task instance
- Sonnet: everything else (default)
- Critical task = files being written contain or interact with auth, tenant isolation, or outbox patterns (risk-based, not stage-based)
- "Sonnet failed" = within current task instance only — not persistent to future tasks

### Decision 3: Claude or Codex?
- Codex: plan complete (all paths named, constraints listed, no TBD) + task qualifies for T3a/T5-boilerplate/T5.5 + files are not critical
- Claude: everything else — all orchestration, planning, review, architectural decisions, all critical files
- When handing to Codex: write tasks/codex-handoff.md first (task + files + constraints + pattern + definition of done)
- Codex loads tasks/codex-handoff.md + tasks/lessons.md before writing

### Decision 4: Parallel dispatch?
- YES if 2+ independent subtasks exist — dispatch subagents simultaneously
- ALWAYS: codebase exploration dispatches a subagent; main context receives only the summary

## Context isolation — hard rule

No research runs in the main context. Before any stage needing codebase understanding, dispatch a subagent to read files and return a structured summary. Main context acts only on the summary.

Session-start reads (lessons/todo/MEMORY.md) are exempt — these are ritual reads with bounded scope.

## T-stage chain

Run only the stages that apply. Determine applicable stages before starting.

| Stage | What | Tool | Model |
|-------|------|------|-------|
| T1 | Contract (OpenAPI/event/governance) | Claude | Opus (new domain), Sonnet (extending) |
| T2 | Backend domain + application layers | Claude | Sonnet |
| T3a | Adapters + transport handlers | Codex | Codex internal |
| T3b | Outbox wiring + readmodel | Claude | Sonnet — NEVER Codex |
| T3.5 | DB migration (.sql) | Claude | Sonnet |
| T4 | SDK generation (script + validate) | Claude | Sonnet |
| T5 | Boilerplate (Codex) + architecture (Claude) | Split | Sonnet (Claude side); Codex internal |
| T5.5 | Tests | Codex | Codex internal |
| T6 | ADR — only when arch decision made | Claude | Opus |
| T7 | Docs update — dispatch $metalshopping-docs as subagent | Subagent | Sonnet |
| Review | $metalshopping-review — always blocking before commit | Claude | Opus (critical), Sonnet (standard) |

T3a and T5.5: Codex. T3b: always Claude, never Codex. T7 and Review: run in parallel (T7 non-blocking, Review blocking).

### Stage selection by task type
- Frontend fix: T5 + T5.5 + Review
- New backend module: T1→T2→T3a→T3b→T3.5→T4→T5→T5.5→T6→T7→Review
- Contract extension: T1→T4→T5.5→Review
- Contract extension touching auth/tenant/outbox: full chain, critical task
- Bug fix: affected stage(s) + T5.5 + Review

## Lesson evaluation — 3 trigger points
1. Codex constraint fix: Claude fixes → run lesson quality filter → write if passes
2. User correction at any point: run filter immediately
3. Post-task sweep: before commit, evaluate full task for uncaught lessons

**Lesson quality filter:** Write to tasks/lessons.md ONLY if ALL three are true:
- Mistake would cause bug/review failure/structural problem if repeated
- Concerns code logic/architecture/backend requirements/module patterns — NOT UI appearance
- Describes a recurring pattern, not a one-off fix

## Memory evaluation (post-task)
After lesson evaluation: did this task establish a durable project-level fact, architectural decision, or user preference change? If yes: write/update memory file + update MEMORY.md index.

## Codex output review
- Standard: Claude reviews against handoff constraints → commit with `reviewed-by: claude` trailer
- High-stakes (auth/tenant/outbox): Claude reviews + user approves → both trailers in commit
- Violation found: Claude fixes directly → lesson filter runs on the fix
- Delete tasks/codex-handoff.md after successful commit

## Agent activity log
Every agent and subagent writes one entry to `logs/agent-activity-YYYY-MM.jsonl` on completion.

Schema:
```json
{
  "timestamp": "<ISO8601>",
  "agent": "<skill name>",
  "parent": "<invoking agent or $ms>",
  "stage": "<T1|T2|T3a|T3b|T3.5|T4|T5|T5.5|T6|T7|review|research>",
  "task": "<task name from tasks/todo.md>",
  "files_changed": ["<path>"],
  "commit": "<hash or null>",
  "decision": "<why this approach was taken>",
  "status": "<success|fix-applied|false-positive|escalated|failed>",
  "issues": "<description or null>"
}
```

Research subagents omit `files_changed` and `commit`, replace with `output_summary`.

## claudewatch interrupts
When claudewatch fires a violation signal during a T-stage: pause, evaluate the signal.
- Confirmed violation: fix it → run lesson quality filter
- False positive: log the dismissal and continue
Stage does not resume until Claude explicitly clears the interrupt.

## $workflow-guardian triggers
Run $workflow-guardian when:
- ADR committed → run before next task starts
- 3+ lessons in same domain → surface in next session WATCH briefing
- Every 10 tasks → periodic health check
- /workflow-check → on-demand

## Analytics routing
All tasks touching packages/feature-analytics/, analytics_serving Go module, analytics_worker, or any of the 11 analytics read surfaces → route to $analytics-orchestrator. Never call analytics sub-agents directly from $ms.

## Post-task sequence (runs via PostTaskComplete hook)
1. Mark [x] in tasks/todo.md
2. Lesson quality filter evaluation
3. Memory evaluation → update MEMORY.md if applicable
4. Commit via commit-commands: `<type>(<scope>): <what>`
5. claude-md-management updates CLAUDE.md
6. claude-mem compresses session context

## Absolute rules (violation = stop and fix immediately)
**Go:** pgdb.BeginTenantTx on every adapter query; current_tenant_id() in every WHERE; PrincipalFromContext→401; TenantFromContext→403; outbox.AppendInTx in same tx as INSERT; every module in composition_modules.go

**Frontend:** sdk.* only (no fetch()); design tokens only; check packages/ui/src/index.ts first; loading+error+empty states; useEffect+cancelled flag

**Process:** build passes + real data verified + commit made = task done; auto-generated SDK directories (packages/generated and packages/generated-types) are never edited manually
