# SDK Boundary

## Purpose

Define explicit ownership and import boundaries between generated artifacts and authored frontend runtime code.

## Ownership model

### `contracts/*`

- source of truth for API and schema semantics

### `packages/generated/*`

- output-only generated artifacts
- no authored transport/runtime behavior
- consumed through `@metalshopping/sdk-types`

### `packages/platform-sdk/*` (current runtime package)

- authored frontend runtime/facade
- composes generated clients and generated types
- owns browser transport policy (`credentials`, CSRF header, trace header, error normalization)
- consumed through `@metalshopping/sdk-runtime`

## Required boundaries

- feature packages consume SDK through stable package names only
- authored runtime code must not depend on deep relative paths into generated output internals
- generated outputs must not include manual runtime behavior

## Anti-patterns

- editing generated files manually as if they were source
- page-level or feature-level custom API clients
- duplicated DTO/type systems parallel to contracts and generated types
- double-cast (`as unknown as`) in auth/session runtime path

## Migration target

This repository currently uses `packages/platform-sdk` as authored runtime. The next hardening step must keep the same conceptual split:

- generated output remains output-only
- runtime remains authored and isolated
- package naming may evolve, but boundary semantics cannot regress
