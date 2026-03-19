# Login MVP Execution Plan

## Objective

Close login with a deterministic, professional baseline and stop reopening the same gaps.

## Tranche T1: boundary and scope freeze

### Deliverables

- `docs/LOGIN_MVP_SCOPE.md` accepted and frozen
- `docs/LOGIN_DOD.md` accepted and frozen
- `docs/SDK_BOUNDARY.md` accepted and frozen
- ADRs accepted for login closure governance and SDK boundary semantics

### Exit criteria

- no implementation branch starts login hardening work without these frozen references

## Tranche T2: SDK/runtime hardening

### Deliverables

- isolate generated output and authored runtime responsibilities
- remove deep relative imports from authored runtime into generated internals
- remove `as unknown as` from auth/session runtime path
- keep frontend thin and package-driven (`apps/web` only composes providers/routes)

### Exit criteria

- `generate_contract_artifacts -Check`, `go test`, `web:typecheck`, and `web:build` are green
- no prohibited runtime casts/imports in auth/session path

## Tranche T3: login closure validation

### Deliverables

- end-to-end local smoke with Keycloak-backed login/session/logout
- DoD checklist fully checked and evidenced
- SoT/progress synced in the same tranche

### Exit criteria

- all `docs/LOGIN_DOD.md` items checked
- no critical login gap remains open

## Non-negotiable sequencing

1. freeze and ADR first
2. boundary/runtime hardening second
3. closure validation third

Do not open new authenticated surfaces before T3 exit criteria are satisfied.
