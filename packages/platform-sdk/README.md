# Platform SDK

`packages/platform-sdk` (workspace package `@metalshopping/sdk-runtime`) owns the thin browser runtime and stable frontend SDK facade.

- It composes clients generated under `packages/generated/sdk_ts/generated/*`.
- It centralizes browser transport defaults (`credentials`, CSRF, trace header, and error mapping).
- It is authoring-owned code, not generator output.

This keeps `scripts/generate_contract_artifacts.ps1` focused on contract artifact orchestration only.
