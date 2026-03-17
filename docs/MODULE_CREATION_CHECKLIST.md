# Module Creation Checklist

## Purpose

Provide a repeatable checklist for creating new `server_core` modules from the frozen standards.

## Before creating a module

- confirm the capability is a true business module, not a platform concern
- confirm the bounded context name is stable
- confirm canonical ownership of write semantics
- confirm whether contracts already exist or must be created first

## Module scaffold

- create the module under `apps/server_core/internal/modules/<module_name>/`
- include:
  - `domain/`
  - `application/`
  - `ports/`
  - `adapters/`
  - `transport/`
  - `events/`
  - `readmodel/`
- start from the `_template/` structure when possible

## Design checks

- domain owns business meaning
- application coordinates but does not replace domain
- ports define dependencies explicitly
- adapters implement ports only
- transport stays interface-only
- events remain aligned with `contracts/events/`
- read models do not become a second canonical truth

## Boundary checks

- no business logic moved into `platform/`
- no semantically heavy code moved into `shared/`
- no hidden cross-module coupling

## Completion checks

- module fits `docs/MODULE_STANDARDS.md`
- module fits `docs/PLATFORM_BOUNDARIES.md`
- module fits `docs/READMODEL_AND_EVENTS_RULES.md`
- any new contract need is reflected in `contracts/`

