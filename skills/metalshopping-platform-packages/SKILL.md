---
name: metalshopping-platform-packages
description: Create or review MetalShopping `server_core` platform packages under `apps/server_core/internal/platform` using the repo platform package standards, boundary rules, and operational baseline. Use when defining runtime capabilities such as auth, tenancy, governance, messaging, observability, security, or audit infrastructure, or when checking whether code belongs in `platform/` instead of `modules/`.
---

# MetalShopping Platform Packages

## Overview

Use this skill to create or review runtime infrastructure packages inside `apps/server_core/internal/platform/` with the repository's frozen platform standards. Keep work anchored to boundary rules and the platform template instead of inventing generic infrastructure layers.

## Workflow

1. Read only the minimum repo context:
   `docs/PLATFORM_PACKAGE_STANDARDS.md`
   `docs/PLATFORM_BOUNDARIES.md`
   `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
   `apps/server_core/internal/platform/_template_package/README.md`
2. Confirm the target capability is runtime infrastructure, not a business module.
3. Confirm the package purpose and ownership boundary.
4. Start from `apps/server_core/internal/platform/_template_package/`.
5. Keep package purpose narrow, explicit, and reusable.
6. Finish with the review checklist in `references/platform-package-checklist.md`.

## Package rules

- platform packages provide runtime capabilities
- platform packages do not own business semantics
- package names should reflect capabilities, not implementation accidents
- observability and security concerns must remain first-class
- avoid super-packages with unrelated responsibilities

## References

- For the repo workflow and file touchpoints, read `references/repo-platform-flow.md`.
- For the final review pass, read `references/platform-package-checklist.md`.

