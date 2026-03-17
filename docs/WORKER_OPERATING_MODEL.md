# Worker Operating Model

## Purpose

Define how workers should operate without violating canonical state ownership or platform boundaries.

## Worker role

Workers exist for:

- compute
- ingestion
- automation
- async delivery
- external integration execution

## Worker rules

- workers do not own canonical product truth
- workers consume contracts and publish governed outputs
- workers should be retry-safe where async execution is expected
- workers should prefer idempotent behavior and explicit correlation metadata
- worker logic should remain replaceable without redefining platform truth

## Worker categories

- `analytics_worker`: heavy analytics, scoring, explainability, projections
- `integration_worker`: connectors, crawlers, imports, exports, normalization
- `automation_worker`: orchestration, triggers, actions, campaigns
- `notifications_worker`: channel delivery and routing

## Worker anti-patterns

- direct ownership of core transactional semantics
- arbitrary database access patterns with no contract discipline
- implicit event semantics
- hidden coupling that makes core request paths depend on worker completion

