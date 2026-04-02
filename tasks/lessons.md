# Lessons Learned

Use this file only for recurring engineering corrections that would cause bugs, review failures, or structural regressions if repeated.

## Rules

- Write a lesson immediately after any meaningful correction.
- Do not log one-off UI tweaks or cosmetic feedback.
- Keep each lesson concrete and reusable.

## Lesson 1 -- Claim methods must return claim status
Date: 2026-04-02 | Trigger: implementation
Wrong:   `ClaimForPromotion` returned `nil` when the row was already claimed, so the consumer could not tell whether it owned the promotion lock.
Correct: `ClaimForPromotion` returns an explicit boolean claim result and the consumer exits early when the claim is not acquired.
Rule:    Reconciliation claim APIs must make ownership visible to the caller instead of silently succeeding on conflicts.
Layer:   Go adapter

## Lesson 2 -- Failure payloads must snapshot the post-claim state
Date: 2026-04-02 | Trigger: implementation
Wrong:   Promotion failure warnings reused the pre-claim reconciliation struct, so the payload still said `pending` after the consumer had already claimed the row.
Correct: Copy the reconciliation result after claim and force `promotion_status=promoting` before building failure or review payloads.
Rule:    Any failure context emitted after a state transition must reflect the post-transition lifecycle state, not the pre-transition read model.
Layer:   Go handler

## Lesson 3 -- Replay-safe promotion must go through the catalog boundary
Date: 2026-04-02 | Trigger: implementation
Wrong:   ERP promotion wrote directly as if the canonical product was always absent, so a replay hit the catalog unique constraint and failed hard.
Correct: Use a catalog-owned create-or-get path that handles `tenant_id + sku` conflicts and returns the existing canonical product ID without duplicate side effects.
Rule:    Retry-safe promotion flows should be idempotent at the owning module boundary, not by bypassing the boundary with ad hoc SQL.
Layer:   Go adapter

## Lesson Template

```text
## Lesson N -- <title>
Date: YYYY-MM-DD | Trigger: <correction | review | build failure>
Wrong:   <exact code or decision>
Correct: <exact code or decision>
Rule:    <one sentence>
Layer:   <Go adapter | handler | worker | frontend | process | docs>
```

