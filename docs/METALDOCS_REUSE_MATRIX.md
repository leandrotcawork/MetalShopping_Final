# MetalDocs Reuse Matrix

## Status

- Type: transitional migration document
- Scope: evaluate selective reuse from `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalDocs`
- Valid only during migration planning and execution
- Delete this document after the migration decisions are executed and absorbed into code or stable SoT docs

## Purpose

Decide what should be reused, adapted, or discarded from `MetalDocs` before implementation begins in `MetalShopping_Final`.

This document is not a permanent architecture source of truth.
It exists to avoid re-reviewing the same legacy surfaces during migration.

## Decision scale

- `reuse as pattern`: keep the idea or structure, but reimplement for MetalShopping
- `adapt before reuse`: salvageable, but only after structural changes
- `do not reuse`: not suitable as a foundation for MetalShopping

## Executive decision

The recommended strategy is selective reuse, not full migration and not full rewrite.

Use `MetalDocs` as:

- a reference for modular boundaries
- a reference for migration discipline
- a reference for operational rigor and hardening culture

Do not use `MetalDocs` as:

- the direct authentication foundation
- the final platform security model
- the final runtime architecture for rate limit and observability
- a drop-in core for MetalShopping

## Matrix

| Area | Current state in MetalDocs | Decision | Why |
| --- | --- | --- | --- |
| modular structure | Good separation across modules and platform packages | `reuse as pattern` | The repo shows healthy separation of domain, application, infrastructure, and delivery concerns |
| Postgres connectivity | Functional config and connection bootstrap | `adapt before reuse` | The base is useful, but MetalShopping needs stronger production defaults, tenancy awareness, and canonical core ownership |
| IAM role persistence | Basic RBAC persistence exists and is understandable | `adapt before reuse` | The repository and schema ideas are reusable, but they must move behind centralized auth, tenancy, and stronger contracts |
| auth middleware | Identity is trusted from `X-User-Id` header | `do not reuse` | This is not a sufficient authentication boundary for MetalShopping |
| dev fallback mode | App can silently boot in memory mode with dev roles | `do not reuse` | MetalShopping core must fail safe, not degrade into a non-canonical runtime by default |
| rate limiting | In-memory, per-process limiter | `do not reuse` | Not enough for distributed or platform-grade runtime behavior |
| observability | Structured logs and local metrics exist | `adapt before reuse` | The direction is useful, but MetalShopping needs stronger tracing, correlation, and operational surfaces |
| security runbooks and hardening discipline | Good project hygiene and release thinking | `reuse as pattern` | The process mindset is valuable and should be carried forward |
| test discipline | There are tests, but the unit suite is not fully green now | `adapt before reuse` | Useful as a signal of intent, but not strong enough to inherit blindly as a quality baseline |

## Recommended migration posture

### Reuse as pattern

- module/package boundary discipline
- migrations as explicit, reviewable artifacts
- repository-based Postgres wiring as a starting shape
- operational runbooks and hardening gates as engineering culture

### Adapt before reuse

- Postgres config and connection package
- IAM persistence and admin flows
- role cache invalidation pattern
- structured HTTP observability pattern

### Do not reuse

- request identity based on `X-User-Id`
- auth toggle semantics as the product foundation
- memory repository fallback as the default runtime path
- noop publisher as a silent runtime downgrade
- in-memory rate limiter as the platform solution

## Concrete extraction targets

If we mine `MetalDocs`, the safest extraction targets are:

1. Postgres package shape
2. migration discipline
3. IAM repository shape
4. operational runbook style

These should be extracted as references, then rebuilt in `MetalShopping_Final` under the frozen architecture:

- `apps/server_core/internal/platform/db/postgres`
- `apps/server_core/internal/platform/auth`
- `apps/server_core/internal/platform/security`
- `apps/server_core/internal/platform/observability`
- `apps/server_core/internal/modules/iam`

## Concrete rejection targets

These should not be copied into the new codebase:

1. header-trusting auth middleware
2. repository defaulting to memory in the main server path
3. disabled-by-default rate limiting
4. `sslmode=disable` as an unqualified default
5. any runtime path that allows canonical behavior to disappear silently

## Migration rule

Any code brought from `MetalDocs` must satisfy all of the following before landing in `MetalShopping_Final`:

- aligned with `docs/SYSTEM_PRINCIPLES.md`
- aligned with `docs/SERVER_CORE_OPERATING_MODEL.md`
- aligned with `docs/OBSERVABILITY_AND_SECURITY_BASELINE.md`
- tenancy-ready where relevant
- contract-first where relevant
- no silent downgrade from canonical runtime behavior

## Exit criteria for deleting this document

Delete this file once all of the following are true:

- the reuse decisions have been executed
- accepted patterns are reimplemented in `MetalShopping_Final`
- rejected patterns are no longer under consideration
- implementation docs or ADRs already capture any durable decisions that remain useful

At that point this document becomes migration residue and should not stay in the repo.
