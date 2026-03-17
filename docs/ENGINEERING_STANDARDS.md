# Engineering Standards

## Purpose

Define the quality bar for MetalShopping so the platform evolves with the discipline expected from a professional, scalable product.

## Core principles

- explicit boundaries over convenience shortcuts
- contract-first where integration matters
- observability by default
- security by default
- repeatable automation over tribal knowledge
- additive evolution over uncontrolled rewrites

## Architecture standards

- `server_core` owns canonical synchronous behavior
- workers stay specialized and async-first
- frontend remains thin
- domain and platform concerns remain separate
- governance decisions must be explainable and auditable

## Contract standards

- all API, event, and governance contracts live in `contracts/`
- generated SDKs and generated types are downstream only
- versioning must be explicit
- compatibility rules must be documented before publishing

## Data standards

- Postgres is canonical
- multitenant data must be tenancy-aware
- large temporal data must have partition and retention strategy
- auditability is a first-class concern

## Security standards

- central auth and authz
- secure defaults for sensitive operations
- least privilege mindset
- explicit operational surfaces in `admin_console`

## Observability standards

- logs, metrics, and traces are part of platform design
- async flows must remain traceable across boundaries
- operational health should be observable without local tribal knowledge

## Delivery standards

- planning docs are updated when a platform rule changes
- architecture changes need ADRs
- generated artifacts must be reproducible
- workflows must be scriptable

## Review standards

- favor small, reviewable increments
- prefer schema and contract review before implementation review
- reject parallel sources of truth
- reject hidden coupling between core and workers

