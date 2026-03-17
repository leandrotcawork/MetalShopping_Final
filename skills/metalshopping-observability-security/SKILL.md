---
name: metalshopping-observability-security
description: Review or shape MetalShopping platform work against the repository observability and security baseline. Use when designing core or worker behavior that affects logs, metrics, traces, correlation, auth, authz, tenancy, auditability, or operational safety, or when checking whether a proposed design meets the platform baseline before implementation.
---

# MetalShopping Observability Security

## Overview

Use this skill to evaluate platform work against the repo's non-negotiable observability and security baseline before implementation spreads.

## Workflow

1. Read only the minimum repo context:
   `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
   `docs/SERVER_CORE_OPERATING_MODEL.md`
   `docs/WORKER_OPERATING_MODEL.md`
2. Confirm the target area:
   core
   worker
   contract
   operational surface
3. Check traceability, auditability, auth, authz, tenancy, and abuse controls.
4. Call out gaps before implementation proceeds.
5. Finish with the review checklist in `references/obssec-checklist.md`.

## References

- For the repo workflow and file touchpoints, read `references/repo-obssec-flow.md`.
- For the final review pass, read `references/obssec-checklist.md`.

