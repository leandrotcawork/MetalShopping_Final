# AGENTS

## Scope

This file applies to `apps/server_core/`.

## Current mode

- planning-first
- structure and boundaries are being frozen
- do not add product implementation unless explicitly requested

## Core rules

- keep `server_core` as a modular monolith
- keep domain in `internal/modules/`
- keep infrastructure in `internal/platform/`
- keep `internal/shared/` small and neutral
- do not recreate legacy bootstrap code unless the plan explicitly asks for it

## Module standard

Each module under `internal/modules/*` must respect:

- `domain/`
- `application/`
- `ports/`
- `adapters/`
- `transport/`
- `events/`
- `readmodel/`

## What to read first

1. `../../docs/PROJECT_SOT.md`
2. `../../ARCHITECTURE.md`
3. relevant ADRs
4. `README.md`

