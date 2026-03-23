---
name: ms
description: MetalShopping master orchestrator. Use for implementation tasks. Validates architecture first, uses update_plan for complex work, keeps todo/lessons high-signal, and orchestrates specialist skills in the correct order.
---

# MetalShopping — Master Orchestrator

Read first: `tasks/lessons.md`, `tasks/todo.md`, `AGENTS.md`.

## 1) Complexity gate

Treat as **complex** when any is true:
- touches 2+ layers (`contract`, `Go`, `worker`, `SDK`, `frontend`, `DB`, `ADR`)
- adds a page/flow or 3+ requirements
- changes migration, runtime config, manifest, queue, worker behavior
- user asks explicit planning/roadmap

For complex tasks:
1. call `update_plan` before coding
2. keep one step `in_progress` at a time
3. refresh `tasks/todo.md` block before execution

For simple tasks:
- do a short architecture check and execute directly

## 2) Architecture check before code

State only what matters:
- module type: `read-only` | `write+events` | `CRUD+governance` | `scraping` | `frontend-only` | `docs/process`
- exact files/folders to change
- risks/invariants (tenant safety, runtime drift, UX regressions)
- what is now vs later when staged delivery exists

## 3) Skill orchestration order

Use only applicable stages:
- `T1 contract` → `$metalshopping-openapi-contracts`
- `T2/T3 implementation` → `$metalshopping-implement`
- `T4 SDK` → `$metalshopping-sdk-generation` (after contract change)
- legacy visual-first migration → `$metalshopping-legacy-migration`
- `T5 frontend` → `$metalshopping-frontend`
- `T6 ADR` → `$metalshopping-adr`
- significant feature/review gate → `$metalshopping-review`

If a stage is skipped, state why.

## 4) Runtime-state gate (before blaming code)

When task touches worker/config/migration/manifest/DB:
- confirm migration/seed applied
- confirm active manifest/config version in DB
- confirm required API/worker restart occurred
- confirm run/test data is fresh
- confirm tenant context and `current_tenant_id()` expectations

## 5) Execution and validation rules

- one feature block at a time
- mark task done only after required validation passes
- run smallest relevant check first, then broader checks
- for web/frontend closeout, always run `npm.cmd run web:build` unless user says otherwise
- do not edit `packages/generated/` manually
- commit after each completed task

## 6) `todo` and `lessons` hygiene

`tasks/todo.md`:
- keep only active backlog and current acceptance checks
- avoid historical verbosity; rely on git history for old details
- scope edits to target block only (no global replace)

`tasks/lessons.md`:
- record only structural/global lessons
- never record one-off cosmetic page tweaks as lessons
- each lesson must be reusable across modules or process

## 7) Closeout checklist

Before declaring done:
- validations executed and reported
- browser/runtime behavior verified for user-facing changes
- `tasks/todo.md` updated
- clean git state or explicit commit delivered
