# Repo Platform Flow

## Read first

1. `docs/PLATFORM_PACKAGE_STANDARDS.md`
2. `docs/PLATFORM_BOUNDARIES.md`
3. `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
4. `apps/server_core/internal/platform/_template_package/README.md`

## Files this skill normally touches

- `apps/server_core/internal/platform/<package_name>/`
- supporting docs when a new platform capability changes the model

## Repo-specific rules

- runtime capabilities belong in `platform/`
- business semantics belong in `modules/`
- `shared/` remains small and neutral
- platform packages must stay observable and secure by default

