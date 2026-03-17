# ADR-0004: Runtime Governance Model

- Status: accepted
- Date: 2026-03-17

## Context

MetalShopping depends on policies, thresholds, feature flags, and configuration decisions that must be traceable, explainable, and consistent across core and workers.

## Decision

- `contracts/governance/*` defines schema
- `bootstrap/seeds/governance/*` defines initial defaults
- effective state lives in the database
- runtime resolution lives in `apps/server_core/internal/platform/governance/*`
- core and workers must share the same semantics
- hardcoded thresholds and policies are not allowed

Resolution hierarchy:

- global
- environment
- tenant
- module
- entity/profile
- feature-target

## Consequences

- governance becomes explicit and auditable
- Go and Python implementations must follow the same resolution contract
- spreading configuration logic across modules becomes a policy violation

