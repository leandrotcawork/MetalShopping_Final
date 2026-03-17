# Engineering Principles

## Purpose

Define the engineering behavior expected across all layers of MetalShopping.

## General principles

- clarity over cleverness
- explicit boundaries over hidden coupling
- repeatable workflows over tribal memory
- reviewable increments over oversized rewrites
- generated artifacts over parallel manual drift
- operational safety over local convenience

## Design principles

- model business meaning explicitly
- keep platform concerns separate from domain concerns
- prefer stable contracts to ad hoc coordination
- prefer idempotent async behavior where retries are expected
- preserve the ability to reason about failure modes

## Evolution principles

- breaking changes require explicit versioning strategy
- architecture changes require a written decision trail
- contract changes should be reviewed before implementation spreads
- new patterns must prove they reduce complexity, not add novelty

## Delivery principles

- scripts define repeatable workflows
- generated outputs must be reproducible
- docs should record the rule, not just the discussion
- quality gates should be introduced before scale forces them

