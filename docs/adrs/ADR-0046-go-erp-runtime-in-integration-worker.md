# ADR-0046: Go ERP Runtime in integration_worker

- Status: accepted
- Date: 2026-04-02

## Context

The approved platform direction (see `docs/PROJECT_SOT.md`) states "Python in workers during transition." This reflects the current reality: the `apps/integration_worker` app contains a Python `shopping_price_worker` that handles crawl and scraping workloads, which is appropriate for that use case.

The ERP Integration Subplatform (v0.1) requires a new synchronization runtime inside `apps/integration_worker`. This runtime must implement:

- A multi-ERP connector model (SAP, TOTVS Protheus, etc.) with pluggable adapters
- A staged ingestion pipeline: raw ingest → staging → reconciliation
- A run ledger with atomic claim semantics (`SELECT ... FOR UPDATE SKIP LOCKED`) to prevent duplicate processing across replicas
- Tenant isolation via RLS, consistent with platform governance
- Observability integration (structured logging, metrics, trace propagation)

These requirements demand:

- Strong type safety across pipeline stages (raw rows, staging rows, reconciliation diffs)
- Native Postgres client patterns that align with `server_core` platform packages (`pgdb`, `outbox`, `governance`)
- Shared idioms, conventions, and library reuse with `apps/server_core`

Python is used by the existing `shopping_price_worker` and remains appropriate for crawl and scraping workloads. However, the ERP sync runtime is a foundational data infrastructure component, not a crawler. The correctness guarantees required for financial and procurement data ingestion make Go the clearly better fit.

## Decision

The `apps/integration_worker` app will contain two runtimes sharing the same Postgres instance:

1. **Python `shopping_price_worker`** — unchanged, continues as the crawl/scraping runtime. No modifications to its entrypoint, dependencies, or deployment.
2. **Go `erp-sync` runtime** at `apps/integration_worker/cmd/erp-sync/` — implements the ERP connector subplatform described in the ERP Integration Subplatform spec.

### Coexistence rules

- **Separate entrypoints:** Python runtime lives under `cmd/shopping-price-worker/` (or equivalent existing layout); Go runtime lives under `cmd/erp-sync/`.
- **Independent builds and dependency management:** Python uses `requirements.txt` or `poetry`; Go uses `apps/integration_worker/go.mod`. The two toolchains do not interact.
- **Shared Postgres instance and schema:** Both runtimes connect to the same database. RLS is enforced per tenant for all table access. Schema migrations are owned by `server_core` (the platform migration authority).
- **No shared runtime state:** There are no inter-process calls between the two runtimes. All coordination happens through database rows only (e.g., run ledger claims, staging state).
- **Independent deployment:** Each runtime is built and deployed as a separate container or process. A change to one does not require redeployment of the other.
- **Go module root:** `apps/integration_worker/go.mod` defines the Go module for this app. It may import shared platform packages from the monorepo root as needed.

### Justification

- Go aligns with `server_core` patterns, enabling reuse of `pgdb`, `outbox`, and `governance` packages without adaptation layers.
- The staged ingestion pipeline, run claim loop, and raw/staging table interactions benefit directly from Go's type system and explicit error handling.
- The `SELECT ... FOR UPDATE SKIP LOCKED` claim pattern is idiomatic in Go with `pgx`; replicating it reliably in Python requires additional infrastructure.
- Python remains the correct and preferred choice for crawl and scraping workloads. This ADR does not change that.

## Contracts (touchpoints)

- `apps/integration_worker/go.mod` — Go module definition (created in Task 21 of the ERP plan)
- `apps/integration_worker/cmd/erp-sync/` — Go entrypoint for the ERP sync runtime
- Postgres schema tables: `erp_raw_*`, `erp_staging_*`, `erp_run_ledger` (owned by platform migrations)
- Platform packages: `apps/server_core/internal/platform/pgdb`, `apps/server_core/internal/platform/outbox`, `apps/server_core/internal/platform/governance`

## Consequences

- `apps/integration_worker/go.mod` must be created (Task 21 of the ERP Integration Subplatform plan).
- CI must be updated to build and test Go code under `apps/integration_worker/cmd/erp-sync/...` in addition to the existing Python worker tests.
- This Go exception is **scoped to the ERP sync runtime only**. New Python workers added to `integration_worker` or other worker apps do not require a separate ADR to remain Python.
- Future Go expansion into other worker apps (outside `integration_worker`) requires a dedicated ADR per worker app.
- The "Python in workers during transition" platform direction remains in effect for all workloads other than the ERP sync runtime. `PROJECT_SOT.md` is updated with a scoped exception note referencing this ADR.

## Alternatives considered

- **Python ERP sync runtime:** Rejected due to the difficulty of expressing the staged pipeline type model and `SELECT ... FOR UPDATE SKIP LOCKED` claim semantics reliably in Python. The correctness bar for financial/procurement data ingestion is higher than for price scraping.
- **ERP runtime inside `server_core`:** Rejected because the ERP sync runtime runs as a background worker with its own scheduling, concurrency, and retry model. Embedding it in the HTTP core would couple deployment, add noise to the core, and violate the "specialized workers outside the core" platform direction.
- **Separate new app (e.g., `apps/erp_worker`):** Considered but rejected for v0.1 to avoid proliferating apps before the integration subplatform is proven. Co-locating with `integration_worker` under a clear coexistence contract is lower overhead. A future ADR may split the app if scope grows.
