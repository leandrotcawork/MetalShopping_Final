# ADR-0014: SoT Documentation Drift Guard

- Status: accepted
- Date: 2026-03-18

## Context

MetalShopping now has faster iteration across frontend auth ownership, generated SDK runtime evolution, Keycloak login theming, and `server_core` composition refactors. These changes can ship in code before `docs/PROJECT_SOT.md` and `docs/PROGRESS.md` are updated.

When that happens, architecture governance weakens:

- onboarding context becomes stale
- implementation gates become ambiguous
- future decisions reopen already-closed boundaries

## Decision

- `docs/PROJECT_SOT.md` and `docs/PROGRESS.md` are mandatory sync documents for structural changes.
- Any structural change in these areas must include SoT/progress updates in the same tranche:
  - `package.json`
  - `apps/web/package.json`
  - `apps/web/tsconfig.json`
  - `apps/web/vite.config.ts`
  - `.github/workflows/`
  - `apps/server_core/cmd/metalshopping-server/`
  - `apps/web/src/app/`
  - `packages/feature-auth-session/package.json`
  - `packages/feature-auth-session/src/`
  - `packages/feature-products/package.json`
  - `packages/generated-types/package.json`
  - `packages/generated-types/src/`
  - `packages/platform-sdk/package.json`
  - `packages/platform-sdk/src/`
  - `ops/keycloak/themes/metalshopping/login/`
  - `scripts/generate_contract_artifacts.ps1`
- The repository now includes `scripts/check_sot_doc_drift.ps1` to fail when structural changes exist without updates to `PROJECT_SOT` and `PROGRESS` in both local and CI diff modes.

## Consequences

- architecture docs become operational artifacts, not optional follow-up notes
- major auth/session and frontend ownership shifts stay visible and reviewable
- reviewers gain a deterministic check for SoT drift in local workflow and CI

## Follow-up

- `check_sot_doc_drift.ps1` is now wired in pull request CI after contract checks
- keep structural path prefixes in the script aligned with future package boundaries
