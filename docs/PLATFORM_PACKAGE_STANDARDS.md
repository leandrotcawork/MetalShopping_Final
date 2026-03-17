# Platform Package Standards

## Purpose

Define how packages under `apps/server_core/internal/platform/` should be created and kept professional, reusable, and bounded.

## Platform package goal

A platform package provides a reusable runtime capability to the system without taking ownership of business semantics.

## What belongs in platform packages

- authentication runtime
- authorization runtime support
- tenancy runtime support
- governance runtime support
- persistence infrastructure
- messaging infrastructure
- jobs and scheduling support
- cache and file abstractions
- delivery channels
- observability
- security
- audit infrastructure

## What does not belong in platform packages

- business workflows
- domain invariants
- business-specific thresholds or policies
- cross-cutting dumping ground utilities with unclear ownership

## Package design rules

- keep package purpose narrow and explicit
- expose stable interfaces where reuse matters
- avoid leaking infrastructure details into module domain code
- prefer composition over large platform super-packages
- preserve clear naming that reflects runtime capability

## Suggested internal shape

Platform packages do not need one universal structure, but they should usually separate:

- public package entrypoints
- internal helpers
- concrete infrastructure clients
- config or runtime options
- tests later when implementation begins

## Platform anti-patterns

- business logic hidden in infrastructure code
- generic helper packages with no clear boundary
- one package owning unrelated operational concerns
- platform APIs that force modules to know low-level infrastructure details

