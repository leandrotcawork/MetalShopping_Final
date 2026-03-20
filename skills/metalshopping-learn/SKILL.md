---
name: metalshopping-learn
description: Capture lessons after any correction, test failure, or review finding. Writes to tasks/lessons.md immediately. Prevents the same mistake from recurring. Use after ANY user correction or Critical/High review finding.
---

# MetalShopping Learn

## Trigger this skill after
- User corrects any pattern, approach, or output
- Review produces Critical or High finding
- Build fails due to wrong pattern
- Same mistake appears twice

## Workflow
1. Identify: what was wrong, what is correct, which layer
2. Write to `tasks/lessons.md` immediately:

```
## Lesson N — <title>
Date: YYYY-MM-DD | Trigger: <correction | review | build failure>

Wrong: <exact code or decision>
Correct: <exact code or decision>
Rule: <one sentence — specific enough to prevent recurrence>
Layer: <Go adapter | handler | worker | frontend | all>
```

3. Update `tasks/todo.md` — reset task to [ ] if needs redo
4. Apply the lesson immediately in the current session

## At session start
Read `tasks/lessons.md` before any code.
If a lesson applies to the current task: apply it proactively.

## Canonical lessons (always active — copy to tasks/lessons.md on project start)
See `references/canonical-lessons.md`
