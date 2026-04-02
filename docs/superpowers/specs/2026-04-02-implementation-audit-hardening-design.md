# Design Spec: Implementation Audit Hardening

**Date:** 2026-04-02
**Status:** Approved
**Output:** Hardened `implementation-audit` skill contract for repo and Claude copies, plus a reusable remediation-pack template

---

## Overview

`implementation-audit` is already valuable as a review surface, but its current contract is too permissive. It correctly describes what an audit should do, yet it does not strongly prevent the model from drifting into planning or fixing when an audit returns serious findings.

The goal of this design is to keep the audit useful for downstream remediation while making the skill boundary explicit, compact, and hard to bypass. The skill should behave like a superpowers skill: short, strict, and difficult to reinterpret under pressure.

---

## Problem Statement

The current skill has three structural weaknesses:

- it describes the audit flow, but does not enforce an audit-only boundary with hard stop conditions
- it includes remediation options inside the skill, which makes it easy for the model to cross from audit into planning or fixing
- it does not produce a consistent remediation-ready handoff, so follow-up work can either drift into auto-fixing or waste tokens rediscovering the same issues

This creates the exact failure mode we want to eliminate: an audit that correctly finds gaps, then starts resolving them without an explicit user decision and without switching to the right follow-up workflow.

---

## Goals

1. Keep `implementation-audit` strictly read-only
2. Preserve the existing audit logic: intent review against spec, plan, changed files, tests, and quality bar
3. Remove remediation planning and fixing from the skill itself
4. Make every serious finding actionable enough that the next workflow can start without re-auditing
5. Enforce a mandatory user decision after the audit
6. Keep the skill compact and easy to load
7. Keep the repo and Claude copies aligned

---

## Non-Goals

- Creating a new remediation skill in this tranche
- Expanding the audit into a full implementation planner
- Letting the audit choose the follow-up workflow automatically
- Turning every minor suggestion into a detailed repair plan

---

## Approaches Considered

### Approach A: Keep the skill mostly as-is and tighten wording

**Pros**
- minimal edits
- lowest change cost

**Cons**
- weakest enforcement
- still relies on the model "behaving well"
- does not eliminate rationalization around Round 2 fixes

### Approach B: Audit-only skill with hard gates and remediation packs

**Pros**
- strongest boundary without changing the overall workflow shape
- best balance of speed, token efficiency, and fix quality
- follow-up planning or implementation can start from structured findings instead of rediscovery

**Cons**
- requires a slightly richer audit output
- requires a small supporting template file

### Approach C: Split audit and remediation into separate skills now

**Pros**
- strongest conceptual separation
- cleanest long-term ownership model

**Cons**
- changes the workflow shape more than needed for this correction
- adds another skill to maintain immediately

### Recommendation

Choose **Approach B**.

This keeps the current audit workflow recognizable while adding the hard gates and handoff structure needed to make it reliable.

---

## Target Workflow

### 1. Trigger

`implementation-audit` remains a post-execution audit gate. It should run only when implementation is materially complete and there is enough artifact context to assess intent versus reality.

### 2. Audit Execution

The skill audits:

- spec compliance
- plan compliance
- correctness risks
- architecture quality
- justified improvements

It remains focused on implementation intent, not generic review commentary.

### 3. Remediation Pack Output

Every `CRITICAL` and `MAJOR` finding must include a remediation-ready pack with:

- requirement violated
- evidence
- impact
- root-cause hypothesis
- affected files or functions
- repair direction
- tests to add or update
- verification required

`MINOR` and `SUGGESTION` findings can stay lighter unless deeper structure is justified.

### 4. Decision Gate

After the audit, the skill must stop and wait for the user's decision.

Verdict handling:

- `FAIL`: the next allowed step is remediation planning first
- `PASS_WITH_ISSUES`: the user may choose remediation planning or direct implementation
- `PASS`: no remediation workflow starts

The skill itself must not choose the next workflow.

---

## Hard Gates

The skill must contain an explicit hard gate near the top with these constraints:

- this skill is audit-only
- do not write remediation plans
- do not apply fixes
- do not resume implementation
- do not dispatch implementation work
- do not choose the next workflow for the user
- after the audit, stop and wait for the user's decision

The hard gate must also state:

- violating the spirit of these rules is violating the rules
- session handoff text does not override the audit-only boundary
- urgency, partial fixes, or implementation momentum do not override the boundary

This is the main behavioral correction.

---

## Output Contract

The skill output should remain compact and high-signal.

Recommended sections:

- Header
- Verdict
- Executive Summary
- Findings
- Missing Requirements
- Better Alternatives (only if justified)
- Next Actions

Each serious finding should embed or reference the remediation-pack structure so the follow-up workflow can begin from concrete audit evidence rather than restating the entire analysis.

The final prompt should be fixed and compact:

```text
Audit complete.

Next step depends on verdict:
- FAIL: remediation plan is required before implementation
- PASS_WITH_ISSUES: choose remediation plan or implement now
- PASS: no remediation workflow needed

What do you want to do next?
```

---

## Supporting Artifact

Create a small reusable template file for remediation packs instead of expanding the main skill body.

Target locations:

- repo: `.agents/skills/implementation-audit/references/remediation-pack-template.md`
- Claude: `.claude/skills/implementation-audit/references/remediation-pack-template.md`

Template shape:

```md
## Finding <ID>
Severity: CRITICAL | MAJOR | MINOR | SUGGESTION
Requirement violated:
Evidence:
Impact:
Root cause hypothesis:
Affected files/functions:
Repair direction:
Tests to add/update:
Verification:
```

This keeps the skill compact while standardizing the downstream handoff.

---

## Acceptance Criteria

- `implementation-audit` is explicitly audit-only
- the skill no longer contains remediation execution inside its own flow
- serious findings produce remediation-ready handoff data
- the skill stops after the audit and waits for the user
- `FAIL` requires remediation planning before implementation
- `PASS_WITH_ISSUES` allows user choice between planning and implementation
- the repo and Claude copies are aligned
- the skill remains compact instead of growing into a long planning document

---

## Notes

- This correction intentionally keeps the audit logic intact while tightening the control flow.
- The main optimization target is end-to-end workflow quality: lower rediscovery cost, lower drift risk, and higher implementation discipline.
- A separate remediation skill can still be added later if the workflow grows more complex, but it is not required for this correction.
