# ADR-0038: Shopping Supplier Selection Cards v1

- Status: draft
- Date: 2026-03-20

## Context

Legacy Shopping uses a supplier card grid for selection, which improves scanning, density, and the "operational control" feeling of the workflow. The target Shopping page currently uses checkbox rows, which is functionally correct but visually divergent and less scalable as suppliers grow.

## Decision

Supplier selection in Step 2 will converge to the legacy card grid UX:

- suppliers render as selectable cards with clear selected state
- a "desmarcar todos" action is present and obvious
- supplier metadata (execution kind, online/offline hint if available) is shown with the same density language as legacy

No contract changes are introduced in this ADR. The cards are a rendering of the `shopping bootstrap` suppliers list.

## Contracts (touchpoints)

- `contracts/api/jsonschema/shopping_bootstrap_v1.schema.json`
- `contracts/api/openapi/shopping_v1.openapi.yaml`

## Implementation Checklist

- Replace checkbox list with card grid rendering.
- Add "desmarcar todos" and keep selection state deterministic.
- Keep selection state stored in page/workflow state only.
- Promote the card primitive to `packages/ui` only if it is reused outside Shopping.

## Acceptance Evidence (for Status: accepted)

- Supplier selection UI matches the legacy card grid intent and density.
- Selection supports select/deselect and "clear all" without regressions in run request payload.
- `web:typecheck` and `web:build` pass.

## Consequences

- Suppliers remain easy to scan and operate on as the supplier list grows.

