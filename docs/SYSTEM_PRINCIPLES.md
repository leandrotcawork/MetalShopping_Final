# System Principles

## Purpose

Define how MetalShopping should behave as a professional, scalable platform before we derive templates, module standards, or implementation workflows.

## Product principles

- MetalShopping is an enterprise decision platform, not a traditional e-commerce app.
- The platform exists to support commercial strategy, pricing, market intelligence, procurement, CRM, automation, and analytics in one coherent system.
- Product design must prefer explicit business semantics over generic technical abstractions.

## Platform principles

- monorepo over disconnected repos
- server-first over client-sovereign logic
- canonical state over duplicated local truth
- contract-first over implicit integration
- governance-first over hardcoded runtime behavior
- operational clarity over hidden convenience

## Ownership principles

- `apps/server_core` owns canonical synchronous behavior.
- workers do not own product truth.
- `contracts/` owns interface truth.
- `bootstrap/` owns defaults, not live state.
- `packages/generated/*` owns generated outputs, not source semantics.

## Scaling principles

- optimize for evolutionary scale, not premature microservice fragmentation
- preserve modular boundaries so scaling options remain open later
- prefer additive evolution over rewrites
- separate compute scale from canonical transaction scale

## Reliability principles

- every important system behavior should be explainable
- every important mutation should be observable
- every important integration should be contract-governed
- every important operational path should be auditable

