# Design Spec: MetalShopping Developer Workflow Guide

**Date:** 2026-03-24
**Status:** Approved
**Output:** `docs/workflow-guide.html` (standalone, no server required)

---

## Overview

A personal reference HTML page for Leandro to understand and use the full MetalShopping development workflow. Opens in any browser from the file system (`file://`). No build step, no server, no dependencies.

**Not** an agent activity dashboard — that is a separate deliverable (`npm run agent-log`).

---

## Format & Theme

- **Single standalone HTML file** — works from `file://` path, zero friction
- **Dark theme** — near-black background, electric blue + cyan accents
- **Color palette:**
  - Background: `#0a0f1e` (deep navy)
  - Surface cards: `#111827`
  - Primary accent: `#3b82f6` (electric blue)
  - Flow connectors: `#06b6d4` (cyan, animated dashes)
  - Session ritual nodes: `#7c3aed` (purple)
  - Decision nodes: `#d97706` (amber)
  - T-stage nodes: `#3b82f6` (blue)
  - Agent nodes: `#059669` (green)
  - Post-task nodes: `#475569` (slate)
  - Warning/claudewatch: `#f59e0b` (amber badge)
  - Analytics bypass: `#0d9488` (teal)
  - Text primary: `#f1f5f9`
  - Text secondary: `#94a3b8`

---

## Layout: B — Flow-first + Drill-down

### Sticky Top Bar
- Left: "MetalShopping Workflow" title + version date
- Right: section tabs — **Skills & Plugins | Examples | Commands | Quick Reference**
- Clicking a tab anchor-scrolls to that section below the diagram

### Hero: Interactive Flow Diagram (100% width, ~60vh)
The centerpiece. All workflow nodes rendered as an interactive SVG/HTML diagram. Clicking any node opens the slide-in detail panel.

### Slide-in Detail Panel (400px, right side)
Appears on node click, overlays content. Dismissed with close button or Escape. Contains per-node detail (see Node Panel Structure below).

### Bottom Sections (below the diagram)
Four tabs: Skills & Plugins | Examples | Commands | Quick Reference. Each is a full-width section with anchor ID matching the top nav.

---

## Flow Diagram: Node Map

Nodes organized in rows, rendered left-to-right with animated cyan connector lines.

### Row 1 — Session Ritual (purple)
```
[Session Start] → [Read lessons + todo + MEMORY.md] → [STATE/NEXT/WATCH Output] → [Wait for direction]
```

### Row 2 — Task Intake (amber)
```
[$ms receives task] → [4 Decisions block]
```
The 4 Decisions block is a single clickable node. Its panel expands all four decisions with decision trees.

### Branch — Plan Mode Gate
```
Plan mode YES → [Write plan to todo.md] → [User: approve / revise / reject] → (back to T1 Contract — start of T-stage chain — on approval)
Plan mode NO  → (straight to T-stages)
```

### Row 3 — T-stage Chain (blue, horizontal)
```
[T1 Contract] → [T2 Backend Domain] → [T3a Adapters (Codex)] → [T3b Outbox+Readmodel (Claude only)]
             → [T3.5 Migration] → [T4 SDK Gen] → [T5 Frontend (split)] → [T5.5 Tests (Codex)]
```
T3a and T5.5 have a Codex badge. T3b has a "NEVER Codex" warning badge. T4 has a "Script" badge — it runs `scripts/generate_contract_artifacts.ps1` (PowerShell), not an AI agent. The T4 panel must make this clear.

### Row 4 — Parallel Gate
```
[T6 ADR] (when applicable) ─┐
[T7 Docs — non-blocking]    ├─ run in parallel
[Review Gate — blocking]    ─┘
```

### Row 5 — Post-task (slate)
```
[Post-task sequence] → [claude-mem compression]
```

### Floating Side Nodes (always visible)
- `⚠ claudewatch` (amber) — positioned beside the T-stage row, connected with a dashed line to all T-stages. Click → explains interrupt flow (confirmed violation → fix + lesson filter; false positive → log + continue)
- `$analytics-orchestrator` (teal) — bypass route arrow from `$ms receives task`, labeled "analytics tasks only"
- `$workflow-guardian` (slate) — positioned beside Post-task, showing 4 trigger labels

---

## Node Panel Structure

Every node's slide-in panel follows this format:

| Field | Content |
|-------|---------|
| **Title** | Node name + type badge (e.g. "T-stage", "Decision", "Agent") |
| **What it is** | 2–3 sentence plain-language explanation |
| **When it fires** | Trigger conditions (bullet list) |
| **What happens** | Step-by-step sequence |
| **Example** | One real concrete scenario |
| **Related commands** | CLI snippets with copy button (if applicable) |
| **Connected to** | Clickable links to adjacent nodes |

The **4 Decisions** node panel is the richest — each decision rendered as a mini decision tree with YES/NO branches and concrete examples:
- Plan mode: "New Go module → YES. Single CSS fix → NO."
- Model: "T1 new domain → Opus. Bug fix → Sonnet."
- Claude vs Codex: "T3a adapters, plan complete → Codex. T3b outbox → always Claude."
- Parallel: "T7 + Review always parallel. Research always subagent."

---

## Tab 1 — Skills & Plugins

### Custom Skills (agent skills)
Card per skill: name, one-line purpose, when to invoke, inputs, outputs, "see in flow" link.

Skills covered:
- `$ms` — master orchestrator, routes everything
- `$metalshopping-docs` — T7 delta-only docs update, non-blocking
- `$analytics-orchestrator` — analytics domain entry point
- `$analytics-intelligence` — STUB: 8-layer intelligence model
- `$analytics-surfaces` — STUB: 11 analytics read surfaces
- `$analytics-campaigns` — STUB: campaigns + alerts
- `$analytics-ai` — STUB: AI copilot layer
- `$workflow-guardian` — inspect-and-report, never auto-updates

### Installed Plugins
Card per plugin: name, what it does, how it integrates with the workflow.

Plugins covered:
- `commit-commands` — `/commit` slash command, structured commit format
- `claude-mem` — cross-session context compression, session memory recovery
- `claudewatch` — real-time monitoring, drift detection, cost velocity, MCP tools
- `code-review` — `/code-review` slash command for PR review
- `security-guidance` — injects security rules into every write operation
- `claude-md-management` — CLAUDE.md update post-task
- `superpowers` — planning, subagent execution, brainstorming, finishing branches
- `skill-creator` — creates and improves skills
- `frontend-design` — production-grade frontend UI generation
- `feature-dev` — guided feature development with architecture focus
- `context7` — up-to-date library documentation lookup
- `figma` — Figma design → code implementation
- `pr-review-toolkit` — comprehensive PR review

---

## Tab 2 — Examples

Six real concrete scenarios. Each shows the full path through the flow diagram, with actual file names, commands, and decisions logged.

1. **New backend module from scratch** — full T1→T7 chain, plan mode YES, Sonnet default, Codex on T3a/T5.5
2. **Frontend bug fix** — T5 + T5.5 + Review only, no plan mode, Sonnet, Claude throughout
3. **Analytics task routing** — $ms → $analytics-orchestrator, how routing decision looks
4. **Codex handoff** — what `tasks/codex-handoff.md` looks like, what Codex does, how Claude reviews output
5. **Session start output** — real STATE/NEXT/WATCH example with in-progress task and relevant lesson
6. **Lesson quality filter** — three examples: one that passes (Go adapter pattern), two that fail (UI color, one-off fix)

---

## Tab 3 — Commands

Grouped cheat sheet with copy button on every snippet.

**Web (npm workspaces, from repo root):**
- Typecheck, build, build:wsl, test, test:ci, dev server

**Go (from repo root):**
- All tests, single package tests

**Workflow:**
- `npm run agent-log` — generate HTML activity dashboard
- `/workflow-check` — trigger $workflow-guardian on-demand
- `/commit` — structured commit via commit-commands
- `claudewatch scan` — project health scan
- `claudewatch gaps` — diagnose friction sources

**Contracts:**
- Generate SDK/types (PowerShell)
- Validate contracts

---

## Tab 4 — Quick Reference

One-screen scannable summary. No scrolling required.

**Decision tables (4 boxes):**
1. Plan mode triggers (YES/NO table)
2. Model selection (Opus vs Sonnet, when)
3. Claude vs Codex (qualifying conditions)
4. T-stage ownership (compact table: stage, tool, model, never)

**Lesson quality filter checklist** (3 checkboxes — all must pass to write)

**Post-task 6-step sequence** (numbered, compact)

**$workflow-guardian 4 triggers** (bullet list)

---

## Acceptance Criteria

- [ ] Opens from `file://` in Chrome/Edge with no errors
- [ ] All flow nodes are clickable — panel slides in with correct content
- [ ] Slide-in panel closes on X click and Escape key
- [ ] All 4 bottom tabs scroll to their correct section
- [ ] Copy buttons work on command snippets
- [ ] "See in flow" links in skill cards: clicking closes the tab section, scrolls to the diagram, opens the node's slide-in panel, and applies a 400ms CSS highlight pulse on the node border
- [ ] Page is fully readable on a 1080p monitor without horizontal scroll
- [ ] claudewatch, $analytics-orchestrator, $workflow-guardian floating nodes are always visible
- [ ] T3a and T5.5 show Codex badge; T3b shows NEVER Codex badge
- [ ] All 13 plugins and 8 skills have entries in Tab 1
