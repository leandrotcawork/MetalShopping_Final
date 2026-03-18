# Frontend Migration Playbook

## Purpose

Freeze the professional migration sequence for MetalShopping frontend surfaces so that legacy visual quality is preserved while the new frontend architecture stays thin, scalable, and package-owned.

This playbook exists because the first `Products` migration proved an important point:

- copying visuals without first extracting the visual system creates churn
- migrating pages before freezing widget, shell, and typography rules creates rework
- preserving the legacy appearance does not mean preserving the legacy implementation

## Core rule

Every frontend migration must happen in two passes:

1. study and extract the legacy surface
2. rebuild it on the target architecture

Do not jump directly from a legacy page to a new page implementation.

## Mandatory pre-migration study

Before migrating any legacy surface, inspect the relevant legacy frontend files and classify what exists.

That study must cover:

- app shell behavior
- typography hierarchy
- layout spacing and alignment
- repeated buttons, badges, filters, cards, and tables
- page-specific widgets
- route-level API communication
- ad hoc DTOs or parsing logic

## Mandatory classification

Each important legacy artifact must be classified as one of:

- preserve visually
- refactor structurally
- reject

Examples:

- preserve visually
  - page shell composition
  - widget hierarchy
  - typography ratios
  - table density
  - color language
- refactor structurally
  - repeated cards
  - repeated dropdowns
  - shared table headers
  - shell pieces that belong in reusable packages
- reject
  - direct `fetch` inside pages
  - manual DTOs in the frontend
  - inline payload normalization in route components
  - generic `shared` dumping grounds

## Required extraction sequence

When a surface is large enough, the migration should follow this order:

1. freeze shell behavior
2. freeze typography and spacing baseline
3. extract reusable widgets from the legacy page
4. define the backend-owned read surface
5. wire generated contract consumption
6. rebuild the page with feature-local adapters and view models

## Target ownership

### `apps/web`

Owns:

- route registration
- providers
- app shell composition
- page-level composition

Does not own:

- transport DTO truth
- reusable cross-feature widgets
- domain payload normalization

### `packages/generated`

Owns:

- generated contract types
- generated SDKs
- transport-facing canonical client artifacts

### `packages/ui`

Owns:

- reusable shell pieces
- table primitives
- filter primitives
- cards
- banners
- status pills
- buttons and shell-adjacent primitives

### `packages/feature-*`

Owns:

- feature adapters
- feature-local view models
- feature composition
- feature widgets that are not cross-feature yet

## Big-tech frontend rules

The target frontend must behave like a professional product frontend, not a page-by-page port.

That means:

- backend owns semantic read composition
- frontend consumes explicit read surfaces
- pages orchestrate and render
- feature adapters handle transport details
- reusable UI lives in explicit packages
- layout systems are frozen before surface proliferation
- a migrated page should become easier to extend than the legacy page, not just prettier

## UI fill strategy

The target should not fill UI by blindly copying static legacy markup.

The correct strategy is:

- preserve the visual language
- rebuild the information flow from generated contracts and backend read surfaces
- introduce placeholders only when they fit the frozen widget language
- avoid one-off CSS or layout exceptions when the repeated pattern should become a reusable primitive

## Review checklist

Before calling a migrated surface ready, confirm:

- the shell matches the approved legacy interaction and alignment baseline
- the typography and spacing feel like MetalShopping
- repeated widgets have been promoted into `packages/ui` where appropriate
- the page does not contain transport parsing
- the page is not inventing DTOs
- generated types are being consumed
- the backend surface is doing the semantic composition the UI needs
- the result is easier to scale than the legacy implementation
