# AGENTS

## Scope

This file applies to `scripts/`.

## Script rules

- scripts exist to automate repository workflows, not to become hidden product logic
- generation and validation scripts must derive from `contracts/`
- scripts should be idempotent when possible
- scripts should be composable and have stable command contracts

## Current phase

- planning-first
- scripts may be placeholders that define workflow contracts
- do not add implementation-heavy automation unless it is explicitly needed

