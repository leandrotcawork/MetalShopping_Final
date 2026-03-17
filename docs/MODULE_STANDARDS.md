# Module Standards

## Purpose

Define the required shape and responsibilities of business modules inside `apps/server_core/internal/modules/`.

## Module goal

Each module represents a bounded business capability with explicit ownership, stable semantics, and clear dependencies.

## Mandatory structure

Every module must use:

```text
domain/
application/
ports/
adapters/
transport/
events/
readmodel/
```

## Directory responsibilities

### `domain/`

Owns:

- entities
- value objects
- invariants
- business rules
- domain errors

Rules:

- no transport concerns
- no infrastructure details
- no framework-coupled logic

### `application/`

Owns:

- use cases
- commands
- queries
- orchestration of domain behavior

Rules:

- coordinates domain logic
- depends on ports, not concrete infrastructure
- does not become a second domain layer

### `ports/`

Owns:

- module-facing interfaces
- required dependencies on other concerns
- contracts for infrastructure interactions

Rules:

- define what the module needs, not how it is implemented
- keep interface boundaries explicit and minimal

### `adapters/`

Owns:

- concrete implementations of ports
- database adapters
- broker adapters
- cache adapters
- external client adapters

Rules:

- implementation details stay here
- avoid leaking infrastructure assumptions back into `domain/`

### `transport/`

Owns:

- HTTP handlers
- gRPC handlers if introduced later
- DTO mapping
- request/response binding

Rules:

- transport is an interface layer only
- transport should not become application logic

### `events/`

Owns:

- event publishing intent
- event consumption mapping where the module reacts to platform events
- event-to-domain translation relevant to the module

Rules:

- event contracts still live in `contracts/events/`
- event semantics must remain business-explicit

### `readmodel/`

Owns:

- read projections
- query-oriented representations
- materialization logic relevant to the module

Rules:

- read models serve consumption needs
- read models do not replace canonical write ownership

## Module dependency rules

- modules may depend on `internal/platform/`
- modules may use `internal/shared/` only for small neutral utilities
- domain logic should not directly depend on another module's transport or adapters
- cross-module interaction should stay explicit and limited

## Module anti-patterns

- generic dumping grounds
- hidden cross-module coupling
- transport-driven domain design
- infrastructure code in `domain/`
- module logic spread across platform folders

