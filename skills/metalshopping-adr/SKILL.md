---
name: metalshopping-adr
description: Full ADR lifecycle for MetalShopping. An ADR is DONE only when documented with a runnable acceptance test, PROJECT_SOT updated, acceptance test passes, and git commit made. Use when freezing any architectural decision.
---

# MetalShopping ADR

## 4 stages — all required, in order

### Stage 1 — Write
File: `docs/adrs/ADR-XXXX-<kebab-title>.md`

Required sections:
- **Status:** Proposed
- **Date:** YYYY-MM-DD
- **Context:** why this decision was needed
- **Decision:** what was decided (precise, not vague)
- **Consequences:** trade-offs
- **Acceptance test:** concrete and runnable — not subjective

Acceptance test examples:
- ✅ `scripts/smoke_shopping_driver_suite_local.ps1` all suppliers PASS
- ✅ `go build ./...` passes after registering module in `composition_modules.go`
- ✅ `GET /api/v1/analytics/overview` returns real data in browser
- ❌ "implementation looks correct"
- ❌ "reviewed by developer"

### Stage 2 — Update SOT
Update `docs/PROJECT_SOT.md` to reflect the accepted decision.

### Stage 3 — Verify
Run the acceptance test from Stage 1.
If fails → fix implementation → re-run. Do not proceed until green.

### Stage 4 — Commit and close
```bash
git commit -m "docs(adr): ADR-XXXX <short title> — verified and closed"
```
Update ADR Status: Proposed → Accepted.
An ADR with Status: Proposed is not done.

## Workflow
1. Read `docs/PROJECT_SOT.md` + relevant existing ADRs
2. Write ADR with acceptance test (Stage 1)
3. Update PROJECT_SOT (Stage 2)
4. Implement and run acceptance test (Stage 3)
5. Commit (Stage 4)

## References
- `references/adr-checklist.md`
