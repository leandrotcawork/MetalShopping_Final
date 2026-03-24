# Developer Workflow Setup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fully implement the MetalShopping developer workflow — plugin configuration, session rituals, $ms orchestrator redesign, new agents, hooks, agent activity log, and workflow guardian — so every future development session runs with full context, correct tool routing, and automatic hygiene.

**Architecture:** Three layered concerns implemented in strict order: (1) plugin environment (nothing works without the right tools installed), (2) session structure (rituals and automation baked into hooks and skills), (3) orchestrator topology ($ms redesign, new agents, log infrastructure). Each layer depends on the previous.

**Tech Stack:** Claude Code settings.json (hooks/plugins), bash scripts, Node.js/npm (log HTML generator), Markdown skill files (.agents/skills/), JSONL (agent activity log)

**Specs:**
- `docs/superpowers/specs/2026-03-23-plugin-set-design.md`
- `docs/superpowers/specs/2026-03-23-session-structure-design.md`
- `docs/superpowers/specs/2026-03-23-orchestrator-agent-topology-design.md`

---

## File Map

### Created
| File | Purpose |
|------|---------|
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/.../memory/project-identity.md` | Seeded memory: platform type, stack, phase |
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/.../memory/engineering-bar.md` | Seeded memory: absolute rules for Go, frontend, process |
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/.../memory/skill-routing.md` | Seeded memory: T1→T7 orchestration chain |
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/.../memory/user-preferences.md` | Seeded memory: production-grade focus, workflow preferences |
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/.../memory/MEMORY.md` | Memory index file |
| `.agents/skills/metalshopping-docs/SKILL.md` | New T7 documentation agent |
| `.agents/skills/analytics-orchestrator/SKILL.md` | New analytics domain orchestrator |
| `.agents/skills/analytics-intelligence/SKILL.md` | Analytics intelligence sub-agent stub |
| `.agents/skills/analytics-surfaces/SKILL.md` | Analytics surfaces sub-agent stub |
| `.agents/skills/analytics-campaigns/SKILL.md` | Analytics campaigns sub-agent stub |
| `.agents/skills/analytics-ai/SKILL.md` | Analytics AI copilot sub-agent stub |
| `.agents/skills/workflow-guardian/SKILL.md` | Workflow inspect-and-report agent |
| `logs/.gitkeep` | Ensures logs/ directory is tracked |
| `scripts/generate-agent-log.mjs` | Node.js script generating HTML from JSONL logs |

### Modified
| File | Change |
|------|--------|
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/blocklist.json` | Remove `code-review@claude-plugins-official` entry |
| `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json` | Disable `data` plugin, add 4 hooks |
| `.agents/skills/ms/SKILL.md` | Full rewrite as 4-decision router with T1→T7 chain |
| `package.json` | Add `agent-log` script |

---

## Task 1: Unblock code-review plugin

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/blocklist.json`

- [ ] **Step 1: Read current blocklist**

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/blocklist.json"
```

Expected: JSON with two entries — `code-review@claude-plugins-official` (test artifact) and `fizz@testmkt-marketplace` (security test).

- [ ] **Step 2: Remove the code-review entry**

Edit the file to remove only the `code-review@claude-plugins-official` entry. Keep the `fizz@testmkt-marketplace` entry. The result should be a valid JSON array with one entry.

- [ ] **Step 3: Verify**

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/blocklist.json"
```

Expected: Valid JSON with only the fizz entry remaining.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "chore(workflow): unblock code-review plugin (test artifact)"
```

---

## Task 2: Disable data plugin

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json`

- [ ] **Step 1: Read current settings**

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json"
```

Note the current `enabledPlugins` array and its format.

- [ ] **Step 2: Remove data from enabled plugins**

Edit `settings.json` to remove `data` from the `enabledPlugins` array (or add it to a `disabledPlugins` array if that is the correct format — check the schema from the current file). Do not remove any other plugins.

- [ ] **Step 3: Verify**

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json"
```

Expected: `data` no longer appears in the enabled plugins list.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "chore(workflow): disable data plugin (no Airflow surface in this repo)"
```

---

## Task 3: Install commit-commands plugin

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json` (updated after install)

- [ ] **Step 1: Install via Claude Code plugin system**

In Claude Code, run:
```
/plugin install commit-commands
```

Or via settings UI: Plugins → Discover → search `commit-commands` → Install.

- [ ] **Step 2: Verify installation**

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/installed_plugins.json" | grep commit-commands
```

Expected: Entry for `commit-commands` present.

- [ ] **Step 3: Verify /commit slash command is available**

In a new Claude Code session, type `/commit` — it should appear as an available command without error.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "chore(workflow): install commit-commands plugin"
```

---

## Task 4: Install claude-mem plugin

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/installed_plugins.json` (updated after install)

- [ ] **Step 1: Add GitHub marketplace source (if not already present)**

In Claude Code settings, add marketplace source:
```
https://github.com/thedotmack/claude-mem
```

- [ ] **Step 2: Install claude-mem**

```
/plugin install claude-mem
```

- [ ] **Step 3: Verify installation**

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/installed_plugins.json" | grep claude-mem
```

- [ ] **Step 4: Verify session compression behavior**

Complete one small task in Claude Code and verify that `claude-mem` produces a session artifact at session end. Check the plugin's documentation for the artifact location.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "chore(workflow): install claude-mem for cross-session memory compression"
```

---

## Task 5: Install claudewatch plugin

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/plugins/installed_plugins.json`

- [ ] **Step 1: Add GitHub marketplace source**

In Claude Code settings, add marketplace source:
```
https://github.com/blackwell-systems/claudewatch
```

- [ ] **Step 2: Install claudewatch**

```
/plugin install claudewatch
```

- [ ] **Step 3: Configure the three alert rules**

After installation, configure these project-specific rules (check plugin docs for config file location — likely `~/.claude/claudewatch-config.json` or similar):

```json
{
  "rules": [
    {
      "name": "frontend-fetch-violation",
      "description": "Direct fetch() call in frontend — must use SDK methods only",
      "pattern": "fetch\\(",
      "file_glob": "packages/feature-*/**/*.{ts,tsx}",
      "exclude_glob": "packages/generated/**",
      "severity": "critical"
    },
    {
      "name": "missing-tenant-tx",
      "description": "Go adapter write without BeginTenantTx",
      "pattern": "func.*Adapter.*\\{",
      "require_pattern": "BeginTenantTx",
      "file_glob": "apps/server_core/internal/modules/*/adapters/*.go",
      "severity": "critical"
    },
    {
      "name": "outbox-after-commit",
      "description": "outbox.AppendInTx called after Commit()",
      "pattern": "Commit\\(\\)",
      "followed_by": "AppendInTx",
      "file_glob": "apps/server_core/**/*.go",
      "severity": "critical"
    }
  ],
  "mode": "active"
}
```

- [ ] **Step 4: Verify get_drift_signal responds**

In a Claude Code session, ask Claude to call `get_drift_signal`. Expected: valid response (even if empty — no drift yet).

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "chore(workflow): install claudewatch with active interrupt mode and 3 alert rules"
```

---

## Task 6: Bootstrap MEMORY.md with 4 seeded files

**Files:**
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/c--Users-leandro-theodoro-MN-NTB-LEANDROT-Documents-MetalShopping-Final-MetalShopping-Final/memory/project-identity.md`
- Create: `.../memory/engineering-bar.md`
- Create: `.../memory/skill-routing.md`
- Create: `.../memory/user-preferences.md`
- Create: `.../memory/MEMORY.md`

Note: the memory directory path uses the repo path encoded as a directory name. Verify the exact path exists:
```bash
ls "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/"
```

- [ ] **Step 1: Create project-identity memory file**

```markdown
---
name: project-identity
description: MetalShopping platform identity, stack, and current development phase
type: project
---

MetalShopping is a server-first B2B platform for commercial strategy, pricing, analytics, procurement, and CRM. Solo developer, production-intended v1.

**Stack:** Go 1.23 backend (modular monolith), React 18 thin-client (Vite + TypeScript), Python workers (skeleton), PostgreSQL (multi-tenant shared-database).

**Phase:** 3A — Foundation Hardening. Analytics legacy migration in progress (Home, Products, Taxonomy, Brands surfaces). Shopping module next.

**Philosophy:** Make it work → make it beautiful → make it fast.

**Why:** This is the first version of intended production software, not a prototype.
```

- [ ] **Step 2: Create engineering-bar memory file**

```markdown
---
name: engineering-bar
description: Absolute rules for Go, frontend, and process — violation = stop and fix
type: project
---

**Go (violation = stop immediately):**
- `pgdb.BeginTenantTx` on every Postgres adapter query — no exceptions
- `current_tenant_id()` in every WHERE on tenant-scoped tables
- `platformauth.PrincipalFromContext` → 401 before any handler operation
- `tenancy_runtime.TenantFromContext` → 403 before any handler operation
- `outbox.AppendInTx` inside same tx as INSERT — never after Commit
- Every new module registered in `composition_modules.go`

**Frontend (violation = stop immediately):**
- Data only via `sdk.*` from `@metalshopping/sdk-runtime` — no raw `fetch()`
- Design tokens only — no hardcoded hex values
- Check `packages/ui/src/index.ts` before creating any component
- Every data-fetching component must have loading + error + empty states
- Fetch pattern: `useEffect + cancelled flag`

**Process:**
- Task done only when: build passes + real data verified + commit made
- `packages/generated/` and `packages/generated-types/` never edited manually
- One commit per completed task — no uncommitted work at session end
- ADR committed only after acceptance test passes
```

- [ ] **Step 3: Create skill-routing memory file**

```markdown
---
name: skill-routing
description: T1→T7 orchestration chain — which skill handles which task type
type: project
---

All tasks enter `$ms` first. `$ms` makes 4 decisions (plan mode / model / Claude vs Codex / parallel) then routes to the correct T-stage.

| Stage | Skill | Tool | When |
|-------|-------|------|------|
| T1 | `$metalshopping-openapi-contracts` | Claude | New or extended contract |
| T2 | `$metalshopping-implement` | Claude | Backend domain + application layers |
| T3a | Codex via handoff | Codex | Adapters + transport (pattern-repetitive) |
| T3b | `$metalshopping-implement` | Claude | Outbox + readmodel (never Codex) |
| T3.5 | `$metalshopping-implement` | Claude | DB migration (.sql) |
| T4 | `$metalshopping-sdk-generation` | Claude | After any contract change |
| T5 | `$metalshopping-design-system` (Claude arch) + Codex (boilerplate) | Split | Frontend |
| T5.5 | Codex via handoff | Codex | Tests |
| T6 | `$metalshopping-adr` | Claude/Opus | When architectural decision made |
| T7 | `$metalshopping-docs` | Subagent | Docs update (parallel with review) |
| Review | `$metalshopping-review` | Claude/Opus | Always blocking before commit |

**Specialty routing:**
- Analytics tasks → `$analytics-orchestrator` (never bypass to sub-agents directly)
- Frontend visual/component → `$metalshopping-design-system`
- Event contracts → `$metalshopping-event-contracts`
- Governance contracts → `$metalshopping-governance-contracts`
- Legacy migration → `$metalshopping-legacy-migration`
```

- [ ] **Step 4: Create user-preferences memory file**

```markdown
---
name: user-preferences
description: User workflow preferences, quality focus, and collaboration style
type: user
---

**Quality standard:** Production-grade from day one. "Would a Stripe or Google senior engineer approve this?" is the bar.

**Workflow:** Task-driven sessions. Always follow tasks/todo.md. Plan mode for non-trivial changes.

**Tool split:** Claude Code for orchestration, planning, architecture, review. OpenAI Codex for implementation volume and token relief.

**Memory needs:** Cross-session context (A) and code-level pattern tracking (C) are priorities.

**On lessons:** Only architectural/structural/logic patterns. Never UI appearance or one-off corrections.

**Response style:** Concise, direct. No walls of text. Lead with the answer.

**Sub-problem order (workflow setup):** Plugins first → session structure → orchestrator → analytics brainstorm (separate session).
```

- [ ] **Step 5: Create MEMORY.md index**

```markdown
# Memory Index

| File | Type | Description |
|------|------|-------------|
| [project-identity.md](project-identity.md) | project | Platform identity, stack, current phase |
| [engineering-bar.md](engineering-bar.md) | project | Absolute rules for Go, frontend, process |
| [skill-routing.md](skill-routing.md) | project | T1→T7 chain and skill-to-task mapping |
| [user-preferences.md](user-preferences.md) | user | Production-grade focus, workflow preferences |
```

- [ ] **Step 6: Verify all 4 files exist with valid frontmatter**

```bash
ls "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/projects/c--Users-leandro-theodoro-MN-NTB-LEANDROT-Documents-MetalShopping-Final-MetalShopping-Final/memory/"
```

Expected: 5 files — MEMORY.md + 4 seeded `.md` files.

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "chore(workflow): bootstrap MEMORY.md with 4 seeded project context files"
```

---

## Task 7: Configure 4 hooks in settings.json

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json`

Use the `update-config` skill for this task: `/update-config`

- [ ] **Step 1: Add SessionStart hook**

Add a hook that fires when a session opens, instructing Claude to read lessons → todo → MEMORY.md and output the STATE/NEXT/WATCH briefing.

Hook config (exact format depends on Claude Code hooks schema — check current settings.json for existing hook format):

```json
{
  "event": "session_start",
  "instruction": "Read tasks/lessons.md, then tasks/todo.md, then MEMORY.md in that order. Output a briefing in exactly this format: STATE: [clean | in-progress: <task> | blocked: <reason>] / NEXT: <next incomplete task> / WATCH: <one relevant lesson for NEXT, or omit if none>. Wait for user direction. Do not start any task autonomously."
}
```

- [ ] **Step 2: Add PostTaskComplete hook**

```json
{
  "event": "task_complete",
  "instruction": "Run the post-task sequence: (1) mark task [x] in tasks/todo.md, (2) evaluate the task against the lesson quality filter — write to tasks/lessons.md only if the correction concerns code logic/architecture/backend requirements/module patterns and describes a recurring pattern, (3) evaluate whether the task established a durable project-level fact for MEMORY.md — write/update memory file and MEMORY.md index if yes, (4) commit via commit-commands with format <type>(<scope>): <what>, (5) invoke claude-md-management to update CLAUDE.md, (6) claude-mem compresses session context. If any step fails, report immediately and stop the sequence."
}
```

- [ ] **Step 3: Add PreFileWrite hook**

```json
{
  "event": "pre_file_write",
  "condition": "file path contains packages/generated/ OR packages/generated-types/",
  "action": "block",
  "message": "BLOCKED: packages/generated/ and packages/generated-types/ are auto-generated. Never edit manually. Run scripts/generate_contract_artifacts.ps1 to regenerate."
}
```

- [ ] **Step 4: Verify or add SecurityCheck hook**

The `security-guidance` plugin may provide this hook natively or it may require an explicit entry.

First, check if the plugin is active and if a hook entry already exists:

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json"
```

Expected: `security-guidance` present in enabled plugins.

Then check whether a `SecurityCheck` hook entry already appears in the hooks array.

- **If a SecurityCheck hook already exists** (provided by the plugin natively): no action needed — proceed to Step 5.
- **If no SecurityCheck hook entry exists**: add it manually to the hooks array in `settings.json`:

```json
{
  "event": "security_check",
  "instruction": "A security-sensitive file edit is in progress. The security-guidance plugin is active — apply its guidance before this write completes. Flag any violation of the absolute rules (BeginTenantTx, PrincipalFromContext, TenantFromContext, outbox.AppendInTx inside transaction, no raw fetch() on frontend)."
}
```

After adding (or confirming native coverage), verify:

```bash
cat "C:/Users/leandro.theodoro.MN-NTB-LEANDROT/.claude/settings.json" | grep -A3 "security"
```

Expected: security-guidance present in enabled plugins AND at least one hook entry covering security-sensitive file writes.

- [ ] **Step 5: Verify all hooks load cleanly**

Start a new Claude Code session. Verify:
- STATE/NEXT/WATCH briefing appears immediately
- No hook errors in startup output

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "chore(workflow): configure SessionStart, PostTaskComplete, and PreFileWrite hooks"
```

---

## Task 8: Rewrite $ms as 4-decision router

**Files:**
- Modify: `.agents/skills/ms/SKILL.md`

- [ ] **Step 1: Read current $ms skill**

```bash
cat ".agents/skills/ms/SKILL.md"
```

Note what the current skill does — we are replacing the routing logic, not starting from scratch.

- [ ] **Step 2: Rewrite SKILL.md**

Replace the skill content with the 4-decision router. The skill must encode:

```markdown
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
- If YES: write plan to tasks/todo.md → present to user → wait for approve/revise/reject → only start T1 on approval

### Decision 2: Which model?
- Opus: T1 for new domain, T6 ADR, critical task at review gate, Sonnet failed on this task instance
- Sonnet: everything else (default)
- Critical task = files being written contain or interact with auth, tenant isolation, or outbox patterns

### Decision 3: Claude or Codex?
- Codex: plan complete (all paths named, constraints listed, no TBD) + qualifies for T3a/T5-boilerplate/T5.5 + files are not critical
- Claude: everything else — all orchestration, planning, review, architectural decisions, all critical files
- When handing to Codex: write tasks/codex-handoff.md first (task + files + constraints + pattern + definition of done)

### Decision 4: Parallel dispatch?
- YES if 2+ independent subtasks exist — dispatch subagents simultaneously
- ALWAYS: codebase exploration dispatches a subagent; main context receives only the summary

## Context isolation — hard rule
No research runs in the main context. Before any stage needing codebase understanding, dispatch a subagent to read files and return a summary. Session-start reads (lessons/todo/MEMORY.md) are exempt.

## T-stage chain

Run only the stages that apply. Determine applicable stages before starting.

| Stage | What | Tool | Model |
|-------|------|------|-------|
| T1 | Contract (OpenAPI/event/governance) | Claude | Opus (new), Sonnet (extending) |
| T2 | Backend domain + application layers | Claude | Sonnet |
| T3a | Adapters + transport handlers | Codex | — |
| T3b | Outbox wiring + readmodel | Claude | Sonnet — NEVER Codex |
| T3.5 | DB migration (.sql) | Claude | Sonnet |
| T4 | SDK generation (script + validate) | Claude | Sonnet |
| T5 | Boilerplate (Codex) + architecture (Claude) | Split | Sonnet (Claude side) |
| T5.5 | Tests | Codex | — |
| T6 | ADR — only when arch decision made | Claude | Opus |
| T7 | Docs update — dispatch $metalshopping-docs as subagent | Subagent | Sonnet |
| Review | $metalshopping-review — always blocking | Claude | Opus (critical), Sonnet (standard) |

T3a and T5.5 go to Codex. T3b never goes to Codex. T7 and Review run in parallel.

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

**Lesson quality filter:** Write to tasks/lessons.md ONLY if ALL true:
- Mistake would cause bug/review failure/structural problem if repeated
- Concerns code logic/architecture/backend requirements/module patterns — NOT UI appearance
- Describes a recurring pattern, not a one-off fix

## Memory evaluation (post-task)
After lesson evaluation: did this task establish a durable project-level fact, architectural decision, or user preference change? If yes: write/update memory file + update MEMORY.md index.

## Codex output review
- Standard: Claude reviews against handoff constraints → commit with `reviewed-by: claude` trailer
- High-stakes (auth/tenant/outbox): Claude reviews + user approves → both trailers in commit
- Violation found: Claude fixes directly → lesson filter runs on the fix

## Agent activity log
Every agent and subagent writes one entry to `logs/agent-activity-YYYY-MM.jsonl` on completion.

## claudewatch interrupts
When claudewatch fires a violation signal during a T-stage: pause, evaluate the signal, (a) confirm → fix → lesson filter, or (b) false positive → log dismissal and continue.

## $workflow-guardian triggers
Run $workflow-guardian when: ADR committed (immediate), 3+ lessons in same domain (queue for next session), every 10 tasks (periodic), or manual /workflow-check.

## Analytics routing
All tasks touching packages/feature-analytics/, analytics_serving, analytics worker, or 11 analytics read surfaces → route to $analytics-orchestrator. Never bypass to sub-agents directly.

## Post-task sequence (runs via PostTaskComplete hook)
1. Mark [x] in tasks/todo.md
2. Lesson quality filter evaluation
3. Memory evaluation → update MEMORY.md if applicable
4. Commit via commit-commands: <type>(<scope>): <what>
5. claude-md-management updates CLAUDE.md
6. claude-mem compresses session context

## Absolute rules (violation = stop and fix)
- Go: pgdb.BeginTenantTx on every adapter query; current_tenant_id() in every WHERE; PrincipalFromContext→401; TenantFromContext→403; outbox.AppendInTx in same tx as INSERT; every module in composition_modules.go
- Frontend: sdk.* only (no fetch()); design tokens only; check packages/ui/src/index.ts first; loading+error+empty states; useEffect+cancelled flag
- Process: build passes + real data verified + commit made = task done; packages/generated/ never edited manually
```

- [ ] **Step 3: Verify skill file is valid markdown with all sections present**

```bash
cat ".agents/skills/ms/SKILL.md" | grep "^## "
```

Expected: Session start, 4 decisions, Context isolation, T-stage chain, Lesson evaluation, Memory evaluation, Codex output review, Agent activity log, claudewatch interrupts, Analytics routing, Post-task sequence, Absolute rules.

- [ ] **Step 4: Commit**

```bash
git add .agents/skills/ms/SKILL.md && git commit -m "feat(workflow): rewrite \$ms as 4-decision router with T1-T7 chain and Codex routing"
```

---

## Task 9: Create $metalshopping-docs skill

**Files:**
- Create: `.agents/skills/metalshopping-docs/SKILL.md`

- [ ] **Step 1: Create skill directory**

```bash
mkdir -p ".agents/skills/metalshopping-docs"
```

- [ ] **Step 2: Write SKILL.md**

```markdown
# $metalshopping-docs — Documentation Agent (T7)

Dispatched as a non-blocking subagent at T7, in parallel with the review gate.
Never blocks the commit. If this agent fails, log a warning and continue.

## Inputs (provided by $ms in the dispatch)
- Task name and description (from tasks/todo.md)
- Committed diff (from last git commit)
- Current docs/PROGRESS.md
- Current docs/PROJECT_SOT.md

## What to do

1. Read the committed diff to understand what changed
2. Read the current docs/PROGRESS.md and docs/PROJECT_SOT.md
3. Write DELTA ONLY — do not rewrite from scratch:
   - In PROGRESS.md: mark the completed feature/task as done, update percentage if applicable
   - In PROJECT_SOT.md: update the relevant section if the completed task changes the current-phase description
4. If nothing in the docs needs changing (e.g. a bug fix or minor tweak), do nothing and report "no docs update needed"
5. Write agent activity log entry to logs/agent-activity-YYYY-MM.jsonl

## Log entry format
```json
{
  "timestamp": "<ISO8601>",
  "agent": "$metalshopping-docs",
  "parent": "$ms",
  "stage": "T7",
  "task": "<task name>",
  "files_changed": ["docs/PROGRESS.md"],
  "commit": null,
  "decision": "<what was updated and why>",
  "status": "success",
  "issues": null
}
```

## Failure behavior
If this agent cannot complete, write a warning entry to tasks/todo.md:
`- [ ] [workflow-docs-retry] Docs update failed for task: <task name> — retry manually`

Never block the main pipeline. Return immediately after writing the log entry or the warning.
```

- [ ] **Step 3: Commit**

```bash
git add .agents/skills/metalshopping-docs/ && git commit -m "feat(workflow): add \$metalshopping-docs T7 documentation agent"
```

---

## Task 10: Create $analytics-orchestrator and 4 sub-agent stubs

**Files:**
- Create: `.agents/skills/analytics-orchestrator/SKILL.md`
- Create: `.agents/skills/analytics-intelligence/SKILL.md`
- Create: `.agents/skills/analytics-surfaces/SKILL.md`
- Create: `.agents/skills/analytics-campaigns/SKILL.md`
- Create: `.agents/skills/analytics-ai/SKILL.md`

- [ ] **Step 1: Create analytics-orchestrator skill**

```bash
mkdir -p ".agents/skills/analytics-orchestrator"
```

```markdown
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
Write agent activity log entry on every dispatch and on completion.

## Note
Full sub-agent definitions are in docs/superpowers/ANALYTICS_BRAINSTORM_BACKLOG.md.
These stubs define routing only. Full capability design requires a dedicated brainstorm session.
```

- [ ] **Step 2: Create 4 sub-agent stubs**

Create `.agents/skills/analytics-intelligence/SKILL.md`:
```markdown
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
```

Create `.agents/skills/analytics-surfaces/SKILL.md`:
```markdown
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
```

Create `.agents/skills/analytics-campaigns/SKILL.md`:
```markdown
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
```

Create `.agents/skills/analytics-ai/SKILL.md`:
```markdown
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
```

- [ ] **Step 3: Commit**

```bash
git add .agents/skills/analytics-orchestrator/ .agents/skills/analytics-intelligence/ .agents/skills/analytics-surfaces/ .agents/skills/analytics-campaigns/ .agents/skills/analytics-ai/ && git commit -m "feat(workflow): add analytics-orchestrator and 4 sub-agent stubs"
```

---

## Task 11: Create $workflow-guardian skill

**Files:**
- Create: `.agents/skills/workflow-guardian/SKILL.md`

- [ ] **Step 1: Create skill directory**

```bash
mkdir -p ".agents/skills/workflow-guardian"
```

- [ ] **Step 2: Write SKILL.md**

```markdown
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
```

- [ ] **Step 3: Commit**

```bash
git add .agents/skills/workflow-guardian/ && git commit -m "feat(workflow): add \$workflow-guardian inspect-and-report agent"
```

---

## Task 12: Add logs/ directory and agent-log HTML generator

**Files:**
- Create: `logs/.gitkeep`
- Create: `scripts/generate-agent-log.mjs`
- Modify: `package.json`

- [ ] **Step 1: Create logs directory**

```bash
mkdir -p logs && touch logs/.gitkeep
```

- [ ] **Step 2: Write the HTML generator script**

Create `scripts/generate-agent-log.mjs`:

```javascript
#!/usr/bin/env node
// Agent Activity Log → HTML Dashboard
// Usage: node scripts/generate-agent-log.mjs
// Output: logs/agent-activity.html

import { readFileSync, writeFileSync, readdirSync } from 'fs';
import { join } from 'path';

const LOGS_DIR = 'logs';
const OUTPUT = join(LOGS_DIR, 'agent-activity.html');

const STATUS_COLOR = {
  success: '#22c55e',
  'fix-applied': '#eab308',
  escalated: '#f97316',
  failed: '#ef4444',
  'false-positive': '#94a3b8',
  default: '#94a3b8',
};

function readAllEntries() {
  const files = readdirSync(LOGS_DIR)
    .filter(f => f.endsWith('.jsonl') && f.startsWith('agent-activity-'));
  const entries = [];
  for (const file of files) {
    const lines = readFileSync(join(LOGS_DIR, file), 'utf-8')
      .split('\n')
      .filter(Boolean);
    for (const line of lines) {
      try { entries.push(JSON.parse(line)); } catch {}
    }
  }
  return entries.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
}

function buildHTML(entries) {
  const agents = [...new Set(entries.map(e => e.agent))].sort();
  const stages = [...new Set(entries.map(e => e.stage))].sort();
  const statuses = [...new Set(entries.map(e => e.status))].sort();

  const rows = entries.map(e => {
    const color = STATUS_COLOR[e.status] || STATUS_COLOR.default;
    const files = (e.files_changed || []).join(', ') || e.output_summary || '—';
    const commit = e.commit ? `<a href="#">${e.commit.slice(0, 7)}</a>` : '—';
    return `
      <tr data-agent="${e.agent}" data-stage="${e.stage}" data-status="${e.status}">
        <td>${new Date(e.timestamp).toLocaleString()}</td>
        <td><code>${e.agent}</code></td>
        <td>${e.stage}</td>
        <td>${e.task}</td>
        <td style="color:${color};font-weight:bold">${e.status}</td>
        <td>${commit}</td>
        <td class="expandable" title="${(e.decision||'').replace(/"/g,'&quot;')}">${(e.decision||'—').slice(0,60)}${(e.decision||'').length>60?'…':''}</td>
      </tr>`;
  }).join('');

  const agentOptions = agents.map(a => `<option value="${a}">${a}</option>`).join('');
  const stageOptions = stages.map(s => `<option value="${s}">${s}</option>`).join('');
  const statusOptions = statuses.map(s => `<option value="${s}">${s}</option>`).join('');

  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Agent Activity Log — MetalShopping</title>
<style>
  body { font-family: -apple-system, sans-serif; background: #0f172a; color: #e2e8f0; margin: 0; padding: 24px; }
  h1 { color: #f8fafc; margin-bottom: 8px; }
  .meta { color: #94a3b8; font-size: 13px; margin-bottom: 24px; }
  .filters { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
  select, input { background: #1e293b; color: #e2e8f0; border: 1px solid #334155; padding: 6px 10px; border-radius: 6px; font-size: 13px; }
  input { width: 240px; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { background: #1e293b; color: #94a3b8; text-align: left; padding: 10px 12px; border-bottom: 1px solid #334155; position: sticky; top: 0; }
  td { padding: 9px 12px; border-bottom: 1px solid #1e293b; vertical-align: top; }
  tr:hover td { background: #1e293b; }
  code { background: #334155; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
  a { color: #38bdf8; }
  .hidden { display: none; }
  .count { color: #94a3b8; font-size: 13px; margin-bottom: 8px; }
</style>
</head>
<body>
<h1>Agent Activity Log</h1>
<div class="meta">Generated ${new Date().toLocaleString()} · ${entries.length} entries</div>
<div class="filters">
  <input id="search" placeholder="Search tasks, agents, decisions…" oninput="filter()">
  <select id="agentFilter" onchange="filter()"><option value="">All agents</option>${agentOptions}</select>
  <select id="stageFilter" onchange="filter()"><option value="">All stages</option>${stageOptions}</select>
  <select id="statusFilter" onchange="filter()"><option value="">All statuses</option>${statusOptions}</select>
</div>
<div class="count" id="count">${entries.length} entries</div>
<table>
  <thead><tr>
    <th>Time</th><th>Agent</th><th>Stage</th><th>Task</th>
    <th>Status</th><th>Commit</th><th>Decision</th>
  </tr></thead>
  <tbody id="tbody">${rows}</tbody>
</table>
<script>
function filter() {
  const search = document.getElementById('search').value.toLowerCase();
  const agent = document.getElementById('agentFilter').value;
  const stage = document.getElementById('stageFilter').value;
  const status = document.getElementById('statusFilter').value;
  let visible = 0;
  document.querySelectorAll('#tbody tr').forEach(row => {
    const text = row.textContent.toLowerCase();
    const show = (!search || text.includes(search))
      && (!agent || row.dataset.agent === agent)
      && (!stage || row.dataset.stage === stage)
      && (!status || row.dataset.status === status);
    row.classList.toggle('hidden', !show);
    if (show) visible++;
  });
  document.getElementById('count').textContent = visible + ' entries';
}
</script>
</body>
</html>`;
}

const entries = readAllEntries();
writeFileSync(OUTPUT, buildHTML(entries));
console.log(`Generated ${OUTPUT} with ${entries.length} entries.`);
```

- [ ] **Step 3: Add npm script to package.json**

```bash
# In package.json scripts section, add:
"agent-log": "node scripts/generate-agent-log.mjs"
```

- [ ] **Step 4: Verify script runs**

```bash
npm run agent-log
```

Expected: `Generated logs/agent-activity.html with 0 entries.` (zero entries is correct — no tasks have run yet).

- [ ] **Step 5: Verify HTML opens in browser**

Open `logs/agent-activity.html` in a browser. Expected: dark-theme dashboard with filter controls, empty table, "0 entries" count.

- [ ] **Step 6: Add logs/agent-activity.html to .gitignore (generated file)**

```bash
echo "logs/agent-activity.html" >> .gitignore
```

- [ ] **Step 7: Commit**

```bash
git add logs/.gitkeep scripts/generate-agent-log.mjs package.json .gitignore
git commit -m "feat(workflow): add agent activity log infrastructure and HTML dashboard generator"
```

---

## Task 13: End-to-end verification

**Files:** None created — verification only.

- [ ] **Step 1: Start a fresh Claude Code session**

Expected: STATE/NEXT/WATCH briefing appears automatically. No manual prompt needed.

- [ ] **Step 2: Verify code-review works**

Ask Claude to run `/code-review` on any file. Expected: runs without blocklist error.

- [ ] **Step 3: Verify PreFileWrite hook blocks generated files**

Ask Claude to edit any file under `packages/generated/`. Expected: blocked with the violation message.

- [ ] **Step 4: Complete one small task and verify post-task sequence**

Pick any minor task. Complete it. Verify:
- `tasks/todo.md` has the task marked `[x]`
- A commit exists with correct format
- `logs/agent-activity-YYYY-MM.jsonl` has a new entry
- Run `npm run agent-log` — HTML shows the new entry

- [ ] **Step 5: Run $workflow-guardian**

```
/workflow-check
```

Expected: `[workflow-check]` entry appears in `tasks/todo.md` with findings by severity.

- [ ] **Step 6: Final commit**

```bash
git add -A && git commit -m "chore(workflow): verify end-to-end workflow setup complete"
```

---

## Summary

| Task | Spec | Acceptance criteria verified |
|------|------|------------------------------|
| 1. Unblock code-review | Plugin set §1 | AC1 |
| 2. Disable data plugin | Plugin set §2 | AC2 |
| 3. Install commit-commands | Plugin set §3 | AC3 |
| 4. Install claude-mem | Plugin set §4 | AC4 |
| 5. Install claudewatch + rules | Plugin set §5 | AC5 |
| 6. Bootstrap MEMORY.md | Plugin set §6, Session §6 | Plugin AC6, Session AC6 |
| 7. Configure 4 hooks | Orchestrator §6 | Orchestrator AC8, AC9 |
| 8. Rewrite $ms | Orchestrator §1-5 | Orchestrator AC1-7, AC10-12 |
| 9. Create $metalshopping-docs | Orchestrator §7 | Orchestrator AC4, AC5 |
| 10. Create analytics-orchestrator | Orchestrator §7 | Orchestrator AC8 |
| 11. Create $workflow-guardian | Orchestrator §10 | Orchestrator AC14 |
| 12. Logs + HTML generator | Orchestrator §9 | Orchestrator AC13 |
| 13. End-to-end verification | All specs | All ACs |
