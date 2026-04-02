# Design Spec: MetalShopping Governance Audit and Documentation Consolidation

**Date:** 2026-04-01
**Status:** Approved
**Output:** Governance-aligned updates to `AGENTS.md`, `CLAUDE.md`, `CODEX.md`, `docs/PROJECT_SOT.md`, `ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md`

---

## Overview

MetalShopping needs a documentation and agent-governance model that scales like a long-lived enterprise platform, not a one-off implementation project. The immediate goal is to make the repository unambiguous for AI agents and humans: one clear operational source of truth, one stable architectural blueprint, short normative entrypoints for agents, and no duplicated rule sets drifting over time.

This work is not a generic doc cleanup. It is a governance hardening pass so future module planning, implementation, and AI-assisted execution happen against stable rules, explicit precedence, and a product direction designed for long-term growth.

---

## Problem Statement

The repository already has strong documentation, but the current governance layer has structural drift:

- `AGENTS.md` and `CLAUDE.md` duplicate core rules
- `CODEX.md` did not exist even though Codex should follow the same repository rules
- `docs/PROJECT_SOT.md` is the best candidate for operational SoT, but does not yet explicitly own hierarchy and precedence
- `fulldocs` is currently the strategic source for this restructuring, but that precedence is not yet codified
- `CLAUDE.md` references `docs/ARCHITECTURE.md`, while the architecture file currently lives at repo root as `ARCHITECTURE.md`
- `AGENTS.md` and `CLAUDE.md` require `tasks/lessons.md` and `tasks/todo.md`, but those files were not present in the repository at the start of this tranche

Without resolving these issues first, every future plan risks being correct in intent but weak in execution discipline.

---

## Goals

1. Make `docs/PROJECT_SOT.md` the explicit operational source of truth for the project
2. Define a professional hierarchy between strategy, architecture, execution, progress, and agent instructions
3. Ensure `AGENTS.md`, `CLAUDE.md`, and `CODEX.md` are short, normative, and aligned
4. Keep architecture stable and separated from operational tracking
5. Create an update model that prevents drift between `fulldocs`, `docs/`, and agent entrypoint files
6. Prepare the repository for a future master orchestration plan and one-module-at-a-time implementation planning

---

## Non-Goals

- Rewriting every domain document in `fulldocs`
- Producing detailed implementation plans for all modules now
- Freezing every future module decision in this pass
- Replacing ADRs as the mechanism for binding architectural decisions

---

## Approaches Considered

### Approach A: Large self-contained agent files

Put most governance, architecture, and operational rules directly inside `AGENTS.md`, `CLAUDE.md`, and `CODEX.md`.

**Pros**
- Fewer indirections for an agent starting work
- Simple mental model at first glance

**Cons**
- High duplication
- Fast drift across files
- Hard to maintain as the platform grows
- Weak separation between project truth and tool-specific behavior

### Approach B: Central SoT with short normative agent entrypoints

Use `docs/PROJECT_SOT.md` as the operational source of truth, keep `ARCHITECTURE.md` as the stable technical blueprint, and make `AGENTS.md`, `CLAUDE.md`, and `CODEX.md` short, strict, and referential.

**Pros**
- Strong hierarchy and low duplication
- Easier to maintain at scale
- Clear precedence for humans and agents
- Best fit for a multi-agent, multi-module, long-lived product

**Cons**
- Requires disciplined cross-document linking
- Requires every entrypoint file to be explicit about precedence

### Approach C: Distributed ownership without one central SoT

Keep each document as the owner of its own area without a strict top-level operational SoT.

**Pros**
- Flexible for ad hoc work
- Lower upfront consolidation effort

**Cons**
- Weakest model for AI agents
- Increases ambiguity during execution
- Makes conflict resolution slower and more subjective

### Recommendation

Choose **Approach B**.

This gives MetalShopping the most scalable and professional governance model: strategy can evolve, architecture can stay stable, execution can remain incremental, and agents can follow a strict hierarchy without inventing local interpretations.

---

## Target Governance Model

### 1. Source of Truth Hierarchy

The repository should adopt this precedence:

1. `docs/PROJECT_SOT.md`
2. `ARCHITECTURE.md`
3. `AGENTS.md`
4. `CLAUDE.md` and `CODEX.md`
5. `docs/IMPLEMENTATION_PLAN.md`
6. `docs/PROGRESS.md`
7. `tasks/todo.md`
8. `tasks/lessons.md`
9. module specs and implementation plans under `docs/superpowers/`

For this restructuring phase, `fulldocs` is the accepted strategic input source. Once decisions are consolidated into repository docs, agents should operate from the repository SoTs, not from `fulldocs` directly.

### 2. Document Roles

`docs/PROJECT_SOT.md`
- central operational truth
- product identity
- state of the platform
- hierarchy and precedence rules
- update discipline
- macro execution direction

`ARCHITECTURE.md`
- stable architecture thesis
- ownership boundaries
- core/worker/client interaction model
- multi-tenant and governance model
- long-term technical direction

`AGENTS.md`
- mandatory session ritual
- absolute project rules
- required read order
- stop-and-fix conditions

`CLAUDE.md` and `CODEX.md`
- agent-specific operating instructions
- same normative core rules
- no divergence on project truth
- only tool-specific differences where strictly necessary

`docs/IMPLEMENTATION_PLAN.md`
- macro phase ordering
- high-level execution roadmap
- no deep module specs

`docs/PROGRESS.md`
- actual completion state
- gates, blockers, next steps
- no strategy duplication

`tasks/todo.md`
- session-control tracker for active state, current tranche, and task checkboxes
- does not override repository SoTs

`tasks/lessons.md`
- session-control tracker for durable corrections and lessons
- does not override repository SoTs

### 3. Update Protocol

Use this flow:

1. New strategic insight starts in `fulldocs` or structured discussion
2. Accepted operational truth moves into `docs/PROJECT_SOT.md`
3. Stable architecture changes go through ADRs and then into `ARCHITECTURE.md` if needed
4. Macro sequencing changes update `docs/IMPLEMENTATION_PLAN.md`
5. Status changes update `docs/PROGRESS.md`
6. Agent behavior changes update `AGENTS.md`, `CLAUDE.md`, and `CODEX.md`

### 4. Anti-Drift Rules

- no document should restate another document in full
- each document must point to the canonical owner of a topic
- `AGENTS.md`, `CLAUDE.md`, and `CODEX.md` must explicitly require reading `docs/PROJECT_SOT.md`
- conflicts must resolve in favor of `docs/PROJECT_SOT.md`
- if a rule is shared by Claude and Codex, both files must stay aligned
- `tasks/todo.md` and `tasks/lessons.md` are lower precedence than repository SoTs

---

## Initial Audit Scope

### Phase 1: Mandatory agent entrypoints

- `AGENTS.md`
- `CLAUDE.md`
- `CODEX.md` (new)
- `docs/PROJECT_SOT.md`

### Phase 2: Structural support documents

- `ARCHITECTURE.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- relevant ADRs that define frozen platform behavior

### Immediate gaps to resolve in Phase 1

- missing `CODEX.md`
- missing `tasks/lessons.md`
- missing `tasks/todo.md`
- mismatched architecture file reference
- undocumented precedence between `fulldocs` and repository docs

---

## Execution Sequence After This Spec

1. Consolidate the normative governance layer
2. Freeze hierarchy and precedence in `docs/PROJECT_SOT.md`
3. Align `AGENTS.md`, `CLAUDE.md`, and `CODEX.md` with the same mandatory read order and project rules
4. Align `ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md` with the new governance model
5. Create a master orchestration plan for sequencing, tracking, and cross-module dependencies
6. Move to one module spec and one implementation plan at a time

---

## Acceptance Criteria

- `docs/PROJECT_SOT.md` explicitly declares documentation precedence and operational SoT ownership
- `AGENTS.md`, `CLAUDE.md`, and `CODEX.md` share the same mandatory read order and absolute rules
- `CLAUDE.md` and `CODEX.md` do not diverge on project truth
- `ARCHITECTURE.md` remains architectural and does not become a progress tracker
- `docs/IMPLEMENTATION_PLAN.md` and `docs/PROGRESS.md` are aligned with the new governance hierarchy
- the repository has a clear starting point for a future master orchestration plan

---

## Notes

- This design intentionally optimizes for long-term scalability and AI execution discipline over local convenience.
- The product objective must remain explicit across the governance layer: MetalShopping is a long-lived enterprise platform for commercial strategy, pricing, procurement, CRM, analytics, automation, and future AI-assisted operations.

