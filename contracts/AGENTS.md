# AGENTS

## Scope

This file applies to `contracts/`.

## Contract rules

- `contracts/` is the single source of truth for API, event, and governance schemas
- generated SDKs and generated types must derive from contracts
- manual parallel shared type systems are not allowed
- event versioning starts in `events/v1`

## Planning focus

- define naming conventions
- define ownership conventions
- define versioning conventions
- define generation targets for TS and Python SDKs

