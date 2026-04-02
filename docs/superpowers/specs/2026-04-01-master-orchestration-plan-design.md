# Design Spec: MetalShopping Master Orchestration Plan

**Date:** 2026-04-01
**Status:** Approved
**Output:** A master orchestration document that acts as the live execution index for MetalShopping across product fronts and transversal fronts

---

## Overview

MetalShopping now has the base governance layer consolidated. The next missing artifact is not another deep module plan. It is a program-level orchestration document that keeps execution ordered, visible, and dependency-aware across the whole platform.

The master orchestration plan should become the live coordination index for the repository. It should show which fronts exist, which are active, which are blocked, which depend on other fronts, and which front should open the next spec.

This document is intentionally one level above module specs and implementation plans. It exists to prevent fragmented planning and to preserve a disciplined execution order as the platform grows.

---

## Problem Statement

MetalShopping already has:

- an operational SoT
- a stable architecture document
- a progress tracker
- multiple focused specs, ADRs, and implementation plans

What it still lacks is a single orchestration layer that answers these questions clearly:

- what major fronts exist across the product and platform
- which fronts are already executed, active, blocked, or waiting
- what depends on what
- which transversal work affects multiple fronts
- which front should be specified next

Without this layer, planning can stay locally correct while becoming globally inefficient. Teams or agents can open deep specs in the wrong order, miss cross-front impacts, or lose visibility of already completed groundwork.

---

## Goals

1. Create a live index of execution across MetalShopping
2. Map both product fronts and transversal fronts in one orchestration view
3. Make dependencies, blockers, and cross-front impact explicit
4. Define which front is recommended to open next for detailed spec work
5. Keep the orchestration document above module specs and implementation plans
6. Preserve disciplined sequencing: one spec at a time, then one implementation plan at a time

---

## Non-Goals

- replacing `docs/PROJECT_SOT.md` as the operational root
- replacing `docs/PROGRESS.md` as the factual progress tracker
- replacing module specs or implementation plans
- describing detailed implementation steps for every front
- producing all front-specific action plans in the same pass

---

## Recommended Model

Use a hybrid orchestration structure:

- organize the document primarily around product fronts
- include a dedicated section for transversal fronts
- make cross-front dependencies explicit between both sections

This is the best fit for MetalShopping because:

- product fronts preserve business and delivery visibility
- transversal fronts preserve structural and platform visibility
- the combination prevents planning blind spots

If the document were organized only by technical layers, product sequencing would become hard to reason about. If it were organized only by business domains, shared platform and governance work would disappear from orchestration. The hybrid model keeps both visible.

---

## Document Role

The master orchestration plan is a coordination document.

It should:

- identify the major execution fronts
- show their state
- record dependency direction
- record impact across fronts
- choose the recommended next front for detailed planning

It should not:

- become a duplicate backlog
- restate architecture in full
- replace progress evidence
- contain step-by-step implementation instructions

---

## Operating Rules

The document must enforce these rules:

1. It does not override `docs/PROJECT_SOT.md`, `ARCHITECTURE.md`, `AGENTS.md`, or `docs/PROGRESS.md`.
2. It decides which front should open the next detailed spec.
3. No front gets an implementation plan before it has an approved spec.
4. Transversal fronts may block or unlock multiple product fronts and that must be stated explicitly.
5. Each front must have a normalized status.
6. Cross-front dependencies must be concrete, not vague.
7. The document must end with a recommended next front or ordered next fronts.

---

## Status Model

Use this normalized status set:

- `done`
- `in progress`
- `ready for spec`
- `waiting on dependency`
- `blocked`

The orchestration document should not invent alternative labels unless there is a durable reason to update the status model globally.

---

## Document Structure

The master orchestration plan should use this high-level structure:

1. `Purpose`
2. `How to use this document`
3. `Execution statuses`
4. `Product fronts`
5. `Transversal fronts`
6. `Cross-front dependency rules`
7. `Recommended next fronts`

This keeps the document readable while preserving enough structure to operate as a real execution index.

---

## Front Template

Each front should use the same concise template:

- `Objective`
- `Current status`
- `Why it matters now`
- `Depends on`
- `Unblocks`
- `Existing artifacts`
- `Next artifact to create`

This template is intentionally short. The orchestration layer should point to deeper artifacts, not become one.

---

## Product Fronts to Map Initially

The initial master orchestration plan should map, at minimum, these product fronts:

- `Home`
- `Shopping`
- `Analytics`
- `CRM`
- `Catalog`
- `Pricing`
- `Inventory`
- `Procurement`

These fronts reflect the current product direction and existing repository language. More fronts may be added later if the platform introduces a clearly separate execution stream.

---

## Transversal Fronts to Map Initially

The initial master orchestration plan should map, at minimum, these transversal fronts:

- `Governance`
- `Contracts`
- `SDK generation`
- `Auth/session`
- `Frontend migration`
- `Workers/integrations`
- `CI/quality gates`
- `Observability/security`

These fronts matter because they either enable multiple product fronts or can block several of them at once.

---

## Cross-Front Expectations

The orchestration document should explicitly recognize cross-front relationships such as:

- `Analytics` depending on contracts, governance, read models, and surface decisions
- `CRM` depending on identity, events, and upstream product signals
- `Shopping` depending on suppliers runtime, procurement boundaries, and frontend parity
- `Procurement` depending on inventory, pricing boundaries, and integration inputs
- `Frontend migration` affecting `Home`, `Shopping`, `Analytics`, and `CRM`
- `SDK generation` affecting every thin-client surface

These are not full implementation details. They are orchestration-level dependencies and should remain stated at that level.

---

## Update Discipline

The master orchestration plan should be updated when:

- a front changes status
- a dependency changes
- a new front becomes relevant
- a spec is opened for a front
- an implementation plan is opened for a front
- a front is blocked or unblocked by transversal work

It should not be updated for every local code change unless that change moves front-level orchestration state.

---

## Acceptance Criteria

- the repository has one explicit master orchestration document
- both product fronts and transversal fronts are mapped
- each mapped front has a normalized status
- dependencies and cross-front impact are explicit
- the document points to existing artifacts where available
- the document identifies the recommended next front for detailed spec work
- the document stays orchestration-level and does not collapse into detailed implementation planning

---

## Notes

- This document should become the bridge between governance and detailed module planning.
- The next stage after this spec is to write the master orchestration plan itself, not yet the detailed action plans for every front.
- Once the orchestration layer exists, the project should continue with one approved front spec at a time and one implementation plan at a time.
