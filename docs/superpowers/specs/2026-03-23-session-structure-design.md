# Session Structure Design

**Date:** 2026-03-23
**Status:** Draft
**Scope:** Claude Code session start/end rituals, task completion automation, lesson quality filtering, and Codex handoff protocol

## Problem

Working solo on MetalShopping v1 — a production-intended platform being built from scratch. Sessions are task-driven (always following a plan from `tasks/todo.md`). Three structural problems exist in the default workflow:

1. Each session starts cold — no automatic briefing on where things stand, what was in progress, what's next
2. File hygiene (`tasks/todo.md`, `tasks/lessons.md`, CLAUDE.md) depends on manual discipline, which erodes over time
3. No defined protocol for handing off to OpenAI Codex when Claude Code reaches capacity, leaving Codex without the context it needs to produce consistent, constraint-compliant code

## Goals

1. Every session starts with an accurate, scannable briefing — no manual context reconstruction
2. `tasks/todo.md` and `tasks/lessons.md` reflect current state after every task; incomplete state is surfaced immediately, not silently lost
3. Lessons captured are high-signal only — architectural, structural, logic-level patterns; never UI appearance or one-off corrections
4. When work moves to Codex, it receives a self-contained context package sufficient to produce code consistent with this repo's absolute rules
5. All Codex output is reviewed by Claude against stated constraints before any commit, with an auditable record

## Out of Scope

- Codex skill definitions and `$ms` orchestrator updates (sub-problem 3)
- Plugin installation and configuration (sub-problem 1)
- Keybindings

**Note on Codex handoff:** The handoff protocol (Section 4) is in scope here because it governs session behavior — when and how Claude pauses and packages work. What Codex does with that package, and how `$ms` is updated to be two-tool-aware, is sub-problem 3.

---

## Supporting systems

This spec depends on three tools defined and installed in sub-problem 1:
- **`claude-mem`** — compresses session context at task completion; rehydrates at session start
- **`claude-md-management`** — updates CLAUDE.md with session-level pattern changes
- **`commit-commands`** — executes git commit with enforced message format and pre-commit hooks

**MEMORY.md** is the index file for the persistent memory system (also defined in sub-problem 1). It lives at the repo root and is populated with structured `.md` files covering project identity, engineering bar, skill routing, and user preferences. Claude reads it at session start to recover non-CLAUDE.md context from prior sessions.

---

## Design

### 1. Session Start Ritual

On session open, Claude reads in this order:
1. `tasks/lessons.md` — apply every active rule before touching code
2. `tasks/todo.md` — identify current state
3. MEMORY.md — recover prior-session context not present in CLAUDE.md

Then outputs a fixed-format briefing:

```
STATE: [clean | in-progress: <task name> | blocked: <reason>]
NEXT: <next incomplete task from todo.md>
WATCH: <lesson relevant to NEXT, if any>
```

**Rationale for these three fields:**
- STATE answers "where am I?" — prevents starting duplicate work on an in-progress task
- NEXT answers "what's next?" — removes the need to scan todo.md manually
- WATCH surfaces the one lesson most likely to matter for the upcoming task, reducing the chance of repeating a known mistake

**WATCH relevance logic:** Claude scans lesson entries for keywords and domain tags matching the next task's type (Go adapter, frontend component, contract, etc.). If multiple lessons match, the most recently added one is shown. If none match, WATCH is omitted.

Claude waits for user direction after the briefing. It does not start any task autonomously.

**Failure recovery:** If any of the three files is missing or unreadable, Claude reports it in the briefing (e.g., `STATE: unreadable — tasks/todo.md not found`) and waits for user resolution before proceeding.

### 2. Post-Task Automation

Every task completion triggers this sequence automatically:

1. **Mark done:** Task marked `[x]` in `tasks/todo.md`
2. **Evaluate for lesson:** Apply quality filter (Section 3); append to `tasks/lessons.md` if it passes
3. **Commit:** `commit-commands` produces one commit with format `<type>(<scope>): <what>`, pre-commit hooks run
4. **Capture learnings:** `claude-md-management` updates CLAUDE.md with session-level patterns
5. **Compress context:** `claude-mem` captures session context in background

**Partial failure handling:** If the sequence fails at any step, the failure is reported immediately and the remaining steps do not run. The incomplete state is visible in the next session's briefing (STATE will show `in-progress` or a specific error). No silent data loss — the user always knows what did and did not complete.

Session end requires no action. Files are current after every task. There is nothing to flush.

### 3. Lesson Quality Filter

A correction is written to `tasks/lessons.md` only if it passes ALL three criteria:

1. If repeated, the mistake would cause a **bug, review failure, or structural/architectural problem**
2. The correction concerns **code logic, architecture, backend requirements, module patterns, or data flow** — not UI appearance
3. It describes a **pattern** — something likely to recur, not an isolated incident

**Not a lesson (silently dropped):**
- Wrong color, label text, padding, or spacing
- Typos or copy errors
- Any correction fully explained by a `git diff`

**Is a lesson:**
- A component duplicated one that already existed in `packages/ui/src/index.ts`
- An SDK type was assumed non-nullable but was nullable at runtime
- An absolute rule (SDK-only, tenant tx, outbox in-tx) was violated during implementation

**Known tradeoff:** A genuine architectural mistake that occurred once and is unlikely to recur will be dropped by criterion 3. This is intentional — `tasks/lessons.md` is a pattern library, not a post-mortem log. One-time architectural decisions belong in ADRs.

Claude applies the filter automatically after every user correction. The user can always add a lesson manually by appending directly to `tasks/lessons.md`.

### 4. Codex Handoff Protocol

#### When to hand off

Hand off to Codex when the task meets ALL of the following:

- The plan is complete: all file paths are named, all constraints are listed, no section is marked TBD
- No architectural decisions remain open for this task
- The task does not involve auth, tenant isolation, or outbox logic (these require Claude review-in-flight)

Additionally hand off when Claude Code's capacity is reached. Since token exhaustion cannot be detected in advance, use this proxy signal: if the task requires creating or modifying more than 3 files or more than 200 lines of net-new code, proactively hand off rather than risk context degradation mid-task.

#### The handoff document

Claude writes `tasks/codex-handoff.md` before any work moves to Codex:

```markdown
# Codex Handoff — <task name>

## Task
<exact task description from todo.md>

## Files to create or modify
- <path>: <what to do>

## Constraints (absolute — do not violate)
<relevant subset of repo absolute rules for this task>

## Pattern to follow
<file path of most relevant existing example>
<description of the pattern to replicate>

## Definition of done
<what Claude will verify on review>
```

This document is self-contained. Codex does not need access to this conversation or prior sessions.

#### On Codex output return — standard tasks

1. Claude reads output against the constraints in the handoff document
2. If all constraints are satisfied: commit via `commit-commands`; commit message includes `reviewed-by: claude` trailer
3. If a constraint is violated: Claude fixes it directly, does not send back to Codex; the fix runs through the lesson quality filter — Codex has access to `tasks/lessons.md` and recurring violations should be captured there
4. Delete `tasks/codex-handoff.md` after successful commit

#### On Codex output return — high-stakes tasks

High-stakes tasks are those touching auth, tenant isolation, or outbox logic. For these:

1. Claude reviews against handoff constraints
2. Claude writes a review note to `tasks/todo.md` under the task: `REVIEW: <what was checked, what passed, what was fixed>`
3. User reviews Claude's note and the diff before approving the commit
4. Commit happens only after explicit user approval
5. Commit message includes both `reviewed-by: claude` and `approved-by: user` trailers

---

## Acceptance Criteria

1. Session start outputs the STATE/NEXT/WATCH briefing before any other action; if a file is missing, the briefing reports the error instead of proceeding silently
2. After every task completion, `tasks/todo.md` has the task marked `[x]` and a commit exists — without user prompting
3. If the post-task sequence fails at step 3 (commit), steps 4 and 5 do not run, and the next session briefing shows `STATE: in-progress: <task name>`
4. After a UI color correction, `tasks/lessons.md` is unchanged
5. After an SDK type violation is caught, `tasks/lessons.md` contains a new entry describing the pattern
6. `tasks/codex-handoff.md` exists and is complete (no TBD fields) before any work moves to Codex
7. Every commit of Codex output contains a `reviewed-by: claude` trailer in the commit message
8. For high-stakes Codex tasks, a REVIEW note exists in `tasks/todo.md` and the commit contains both trailers before the commit is made
9. `tasks/codex-handoff.md` does not exist after a successful commit of Codex output
