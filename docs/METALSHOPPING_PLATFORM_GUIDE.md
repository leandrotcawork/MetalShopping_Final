# MetalShopping Platform Guide

State observed on: 2026-03-25

This document consolidates what MetalShopping is today, which modules actually exist in the repository, how those modules work together, and which future directions are most consistent with the codebase, the documents under `docs/`, and the `MetalShopping_Project_Bible.docx` file.

## 1. What MetalShopping is

MetalShopping is not a traditional e-commerce product. It was designed as an enterprise operational intelligence platform for finishing-material companies.

In practice, the system aims to concentrate the following in a single product:

- canonical product catalog
- pricing
- inventory
- market monitoring
- shopping workflows and supplier signal collection
- operational and strategic analytics
- future CRM and automation capabilities
- governance for rules, thresholds, and feature flags

The currently frozen identity of the project is:

- monorepo
- server-first
- `apps/server_core` as the canonical core
- specialized workers outside the core
- Postgres as the canonical transactional state
- Go in the core
- Python in integration and compute workers
- thin-client frontend consuming generated SDKs

In delivery terms, the project explicitly follows:

`make it work -> make it clean -> make it fast`

## 2. Architectural thesis of the product

MetalShopping was organized to avoid two common failure modes:

- spreading business rules across frontend, backend, and automation layers
- letting workers become owners of product state

Because of that, the current architecture is based on these principles:

- `contracts/` is the source of truth for APIs, events, and governance
- `server_core` implements public contracts, auth, tenancy, governance, and canonical mutations
- workers handle integration, scraping, compute, and asynchronous delivery
- the frontend does not talk to endpoints through manual `fetch()`; it consumes `@metalshopping/sdk-runtime`
- relevant events leave the core through outbox and feed workers or future read models

The standard product flow today looks like this:

```text
contracts/api/openapi/*.yaml
    -> packages/generated/sdk_ts/generated/*
    -> packages/platform-sdk
    -> apps/web

contracts/events/v1/*.json
    -> outbox in server_core
    -> workers / async consumers

contracts/governance/*
    -> bootstrap/seeds/governance/*
    -> runtime in apps/server_core/internal/platform/governance/*
```

## 3. Current repository structure

The repository is divided into blocks with well-defined responsibilities:

| Block | Current role |
|---|---|
| `apps/server_core` | Main backend, auth, tenancy, governance, APIs, transactional state, and outbox |
| `apps/web` | Thin web client with shell, routes, and page composition |
| `apps/integration_worker` | Python integration and scraping worker, currently with real Shopping Price runtime |
| `apps/analytics_worker` | Reserved for heavy analytics compute and future projections |
| `apps/automation_worker` | Reserved for campaigns, triggers, and future automations |
| `apps/notifications_worker` | Reserved for email, SMS, WhatsApp, and webhook delivery |
| `apps/admin_console` | Planned thin admin console for governance and operations |
| `apps/desktop` | Planned thin desktop client |
| `contracts/` | HTTP contracts, versioned events, and governance artifacts |
| `packages/ui` | Shared visual components |
| `packages/platform-sdk` | Author-owned runtime that encapsulates generated clients |
| `packages/feature-*` | Per-feature adapters, view models, and composition |
| `packages/generated` | Generated SDKs and types; never edited manually |
| `docs/` | SoT, ADRs, plans, operating rules, and domain visions |

## 4. Modules that exist today in the core

The directories under `apps/server_core/internal/modules/` represent the target product topology. As of 2026-03-25, the real state is not uniform: some modules already have functional slices, while others are still architectural placeholders.

### 4.1 Modules with an observed functional slice

| Module | Current state | Real role today |
|---|---|---|
| `iam` | Implemented | Internal roles and permissions for MetalShopping |
| `catalog` | Implemented | Canonical product, taxonomy, identifiers, and portfolio base |
| `pricing` | Implemented | Internal price, replacement cost, average cost, and price history |
| `inventory` | Implemented | Current inventory position and position history |
| `home` | Implemented | Entry-level summary KPIs |
| `shopping` | Implemented | Collection request orchestration, run reads, and exports |
| `suppliers` | Implemented | Supplier directory and driver manifests |
| `analytics_serving` | Implemented | Analytics Home read surface |

### 4.2 Modules reserved in the architecture but without a relevant functional slice yet

| Module | Correct reading of the current state |
|---|---|
| `tenant_admin` | Reserved for tenant administration |
| `sales` | Reserved for future commercial domain work |
| `customers` | Foundation for future CRM |
| `procurement` | Planned, with frozen documentary boundary, but no material runtime slice yet |
| `market_intelligence` | Planned for future evolution |
| `crm` | Planned for a future layer |
| `automation` | Planned for governed automations |
| `integrations_control` | Planned for connector operations |
| `alerts` | Planned for operational and strategic alerts |

## 5. Platform foundation inside `server_core`

Before the business modules, the product already has a fairly well-defined platform layer.

### 5.1 Auth and web session

The current login uses Keycloak as the initial IdP, but the session boundary belongs to `server_core`.

Real flow:

1. The browser starts with `GET /api/v1/auth/session/login`.
2. `server_core` redirects to the OIDC provider.
3. The callback returns to `GET /api/v1/auth/session/callback`.
4. The backend validates state, exchanges the code for a token, creates a session in Postgres, and sets an `HttpOnly` cookie.
5. The frontend bootstraps through `GET /api/v1/auth/session/me`.
6. Session mutations use CSRF cookie + header.

This prevents the browser from becoming the owner of tokens, permissions, or tenancy.

### 5.2 Tenancy runtime

The platform is multi-tenant from the foundation upward. The current pattern combines:

- `tenant_id` in the data model
- runtime tenant in request context
- tenant-aware Postgres session
- `current_tenant_id()` as a filter in tenant-scoped tables
- RLS as the initial isolation baseline

From a design perspective, this prepares the system for growth without duplicating the application per customer.

### 5.3 Runtime governance

Product governance was not left as scattered configuration in code.

Today there are:

- feature flags in `contracts/governance/feature_flags`
- thresholds in `contracts/governance/thresholds`
- policies in `contracts/governance/policies`
- initial seeds in `bootstrap/seeds/governance`
- resolvers in `internal/platform/governance/*`

This layer already affects real runtime behavior. Examples:

- product creation enablement
- session timeout controls
- manual price override policy

### 5.4 Outbox and messaging

Relevant mutations should not depend on synchronous worker calls. The project already implements the outbox foundation in `internal/platform/messaging/outbox`.

In practice, the expected flow is:

- the module writes to the database
- in the same transactional context it records the event in the outbox
- the dispatcher publishes or hands it off to asynchronous consumers

This already supports events such as `catalog_product_created`, `pricing_price_set`, `inventory_position_updated`, and `shopping_run_requested`.

## 6. Current business modules and how they work

### 6.1 IAM

The `iam` module is responsible for the product's internal permissions. It does not replace the IdP; it complements external identity with internal authorization semantics.

What it does today:

- upserts role assignments by user
- responds to permission checks used by other modules
- governs access to catalog, pricing, inventory, and administrative operations

Architectural role:

- Keycloak authenticates
- `iam` decides what the user is allowed to do inside MetalShopping

### 6.2 Catalog

`catalog` is the first strong canonical module of the product. It owns product identity.

What the module already has:

- initial product CRUD
- taxonomy
- product identifiers
- description
- product status
- the base portfolio surface for `Products`

Real responsibility:

- `catalog` is not just registration; it defines the base on top of which `pricing`, `inventory`, `shopping`, `analytics`, and `procurement` operate

Important details:

- product creation is already governed by feature flags and thresholds
- the module already publishes an event through outbox

### 6.3 Pricing

`pricing` owns the semantics of internal price.

Current capabilities:

- register product price
- list price history
- read current price
- store replacement cost
- store average cost when available
- apply governance guards for manual override
- avoid artificial history on reruns without real changes

Role of the module:

- turn price into governed operational data instead of a field scattered across the system

### 6.4 Inventory

`inventory` owns live inventory position.

Current capabilities:

- register inventory position per product
- list position history
- read current position
- store `on_hand_quantity`
- store `last_purchase_at`
- store `last_sale_at`
- maintain position status

Role of the module:

- act as the canonical truth of operational inventory

Important detail:

- the official plan prevents purchasing semantics, lead time, or ERP leakage from invading `inventory`

### 6.5 Home

`home` is the product's first entry surface.

Today the summary endpoint returns:

- total product count
- active product count
- priced product count
- inventory-tracked product count
- last updated timestamp

Role of the module:

- act as an initial operational summary of the platform
- prove the `contract -> backend -> sdk -> screen` thesis

### 6.6 Shopping

`shopping` is one of the most important modules today because it materializes the `server_core + worker` pattern.

What it currently does:

- expose operational bootstrap
- expose run summary
- create `run requests`
- list runs
- detail a run
- list items by run
- summarize status by item and by supplier
- read latest snapshot by product
- manage supplier signals
- list manual URL candidates
- export results to XLSX

Current functional flow:

1. The user starts a collection from the frontend.
2. `server_core` validates auth and tenant.
3. The `shopping` module creates a `run_request`.
4. The core records the `shopping.run_requested` event in the outbox.
5. `integration_worker` processes it in queue mode or event mode.
6. The worker writes results into Postgres.
7. `server_core` serves summary, runs, items, latest snapshots, and exports back to the client.

This flow matters because it represents the desired model for future integrations:

- Python performs field scraping and processing
- Go remains the canonical API boundary
- state lives in Postgres

### 6.7 Suppliers

`suppliers` is no longer just a passive registry. It works as the operational governance layer for supplier runtime behavior.

What already exists:

- supplier directory
- supplier enable/disable controls
- versioned driver manifests
- manifest validation by family
- manifest activation

Observed runtime families:

- `http`
- `playwright`

Strategies already registered:

- `http.mock.v1`
- `http.vtex_persisted_query.v1`
- `http.html_search.v1`
- `http.leroy_search_sellers.v1`
- `http.html_dom_first_card.v1`
- `playwright.mock.v1`
- `playwright.pdp_first.v1`

In practice, `suppliers` became the layer that allows Shopping to scale without hardcoded supplier logic spread across the worker runtime.

### 6.8 Analytics Serving

`analytics_serving` is the first backend interface for the analytics domain.

Today it provides:

- an `Analytics Home` endpoint
- read snapshot metadata
- structured blocks such as KPIs, actions for today, alerts, portfolio distribution, and timeline

Current role:

- serve a tenant-safe read surface
- feed the analytics frontend without pushing heavy analytics logic into the browser

Future role:

- become the serving layer of the commercial intelligence module

## 7. Current workers and how they enter the flow

### 7.1 Integration Worker

This is the most concrete Python worker in the repository today.

It works like this:

- it can run in `queue` mode or `event` mode
- it processes `shopping_price_run_requests`
- it uses `http` and `playwright` strategies
- it decides the lookup term according to lookup policy
- it executes the strategy
- it writes observations into Shopping tables in Postgres

Key architectural point:

- it does not call core HTTP endpoints to complete the job
- it writes to the database; the core reads and exposes the results

This is the official level-1 pattern for asynchronous integration.

### 7.2 Analytics Worker

Today it is more of a reserved boundary than an active operational slice, but its mission is very clear in the docs:

- scoring
- projections
- explainability
- simulations
- heavy processing
- analytics read model refresh

### 7.3 Automation Worker

Planned for:

- triggers
- campaigns
- governed actions
- asynchronous follow-ups

### 7.4 Notifications Worker

Planned for:

- email
- SMS
- WhatsApp
- webhooks

The alert domain remains in the core; this worker only delivers.

## 8. Current frontend and thin-client boundary

The frontend was explicitly designed not to become a second business application.

### 8.1 `apps/web`

This is the main thin client today.

Observed routes:

- `/home`
- `/products`
- `/shopping`
- `/analytics`
- `/analytics/products/:pn/*`
- `/login`

It owns:

- shell
- routing
- providers
- page composition

It should not own:

- canonical contracts
- business rules
- manual HTTP calls

### 8.2 `packages/platform-sdk`

This package is central to the current web architecture.

It:

- encapsulates generated clients
- standardizes headers, trace id, CSRF, and credentials
- maps HTTP errors into a consistent runtime layer
- expands the surface used by the frontend with more ergonomic methods

Without it, the frontend would depend directly on generated code; with it, there is a stable and controlled author-owned layer.

### 8.3 `packages/ui`

It centralizes shared components already reused across surfaces, such as:

- `AppFrame`
- `Button`
- `Checkbox`
- `FilterDropdown`
- `MetricCard`
- `MetricChip`
- `SelectMenu`
- `SortHeaderButton`
- `StatusBanner`
- `StatusPill`
- `SurfaceCard`

### 8.4 Feature packages

`feature-auth-session`

- login
- session bootstrap
- authenticated route
- redirect/auth bootstrap screens

`feature-products`

- adapters and widgets for the Products surface
- portfolio composition

`feature-analytics`

- compatibility adapters and DTOs for the legacy frontend
- current Analytics surfaces
- product workspace
- visual migration track focused on parity

## 9. Current contracts and integration state

### 9.1 Observed OpenAPI surfaces

Today the repository already has contracts for:

- `analytics`
- `auth_session`
- `catalog`
- `home`
- `iam`
- `inventory`
- `pricing`
- `products`
- `shopping`
- `suppliers`

### 9.2 Observed events

There are already versioned events for:

- `catalog_product_created`
- `iam_role_assigned`
- `inventory_position_updated`
- `pricing_price_set`
- `shopping_run_requested`

### 9.3 Observed governance

There are already artifacts for:

- feature flags
- thresholds
- policies

In other words, the project has already moved beyond the "folder structure only" phase and entered a phase where contracts, persistence, frontend, and asynchronous flow are actually connected.

## 10. How the system works end to end today

### 10.1 Standard synchronous HTTP flow

1. The user enters through `apps/web`.
2. The session is authenticated via `auth/session`.
3. The frontend calls `sdk.*` through `@metalshopping/sdk-runtime`.
4. `server_core` authenticates the principal and resolves the tenant.
5. The business module executes validations and rules.
6. Data access happens through tenant-aware Postgres paths.
7. The backend responds with contract-aligned DTOs.

### 10.2 Mutation flow with an event

1. The client calls a core mutation.
2. The module writes transactional state.
3. The corresponding event enters the outbox.
4. The async consumer processes it later.
5. The system avoids synchronous coupling between request and worker.

### 10.3 Composed read-surface flow

The best example today is `Products`.

It exists because the backend consolidates:

- product identity through `catalog`
- current price through `pricing`
- current inventory position through `inventory`

That way, the frontend does not need to stitch three modules together manually.

### 10.4 Shopping Price flow

This is the most representative flow of the target model:

```text
apps/web
  -> sdk-runtime.shopping.createRunRequest()
  -> server_core/shopping
  -> outbox: shopping.run_requested
  -> integration_worker
  -> Postgres: runs, items, snapshots, signals
  -> server_core/shopping read APIs
  -> apps/web shows progress and results
```

## 11. Ownership boundaries that are already frozen

One of MetalShopping's strengths is that several ownership boundaries were frozen early. This reduces the risk of semantic drift.

Important examples:

- `catalog` owns product identity
- `pricing` owns internal price and cost semantics
- `inventory` owns inventory position
- `procurement` will own replenishment and lead-time semantics, not `inventory`
- `analytics` will own derived metrics, scoring, recommendations, and explainability, not canonical writes
- workers must not become owners of business truth

## 12. Where the project stands today

Crossing the codebase with the SoT shows a very clear state:

- the auth, tenancy, governance, and outbox foundation is in place
- `catalog`, `pricing`, and `inventory` already exist as real modules
- `Products` is already a real surface using backend + SDK
- `Home` already has an endpoint and page
- `Shopping` already has a functional backend and a real worker
- `Analytics` already has an initial serving layer and a strong frontend migration in progress
- `CRM`, `procurement`, `alerts`, `automation`, and the stronger AI layers still live in the future boundary

In short, the product has already left the purely conceptual phase. It is in a foundation-implemented stage with real product slices.

## 13. Future projections of how the platform should work

The most consistent projections are not guesswork; they appear repeatedly in the codebase, ADRs, SoT, and Project Bible.

### 13.1 Immediate horizon

The immediate horizon of the product is to close the operational Layer 1.

That means:

- consolidating `Home`
- closing `Shopping` with stronger operational parity
- continuing the rise of `Analytics` on top of real endpoints
- opening `CRM v1`

The correct reading is: the project first wants to put the modules used daily by the internal team on solid ground.

### 13.2 Domain horizon

After the initial operational layer, the clearest trend in the repository is:

- `procurement` is born on top of published inputs, not direct ERP reads
- `analytics` evolves from dashboard to decision engine
- `suppliers` evolves from directory/manifest into a governed connector mesh
- `shopping` stops being just a page and becomes a complete operational workflow
- `products` remains a consolidated read surface on top of canonical domains

### 13.3 Analytics horizon

The repository already describes `analytics` as an intelligence module in eight layers:

- metrics
- rules
- calculations
- classifications
- recommendations
- campaigns
- alerts
- AI

If that vision is followed, the future of the product is:

- showing what is getting worse and why
- suggesting actions by SKU, brand, and taxonomy
- prioritizing work queues
- generating traceable explainability
- producing buying and pricing suggestions backed by evidence

### 13.4 Automation and AI horizon

The Project Bible points to a very clear line:

- automatic campaigns
- marketplace agents
- AI-powered market research
- predictive customer profiling
- buying recommendations
- intelligent alerts

Based on the current repository design, this will probably happen like this:

- `analytics_worker` produces scoring and recommendations
- `automation_worker` turns recommendations into governed actions
- `notifications_worker` delivers alerts or campaign outputs
- `server_core` remains the official boundary for reads, approval, and auditability

### 13.5 Multi-customer horizon

Today the focus is the internal team, but the architecture was already designed for a broader multi-tenant future.

The strongest signals are:

- tenancy from the foundation upward
- runtime governance
- versioned contracts
- thin clients
- separation between external identity and internal IAM

That prepares the product to move from internal operational software to a commercial platform without rewriting the foundations.

## 14. Final executive reading

MetalShopping today is already an enterprise platform under advanced construction, not a disorganized prototype.

Its current shape can be summarized as follows:

- the Go core already controls auth, tenancy, governance, contracts, and canonical data
- the web app already works as a thin client on top of generated SDKs
- Shopping already proves the `Go + Python + Postgres + outbox` model
- the foundational domains `catalog`, `pricing`, and `inventory` already support the first strong product surface
- the next major evolution is to turn `Analytics`, `Shopping`, `CRM`, and then `procurement` into complete operational engines

If the plan remains coherent with what has already been frozen, the result will not be only a system for registration and dashboards. It will become a platform for commercial intelligence, operational execution, and governed automation for the finishing-material retail sector.
