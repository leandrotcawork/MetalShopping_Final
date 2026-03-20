# ADR-0036: Shopping Frontend Parity Baseline v1

- Status: accepted
- Date: 2026-03-20

## Context

The legacy Shopping surface has a strong operational workflow and a mature visual language. The target repo already preserves the 3-step wizard concept (Upload, Configurar, Executar), but the current implementation diverges in several UX-critical areas:

- duplicated headers (AppFrame hero + page header)
- upload UX shows backend-oriented fields instead of user-oriented pickers
- supplier selection is rendered as checkbox rows instead of card selection
- manual URL management is a small form instead of a dense operational panel

This ADR exists to freeze a single baseline rule: we preserve the legacy visual language while enforcing the thin-client architecture and package ownership rules.

Legacy reference:

- `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\Nova pasta\MetalShopping\frontend\apps\web\src\pages\shopping\ShoppingPage.tsx`
- `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\Nova pasta\MetalShopping\frontend\apps\web\src\pages\shopping\shopping.css`

Target reference:

- `apps/web/src/pages/ShoppingPage.tsx`
- `apps/web/src/pages/ShoppingPage.module.css`

## Decision

The Shopping frontend will converge to a parity baseline with the legacy surface, while keeping the new architecture boundaries frozen:

- Preserve the 3-step wizard and operational density.
- Remove duplicated headers (a single hero/header region per page).
- If `AppFrame` hero is present, the Shopping page must not render an additional top header.
- Thin-client boundary: no direct `fetch()` in pages.
- Thin-client boundary: consume generated SDK/runtime only (`@metalshopping/platform-sdk`, `@metalshopping/sdk-runtime`, `@metalshopping/sdk-types`).
- Thin-client boundary: no manual DTOs or transport parsing inside the page.
- Ownership: `apps/web` owns route composition and page orchestration.
- Ownership: `packages/ui` owns reusable primitives and shell-adjacent widgets.
- Ownership: Shopping-specific workflow widgets may move into a new `packages/feature-shopping` once there is clear complexity pressure.

## Scope

This ADR freezes the baseline rules and does not by itself introduce contract changes.

Follow-up ADRs will cover:

- Upload UX capability (desktop picker vs web upload)
- Supplier selection cards
- Manual URL operational panel
- Report generation and URL automation controls

## Implementation Checklist

- Run `metalshopping-frontend-migration-guardrails` classification for each legacy artifact touched.
- Refactor the Shopping page so header is single-source (no double hero).
- Keep existing SDK-driven calls intact; only UI composition changes are allowed in this tranche.
- Promote any widget reused across surfaces into `packages/ui`, otherwise keep it local.

## Acceptance Evidence (for Status: accepted)

- `apps/web/src/pages/ShoppingPage.tsx` renders a single header/hero region.
- Visual parity checklist for the tranche is recorded in `docs/SHOPPING_FRONTEND_PARITY_ACCEPTANCE.md`.
- `npm.cmd run web:typecheck` passes.
- `npm.cmd run web:build` passes.

## Consequences

- The Shopping surface becomes visually consistent with legacy without reintroducing legacy frontend debt.
- Future UI work has a stable baseline to build on without ad hoc design drift.
