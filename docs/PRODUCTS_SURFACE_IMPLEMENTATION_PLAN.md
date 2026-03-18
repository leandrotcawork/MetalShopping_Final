# Products Surface Implementation Plan

## Goal

Implement the first real MetalShopping frontend surface by rebuilding the legacy `Products` experience on top of the new canonical backend.

This implementation must follow `docs/FRONTEND_MIGRATION_CHARTER.md`.

## Why `Products` first

`Products` is the best first surface because it can already be grounded in the current backend foundation:

- `catalog` owns product identity
- `pricing` owns current price and cost semantics
- `inventory` owns current stock position

That makes `Products` the highest-value bridge between the backend we already built and a usable product UI.

## Legacy visual reuse policy

The target should preserve the legacy visual context where it was already strong:

- portfolio table layout
- filters and search flow
- operational emphasis on current product state
- route from product list into downstream actions

But the implementation must not preserve:

- page-local coupling to backend envelopes
- manual DTO duplication
- hard dependency on legacy sidecar contracts
- ambiguous frontend ownership such as reusable widgets buried inside app-local folders

## Frontend target structure

### `apps/web`

Owns:

- app shell
- route wiring
- providers
- page-level composition

### `packages/generated`

Owns:

- generated SDK client
- generated request and response types

### `packages/ui`

Owns:

- table primitives
- filter widgets
- badges
- layout primitives
- design tokens

### `packages/feature-products`

Owns:

- feature-specific query adapters
- view-model builders
- `Products` page widgets
- feature-local composition helpers

## Backend target for the first slice

The first `Products` surface should not force the frontend to manually stitch three separate domains.

The backend should expose a dedicated read surface that consolidates:

- product identity from `catalog`
- current price and cost from `pricing`
- current stock position from `inventory`

This should be a readmodel-style operational endpoint, not a new canonical write owner.

## First user-visible scope

The first `Products` surface should include:

- portfolio list
- search by SKU, description, `pn_interno`, `ean`, and reference
- filters by brand, taxonomy, and status
- current price column
- replacement and average cost visibility
- current stock quantity
- product lifecycle status

## Explicit non-goals

Do not include in the first `Products` slice:

- full product editing studio
- analytics-heavy portfolio diagnostics
- procurement actions
- shopping execution flow

## Frontend quality rules

- page component orchestrates only
- API calling stays in feature adapters or generated SDK
- DTOs come from generated artifacts
- view models are derived locally inside feature packages
- CSS Modules for page and feature styling
- reusable widgets move to `packages/ui`

## Sequence

1. scaffold `apps/web`
2. define `Products` read surface contract
3. scaffold `packages/ui`
4. scaffold `packages/feature-products`
5. implement first `Products` page with legacy visual context
6. validate route, auth/session, and data loading end-to-end
