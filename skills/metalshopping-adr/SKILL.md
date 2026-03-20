---
name: metalshopping-adr
description: Full ADR lifecycle for MetalShopping. An ADR is DONE only when the decision is documented with an acceptance test, PROJECT_SOT is updated, the test passes, and a git commit is made. Use when freezing any architectural decision.
---

# MetalShopping ADR

## The 4 stages (all required)

**Stage 1 — Write**
File: `docs/adrs/ADR-XXXX-<kebab-title>.md`
Required sections: Status (Proposed), Date, Context, Decision, Consequences, Acceptance test

Acceptance test must be concrete and runnable — not "looks correct":
- Good: `smoke_shopping_driver_suite_local.ps1` all suppliers PASS
- Good: `go build ./...` passes after wiring in composition
- Bad: "reviewed by developer" / "implementation looks correct"

**Stage 2 — Update SOT**
Update `docs/PROJECT_SOT.md` to reflect the accepted decision.

**Stage 3 — Verify**
Run the acceptance test from Stage 1. Fix if fails. Re-run. Do not proceed until green.

**Stage 4 — Commit**
```
git commit -m "docs(adr): ADR-XXXX <short title> — verified and closed"
```
Update ADR Status: Proposed → Accepted.

## Workflow
1. Read `docs/PROJECT_SOT.md` + relevant existing ADRs
2. Write ADR with acceptance test (Stage 1)
3. Update PROJECT_SOT (Stage 2)
4. Implement and run test (Stage 3)
5. Commit (Stage 4)

## References
- `references/adr-checklist.md`
