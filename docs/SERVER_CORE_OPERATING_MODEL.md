# Server Core Operating Model

## Purpose

Define how `apps/server_core` should operate as the canonical center of the platform.

## Core role

`server_core` is the modular monolith that owns:

- auth
- authz
- tenancy runtime
- governance runtime
- canonical transactional state
- synchronous serving
- public API behavior
- read serving for clients

## Core rules

- normal requests must be answerable without synchronous worker dependence
- canonical writes happen in the core
- business invariants are enforced in the core
- public contract implementation starts from `contracts/`
- outbox publication happens from core-owned mutation flows

## Core data model behavior

- Postgres is the canonical write model
- some read models may live in Postgres
- materialized read models may be served through the core
- historical data remains owned by the domain that produced it

## Core service behavior

- core should optimize for predictability, auditability, and correctness
- core should expose administrative and governance-safe operational surfaces
- core should not accumulate compute-heavy asynchronous responsibilities that belong in workers

## Core anti-patterns

- calling workers synchronously for normal request completion
- storing canonical product state in clients
- scattering runtime policies across modules
- using the core as a generic passthrough without owned semantics

