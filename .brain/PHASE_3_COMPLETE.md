---
name: Phase 3 Complete
description: ForgeFlow Mini Phase 3 — Intelligence Layer
date: 2026-03-24T21:15:00Z
status: complete
---

# ✅ Phase 3: Intelligence Layer — Complete

**Date:** 2026-03-24 | **Status:** Implementation complete | **Next:** Phase 4 (Visualization)

---

## Deliverables

### 1. brain-mckinsey.md ✅
**File:** `forgeflow-mini/skills/brain-mckinsey.md`
- McKinsey Layer: Strategic decision analysis for high-stakes architectural choices
- **Workflow:** Load strategy context → internal scoring (4 axes, 40/20/20/20 weights) → parallel external research (3 sub-agents) → synthesize 3 alternatives → output recommendation card
- **Output:** `working-memory/mckinsey-output.md` with decision, business impact, benchmarks, alternatives, recommendation, ROI estimate, risk flags, ADR conflicts
- **Token budget:** 20k in / 8k out
- **Integration:** Invoked by brain-task when classification = "architectural"

### 2. brain-consolidate.md ✅
**File:** `forgeflow-mini/skills/brain-consolidate.md`
- Full consolidation cycle: batch process completed tasks, review sinapse updates, surface escalations, generate brain-health.md
- **7-step workflow:**
  1. Inventory completed work
  2. Collect all sinapse update proposals
  3. Batch review display (developer approves/rejects/modifies each)
  4. Escalation check (3+ same-pattern lessons → propose convention)
  5. Generate brain-health.md (staleness, coverage gaps, orphaned sinapses, top 10 weights)
  6. Update brain.db weights (approve +0.02, decay on disuse −0.005/day)
  7. Clear working-memory and archive to progress/activity.md
- **Token budget:** 15k in / 8k out
- **Trigger:** `/brain-consolidate` or auto-suggested after 5 completed tasks

### 3. brain-lesson.md (Updated) ✅
**File:** `forgeflow-mini/skills/brain-lesson.md`
- **Step 4 expanded:** Full pattern escalation workflow
  - Query brain.db for 3+ lessons with same (domain, tag)
  - Extract common pattern/rule
  - Draft convention in hippocampus/conventions.md style
  - Create escalation proposal: `lessons/inbox/escalation-PROPOSAL-[timestamp].md`
  - Output escalation message to developer
  - Wait for approval before writing to hippocampus (immutable)
- **Distributed lesson architecture reference:**
  - Domain-specific: `cortex/<domain>/lessons/`
  - Cross-domain: `lessons/cross-domain/`
  - Temporary: `lessons/inbox/` (for escalation proposals)
- **Curation rule:** Lessons survive only if they meet all 3 criteria (cross-domain, prevents mistakes, architectural)

### 4. brain-document.md (Updated) ✅
**File:** `forgeflow-mini/skills/brain-document.md`
- **Step 2 enhanced:** Guidance on anti-pattern detection
  - If new anti-pattern discovered → do NOT document in cortex sinapses
  - Route to `/brain-lesson` instead
  - Anti-patterns are failures, not architectural patterns
  - brain-lesson handles escalation if 3+ same-type anti-patterns
- **Anti-Patterns table expanded:** Added 3 rows for lesson handling and pattern/failure separation
- **Distributed lesson architecture references:**
  - Lessons live in domain-specific directories or cross-domain
  - Never edit lessons/ directly → use brain-lesson skill
  - Lessons are separate from sinapses (patterns)

---

## Architecture Overview

### Intelligence Layer Components

```
McKinsey Layer (brain-mckinsey.md)
  │
  ├─ Internal Scoring: 4 axes (B/R/E/A), composite formula
  ├─ External Research: 3 parallel sub-agents (web, benchmarks, docs)
  ├─ Synthesis: 3 alternatives with ROI analysis
  └─ Output: working-memory/mckinsey-output.md (recommendation card)

Escalation System (brain-lesson.md + brain-consolidate.md)
  │
  ├─ Query: brain.db for 3+ same-pattern lessons
  ├─ Draft: convention text in hippocampus/conventions.md style
  ├─ Propose: create escalation-PROPOSAL-*.md in lessons/inbox/
  ├─ Review: developer approves/rejects in consolidation cycle
  └─ Promote: move to hippocampus on approval (immutable layer)

Consolidation Cycle (brain-consolidate.md)
  │
  ├─ Batch: collect sinapse update proposals from completed tasks
  ├─ Review: unified diff display, developer approval (A/R/M)
  ├─ Escalate: surface pending escalation proposals
  ├─ Health: generate progress/brain-health.md (staleness, coverage, orphans)
  ├─ Weights: update brain.db (+0.02 approved, −0.005/day decay)
  └─ Archive: clear working-memory, move to progress/activity.md

Pattern Escalation Workflow
  │
  ├─ Trigger: 3+ lessons with same (domain, tag)
  ├─ Evidence: extract common rule from all 3+ lessons
  ├─ Propose: draft convention + create escalation-PROPOSAL-*.md
  ├─ Await: developer review in consolidation cycle
  ├─ Approve: write to hippocampus/conventions.md (immutable)
  └─ Mark: escalated=1 for all source lessons in brain.db
```

### Workflow Integration

```
Task Completes
  ├─ brain-document proposes sinapse updates → working-memory/sinapse-updates.md
  ├─ brain-lesson (if failure) captures anti-pattern → cortex/<domain>/lessons/lesson-XXXXX.md
  ├─ 3+ same-pattern lessons? → brain-lesson creates escalation-PROPOSAL-*.md
  └─ [Repeat for 5 tasks or developer request]

Developer triggers /brain-consolidate
  ├─ Review all sinapse update proposals
  │  └─ Approve (A) / Reject (R) / Modify (M) each
  ├─ Review all pending escalation proposals
  │  └─ Approve → moves convention to hippocampus/conventions.md
  │  └─ Reject → discards proposal, keeps lessons
  ├─ Generate progress/brain-health.md health report
  ├─ Update brain.db weights
  └─ Archive working-memory to progress/activity.md
```

---

## Key Concepts

### McKinsey Layer (Strategic Intelligence)

**When to use:** High-stakes decisions only
- New module / major refactor
- Architecture choice (REST vs gRPC, sync vs async)
- Tech selection (database, caching, messaging)
- Major system boundary change

**Scoring formula:**
```
Composite = (Business Impact × 0.4) + (Tech Risk_inverted × 0.2) + (Effort_inverted × 0.2) + (Strategic Alignment × 0.2)
```

**Output:** Recommendation card with business impact score, external benchmarks, 3 alternatives (A/B/C), ROI estimate, risk flags, ADR conflicts.

### Pattern Escalation (Lessons → Conventions)

**Trigger:** 3+ lessons with same (domain, tag)

**Workflow:**
1. Query brain.db
2. Extract common rule/pattern
3. Draft convention in hippocampus/conventions.md style
4. Create escalation proposal (lessons/inbox/escalation-PROPOSAL-*.md)
5. Wait for developer approval
6. On approval: move to hippocampus/conventions.md (immutable)

**Why:** Prevents lesson duplication. Promotes proven patterns to strategic layer.

### Consolidation Cycle (Batch Knowledge Update)

**Trigger:** `/brain-consolidate` or after 5 completed tasks

**Actions:**
- Review sinapse update proposals (developer A/R/M)
- Surface escalation candidates
- Generate brain-health.md (staleness, coverage, weights)
- Update brain.db weights
- Archive working-memory

**Output:** Summary of approvals, escalations, health report, next consolidation date.

---

## Token Efficiency

| Component | In | Out | Notes |
|---|---|---|---|
| brain-mckinsey | 20k | 8k | High-stakes only; parallel sub-agents |
| brain-consolidate | 15k | 8k | Per consolidation cycle (5+ tasks) |
| brain-lesson (escalation) | 5k | 2k | Only if 3+ lessons exist |
| brain-document (anti-patterns) | 2k | 1k | Routed to brain-lesson, not documented in cortex |

**Typical task flow token budget:**
- Standard task: 100k (map + plan + implement + document)
- High-stakes task: 128k (+ mckinsey layer)
- Consolidation (5 tasks): 15–25k (one-time batch)

---

## Next: Phase 4 (Visualization + Learning)

**Deliverables:**
1. `scripts/visualize.js` — D3.js brain-graph.html generation
2. `brain-status` upgrade — triggers visualization, shows health dashboard
3. `progress/brain-health.md` generation — automated at consolidation
4. `progress/activity.md` — running agent activity log with token costs
5. Cortex evolution — staleness marking, update proposals post-task
6. Context packet trace — working-memory/context-packet.md per task

**Estimated scope:** 2–3 weeks

---

## Verification Checklist

Phase 3 implementation verified when:

- [ ] `/brain-mckinsey "Should we adopt [technology]?"` runs successfully
- [ ] Output produces recommendation card with 3 alternatives + ROI estimate
- [ ] External research benchmarks are from actual searches (not hallucinated)
- [ ] 3+ same-pattern lessons trigger escalation proposal creation
- [ ] Escalation proposal has correct format: frontmatter + proposed convention + evidence
- [ ] `/brain-consolidate` surfaces all pending escalations
- [ ] Developer can A/R/M sinapse update proposals in batch
- [ ] `progress/brain-health.md` is generated with: staleness, coverage, weights, orphans
- [ ] brain.db weights are updated (+0.02 approved, −0.005/day decay)
- [ ] `working-memory/` is cleared after consolidation
- [ ] brain-document routes anti-patterns to brain-lesson (not cortex sinapses)
- [ ] brain-lesson Step 4 has full escalation workflow (7 detailed steps)

---

## Created Files

```
forgeflow-mini/skills/
  ├── brain-mckinsey.md (6.1k) — McKinsey Layer skill
  ├── brain-consolidate.md (11k) — Consolidation cycle skill
  ├── brain-lesson.md (8.6k) — Updated Step 4 with escalation
  └── brain-document.md (6.9k) — Updated Step 2 + anti-patterns table
```

---

**Status:** ✅ Phase 3 Intelligence Layer complete
**Next:** Phase 4 Visualization + Learning
**Created:** 2026-03-24T21:15:00Z
