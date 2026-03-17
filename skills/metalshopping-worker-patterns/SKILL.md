---
name: metalshopping-worker-patterns
description: Create or review MetalShopping worker structure, responsibilities, and async behavior using the worker operating model, contract-first rules, and event-driven boundaries. Use when shaping `analytics_worker`, `integration_worker`, `automation_worker`, or `notifications_worker`, or when checking whether a responsibility belongs in a worker instead of `server_core`.
---

# MetalShopping Worker Patterns

## Overview

Use this skill to create or review worker responsibilities and boundaries with the repository's async-first operating model. Keep work anchored to the worker model instead of letting workers become shadow backends.

## Workflow

1. Read only the minimum repo context:
   `docs/WORKER_OPERATING_MODEL.md`
   `docs/CONTRACT_EVOLUTION_RULES.md`
   `docs/READMODEL_AND_EVENTS_RULES.md`
   `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
2. Confirm the target worker and its category.
3. Confirm the responsibility belongs in async compute, ingestion, automation, or delivery rather than canonical core logic.
4. Keep contract consumption and publication explicit.
5. Keep retry, idempotency, and correlation expectations visible.
6. Finish with the review checklist in `references/worker-checklist.md`.

## Worker rules

- workers do not own canonical truth
- workers consume contracts and publish governed outputs
- workers stay replaceable without redefining product truth
- workers should be retry-safe and correlation-aware
- workers must not become synchronous request dependencies

## References

- For the repo workflow and file touchpoints, read `references/repo-worker-flow.md`.
- For the final review pass, read `references/worker-checklist.md`.

