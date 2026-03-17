# AGENTS

## Purpose

This repository is in planning-first mode. Before writing product code, agents must use the architecture and planning docs as the source of truth so the team does not drift or reopen base decisions.

## Current phase

- Current phase: planning and definition
- Default work: architecture, ADRs, contracts planning, delivery plans, progress tracking
- Do not implement product code unless the user explicitly asks to leave planning mode for a specific task

## Reading order

Read only what is needed, in this order:

1. `README.md`
2. `docs/PROJECT_SOT.md`
3. `ARCHITECTURE.md`
4. Relevant ADRs in `docs/adrs/`
5. The nearest local `AGENTS.md`

## Authority order

When documents disagree, follow this order:

1. `docs/adrs/*.md`
2. `docs/PROJECT_SOT.md`
3. `ARCHITECTURE.md`
4. `README.md`

## Token discipline

- Do not scan the whole repository unless the task truly needs it.
- Prefer reading the smallest relevant document set.
- Do not restate large architecture texts when a short reference is enough.
- Update the SoT docs when a planning decision changes instead of duplicating the same guidance in many places.

## Planning rules

- Keep `server_core` as the canonical sync core.
- Keep workers separate from canonical state ownership.
- Keep `contracts/` as the only source of truth for APIs, events, and governance schemas.
- Keep frontend surfaces thin.
- Keep planning artifacts explicit: SoT, ADRs, implementation plan, and progress.

