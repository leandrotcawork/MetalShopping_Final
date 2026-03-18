---
name: metalshopping-frontend-migration-guardrails
description: Review or implement MetalShopping frontend migration work when preserving the legacy visual language while replacing weak frontend architecture with thin-client package boundaries, generated contracts, reusable widgets, and feature-local adapters. Use when porting legacy screens, validating `apps/web` and `packages/*` structure, deciding whether a legacy frontend pattern should be reused, or preventing DTO, CSS, page-logic, and API-communication drift during the frontend rebuild.
---

# MetalShopping Frontend Migration Guardrails

## Overview

Use this skill whenever frontend migration work touches legacy reuse, package boundaries, widgets, DTOs, CSS, feature composition, or API communication.

This skill is for preserving what the legacy frontend got right visually while refusing to carry forward the parts that were not scalable.

## Workflow

1. Read only the minimum frozen context:
   `docs/PROJECT_SOT.md`
   `docs/OPERATIONAL_SURFACES_PLAN.md`
   `docs/PRODUCTS_SURFACE_IMPLEMENTATION_PLAN.md`
   `docs/PRODUCTS_READMODEL_OWNERSHIP.md`
   `docs/FRONTEND_QUALITY_GATES.md`
   `docs/FRONTEND_MIGRATION_CHARTER.md`
   `docs/FRONTEND_MIGRATION_PLAYBOOK.md`
   `docs/adrs/ADR-0005-thin-clients-and-generated-sdks.md`
2. If a legacy surface is involved, inspect the shell, typography, widget, and page files needed to understand the surface end-to-end before porting it.
3. Extract the reusable baseline first:
   - shell behavior
   - typography hierarchy
   - repeated widgets
   - spacing and table density
4. Classify each reused legacy artifact as one of:
   - preserve visually
   - refactor structurally
   - reject
5. Map the target ownership explicitly:
   - `apps/web`
   - `packages/generated`
   - `packages/ui`
   - `packages/feature-*`
6. Reject any change that reintroduces manual DTOs, page-local transport parsing, direct `fetch` in pages, or ambiguous shared folders.
7. End with concrete findings or a concrete implementation move, not general advice.

## Guardrails

- preserve the legacy visual language where it is already strong
- do not start a big surface port before freezing the shell and shared widget baseline it depends on
- do not copy legacy contracts as the new source of truth
- do not let page files become API adapters
- do not let `apps/web` become the home of reusable widgets
- prefer `packages/ui` for cross-feature widgets and `packages/feature-*` for feature-local composition
- keep frontend thin and backend-owned for semantic read composition
- if a folder or package has ambiguous ownership, tighten it before adding more code

## Output expectations

When using this skill:

- state what from the legacy surface is being preserved visually
- state what shell and widget baseline was extracted before the port
- state what is being rejected structurally
- cite the target package ownership being enforced
- call out when a proposed reuse would create long-term drift
- recommend the next best move that keeps the frontend professional and scalable

## References

- For the frozen migration rules and folder ownership, read `references/frontend-migration-rules.md`.
