# Next Execution Decision

## Decision

The next implementation area should be `pricing`.

## Why

- the platform foundation is now strong enough
- the canonical `catalog` is now ready to support it
- legacy semantics show pricing is central to long-term value
- analytics, procurement, market intelligence, and CRM all benefit from a strong internal pricing owner

## Constraints

This decision is valid only if implementation follows:

- `docs/PRICING_READINESS_REVIEW.md`
- `docs/PRICING_CANONICAL_MODEL.md`
- `docs/PRICING_IMPLEMENTATION_PLAN.md`

## Explicit rejection

Do not jump next to:

- `inventory`
- `market_intelligence`
- `crm`
- `procurement`

until the first pricing slice is implemented, validated, and published through contracts and outbox.
