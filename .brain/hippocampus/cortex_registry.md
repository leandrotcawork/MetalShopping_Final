---
id: hippocampus-cortex-registry
title: Cortex Registry
region: hippocampus
tags: [registry, cortex, domains]
links:
  - hippocampus/architecture
weight: 0.9
updated_at: 2026-03-24T10:00:00Z
---

# Cortex Registry

Map of all cortex regions in MetalShopping's Brain.

| Region | Purpose | Coverage |
|--------|---------|----------|
| **cortex/backend** | Go modular monolith, modules, patterns | API contracts, service layer, models, auth, outbox, events |
| **cortex/frontend** | React thin-client, features, patterns | Components, routes, SDK usage, state, styling |
| **cortex/database** | PostgreSQL, multi-tenant patterns | Schema, migrations, tenant isolation, queries |
| **cortex/infra** | Deployment, CI/CD, operations | Docker, Kubernetes, environments, observability |
| **sinapses/** | Cross-cutting flows | Tenant isolation, outbox events, SDK data flow, analytics routing |
| **lessons/** | Captured failures & patterns | Anti-patterns, corrections, architectural decisions |

## Region Details

### cortex/backend
- API contract patterns (OpenAPI)
- Service layer conventions
- Entity/domain model patterns
- Authentication & authorization
- Outbox event publishing
- Module structure (domain → application → ports → adapters → transport)

### cortex/frontend
- Component library usage (`packages/ui`)
- React hooks & patterns
- SDK method usage (`@metalshopping/sdk-runtime`)
- Styling (tokens, CSS modules)
- State management
- Routing structure

### cortex/database
- PostgreSQL multi-tenant schema
- Tenant isolation queries
- Migration patterns
- Index strategies
- Read models (denormalized views)

### cortex/infra
- Docker image builds
- Kubernetes deployments
- Environment configuration (dev, staging, prod)
- CI/CD pipeline (GitHub Actions)
- Observability (logging, tracing, monitoring)

### sinapses/ (Cross-Cutting)
- **Tenant isolation flow** — How auth → tenancy → database isolation works
- **Outbox event flow** — How write + event publish work atomically
- **SDK data flow** — How contracts → generated SDK → frontend data access works
- **Analytics routing** — How $analytics-orchestrator routes tasks

### lessons/ (Distributed Knowledge Database)

**Lesson curation** uses a three-tier classification:

1. **Domain-specific lessons** — stored in `cortex/<domain>/lessons/`
   - `cortex/backend/lessons/` — 5 lessons (tenant safety, handlers, outbox, workers, observability)
   - `cortex/frontend/lessons/` — 5 lessons (SDK contracts, design system, legacy migration, UI imports)

2. **Cross-domain lessons** — stored in `lessons/cross-domain/`
   - 3 lessons (generated artifacts read-only, completion validation, todo.md editing)

3. **Archived lessons** — stored in `lessons/archived/`
   - 23 lessons (cosmetic frontend tweaks, overly-specific fixes, phase-specific procedures)

**Lesson criteria:** Kept lessons satisfy ALL three:
1. **Cross-domain applicability** — applies to multiple features, not just one
2. **Prevents repeated mistakes** — captures a pattern that causes failures if violated
3. **Architectural, not cosmetic** — describes structure/boundaries, not styling/implementation

**Escalation rule:** When 3+ lessons capture the same pattern, escalate to `hippocampus/conventions.md` as an absolute rule.

---

**Created:** 2026-03-24 | **Updated:** 2026-03-24 | **Weight:** 0.9
