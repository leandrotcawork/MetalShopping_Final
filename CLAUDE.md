# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Session start (mandatory)

1. Read `tasks/lessons.md` — apply every rule before touching code
2. Read `tasks/todo.md` — know the current state

## Commands

### Web (npm workspaces, run from repo root)

```bash
# Typecheck
npm run web:typecheck

# Build
npm run web:build
# WSL2 workaround (esbuild EPERM on Windows)
npm --workspace @metalshopping/web run web:build:wsl

# Tests (Vitest)
npm run web:test
# Run tests for a specific file
npm --workspace @metalshopping/web exec -- vitest run src/path/to/file.test.ts -c vitest.config.mjs
# Verbose (CI style)
npm --workspace @metalshopping/web run test:ci

# Dev server (localhost:5173)
npm --workspace @metalshopping/web run dev
```

### Backend (Go, from repo root)

```bash
# Run all tests
go test ./apps/server_core/...

# Run tests in a single package
go test ./apps/server_core/internal/modules/<module>/...
```

### Contract generation & validation (PowerShell scripts)

```powershell
# Generate SDK/types from OpenAPI + JSON Schema contracts (requires Docker)
./scripts/generate_contract_artifacts.ps1 -Target all

# Validate all contracts
./scripts/validate_contracts.ps1 -Scope all
```

## Architecture overview

MetalShopping is a server-first B2B platform (commercial strategy, pricing, analytics, procurement, CRM). It is a monorepo with a Go backend, a React thin-client, Python workers, and shared contracts.

### Source of truth hierarchy

```
contracts/          ← hand-authored (OpenAPI specs + JSON Schema + event schemas)
    ↓ generate
packages/generated/ ← NEVER edit manually — auto-generated SDK + types
packages/generated-types/
```

The CI guards reject manual edits to `packages/generated/` unless the PR is labeled `codegen`.

### Monorepo layout

| Path | Purpose |
|------|---------|
| `apps/server_core/` | Go modular monolith (canonical state authority) |
| `apps/web/` | React 18 thin-client (Vite + TypeScript) |
| `apps/analytics_worker/` | Python compute/scoring (skeleton) |
| `contracts/api/openapi/` | OpenAPI 3.0 specs (one per domain) |
| `contracts/api/jsonschema/` | JSON Schema payloads |
| `contracts/events/` | Versioned event schemas |
| `contracts/governance/` | Feature flags, policies, thresholds schemas |
| `packages/ui/` | Shared React UI primitives (CSS modules) |
| `packages/platform-sdk/` | Generated SDK runtime (`@metalshopping/sdk-runtime`) |
| `packages/feature-analytics/` | Analytics surface package |
| `packages/feature-products/` | Products surface package |
| `packages/feature-auth-session/` | Auth/session package |
| `docs/` | Architecture blueprint, ADRs, implementation plan |
| `tasks/` | Active sprint backlog (`todo.md`) and lessons (`lessons.md`) |
| `bootstrap/seeds/` | Governance + tenant seed data |
| `ops/` | Docker, Kubernetes, Keycloak, observability |

### Go backend (server_core)

Multi-tenant shared-database architecture. Every module lives under `internal/modules/<name>/` and follows this layer structure:

```
domain/       business entities & logic
application/  use-case / command handlers
ports/        interfaces (in + out)
adapters/     Postgres persistence
transport/    HTTP handlers & serialization
events/       event definitions & emission
readmodel/    denormalized query views
```

Platform infrastructure is in `internal/platform/`:
- `db/postgres/` — tenant-aware Postgres helpers
- `auth/` — JWT/OIDC, principal context
- `tenancy_runtime/` — tenant context middleware
- `governance/` — feature flags / policies / thresholds resolution
- `outbox/` — transactional event publishing

Every new module must be registered in `composition_modules.go`.

### Frontend (apps/web)

Thin-client pattern — all data flows through the generated SDK. Feature logic lives in the `packages/feature-*` packages; `apps/web` is the shell that assembles them.

## Absolute rules

### Go
- `pgdb.BeginTenantTx` on **every** Postgres adapter query — no exceptions
- `current_tenant_id()` in every WHERE clause on tenant-scoped tables
- `platformauth.PrincipalFromContext` → 401 before any handler operation
- `tenancy_runtime.TenantFromContext` → 403 before any handler operation
- `outbox.AppendInTx` inside the same transaction as INSERT — never after Commit
- Every new module registered in `composition_modules.go`

### Python worker
- `set_config('app.current_tenant_id', %s, true)` before every write transaction
- `ON CONFLICT ... DO UPDATE` on every insert (idempotency)
- Never call `server_core` HTTP endpoints (one-way dependency)

### Frontend
- Data only via `sdk.*` methods from `@metalshopping/sdk-runtime` — no raw `fetch()`
- Design tokens only — no hardcoded hex values (`$metalshopping-design-system`)
- Check `packages/ui/src/index.ts` before creating any component
- Every data-fetching component must have loading + error + empty states
- Fetch pattern: `useEffect + cancelled flag`

### Process
- A task is done only when: build passes + real data verified + commit made
- `packages/generated/` and `packages/generated-types/` are never edited manually
- One commit per completed task — no uncommitted work at session end
- ADR committed only after the acceptance test passes

## Commit format

```
<type>(<scope>): <what>
```

Types: `feat | fix | docs | chore | refactor`

## Engineering bar

Every decision passes this filter: *"Would a Stripe or Google senior engineer approve this in code review?"*

- Names are self-documenting
- Errors carry structured codes: `MODULE_ENTITY_REASON`
- Every handler logs `trace_id`, `action`, `result`, `duration_ms`
- Every write is idempotent and retry-safe
- No query ever returns cross-tenant data

## Skills (agent task routing)

| Task | Skill |
|------|-------|
| Any implementation (default) | `$ms` |
| OpenAPI contract | `$metalshopping-openapi-contracts` |
| Event contract | `$metalshopping-event-contracts` |
| Governance contract | `$metalshopping-governance-contracts` |
| SDK generation | `$metalshopping-sdk-generation` |
| ADR lifecycle | `$metalshopping-adr` |
| Frontend visual / component | `$metalshopping-design-system` |

## Brain Workflow File Locations

When using `/brain-task`, all task artifacts follow this structure:

### During Task Execution (brain-task Steps 1-5)

| Step | File Location | Purpose |
|------|---------------|---------|
| 1 | `.brain/working-memory/context-packet-[id].md` | Assembled sinapses (context) |
| 2 | `.brain/working-memory/codex-context.md` OR `.brain/working-memory/opus-context.md` | Execution context (sent to model) |
| 4 | `.brain/working-memory/task-completion-[id].md` | Outcome + files + tests + lessons |
| 5 | `.brain/working-memory/sinapse-updates-[id].md` | Proposed sinapse updates (awaiting approval) |

⚠️ **Rule:** ALL task artifacts during execution go to `.brain/working-memory/` — NOT to `tasks/`

### After Task Completion (brain-task Step 6)

| Location | Purpose |
|----------|---------|
| `.brain/progress/completed-contexts/[task-id]-codex-context.md` | Archived execution context |
| `.brain/progress/completed-contexts/[task-id]-completion-record.md` | Archived completion record |
| `.brain/progress/completed-contexts/[task-id]-OUTCOME.md` | Outcome analysis for pattern matching |
| `.brain/progress/activity.md` | Activity log (all tasks appended) |

### Sprint Backlog (manual, not part of brain-task)

| Location | Purpose |
|----------|---------|
| `tasks/todo.md` | Current sprint backlog + action items |
| `tasks/lessons.md` | Permanent lessons learned |

---

## Key reference docs

- `docs/ARCHITECTURE.md` — frozen architecture blueprint
- `docs/PROJECT_SOT.md` — current-phase source of truth
- `docs/IMPLEMENTATION_PLAN.md` — phased roadmap
- `docs/PROGRESS.md` — completion status per feature
- `docs/adrs/` — architecture decision records
