---
name: metalshopping-event-contracts
description: Create or review MetalShopping versioned event contracts under `contracts/events/v1` using the repo event template, naming rules, metadata, payload schema references, compatibility policy, and async integration discipline. Use when implementing or revising domain events such as `pricing_price_changed.v1.json`, standardizing event envelopes, or checking event contracts against the repo SoT and versioning rules.
---

# MetalShopping Event Contracts

## Overview

Use this skill to create or review MetalShopping event contracts with the repository's versioned async integration standards. Keep the work anchored to the repo template and event conventions instead of inventing per-event structure.

## Workflow

1. Read only the minimum repo context:
   `docs/PROJECT_SOT.md`
   `docs/CONTRACT_CONVENTIONS.md`
   `contracts/events/v1/_template.event.json`
2. Confirm the bounded context, event name, and target filename under `contracts/events/v1/`.
3. Start from the template structure instead of drafting event JSON ad hoc.
4. Point payloads to shared schemas in `contracts/api/jsonschema/` whenever a standalone schema is appropriate.
5. Keep versioning explicit in both filename and event metadata.
6. Finish with the review checklist in `references/event-checklist.md`.

## Contract rules

- Use versioned event files under `contracts/events/v1/` first.
- Use stable semantic names such as `pricing_price_changed.v1.json`.
- Keep event meaning stable after publication.
- Prefer additive compatible evolution over semantic mutation.
- Treat events as integration contracts, not internal shortcuts.
- Keep envelopes explicit and traceable across async boundaries.

## Output expectations

When creating or updating an event contract:

- preserve the repository naming convention
- preserve explicit versioning
- keep producer, trigger, and semantic meaning explicit
- use schema references where payloads need validation or reuse
- note any required companion schema files in `contracts/api/jsonschema/`

## References

- For the exact repo workflow and file touchpoints, read `references/repo-event-flow.md`.
- For the final review pass, read `references/event-checklist.md`.

