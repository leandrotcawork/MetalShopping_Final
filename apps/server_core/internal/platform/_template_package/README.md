# Platform Package Template

Use this folder as the conceptual starting point for new packages under `internal/platform/`.

Read first:

- `docs/PLATFORM_BOUNDARIES.md`
- `docs/PLATFORM_PACKAGE_STANDARDS.md`
- `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`

Checklist:

- confirm the capability is runtime infrastructure, not business domain
- keep package purpose narrow and explicit
- expose stable interfaces only where reuse matters
- avoid semantically heavy business logic
- keep operational behavior observable and secure by default

