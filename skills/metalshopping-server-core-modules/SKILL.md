---
name: metalshopping-server-core-modules
description: Create or review MetalShopping `server_core` business modules under `apps/server_core/internal/modules` using the repo module standards, boundary rules, readmodel and event rules, and the module template. Use when instantiating a new bounded-context module, reviewing module structure, or checking whether code belongs in `modules/`, `platform/`, or `shared/`.
---

# MetalShopping Server Core Modules

## Overview

Use this skill to create or review business modules inside `apps/server_core/internal/modules/` with the repository's frozen modular monolith standards. Keep work anchored to the module template and boundary rules instead of improvising module shape.

## Workflow

1. Read only the minimum repo context:
   `docs/MODULE_STANDARDS.md`
   `docs/PLATFORM_BOUNDARIES.md`
   `docs/READMODEL_AND_EVENTS_RULES.md`
   `docs/MODULE_CREATION_CHECKLIST.md`
   `apps/server_core/internal/modules/_template/README.md`
2. Confirm the target capability is truly a business module, not a platform concern.
3. Confirm the module name and ownership boundary.
4. Start from `apps/server_core/internal/modules/_template/`.
5. Keep `domain`, `application`, `ports`, `adapters`, `transport`, `events`, and `readmodel` responsibilities aligned with the standards.
6. Finish with the review checklist in `references/module-checklist.md`.

## Module rules

- modules own business meaning
- modules do not become generic dumping grounds
- `domain/` contains business semantics only
- `application/` orchestrates without replacing domain ownership
- `ports/` define explicit dependencies
- `adapters/` implement ports, not business truth
- `transport/` remains interface-only
- `events/` stays aligned with `contracts/events/`
- `readmodel/` does not become a second canonical model

## Output expectations

When creating or reviewing a module:

- preserve the canonical module structure
- keep platform and shared boundaries clean
- make cross-module dependency choices explicit
- note any required contract work in `contracts/`
- note any reason the capability should live in `platform/` instead of `modules/`

## References

- For the exact repo workflow and file touchpoints, read `references/repo-module-flow.md`.
- For the final review pass, read `references/module-checklist.md`.

