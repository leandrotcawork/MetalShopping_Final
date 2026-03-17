# ADR-0006: Versioned Async Integration

- Status: accepted
- Date: 2026-03-17

## Context

Workers are required for analytics, integrations, automation, and delivery, but the core must stay responsive and in control of canonical synchronous behavior.

## Decision

- relevant mutations publish events
- events are versioned under `contracts/events/v1` first
- outbox and inbox patterns are part of the platform foundation
- workers consume through broker or queue semantics
- workers do not become direct synchronous dependencies of normal core requests

## Consequences

- the system stays decoupled at the async boundary
- event contracts become long-lived platform assets
- versioning discipline is required from the start

