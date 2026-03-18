# Frontend Migration Charter

## Purpose

Freeze how MetalShopping should migrate frontend surfaces from the legacy application without losing the visual quality that already worked and without reintroducing the weak architectural patterns that made the old stack hard to scale.

This charter exists to make one rule explicit:

- preserve the legacy visual language
- do not preserve the legacy architectural shortcuts

## Frozen decision

The legacy frontend is the visual reference for the target product.

That means we intentionally preserve:

- layout feel
- widget language
- color direction
- CSS identity
- successful workflow composition

But we do not preserve:

- manual frontend contracts and DTO duplication
- page-local backend parsing
- direct `fetch` usage inside pages
- generic dumping grounds such as catch-all `shared` buckets
- app-local UI component sprawl when a reusable package boundary should exist

## What the legacy frontend got right

The legacy stack already proved useful direction in several areas:

- a dedicated `apps/web` app boundary
- a package-oriented split under `packages/`
- strong operational page intent in `Products`, `Shopping`, and `Home`
- a visual language the user already trusts

These parts should be reused selectively.

## What must change in the target

The target frontend must be more professional than the legacy stack in the following ways:

- backend contracts are generated, not hand-maintained in frontend packages
- pages orchestrate, but do not normalize domain payloads
- feature packages own adapters and view-model mapping
- reusable widgets live in `packages/ui`
- global CSS stays minimal; feature and page styling stays local through CSS Modules
- the web app stays thin and depends on `server_core` read surfaces instead of hidden frontend composition

## Folder and module validation

## Legacy structure verdict

The legacy folder structure is directionally good, but only after refinement.

### Good legacy patterns to preserve

- `apps/web`
- `packages/feature-*`
- `packages/ui-kit` as the intuition that a shared UI package should exist

### Legacy patterns to refine or reject

- `packages/api-client`
  - replace with generated clients and generated types under `packages/generated`
- `packages/shared`
  - do not recreate a generic shared bucket
- `apps/web/src/components`
  - do not let app-local components become the long-term home of reusable widgets
- app-local manual contracts and parser helpers
  - move transport ownership to generated packages and feature adapters

## Target frontend structure

### `apps/web`

Owns:

- app shell
- route registration
- providers
- page composition
- app-level global CSS and bootstrap wiring

Does not own:

- canonical DTOs
- direct backend normalization logic
- reusable cross-feature widgets

### `packages/generated`

Owns:

- generated SDKs
- generated request and response types
- transport-facing canonical contract artifacts

### `packages/ui`

Owns:

- reusable widgets
- visual primitives
- table and filter building blocks
- badges, pills, cards, shells, and tokens

### `packages/feature-*`

Owns:

- feature-local API adapters
- feature view models
- feature composition helpers
- feature widgets that are not yet cross-feature reusable

## React rules

- page files orchestrate only
- business rules do not migrate into page state
- normalization logic belongs in feature adapters or view-model builders
- repeated visual patterns should be promoted into `packages/ui`
- feature state is allowed; hidden domain semantics in UI state are not

## API and DTO rules

- pages do not call `fetch` directly
- frontend does not define parallel DTOs when generated types already exist
- feature adapters may derive UI view models from generated contracts
- transport envelopes stay in generated SDK or feature adapter layers
- frontend must not become a second source of contract truth

## Styling rules

- preserve the legacy visual identity where it is already strong
- use CSS Modules for feature-local and page-local styles
- keep `global.css` for reset, shell, typography baseline, and tokens only
- promote repeated visual patterns into `packages/ui`
- do not fork colors, spacing, or widget variants casually between pages

## Migration execution rule

Each legacy frontend surface must be migrated through this sequence:

1. inspect the legacy page or widget
2. classify each part as:
   - preserve visually
   - refactor structurally
   - reject
3. map the target ownership:
   - `apps/web`
   - `packages/generated`
   - `packages/ui`
   - `packages/feature-*`
4. implement the target using generated contracts and thin-client rules
5. verify that the result still feels like MetalShopping visually

## Explicit prohibitions

Do not:

- copy legacy frontend contracts into the new repo as a second source of truth
- recreate a generic `shared` package without explicit ownership
- keep parser-heavy `tsx` pages that normalize raw payloads inline
- keep API request assembly scattered across route components
- rewrite the product into a generic SaaS visual style that loses the MetalShopping identity

## Review checklist

Before accepting frontend migration work, confirm:

- the visual direction still feels like the approved legacy MetalShopping
- package ownership is explicit
- reusable widgets moved out of page-local code where appropriate
- DTOs come from generated artifacts
- no page is acting as an API adapter
- no frontend folder became a convenience dump

## Relationship to other docs

This charter complements, and does not replace:

- `docs/OPERATIONAL_SURFACES_PLAN.md`
- `docs/PRODUCTS_SURFACE_IMPLEMENTATION_PLAN.md`
- `docs/PRODUCTS_READMODEL_OWNERSHIP.md`
- `docs/FRONTEND_QUALITY_GATES.md`
- `docs/adrs/ADR-0005-thin-clients-and-generated-sdks.md`
