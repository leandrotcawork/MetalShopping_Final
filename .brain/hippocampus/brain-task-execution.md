---
id: hippocampus-brain-task-execution
title: Brain-Task Execution Checklist & Scaffolding
region: hippocampus
tags: [workflow, brain-task, execution, checklist, scaffolding, template]
links:
  - hippocampus/brain-workflow
  - hippocampus/conventions
weight: 0.75
updated_at: 2026-03-25T03:50:00Z
---

# Brain-Task Execution: Checklist & Scaffolding

Reference this during `/brain-task` execution. Copy the scaffolding structure for each phase.

---

## Pre-Execution Checklist

Before running `/brain-task [description]`:

- [ ] Read `hippocampus/brain-workflow.md` (file location rules)
- [ ] Task description is clear
- [ ] Understand if this is: code review | implementation | debugging
- [ ] Know the task domain: backend | frontend | database | infra

---

## Execution Scaffolding

### Step 1: Load Context (Tier 1+2)

**Create:** `.brain/working-memory/context-packet-{task_id}.md`

```yaml
---
task_id: {domain}-{task-type}-{date}
status: in-progress
created_at: {ISO8601}
---

# Task: [description]

## Tier 1: Hippocampus Summary
- Architecture: [from hippocampus/architecture.md — condensed]
- Conventions: [from hippocampus/conventions.md — condensed]
- Top 3 lessons: [from brain.db matching domain]

## Tier 2: Domain Sinapses
- Sinapse 1: [relevant pattern]
- Sinapse 2: [relevant pattern]
- Sinapse 3: [relevant pattern]
```

**Verify:** File is in `.brain/working-memory/`, not `tasks/` or `docs/`

---

### Step 2: Generate Execution Context

#### **For Codex (Implementation/Features)**

**Create:** `.brain/working-memory/codex-context-{task_id}.md`

```yaml
---
task_id: {task_id}
model: codex
domain: [backend|frontend|database|infra]
created_at: {ISO8601}
---

# Task: [description]

## Acceptance Criteria
- [ ] Requirement 1
- [ ] Requirement 2
- [ ] Build passes
- [ ] Tests passing
- [ ] Follows conventions

## Context: Relevant Sinapses
### Architecture Patterns
- [[sinapse-1]] Pattern Name: [how to apply]
- [[sinapse-2]] Pattern Name: [how to apply]

### Code Examples from Codebase
#### Example 1: [Pattern]
\`\`\`[language]
// From: path/to/file.ts
// This shows the CORRECT pattern for X:

[actual code snippet — 5-10 lines]
\`\`\`

#### Example 2: [Pattern]
\`\`\`[language]
// From: path/to/file.go
// This is how we do Y:

[actual code snippet — 5-10 lines]
\`\`\`

## Common Mistakes (DO NOT DO)
- ❌ [Anti-pattern 1 from conventions]
  → Instead, use [correct pattern from sinapse]
- ❌ [Anti-pattern 2 from lesson]
  → See [sinapse] for correct approach

## Previous Similar Work
- lessons/lesson-00XX.md — "Similar pattern, different context"
- cortex/[domain]/lessons/lesson-00YY.md — "Similar implementation"

## Brain Health
- Region: cortex/[domain]
- Sinapses in region: [N]
- Average weight: [0.XX]
- Staleness: [healthy|stale|very stale]
```

**Verify:** File is in `.brain/working-memory/`, contains 2+ real code examples, lists common mistakes

---

#### **For Opus (Debugging Only)**

**Create:** `.brain/working-memory/opus-context-{task_id}.md`

```yaml
---
task_id: {task_id}
model: opus
debug: true
created_at: {ISO8601}
---

# Problem: [error message]

## Stack Trace
[Full stack trace]

## What Was Attempted
- Attempt 1: [result]
- Attempt 2: [result]
- Attempt 3: [result]

## Context: Related Patterns & Lessons
### Debugging Patterns
- [[sinapse-X]] Pattern: [how relates to error]

### Similar Issues
- lessons/lesson-0035.md — "Same error, different module"
- cortex/backend/lessons/lesson-0042.md — "Root cause was X"
```

---

### Step 3: Execute Implementation

*No files created in this step. Implement in source code.*

---

### Step 4: Document Outcomes

**Create:** `.brain/working-memory/task-completion-{task_id}.md`

```yaml
---
task_id: {task_id}
description: [task description]
status: [success|failed]
model_used: [codex|opus|haiku]
complexity_score: [0-100]
duration_minutes: [N]
tokens_estimated: [N]
files_changed: [count]
tests_passed: [yes|no]
created_at: {ISO8601}
---

# Task Completion: [description]

## What Was Built

[Brief summary of what was implemented]

## Files Changed

- path/to/file.tsx — +N lines, -M lines
- path/to/file.go — +P lines
- [...]

## Tests

- Test 1: ✅ PASS
- Test 2: ✅ PASS
- Test 3: ❌ FAIL (reason: [])

## Sinapses Used

[Which sinapses were loaded and applied?]
- [[sinapse-1]] — Used for [pattern X]
- [[sinapse-2]] — Used for [pattern Y]

## Lessons Identified

[During implementation, did you discover something new?]
- New pattern: [description] → should become lesson-XXXX.md
- Mistake found: [description] → document as lesson-XXXX.md
```

---

### Step 5: Propose Sinapse Updates

**Create:** `.brain/working-memory/sinapse-updates-{task_id}.md`

```yaml
---
task_id: {task_id}
created_at: {ISO8601}
status: awaiting-approval
---

# Sinapse Update Proposals

## For Each Sinapse Used:

### Proposal 1: UPDATE: [Sinapse Title]
**Current:** [Existing content]
**Proposal:** [Changes with diffs]
**Reason:** [Why this update is needed]

Status: ✅ APPROVED / ❌ REJECTED / 🤔 REVISE

### Proposal 2: NEW SINAPSE: [Title]
**Location:** cortex/[domain]/
**Content:** [New sinapse content]
**Reason:** [Why this sinapse is needed]

Status: ✅ APPROVED / ❌ REJECTED / 🤔 REVISE
```

---

### Step 6: Archive & Clear Working Memory

**Archive files:**
```bash
# Move from working-memory to permanent location
mv .brain/working-memory/context-packet-{id}.md \
   .brain/progress/completed-contexts/{id}-context-packet.md

mv .brain/working-memory/codex-context-{id}.md \
   .brain/progress/completed-contexts/{id}-codex-context.md

mv .brain/working-memory/task-completion-{id}.md \
   .brain/progress/completed-contexts/{id}-completion-record.md
```

**Create outcome analysis:**
```bash
cat > .brain/progress/completed-contexts/{id}-OUTCOME.md << 'EOF'
---
task_id: {id}
status: [success|failed]
root_cause: [if failed]
created_at: {ISO8601}
---

# Task Outcome: {id}

## Result
[One-sentence summary]

## What Context Had
- Sinapses: [N] (domains: backend, frontend)
- Lessons: [N] (domains: backend, cross-domain)
- Code examples: [N] patterns

## What Worked Well
- [Pattern A from context was helpful]
- [Pattern B prevented mistake X]

## What Was Missing (if any)
- [Missing context: should have loaded lesson-00NN]
- [Missing pattern: new sinapse needed for cortex/domain]

## For Future
- Create lesson-XXXX.md: [pattern found]
- Update sinapse-YYYY.md: [information became stale]
EOF
```

**Update activity log:**
```bash
cat >> .brain/progress/activity.md << 'EOF'

## [timestamp] Task {id}

- **Description:** {description}
- **Model:** [codex|opus|haiku]
- **Status:** [success|failed]
- **Duration:** [M] min
- **Tokens:** ~[N]k
- **Files:** [P] changed
- **Context:** [N sinapses] + [N lessons]
- **Archive:** progress/completed-contexts/{id}-*.md
EOF
```

**Clear working-memory (remove stale files):**
```bash
rm .brain/working-memory/context-packet-{id}.md
rm .brain/working-memory/codex-context-{id}.md
rm .brain/working-memory/task-completion-{id}.md
rm .brain/working-memory/sinapse-updates-{id}.md
```

**Verify:**
- [ ] working-memory/ is empty (or has only other tasks' files)
- [ ] progress/completed-contexts/ has all 4 archived files
- [ ] progress/activity.md updated
- [ ] Source code committed

---

## Verification Checklist (After Task Complete)

- [ ] All working-memory files for THIS task are archived or deleted
- [ ] No stale files in `.brain/working-memory/` from this task
- [ ] `progress/completed-contexts/` has all 4 archived files: context-packet, codex-context, completion-record, OUTCOME
- [ ] `progress/activity.md` appended with task entry
- [ ] `sinapse-updates-{id}.md` awaiting developer approval
- [ ] Source code committed with proper commit message
- [ ] Build passes + tests passing
- [ ] No uncommitted changes

---

## Anti-Patterns Checklist

- [ ] ❌ Did NOT create files in `tasks/`
- [ ] ❌ Did NOT save to `docs/`
- [ ] ❌ Did NOT put in `cortex/` (except sinapses)
- [ ] ❌ Did NOT skip Step 6 cleanup
- [ ] ❌ Did NOT delete working-memory files (archived instead)
- [ ] ✅ All artifacts in `.brain/working-memory/` during execution
- [ ] ✅ All artifacts in `.brain/progress/` after completion

---

## Quick Copy-Paste Commands

**Start task:**
```bash
task_id="domain-task-$(date +%Y-%m-%d)"
# Create: .brain/working-memory/context-packet-${task_id}.md
# Create: .brain/working-memory/codex-context-${task_id}.md
```

**End task:**
```bash
# Archive
mv .brain/working-memory/context-packet-${task_id}.md \
   .brain/progress/completed-contexts/${task_id}-context-packet.md

# Clear
rm .brain/working-memory/*-${task_id}.md
```

---
