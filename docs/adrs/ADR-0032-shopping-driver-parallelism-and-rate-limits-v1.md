# ADR-0032: Shopping Driver Parallelism and Rate Limits v1

- Status: draft
- Date: 2026-03-20

## Context

Legacy drivers already use workers/tabs/pools to keep runs fast while staying stable under marketplace protections.

In the target repo, the Shopping Price worker is correct but still runs effectively sequential execution per item. This will not scale when we add the legacy supplier set (OBRA_FACIL, DEXCO, TELHA_NORTE, CONDEC, etc.), and it will increase:

- runtime duration for medium/large runs
- risk of being rate-limited / blocked
- noisy retries and unstable run quality

We need deterministic, bounded parallelism rules that remain safe under multi-tenant execution and do not depend on UI behavior.

## Decision

Introduce a family-aware concurrency model with explicit per-tenant and per-supplier caps, configurable via manifest config (bounded ranges validated by `server_core`).

Rules:

- Parallelism is controlled by the worker runtime, not by the UI.
- Concurrency defaults must be conservative, with safe upper bounds validated during manifest validation.
- HTTP and Playwright have separate concurrency controls:
  - HTTP: `maxConcurrency`, `requestsPerSecond` (token bucket), retry/backoff policy
  - PLAYWRIGHT: `tabs`, `headless`, per-run circuit breaker and navigation timeouts
- Concurrency must never bypass tenancy isolation:
  - all persistence remains through tenant-scoped transactions
- A run must be reproducible and bounded:
  - do not spawn unbounded threads/tasks
  - enforce timeouts per request and per run

### Manifest config keys (v1)

HTTP (optional):

- `maxConcurrency` (1..16)
- `requestsPerSecond` (0.1..10.0)
- `maxRetries` (1..8) (already exists)
- `timeoutSeconds` (1..60) (already exists)
- `retryHttpStatuses` (array of integers, bounded)

PLAYWRIGHT (optional):

- `tabs` (1..10)
- `headless` (bool) (already exists)
- `timeoutSeconds` (1..120) (already exists)
- `circuitBreakerThreshold` (1..10)

## Contracts (touchpoints)

- JSON Schemas:
  - evolve `contracts/api/jsonschema/supplier_driver_manifest_http_v1.schema.json`
  - evolve `contracts/api/jsonschema/supplier_driver_manifest_playwright_v1.schema.json`
- Deterministic validation:
  - evolve `apps/server_core/internal/platform/suppliers/driver_family_registry.go`

## Implementation Checklist

1. Contracts + schema evolution for concurrency keys
   Skill: `metalshopping-contract-authoring` + `metalshopping-adr-updates`
2. Go validation for ranges/types
   Skill: `metalshopping-platform-packages`
3. Worker runtime implementation (token bucket + bounded executor)
   Skill: `metalshopping-worker-patterns` + `metalshopping-worker-scaffold`
4. Observability/security review (rate limiting + block detection signals)
   Skill: `metalshopping-observability-security`
5. Smoke suite adds concurrency scenario
   Skill: `metalshopping-worker-patterns`

## Implementation Snapshot (2026-03-20)

- Worker runtime now executes supplier observations via bounded thread pool with per-supplier semaphores:
  - HTTP cap from `maxConcurrency` (default 4, bounded to 1..16)
  - PLAYWRIGHT cap from `tabs` (default 1, bounded to 1..10)
- HTTP rate limit token bucket added per supplier using `requestsPerSecond` when configured.
- HTTP retry policy now supports `retryHttpStatuses` (with sane defaults when missing).
- Manifest contracts evolved:
  - HTTP schema: `maxConcurrency`, `requestsPerSecond`, `retryHttpStatuses`
  - PLAYWRIGHT schema: `tabs`, `circuitBreakerThreshold`
- Server-side deterministic validation evolved in `driver_family_registry.go` for all new keys and ranges.
- Validation evidence:
  - `go build ./apps/server_core/...` -> pass
  - `.venv\Scripts\python.exe -m py_compile apps/integration_worker/shopping_price_worker.py` -> pass
  - `scripts/smoke_shopping_event_local.ps1` (outside sandbox) -> pass

Pending for acceptance:
- multi-supplier bounded-runtime smoke evidence (ADR-0033 suite)

## Acceptance Evidence (for Status: accepted)

- Build/test:
  - `go build ./apps/server_core/...` -> pass
  - `python -m py_compile apps/integration_worker/shopping_price_worker.py` -> pass
- Smoke:
  - multi-supplier smoke (ADR-0033) shows bounded runtime time and no unbounded task growth

## Consequences

- Runs become stable under growth in number of suppliers/products.
- Concurrency is deterministic and reviewable via manifest config.
- Some suppliers may require tighter defaults; this becomes a config change, not bespoke code.
