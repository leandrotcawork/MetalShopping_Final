# Operational Surfaces Plan

## Purpose

Define how MetalShopping should recover its first real product surfaces from the legacy application without reintroducing legacy coupling or weakening the new platform architecture.

This plan should now be read together with `docs/FRONTEND_MIGRATION_CHARTER.md`, which freezes the rule that legacy visual language is preserved while legacy frontend shortcuts are rejected.

The immediate target surfaces are:

- `Home`
- `Shopping`
- `Analytics`

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

### 1. `Home` first

`Home` is now the first operational module under the make-it-work-first mode.

The first Home slice must stay simple:

- real tenant-scoped KPIs
- no mock data
- no frontend-side aggregation
- no over-engineering before real usage

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

### 3. `Analytics` third

`Analytics` should be implemented after `Home` and the first `Shopping` read flow are operational.

The first analytics slice should map existing front-end visual needs to explicit backend read endpoints before adding advanced computation.

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

- visual composition patterns from legacy workspace shell and operations center
- app shell/provider split from the legacy web app
- package decomposition concept from the legacy frontend
- good reusable widget patterns already proven in operational flows

### Do not reuse blindly

- legacy sidecar endpoint shapes as the new contract source of truth
- manual contract duplication under frontend packages
- implicit runtime dependencies on local desktop behavior
- page-local coupling to raw backend envelopes

## Immediate next implementation slice

The next real implementation slices should follow:

1. freeze Shopping read/write contracts and worker-to-postgres flow
2. implement Shopping page binding using generated SDK and backend-owned reads
3. freeze Analytics read contracts based on real screen needs before implementation

Home Level 1 acceptance is closed and recorded in:

- `docs/HOME_LEVEL1_ACCEPTANCE.md`
