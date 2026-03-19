---
name: metalshopping-page-delivery
description: Implement or migrate a MetalShopping React page using the frozen thin-client delivery rules, legacy visual preservation, and SDK-bound data fetching. Use when implementing the Home page, Shopping Price page, Analytics page, CRM page, or any new route that binds to a Go backend via generated SDK.
---

# MetalShopping Page Delivery

## Overview

Use this skill to implement a React page that binds to a Go backend.
Preserve the legacy visual language. Use the generated SDK. Keep the
page thin. Do not invent new patterns.

## Workflow

1. Read the minimum frozen context:
   `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`
   `docs/FRONTEND_MIGRATION_CHARTER.md`
   `docs/FRONTEND_MIGRATION_PLAYBOOK.md`

2. Confirm the SDK has been regenerated after the contract was updated.
   Do not start the page before generation is complete.

3. Identify the legacy page for this surface if it exists.
   Extract from it:
   - visual layout and composition
   - typography and spacing patterns
   - repeated widgets (if 3+ occurrences across pages, extract to `packages/ui`)

4. Implement the page in `apps/web/src/pages/<module>/`:
   - import data via `@metalshopping/platform-sdk`
   - no direct `fetch()` in the page file
   - no manual DTO types (use generated types only)
   - loading state: simple spinner or skeleton is enough at Level 1
   - error state: plain error message is enough at Level 1

5. Widget extraction rule:
   - component used in 3+ places across pages → extract to `packages/ui`
   - component used in 1–2 places only → leave it local, do not extract

6. Confirm the page renders real data in the browser.

7. Run `pnpm tsc --noEmit` and confirm it passes.

8. Finish with the checklist in `references/page-checklist.md`.

## What not to add at Level 1

- no animations or transitions
- no complex empty states
- no mobile-specific layouts
- no i18n
- no localStorage usage for API data
- no direct Keycloak or auth logic in the page
  (use the session context already provided by the shell)

## References

- For the repo page delivery flow, read `references/page-flow.md`.
- For the final review pass, read `references/page-checklist.md`.
