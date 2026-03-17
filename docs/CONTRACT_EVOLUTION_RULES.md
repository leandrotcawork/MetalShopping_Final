# Contract Evolution Rules

## Purpose

Define how APIs, events, governance artifacts, and generated outputs evolve without destabilizing the platform.

## Source-of-truth rules

- `contracts/` is the only contract source of truth
- app code implements contracts but does not redefine them
- generated artifacts are downstream outputs only

## API evolution

- additive change is preferred
- breaking route or payload semantics require explicit version strategy
- OpenAPI changes should be reviewed before implementation spreads to apps

## Event evolution

- published meaning must remain stable
- additive optional payload growth is preferred
- semantic breaks require explicit new versioning
- event naming should remain business-semantic, not implementation-specific

## Governance evolution

- policy, threshold, and feature flag semantics must remain explainable
- resolution hierarchy changes are platform-level decisions
- core and workers must preserve shared interpretation

## Generated artifact evolution

- generated output changes must follow contract changes
- manual edits to generated outputs are not authoritative
- migration overlap may exist temporarily, but only with explicit intent

