---
name: metalshopping-event-contracts
description: Create or review MetalShopping versioned event contracts under contracts/events/v1 using the repo event template, naming rules, and async integration discipline. Use when implementing outbox events or reviewing event envelope consistency.
---

# MetalShopping Event Contracts

## Workflow
1. Read `docs/CONTRACT_CONVENTIONS.md` + `contracts/events/v1/_template.event.json`
2. Confirm bounded context, event name, target filename
3. Start from template — do not draft event JSON from scratch
4. Reference payload schemas from `contracts/api/jsonschema/` when reusable
5. Finish with `references/event-checklist.md`

## Rules
- File: `contracts/events/v1/<domain>_<event_name>.v1.json`
- Semantic name stable after publication — do not rename published events
- Prefer additive compatible evolution over semantic mutation
- Keep producer, trigger, and semantic meaning explicit
- Idempotency key in implementation: `"event_name:aggregate_id"`

## References
- Template: `contracts/events/v1/_template.event.json`
- Checklist: `references/event-checklist.md`
