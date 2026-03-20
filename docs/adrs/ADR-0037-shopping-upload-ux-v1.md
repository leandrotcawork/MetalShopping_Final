# ADR-0037: Shopping Upload UX v1 (Desktop Picker First)

- Status: accepted
- Date: 2026-03-20

## Context

Legacy Shopping provides a user-oriented XLSX importer UX:

- clear dropzone
- file picker affordance (folder icon)
- minimal user-facing inputs

In the target repo, the contract currently uses `xlsxFilePath` (server-side/worker-readable path) and optional `xlsxScopeIdentifiers`. The current UI exposes those as raw text fields, which feels unlike the legacy workflow and invites user error.

The browser cannot provide a real absolute file path for security reasons, so legacy-style "path pickers" only work reliably in a desktop runtime.

## Decision

Shopping upload will be implemented as a capability-based UX:

- Desktop-first behavior: when a desktop runtime exists, the UI uses a platform file picker to set `xlsxFilePath` without manual typing.
- Web fallback behavior: the UI keeps a manual fallback input for `xlsxFilePath`, but it is not the primary workflow.
- Primary UI: the upload step is driven by dropzone + picker affordance (folder icon) instead of raw path typing.
- Scope identifiers (`xlsxScopeIdentifiers`) are not displayed by default and remain advanced-only for troubleshooting.

The Shopping API contract is not changed in this ADR.

## Contracts (touchpoints)

- `contracts/api/jsonschema/shopping_create_run_request_v1.schema.json`
- `contracts/api/openapi/shopping_v1.openapi.yaml`

## Implementation Checklist

- Replace the "backend path" UX with a legacy-like dropzone.
- Add a folder affordance (icon/emoji) consistent with legacy importer.
- Gate desktop picker behavior behind capability detection.
- Move `xlsxScopeIdentifiers` behind an advanced toggle with explicit caution text.

## Acceptance Evidence (for Status: accepted)

- Upload step does not show raw scope identifiers or backend path fields by default.
- Desktop-capable environment can pick XLSX and populates `xlsxFilePath` reliably.
- Web fallback still allows manual `xlsxFilePath` for local/dev without breaking contract validation.
- `npm.cmd run web:typecheck` passes.
- `npm.cmd run web:build` passes.

## Consequences

- The UX matches legacy intent while preserving the contract-first backend model.
- We avoid promising web upload semantics that the current architecture does not provide yet.
