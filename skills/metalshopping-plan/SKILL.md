---
name: metalshopping-plan
description: Plan any MetalShopping feature before writing code. Mandatory for tasks with 3+ steps or any architecture decision. Writes tasks/todo.md with ordered tasks, contracts, skills, acceptance tests, and commit messages per task. Nothing is implemented without an approved plan.
---

# MetalShopping Plan

## Workflow
1. Read `ARCHITECTURE.md` + `docs/PROJECT_SOT.md`
2. Classify: layers, module type, events, ADR needed
3. Write `tasks/todo.md` — use `references/todo-template.md`
4. Validate: task order correct? tests runnable, not subjective?
5. Wait for approval. Then begin T1.

## Module type decides Go structure
- **read-only** → Reader + postgres reader — ref: `modules/home/`
- **write+events** → Writer + AppendInTx + events/ — ref: `modules/shopping/`
- **CRUD+governance** → domain + Repository + gov adapters — ref: `modules/catalog/`
- **scraping** → Python worker + Go reader — ref: `integration_worker/`

## Frozen task order
T1 contract → T2 Go → T3 worker (optional) → T4 SDK → T5 frontend → T6 ADR (optional)

## References
- `references/todo-template.md`
