# ADR-0033: Shopping Driver Smoke Suite v1 (multi-supplier)

- Status: draft
- Date: 2026-03-20

## Context

We already have reproducible smoke entrypoints for:

- publishing a run request + outbox event (`apps/server_core/cmd/smoke-shopping-event`)
- running the worker once in event mode (`scripts/smoke_shopping_event_local.ps1`)

But we do not have a single "driver suite" smoke that:

- runs a fixed set of suppliers (legacy parity set) in one command
- records objective evidence for each supplier
- provides a stable acceptance gate before expanding UI and user workflow

Without this, we will keep "it works on my machine" drift as suppliers grow.

## Decision

Add a deterministic smoke suite script that runs the accepted supplier set and produces a single report.

Rules:

- The suite must be deterministic and not depend on the web UI.
- Each supplier run must:
  - seed supplier + manifest (or require they exist)
  - execute via event flow (ADR-0025) or queue flow (ADR-0018) with explicit mode selection
  - query the DB after execution to capture objective evidence:
    - run_id, counts by (channel,item_status), sample observed prices, HTTP status, lookup_term
- Evidence must be written to a markdown report under `docs/` for review.
- The suite must support "expected non-OK outcomes" per supplier when the supplier is known to block or return no results (explicitly configured).

## Contracts (touchpoints)

- No new external contracts.
- Uses existing DB read surfaces as evidence source.

## Implementation Checklist

1. Add suite script:
   - `scripts/smoke_shopping_driver_suite_local.ps1`
   Skill: `metalshopping-worker-patterns`
2. Add a DB evidence helper:
   - either in PowerShell (via Python + psycopg) or a small Go helper binary
   Skill: `metalshopping-worker-scaffold`
3. Add a standard report template:
   - `docs/SHOPPING_DRIVER_SUITE_ACCEPTANCE.md`
   Skill: `metalshopping-adr-updates`
4. Observability/security review (no credentials leakage; tenant-safe queries)
   Skill: `metalshopping-observability-security`

## Implementation Snapshot (2026-03-20)

- Added deterministic suite runner:
  - `scripts/smoke_shopping_driver_suite_local.ps1`
  - supports supplier matrix + per-supplier expected statuses + explicit allow-empty behavior
  - executes event flow per supplier and records run evidence
- Added DB evidence helper:
  - `scripts/smoke_shopping_driver_suite_report.py`
  - captures `run_request`, `run_id`, `total_items`, counts by `(channel,item_status)`, and samples (`observed_price`, `http_status`, `lookup_term`)
- Added standard report artifact:
  - `docs/SHOPPING_DRIVER_SUITE_ACCEPTANCE.md`
- Validation evidence:
  - `.venv\Scripts\python.exe -m py_compile scripts/smoke_shopping_driver_suite_report.py` -> pass
  - `scripts/smoke_shopping_driver_suite_local.ps1` (outside sandbox) -> pass and writes report

Pending for acceptance:
- rerun suite with non-empty catalog scope (`InputMode=catalog`) so evidence includes non-zero item counts/samples
- complete OBRA_FACIL Playwright non-mock evidence (ADR-0034)

## Acceptance Evidence (for Status: accepted)

- Smoke suite run produces a report with:
  - DEXCO (VTEX non-mock)
  - TELHA_NORTE (VTEX non-mock)
  - CONDEC (HTML search non-mock)
  - OBRA_FACIL (Playwright PDP-first non-mock)

## Consequences

- The repo gains a single command to validate driver runtime across suppliers.
- Backend completion can be proven before UI investment.
