---
id: hippocampus-strategy
title: MetalShopping Strategy
type: hippocampus
tags: [strategy, product, phase, goals]
updated_at: 2026-03-26
---

# MetalShopping Strategy

## Current Phase

**Foundation implementation** — make it work first

- Architecture: approved and frozen
- Code status: core foundation running with initial platform + business slices
- Delivery mode: close structural gaps while staying aligned with frozen architecture
- Legacy backend: intentionally not in use
- Next gate: complete remaining business modules and analytics surfaces

## Product Vision

MetalShopping is the operational intelligence layer for metal and construction material commerce. It enables:

1. **Commercial strategy** — pricing intelligence, margin control, competitive positioning
2. **Procurement** — supplier management, negotiation tracking, purchase orders
3. **Market monitoring** — price signals, market index, competitor tracking
4. **CRM** — customer portfolio, opportunity management
5. **Automations** — workflow triggers, notification rules
6. **Analytics** — operational and strategic dashboards for all domains

## Platform Direction

- Server-first: Go monolith owns canonical state
- Thin clients (web, desktop, admin_console) consume via SDK
- Explicit contracts drive the SDK boundary — no ad-hoc API calls
- Workers (Python) handle compute-heavy tasks without calling back into server_core
- Governance layer controls feature flags, policies, thresholds at runtime

## Non-Goals (Current Phase)

- One database per tenant
- Legacy backend reuse
- Premature optimization or caching
- Feature flags for backwards compatibility

## Priorities

1. Structural completeness — all 18 modules minimally functional
2. Tenant isolation correctness — no cross-tenant data leaks ever
3. Contract parity — SDK and contracts reflect actual behavior
4. Analytics surfaces — all 11 read surfaces operational
5. Frontend migration — feature-* packages replace legacy UI

## Decision Filter

Every decision passes: *"Would a Stripe or Google senior engineer approve this in code review?"*
