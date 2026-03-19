# ADR-0016: Generated SDK Versus Authored Runtime Boundary

- Status: accepted
- Date: 2026-03-18

## Context

The repository already moved from handwritten SDK emission to OpenAPI Generator orchestration. However, runtime hardening reviews still found boundary ambiguity between generated outputs and authored frontend runtime/facade behavior, especially in auth/session flows.

## Decision

- `contracts/*` remains the only source of truth.
- `packages/generated/*` remains output-only.
- authored frontend transport/runtime logic remains isolated from generated output ownership.
- feature packages consume stable package boundaries and do not import deep generated internals.
- auth/session runtime path must not rely on `as unknown as` style double-casts.

## Consequences

- Generated artifacts can be regenerated safely without risking authored runtime drift.
- Frontend auth/session behavior becomes easier to test and reason about.
- Future package renames are allowed, but the generated-versus-authored boundary semantics are frozen and must not regress.

## Follow-up

- Enforce boundary through runtime refactoring and lint/check rules in the login hardening tranche.
- Keep this ADR, `docs/SDK_BOUNDARY.md`, and `docs/SDK_GENERATION_STRATEGY.md` aligned.
