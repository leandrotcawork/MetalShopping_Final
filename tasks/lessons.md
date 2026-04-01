# Lessons Learned

Use this file only for recurring engineering corrections that would cause bugs, review failures, or structural regressions if repeated.

## Rules

- Write a lesson immediately after any meaningful correction.
- Do not log one-off UI tweaks or cosmetic feedback.
- Keep each lesson concrete and reusable.

## Lesson Template

```text
## Lesson N — <title>
Date: YYYY-MM-DD | Trigger: <correction | review | build failure>
Wrong:   <exact code or decision>
Correct: <exact code or decision>
Rule:    <one sentence>
Layer:   <Go adapter | handler | worker | frontend | process | docs>
```
