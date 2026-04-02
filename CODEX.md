# CODEX.md

This file provides guidance to Codex when working in this repository.

## Mandatory startup

Before any planning, implementation, or review:

1. Read `docs/PROJECT_SOT.md`
2. Read `ARCHITECTURE.md`
3. Read `AGENTS.md`
4. Read `tasks/todo.md`
5. Read `tasks/lessons.md`

If documents conflict, follow `docs/PROJECT_SOT.md` first, then `ARCHITECTURE.md`, then `AGENTS.md`.

## Scope

Use this file for Codex-specific operating guidance only. Do not treat it as a competing source of project truth.

## Commands

### Web
```powershell
npm run web:typecheck
npm run web:build
npm run web:test
npm --workspace @metalshopping/web run dev
```

### Backend
```powershell
go test ./apps/server_core/...
go test ./apps/server_core/internal/modules/<module>/...
```

### Contracts
```powershell
./scripts/generate_contract_artifacts.ps1 -Target all
./scripts/validate_contracts.ps1 -Scope all
```

## Working rules

- Prefer updating repository SoTs instead of creating duplicate planning files.
- Never edit `packages/generated/` or `packages/generated-types/` manually.
- Keep architecture changes in ADRs and SoT documents, not only in agent notes.
- Use the skill map and workflow defined by `AGENTS.md`.

## Key references

- `docs/PROJECT_SOT.md`
- `ARCHITECTURE.md`
- `AGENTS.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- `docs/adrs/`
