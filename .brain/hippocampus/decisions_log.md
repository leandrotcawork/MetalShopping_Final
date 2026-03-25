---
id: hippocampus-decisions-log
title: Decisions Log
region: hippocampus
tags: [adr, decisions, architecture]
links:
  - hippocampus/architecture
weight: 0.8
updated_at: 2026-03-24T10:00:00Z
---

# Architecture Decision Records (ADRs)

This log captures all significant architectural decisions that shaped MetalShopping.

## ADR-0001: Multi-Tenant Shared-Database Architecture

**Decision:** Use shared-database multi-tenancy (single PostgreSQL, multiple tenants).

**Rationale:**
- Simpler operational model than separate databases
- Easier analytics (single warehouse for all tenants)
- Lower infrastructure cost
- Requires strict tenant isolation at application level

**Consequences:**
- Must enforce tenant checks everywhere (high discipline)
- Tenant data leaks are critical bugs
- All queries must filter by `current_tenant_id()`
- Database schema is shared; multi-tenant constraints needed

**Status:** Accepted | **Date:** 2026-01 | **Related:** [[cortex/database/schema.md]]

---

## ADR-0002: Contract-Driven SDK Generation

**Decision:** All API data flows generated from hand-authored contracts (OpenAPI + JSON Schema).

**Rationale:**
- Single source of truth for data structures
- Type-safe frontend code (no manual type definitions)
- Automatic SDK regeneration when contracts change
- Enforces API-first thinking

**Consequences:**
- Contracts must be maintained carefully
- Manual edits to generated code are forbidden
- Contract changes ripple to frontend automatically
- Build step required: contract → SDK generation

**Status:** Accepted | **Date:** 2026-01 | **Related:** [[cortex/backend/api.md]]

---

## ADR-0003: Transactional Outbox Pattern for Event Publishing

**Decision:** Use outbox pattern for all event publishing. Events appended within same transaction as domain writes.

**Rationale:**
- Guarantees no lost events (atomicity)
- Decouples event producers from consumers
- Handles failure scenarios cleanly (retry-safe)
- Standard pattern at Stripe, Netflix, etc.

**Consequences:**
- Every write must append events in same transaction
- Events processed asynchronously (eventual consistency)
- Duplicate event handling required (idempotency)
- Outbox cleanup/archival needed

**Status:** Accepted | **Date:** 2026-01 | **Related:** [[cortex/backend/outbox.md]]

---

## ADR-0004: React Thin-Client with SDK-Only Data Access

**Decision:** Frontend accesses all data through generated SDK only. No raw HTTP calls.

**Rationale:**
- Enforces contract compliance
- Type-safe data access
- Single source of truth for data structure changes
- Easier to add auth/validation at SDK level

**Consequences:**
- All features must have corresponding SDK methods
- Contract changes require frontend code changes
- Complex queries must be handled by backend (SDK methods)
- Difficult to bypass SDK for debugging

**Status:** Accepted | **Date:** 2026-02 | **Related:** [[cortex/frontend/sdk.md]]

---

## ADR-0005: Modular Monolith Backend with Feature Ownership

**Decision:** Go backend structured as modular monolith. Each feature owns its domain → application → adapters layers.

**Rationale:**
- Monolith simplicity (single deployment unit)
- Module boundaries allow independent testing
- Clear ownership (who owns which feature)
- Easy to split to microservices later if needed

**Consequences:**
- No inter-module dependencies (strict boundaries)
- Shared platform infrastructure (auth, tenancy, outbox)
- Feature teams can work independently
- Requires clear module naming and layering

**Status:** Accepted | **Date:** 2026-01 | **Related:** [[cortex/backend/index.md]]

---

## ADR-0006: ForgeFlow Mini as Development Brain

**Decision:** Use ForgeFlow Mini as persistent knowledge system for project context, patterns, and lessons.

**Rationale:**
- Captures lessons so they aren't re-discovered
- Patterns accumulate and guide future work
- Brain visualizes architecture and dependencies
- Reduces onboarding friction for new developers

**Consequences:**
- Must maintain Brain as project evolves
- Brain becomes source of truth for conventions
- Consolidation cycles required regularly
- Brain-driven task routing enforces patterns

**Status:** Accepted | **Date:** 2026-03 | **Related:** [[.brain]] directory structure

---

**Note:** Additional ADRs should be added as major architectural decisions are made. See [[docs/adrs/]] for detailed decision records.

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.8
