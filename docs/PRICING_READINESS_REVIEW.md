# Pricing Readiness Review

## Purpose

Record whether the current canonical `catalog` is strong enough to support the first `pricing` slice without reopening product ownership or adding convenience-driven coupling.

This document is the gate between foundation hardening and pricing expansion.

## Verdict

`catalog` is ready for the first `pricing` slice.

This verdict applies only to the first pricing implementation wave described in `docs/PRICING_CANONICAL_MODEL.md`.

## What `pricing` needs from `catalog`

The first pricing slice requires a stable product reference with:

- canonical `product_id`
- canonical `tenant_id`
- canonical `sku`
- product naming for operational and UI surfaces
- stable product active status
- stable taxonomy linkage for future pricing segmentation
- stable identifiers for integrations and matching

The current `catalog` already provides these foundations.

## What is already in place

### Canonical product ownership

`catalog` owns canonical product identity and `pricing` does not redefine product truth.

The current state already includes:

- `catalog_products`
- `catalog_product_identifiers`
- `catalog_taxonomy_nodes`
- `catalog_taxonomy_level_defs`
- `description`
- `brand_name`
- `stock_profile_code`

This is sufficient for the first pricing slice because price can attach to canonical product identity without ambiguity.

### Multi-tenant safety

The current catalog stack already proves:

- `tenant_id` on canonical product state
- runtime tenant context
- RLS-backed catalog tables
- tenant-aware HTTP and persistence paths

This satisfies the minimum safety bar for a pricing module that must never cross tenant boundaries.

### Runtime governance

The platform now already proves governance affecting live runtime behavior through:

- feature flags
- thresholds
- policies

This is important because `pricing` should start governed, not grow governance later as a refactor.

### Async boundary

The platform now already proves:

- versioned event contract
- transactional outbox append
- dispatcher publication path

That means the first pricing mutation can also be required to emit a real event from day one.

## What does not block pricing

The following are intentionally not blockers for the first pricing slice:

- full master-data ingest from legacy `products_master`
- no-null-overwrite import semantics in production flows
- richer product dimensions beyond the current canonical product shape
- broker-backed worker delivery beyond the current outbox publication path
- admin-console mutation surface for governance values

These matter, but they are not required to open the first well-modeled pricing slice.

## What pricing must not push back into catalog

The first pricing slice must not cause `catalog` to absorb:

- internal price values
- cost basis values
- margin rules
- pricing validity windows
- price origin metadata
- pricing history

Those belong to `pricing`.

## Required constraints before pricing implementation

The first pricing slice must satisfy all of the following:

- `pricing` references canonical `product_id`
- `pricing` does not duplicate product master fields
- `pricing` is tenant-aware from the first migration
- `pricing` uses runtime governance for at least one threshold and one policy
- `pricing` publishes a versioned event through outbox
- `pricing` keeps write ownership explicit and does not leak into analytics or integration workers

## Conclusion

The current `catalog` is no longer a minimal blocker for pricing.

It is now a sufficient canonical foundation for:

- the first internal price write path
- cost and margin-aware pricing rules
- governed pricing decisions
- future pricing events and downstream consumption

The next correct move is to freeze the first canonical pricing model and contract scope, then implement the first slice without reopening `catalog`.
