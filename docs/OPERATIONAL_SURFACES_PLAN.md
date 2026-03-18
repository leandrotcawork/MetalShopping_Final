# Operational Surfaces Plan

## Purpose

Define how MetalShopping should recover its first real product surfaces from the legacy application without reintroducing legacy coupling or weakening the new platform architecture.

This plan should now be read together with `docs/FRONTEND_MIGRATION_CHARTER.md`, which freezes the rule that legacy visual language is preserved while legacy frontend shortcuts are rejected.

The immediate target surfaces are:

- `Products`
- `Shopping`
- `Home`

## Recovery principle

We should preserve the legacy visual context and successful workflow patterns, but not copy:

- sidecar-specific contracts as the new platform truth
- manual parallel DTO systems
- page-local business logic
- direct coupling from frontend to legacy runtime assumptions

The target frontend must stay:

- thin
- contract-driven
- reusable by design
- aligned with generated SDKs and generated types
- aligned with the frontend migration charter for visual reuse and package ownership

## Execution order

### 1. `Products` first

`Products` is the best first surface because the new backend already owns the core product semantics:

- `catalog` for identity
- `pricing` for current price and cost semantics
- `inventory` for stock position

What this first surface should provide:

- product portfolio list
- search and filtering
- current price visibility
- current cost visibility
- current stock visibility
- route to product-level actions later

### 2. `Shopping` second

`Shopping` must be rebuilt as an operational workflow, not just a page.

The legacy flow is strong and should be preserved conceptually:

- bootstrap
- input preparation
- run submission
- progress
- result
- history

But the new implementation must sit on explicit runtime and integration contracts instead of old sidecar-first assumptions.

### 3. `Home` third

`Home` should come after `Products` and the first operational `Shopping` foundation because it depends on aggregated read models such as:

- recent runs
- supplier health
- operational counters
- portfolio and run summaries

It should not be implemented first as a static dashboard shell without trusted data.

## Frontend architecture rules

### App boundary

`apps/web` should contain:

- app shell
- routing
- providers
- feature composition
- page-level layout composition

It should not become the place where DTO definitions, ad hoc endpoint logic, or business rules accumulate.

### Package split

The legacy frontend package split is directionally good and should be reused selectively:

- `packages/generated`
  - generated SDK and generated types only
- `packages/ui`
  - reusable visual primitives and tokens
- `packages/feature-products`
  - feature-local view models, adapters, UI composition
- `packages/feature-shopping`
  - workflow orchestration and shopping-specific UI composition
- future `packages/feature-home`
  - home readmodel composition only

### DTO and type rules

- canonical API DTOs come from generated artifacts
- feature packages may define view models derived from canonical DTOs
- frontend must not invent parallel contract types when a generated type already exists
- transport envelopes stay in generated/client layers, not page components

### API rules

- pages do not call `fetch` directly
- pages use feature APIs or generated SDK clients
- query normalization and request assembly live in feature/api or generated client layers
- session/bootstrap concerns stay in app providers

### UI and CSS rules

- preserve the legacy visual context where it is already strong
- move reusable primitives to `packages/ui`
- use CSS Modules for feature/page-local styling
- keep only app-level reset and shell CSS global
- avoid one-off page widgets when the pattern clearly repeats

### React structure rules

- page files orchestrate, not own business rules
- data normalization belongs in feature adapters/view-model builders
- widgets should be composable and testable
- local state is acceptable for interaction state, not for hidden domain semantics

## Selective legacy reuse

### Reuse

- visual composition patterns from legacy `HomePage`, `ProductsPage`, and `ShoppingPage`
- app shell/provider split from the legacy web app
- package decomposition concept from the legacy frontend
- good reusable widget patterns already proven in analytics and shopping flows

### Do not reuse blindly

- legacy sidecar endpoint shapes as the new contract source of truth
- manual contract duplication under frontend packages
- implicit runtime dependencies on local desktop behavior
- page-local coupling to raw backend envelopes

## Immediate next implementation slice

The next real implementation slice should be:

1. freeze `Products` surface contract and readmodel shape
2. scaffold `apps/web` with the correct thin-client package boundaries
3. implement the first `Products` page using the legacy visual context as reference
4. only then open `Shopping` runtime and UI recovery
