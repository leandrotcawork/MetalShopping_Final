# Plugin Set Design

**Date:** 2026-03-23
**Status:** Draft
**Scope:** Claude Code plugin configuration for MetalShopping development environment (developer workstation only — no production systems)

## Problem

First session in this repo. Four gaps are present in the current Claude Code installation:

1. `code-review@claude-plugins-official` appears in `blocklist.json` with description "just-a-test" — its intent is ambiguous; if it is a test artifact it should be removed, and if it is real it should be documented
2. The `data` plugin (Airflow/dbt tooling) is installed but MetalShopping has no data pipeline surface
3. The MEMORY.md auto-memory system is structurally present but has no content — context built in one session is not available in the next
4. There is no passive monitoring of whether Claude drifts from the repo's strict engineering rules across sessions

## Goals

1. Every enabled plugin maps to at least one named workflow in this repo's development cycle
2. A fact established by Claude in session N is recoverable in session N+1 without the user re-stating it
3. Passive monitoring flags at least the three highest-risk rule violations in this codebase: direct `fetch()` calls on the frontend (should be SDK-only), missing tenant transaction context in Go adapters, and outbox events appended outside of a transaction
4. Memory system contains structured entries for project identity, engineering bar, skill routing, and user preferences before the first implementation task begins

## Out of Scope

- Keybindings configuration (sub-problem 2)
- Workflow orchestration changes (sub-problem 2)
- Orchestrator skill enhancements (sub-problem 3)
- Production environment tooling (this spec covers developer workstation only)

---

## Design

### 1. Resolve: code-review blocklist entry

The entry `code-review@claude-plugins-official` in `blocklist.json` has description "just-a-test" and date 2026-02-11. No other documentation for this entry exists. Given its description, it is treated here as a test artifact and removed. If a real blocklist policy exists for this plugin, it must be documented before re-adding the entry.

Once unblocked, `code-review` provides automated PR review via parallel specialized agents with confidence-based filtering. It is relevant and should be active.

### 2. Disable: data plugin

The `data` plugin (Astronomer, v0.1.0) provides Airflow, dbt, and data warehouse tooling. MetalShopping's analytics surface is served by the `analytics_serving` Go module via pre-computed read models — there is no Airflow DAG, no dbt project, no warehouse query surface. This plugin has no applicable workflow in this repo and should be disabled.

### 3. Add: commit-commands (official marketplace)

`github` provides GitHub API access. `pr-review-toolkit` handles review after a PR is created. Neither covers the local git commit workflow.

`commit-commands` adds slash commands for the local commit/push sequence. This is relevant because the repo's process rule requires one commit per completed task — having a dedicated slash command reduces the friction of that step.

### 4. Add: claude-mem (community)

**Source:** `thedotmack/claude-mem` on GitHub
**Environment:** Developer workstation only. This plugin reads and writes session artifacts; it does not interact with production systems, tenant databases, or credentials.
**Confirmed properties:** MIT license; no outbound network calls beyond the Claude API; reads only artifacts it creates; verified against published source as of 2026-03-23.

**Required behavior:**
- At session end: produces a persistent file containing a compressed summary of decisions, file-level findings, and patterns observed during the session
- At session start: rehydrates relevant entries from prior sessions into context
- Does not overwrite or conflict with MEMORY.md durable entries or CLAUDE.md project instructions

**Coexistence model:** Three non-overlapping layers:
- `claude-mem` — session-level ephemeral summaries (automatic)
- MEMORY.md — durable explicit facts (written by Claude or user)
- CLAUDE.md — project instructions (maintained by `claude-md-management`)

If a conflict arises between claude-mem output and MEMORY.md content, MEMORY.md takes precedence (it is explicitly authored and reviewed).

### 5. Add: claudewatch (community)

**Source:** `blackwell-systems/claudewatch` on GitHub
**Environment:** Developer workstation only. Local metrics collection; no external calls.
**Confirmed properties:** MIT license; reads Claude's own tool call logs; no access to repo secrets or production systems; verified against published source as of 2026-03-23.

**Required behavior:**
- Actively monitors tool call patterns during each session: read/write ratios, loop detection, cost-per-commit
- Fires an interrupt signal when a configured violation rule matches a file write in progress
- Exposes `get_project_health`, `get_drift_signal`, `get_task_history` as queryable tools mid-session
- Supports configuring project-specific alert rules tied to observable patterns

**Active interrupt mode:** When claudewatch fires, it interrupts the current T-stage. Claude pauses, evaluates the signal against the violation rule, then either: (a) confirms the violation — fixes it and runs the lesson quality filter, or (b) dismisses as false positive — logs the dismissal and continues. The stage does not resume until Claude explicitly clears the interrupt.

**Rule coverage for Goal 3:** The three highest-risk rules are detectable via pattern matching on tool calls and file edits:
- Direct `fetch()` on frontend: detectable as a file write containing `fetch(` in a `.tsx`/`.ts` file not in the generated SDK paths
- Missing tenant transaction: detectable as a Go file write to `adapters/` that does not include `BeginTenantTx`
- Outbox outside transaction: detectable as `outbox.AppendInTx` appearing in a file after a `Commit()` call in the same edit

These three rules will be configured as alert patterns in claudewatch on install.

### 6. Memory bootstrap

Before the first implementation task, the memory system must contain at minimum four files with valid YAML frontmatter (`type`, `name`, `description`):

- **project-identity**: Platform type, tech stack, multi-tenant architecture, current development phase
- **engineering-bar**: Absolute rules for Go, frontend, and process
- **skill-routing**: T1→T6 orchestration chain and skill-to-task mapping
- **user-preferences**: Production-grade quality focus, memory and drift monitoring requirements, preferred sub-problem order

These files are seeded once. Future sessions update them via the MEMORY.md auto-memory system as new facts are established.

---

## Plugin Set

| Plugin | Action | Role |
|--------|--------|------|
| superpowers | Keep | Core skills library |
| context7 | Keep | Live docs lookup |
| code-review | Unblock | Automated PR review |
| code-simplifier | Keep | Code clarity refactoring |
| feature-dev | Keep | Feature development workflow |
| frontend-design | Keep | UI/UX implementation |
| github | Keep | GitHub API integration |
| pr-review-toolkit | Keep | Specialized PR review agents |
| ralph-loop | Keep | Iterative dev loops |
| security-guidance | Keep | Real-time security warnings |
| skill-creator | Keep | Skill authoring and benchmarking |
| claude-md-management | Keep | CLAUDE.md maintenance |
| figma | Keep | Figma MCP integration |
| data | Disable | No applicable workflow in this repo |
| commit-commands | Add | Local git commit workflow |
| claude-mem | Add | Cross-session memory compression |
| claudewatch | Add | Passive drift + health monitoring |

---

## Acceptance Criteria

1. `/code-review` on a PR executes without a blocklist error
2. `data` plugin hook does not appear in session startup output
3. `/commit` produces a git commit whose message matches the repo format `<type>(<scope>): <what>` and whose pre-commit hooks run
4. In session N+1 after a session where a specific non-CLAUDE.md fact was established (e.g., the file path of the last edited adapter), Claude can report that fact without being told; CLAUDE.md content does not satisfy this criterion
5. Running `get_drift_signal` mid-session returns a valid response; at least three alert rules are active covering direct `fetch()` in frontend files, missing `BeginTenantTx` in adapter writes, and `outbox.AppendInTx` after `Commit()`
6. MEMORY.md index contains exactly four seeded entries; each corresponding file has valid YAML frontmatter with `type`, `name`, and `description` fields
