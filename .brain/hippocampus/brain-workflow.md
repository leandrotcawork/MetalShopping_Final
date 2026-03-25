---
id: hippocampus-brain-workflow
title: Brain Workflow & File Structure
region: hippocampus
tags: [workflow, brain-task, file-structure, execution, guardrails]
links:
  - hippocampus/conventions
  - hippocampus/strategy
weight: 0.80
updated_at: 2026-03-25T03:50:00Z
---

# Brain Workflow: File Locations & Execution

This guide prevents errors when using `/brain-task` — defines where artifacts go at each step.

---

## File Location Map

### ✅ Correct Locations (During brain-task Steps 1-6)

| Step | File Pattern | Location | Purpose |
|------|------|----------|---------|
| 1 | `context-packet-{id}.md` | `.brain/working-memory/` | Assembled sinapses from Tier 1+2 |
| 2 | `codex-context-{id}.md` | `.brain/working-memory/` | Codex execution context |
| 2 | `opus-context-{id}.md` | `.brain/working-memory/` | Opus execution context (debugging) |
| 4 | `task-completion-{id}.md` | `.brain/working-memory/` | Task outcome + files + tests + lessons |
| 5 | `sinapse-updates-{id}.md` | `.brain/working-memory/` | Proposed sinapse updates (awaiting approval) |
| 6 | `{id}-*.md` | `.brain/progress/completed-contexts/` | Archived execution artifacts |

### ❌ Anti-Patterns (Common Mistakes)

| ❌ Wrong | ✅ Right | Why |
|---|---|---|
| `tasks/analytics-home-review.md` | `.brain/working-memory/task-completion-[id].md` | Task artifacts are ephemeral |
| `tasks/implementation-plan.md` | `.brain/working-memory/codex-context-[id].md` | Plans stay in working-memory during execution |
| `docs/brain-review.md` | `.brain/working-memory/task-completion-[id].md` | Docs are permanent; task artifacts are temporary |
| `.brain/cortex/task-xyz.md` | `.brain/working-memory/` | cortex stores sinapses, not tasks |

---

## Task ID Naming

Format: `{domain}-{task-type}-{date}`

**Examples:**
- `analytics-home-review-2026-03-25` (code review)
- `add-button-feature-2026-03-25` (feature implementation)
- `debug-crash-fix-2026-03-25` (debugging)

---

## Pre-Execution Checklist

Before running `/brain-task [description]`:

- [ ] I've read this file (hippocampus/brain-workflow.md)
- [ ] Task is clear (what are we building/reviewing/fixing?)
- [ ] I know which steps apply (code review? implementation? debugging?)

---

## File Lifecycle

### Phase 1: Execution (Steps 1-5) — In working-memory/

Files are **temporary, task-specific**:
```
.brain/working-memory/
├── context-packet-[id].md
├── codex-context-[id].md
├── task-completion-[id].md
└── sinapse-updates-[id].md
```

**Lifetime:** During task execution only. Cleaned up in Step 6.

### Phase 2: Archive (Step 6) — In progress/completed-contexts/

Files are **permanent, historical**:
```
.brain/progress/completed-contexts/
├── [id]-context-packet.md
├── [id]-codex-context.md
├── [id]-completion-record.md
└── [id]-OUTCOME.md
```

**Lifetime:** Never deleted. Data for future pattern learning.

### Phase 3: Activity Log — In progress/activity.md

Single file tracking all tasks:
```
progress/activity.md

## [timestamp] Task [id]
- Description: [task]
- Model: [codex|opus|haiku]
- Status: [success|failed]
- Duration: [N] min
- Files: [N] changed
```

---

## Step-by-Step Execution

### Step 1: Load Context
```bash
# Create: .brain/working-memory/context-packet-{task_id}.md
# Contents: Hippocampus summary + top lessons + domain sinapses
```

### Step 2: Generate Execution Context
```bash
# Create: .brain/working-memory/codex-context-{task_id}.md
# OR: .brain/working-memory/opus-context-{task_id}.md
# Contents: Task + acceptance criteria + code examples + mistakes to avoid
```

### Step 3: Execute (No files created)
Implement feature / fix bug / review code in source tree.

### Step 4: Document Outcomes
```bash
# Create: .brain/working-memory/task-completion-{task_id}.md
# Contents: Summary + files changed + test results + lessons identified
```

### Step 5: Propose Sinapse Updates
```bash
# Create: .brain/working-memory/sinapse-updates-{task_id}.md
# Contents: For each sinapse used, propose updates (developer reviews)
```

### Step 6: Archive & Clear
```bash
# Move from working-memory → progress/completed-contexts/
mv working-memory/context-packet-{id}.md \
   progress/completed-contexts/{id}-context-packet.md

# Create outcome analysis
cat > progress/completed-contexts/{id}-OUTCOME.md << EOF
---
task_id: {id}
status: [success|failed]
---
# Task Outcome
## Result
[Brief summary]
## What Context Had
[Sinapses loaded]
## What Worked Well
[Patterns that helped]
EOF

# Clear working-memory (no stale files)
rm working-memory/context-packet-{id}.md
rm working-memory/codex-context-{id}.md
rm working-memory/task-completion-{id}.md
rm working-memory/sinapse-updates-{id}.md

# Update activity log
echo "## [timestamp] Task {id}
- Description: [task]
- Model: [codex]
- Status: [success]
- Duration: [N] min
- Files: [N] changed
" >> progress/activity.md
```

---

## Guardrails Summary

| Question | Answer | Rule |
|----------|--------|------|
| Where do I create execution artifacts? | `.brain/working-memory/` | Always during Steps 1-5 |
| When do I archive them? | Step 6 (after completion) | Move to `progress/completed-contexts/` |
| Should I commit working-memory files? | No | They're ephemeral; commit source code only |
| Can I put reviews in `tasks/`? | No | Reviews go to `working-memory/task-completion-[id].md` |
| When do I delete working-memory files? | Never delete—archive them | Archive in Step 6 to `progress/completed-contexts/` |

---

## Why This Structure?

**working-memory/ = Temporary**
- Current task context only
- Cleaned up after completion
- Not persisted (not committed to git)

**progress/completed-contexts/ = Permanent**
- Historical record for pattern learning
- Archived with task-id prefix
- Data for "did we solve this before?"

**tasks/ = Sprint Backlog**
- Manual action items (todo.md)
- Permanent lessons (lessons.md)
- NOT for task execution artifacts

---
