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

## Lesson Template

```text
## Lesson N -- <title>
Date: YYYY-MM-DD | Trigger: <correction | review | build failure>
Wrong:   <exact code or decision>
Correct: <exact code or decision>
Rule:    <one sentence>
Layer:   <Go adapter | handler | worker | frontend | process | docs>
```

