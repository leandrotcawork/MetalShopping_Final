# Frontend Quality Gates

## Purpose

Define the minimum quality bar for `apps/web` and feature packages before the first operational MetalShopping UI surfaces expand.

## Gates

### 1. Contract gate

- `contracts/` must validate before frontend changes are considered ready
- generated SDK and generated types must be refreshed from repo scripts
- frontend code must consume generated artifacts instead of parallel manual DTOs

### 2. Type gate

- frontend packages must pass TypeScript typecheck
- public package exports must be explicit
- `any`-style drift must not be used as a shortcut for feature delivery

### 3. Build gate

- `apps/web` must build reproducibly from scripts
- feature packages must resolve through the app without local path hacks

### 4. Widget gate

- reusable widgets in `packages/ui` must have at least minimal unit coverage or story-level verification later
- feature view-model builders must be testable without page runtime

### 5. Boundary gate

- page files orchestrate only
- direct `fetch` usage in pages is not allowed
- generated artifacts remain downstream only
- business rules do not migrate into frontend state containers

### 6. Styling gate

- global CSS is limited to app shell and resets
- feature and page styling uses CSS Modules
- repeated UI patterns graduate to `packages/ui`
- shell, typography, and repeated widget baselines must be frozen before large surface proliferation

### 7. Legacy study gate

- each migrated surface must cite the legacy files inspected
- each migration must classify legacy artifacts as preserve visually, refactor structurally, or reject
- if a migration reveals repeated widget debt, the widget must be extracted or explicitly queued before the next surface starts

## Initial enforcement workflow

The first frontend slice should be considered ready only when all of the following pass:

1. `powershell -ExecutionPolicy Bypass -File .\\scripts\\validate_contracts.ps1 -Scope all`
2. `powershell -ExecutionPolicy Bypass -File .\\scripts\\generate_contract_artifacts.ps1 -Target all -Check`
3. frontend typecheck script
4. frontend build script
5. legacy study and ownership mapping documented for the surface

## Future enforcement

These gates should move into team workflow and CI as soon as `apps/web` has a runnable baseline.
