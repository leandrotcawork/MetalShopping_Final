# ADR-0013: OpenAPI Generator SDK Pipeline

- Status: accepted
- Date: 2026-03-18

## Context

The repository already froze thin clients and generated SDKs as a direction, but the current TypeScript SDK flow still emits TypeScript manually from `scripts/generate_contract_artifacts.ps1`. That approach helped bootstrap the first web slices, but it is too brittle for a growing platform with more APIs, more auth/runtime behavior, and more frontend consumers.

We need a generation path that is:

- contract-first
- reproducible across machines
- independent from local Java drift
- strong enough for browser clients with centralized auth/session behavior

## Decision

- `contracts/api/openapi/*.yaml` remain the only source of truth for HTTP contracts
- `packages/generated/sdk_ts` remains output-only
- the TypeScript SDK flow now uses the official OpenAPI Generator `typescript-fetch` generator
- generation runs through the official Docker image `openapitools/openapi-generator-cli`
- `scripts/generate_contract_artifacts.ps1` stays as the repo orchestration entrypoint, but no longer emits SDK transport code manually
- the repo may keep a thin generated facade over the official generated clients to preserve canonical query shapes, centralized browser runtime composition, and stable frontend consumption

## Consequences

- the SDK pipeline is now tied to an official, externally validated generator instead of handwritten TS emission
- local Java version differences no longer define whether SDK generation works
- frontend runtime concerns stay centralized and thin
- future front+back integration work must start from contracts, generation, and the shared runtime pattern rather than page-local HTTP code or feature-local clients
- the current PowerShell script is now orchestration only; it is not the long-term justification for handwritten SDK logic
- contract hygiene still needs a follow-up pass because some OpenAPI and schema identifiers emit noisy remote-resolution warnings during generation even though the pipeline succeeds

## Follow-up

- normalize contract identifiers and related schema references so the OpenAPI Generator run becomes silent as well as reproducible
- re-evaluate whether the repo should keep PowerShell as the top-level orchestration entrypoint or move generation orchestration to a more portable task runner once the broader platform toolchain is frozen
