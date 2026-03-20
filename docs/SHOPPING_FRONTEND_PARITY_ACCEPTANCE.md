# Shopping Frontend Parity Acceptance (ADR-0036..ADR-0039)

- Date: 2026-03-20
- Scope: Shopping frontend parity baseline + upload UX + supplier cards + manual URL panel.

## Evidence

- Build:
  - `npm.cmd run web:typecheck` -> pass
  - `npm.cmd run web:build` -> pass

## Notes

- This acceptance is UI/UX parity oriented and does not change API contracts.
- Runtime behavior depends on existing Shopping bootstrap + supplier signals endpoints.

