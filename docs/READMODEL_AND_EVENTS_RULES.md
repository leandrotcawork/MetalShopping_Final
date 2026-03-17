# Readmodel And Events Rules

## Purpose

Define when and how `readmodel/` and `events/` should be used in the core and across async boundaries.

## Read model principles

- canonical writes remain in the owning module
- read models exist to optimize consumption, not to redefine truth
- read models may be derived, materialized, cached, or query-focused
- read models must stay explainable from owned state and governed async outputs

## When to use `readmodel/`

Use `readmodel/` when:

- the UI needs query-friendly shapes
- the API needs aggregated views
- analytics outputs need serving-friendly representations
- temporal or ranking views are derived from canonical state and governed outputs

Do not use `readmodel/` to:

- bypass domain ownership
- store ad hoc business logic that belongs in `domain/`
- invent a second canonical model

## Event principles

- events express important domain or platform facts
- events are versioned contracts
- events drive async decoupling
- events must remain semantically stable after publication

## When to publish events

Publish events when:

- a relevant mutation changes business-significant state
- another bounded capability needs async reaction
- downstream workers need to compute, ingest, automate, or deliver

Do not publish events for:

- meaningless internal noise
- transport-level convenience only
- transient local steps with no integration value

## Core rules for events

- the core publishes events from owned mutation flows
- event publication should align with outbox discipline
- normal request completion must not depend on async consumer completion
- event contracts live in `contracts/events/`, not inside modules

## Consumer rules for events

- workers consume events through governed async infrastructure
- consumers should be idempotent when retries are possible
- correlation and trace metadata should survive async boundaries

## Read model and event relationship

- events may trigger read model updates
- read models may consume canonical state plus materialized async outputs
- read models should remain consistent with the chosen eventual consistency model

## Anti-patterns

- using events as hidden RPC
- building canonical truth from unmanaged read models
- letting read models drift semantically from owned state
- mixing transport DTOs and event semantics carelessly

