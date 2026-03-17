# Implementation Plan

## Goal

Sequence the work so the team can move from planning to implementation without reopening base architecture decisions.

## Phase 0: Freeze the base

Deliverables:

- architecture validated and updated
- global system and engineering principles frozen
- ADRs for the critical freezes
- AGENTS files for repo guidance
- progress tracking in place

Exit criteria:

- tenant isolation rule is frozen
- history ownership rule is frozen
- runtime governance rule is frozen
- frontend thin-client rule is frozen
- async integration rule is frozen
- global operating principles are explicit

## Phase 1: Platform contracts and governance foundation

Deliverables:

- initial `contracts/api`, `contracts/events`, and `contracts/governance` skeletons
- official contract naming and ownership conventions
- initial contract templates
- governance schema strategy
- event versioning conventions
- SDK generation strategy

Exit criteria:

- contract folders have explicit ownership and naming conventions
- initial contract templates are available for repeatable authoring
- governance model is described end-to-end
- no parallel manual type source is planned
- generated targets for TS and Python are defined

## Phase 2: Core platform skeleton

Deliverables:

- `server_core` package boundaries
- module-by-module ownership map
- platform package boundaries for auth, tenancy, governance, db, messaging, observability, security, audit
- module standards and readmodel/event rules
- creation checklists and structural templates
- initial migration strategy for multitenant canonical data

Exit criteria:

- module pattern is fixed
- platform versus domain boundaries are documented
- readmodel and event usage rules are explicit
- templates exist for repeatable module and platform package creation
- no worker owns canonical state

## Phase 3: First implementation wave

Priority order:

1. tenancy and IAM foundation
2. governance runtime skeleton
3. contracts and generated SDK flow
4. first business modules with canonical writes

Exit criteria:

- core can expose basic health and governance-safe bootstrap surfaces
- workers have clear consumption boundaries
- frontend can depend on generated contracts only

## Phase 4: Domain expansion

Candidate order:

1. catalog
2. pricing
3. market_intelligence
4. analytics_serving
5. procurement
6. CRM
7. alerts and notifications

## Ongoing workstreams

- architecture governance
- contract governance
- operational observability and security
- domain sequencing
- migration planning from legacy
