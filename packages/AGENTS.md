# AGENTS

## Scope

This file applies to `packages/`.

## Package rules

- `packages/ui` is for reusable UI primitives and shared visual assets
- `packages/generated/*` contains generated artifacts only
- generated outputs must originate from `contracts/`
- do not create manual fallback types that compete with generated artifacts

## Planning focus

- define how generation works before adding generated files
- define ownership and publish targets for TS and Python consumers

