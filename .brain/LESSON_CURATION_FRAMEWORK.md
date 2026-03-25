---
name: Lesson Curation Framework
description: Criteria for what constitutes an architectural lesson vs. a temporary note
created_at: 2026-03-24T21:00:00Z
---

# Lesson Curation Framework

**Purpose:** Define the bar for what should be permanently archived as a "lesson" in the Brain vs. what is a one-off fix or cosmetic tweak.

## Lesson Criteria (MUST satisfy ALL three)

A lesson is architectural and worth archiving only if it meets **ALL** of these criteria:

### 1. Cross-Domain Applicability
**Definition:** The lesson applies to multiple features, modules, or domains — not just a single surface or one-off scenario.

**Examples of KEEP:**
- "Tenant-safe DB access is mandatory" — applies to every Go adapter query
- "Frontend data flow must use platform SDK contracts" — applies to every data-fetching component
- "Outbox must be atomic with writes" — applies to any domain event emission

**Examples of REMOVE:**
- "Hover parity requires selector specificity on label elements" — only applies to legacy migration table headers
- "Workspace root must provide token fallbacks" — only applies to root workspace layout
- "First-fold workspace KPIs must be rendered in ProductHero" — only applies to ProductHero component

### 2. Prevents Repeated Mistakes
**Definition:** The lesson captures a pattern that has caused failures in the past OR is a critical architectural constraint that, if violated, causes system-wide failures.

**Examples of KEEP:**
- "Handlers must fail fast on auth and tenancy" — violating this causes security/data leaks
- "Generated artifacts are read-only outputs" — violating this breaks the build system
- "Completion requires validation + commit" — violating this leaves work in limbo

**Examples of REMOVE:**
- "Preserve UTF-8 when patching legacy-copied frontend files" — one-time encoding quirk during legacy migration
- "Mock semantics must match UI contract keys" — testing implementation detail
- "Remove ts-nocheck with minimal explicit callback typing" — TypeScript syntax cleanup

### 3. Is Not Cosmetic or Implementation-Specific
**Definition:** The lesson describes a **structural pattern or boundary** — not styling, layout, naming, or implementation detail.

**Examples of KEEP:**
- "Observability is part of the contract" — structural (logging/tracing at API boundary)
- "Reuse design system before adding UI primitives" — structural (component architecture decision)

**Examples of REMOVE:**
- "Table header hover must bind to explicit label node" — CSS selector choice (cosmetic)
- "Portaled UI must redeclare local CSS tokens" — styling implementation (cosmetic)
- "Workspace top bars that belong to the shell must be rendered as shell strips" — layout cosmetic
- "Similar surfaces should share the same base fill" — color/styling choice (cosmetic)

---

## Gray Zone: Operational Procedures vs. Architectural

Some lessons describe **operational procedures** that don't clearly fit "architectural." These are conditionally kept if they prevent widely-repeated mistakes:

**KEEP (Operational but prevents repeated failure):**
- "Legacy migration follows parity-first sequencing" — prevents functionality regression
- "Completion requires validation + commit" — prevents half-finished work
- "tasks/todo.md edits must be block-scoped" — prevents merge conflicts
- "Keep this file high-signal" — prevents noise in lessons

**REMOVE (Operational but too narrow or specific):**
- "Second-pass parity must diff against legacy snapshot" — only for legacy migration
- "Do not downgrade legacy charts in parity phase" — only for legacy migration
- "Hero KPI order must be explicit, not payload-driven" — only for ProductHero

---

## Classification Results (36 Lessons)

| # | Title | Category | Verdict | Reason |
|---|-------|----------|---------|--------|
| 0001 | Tenant-safe DB access is mandatory | Architectural | **KEEP** | Cross-domain (all Go), prevents data leaks |
| 0002 | Handlers must fail fast on auth and tenancy | Architectural | **KEEP** | Cross-domain (all handlers), security critical |
| 0003 | Outbox must be atomic with writes | Architectural | **KEEP** | Cross-domain (any event emission), prevents data loss |
| 0004 | Worker writes require tenant context and idempotency | Architectural | **KEEP** | Cross-domain (all workers), prevents data loss |
| 0005 | Frontend data flow must use platform SDK contracts | Architectural | **KEEP** | Cross-domain (all data fetch), ensures type safety |
| 0006 | Reuse design system before adding UI primitives | Architectural | **KEEP** | Structural (component architecture) |
| 0007 | Generated artifacts are read-only outputs | Architectural | **KEEP** | Prevents build system breakage |
| 0008 | Completion requires validation + commit | Operational | **KEEP** | Prevents unvalidated work from being lost |
| 0009 | tasks/todo.md edits must be block-scoped | Operational | **KEEP** | Prevents merge conflicts |
| 0010 | Legacy migration follows parity-first sequencing | Operational | **KEEP** | Prevents functionality regression (if migration still active) |
| 0011 | Legacy CSS must define safe token fallbacks | Frontend-Specific | **REMOVE** | Too narrow (legacy migration CSS only) |
| 0012 | Runtime behavior changes require operational verification | Vague | **REMOVE** | Too vague; duplicates other rules |
| 0013 | Observability is part of the contract | Architectural | **KEEP** | Structural (logging at API boundary) |
| 0014 | Keep this file high-signal | Meta | **REMOVE** | Meta-lesson about lessons (not architectural) |
| 0015 | Legacy migration must preserve interactive behavior | Operational | **CONDITIONAL KEEP** | Only if legacy migration is active; otherwise **REMOVE** |
| 0016 | Mock semantics must match UI contract keys | Testing | **REMOVE** | Testing implementation detail |
| 0017 | Hover parity requires selector specificity on label elements | Cosmetic | **REMOVE** | CSS selector choice (not structural) |
| 0018 | Table header hover must bind to explicit label node | Cosmetic | **REMOVE** | CSS/HTML structure detail |
| 0019 | Portaled UI must redeclare local CSS tokens | Cosmetic | **REMOVE** | CSS implementation detail |
| 0020 | Header hover must be bound to interactive target only | Cosmetic | **REMOVE** | CSS interaction detail |
| 0021 | Feature CSS modules need local token baselines | Cosmetic | **REMOVE** | Styling implementation detail |
| 0022 | Similar surfaces should share the same base fill | Cosmetic | **REMOVE** | Color/styling choice |
| 0023 | Feature code must import shared UI from package entrypoint | Architectural | **KEEP** | Enforces module structure |
| 0024 | Remove local wrappers after migration to shared UI | Operational | **CONDITIONAL REMOVE** | Only relevant during migration; too narrow |
| 0025 | Delete orphan facade files when usage hits zero | Operational | **REMOVE** | Hygiene task; not a pattern |
| 0026 | Workspace root must provide token fallbacks | Cosmetic | **REMOVE** | CSS/layout implementation detail |
| 0027 | Second-pass parity must diff against legacy snapshot | Operational | **REMOVE** | Only for legacy migration (phase-specific) |
| 0028 | Do not downgrade legacy charts in parity phase | Operational | **REMOVE** | Only for legacy migration (phase-specific) |
| 0029 | Legacy copy must be normalized to local DTO shapes | Operational | **REMOVE** | Only for legacy migration (phase-specific) |
| 0030 | Preserve UTF-8 when patching legacy-copied frontend files | Technical-Quirk | **REMOVE** | One-off encoding workaround |
| 0031 | Remove ts-nocheck with minimal explicit callback typing | Technical | **REMOVE** | TypeScript syntax cleanup |
| 0032 | Simulator hero metrics need tolerant alias mapping | Feature-Specific | **REMOVE** | Only for simulator feature |
| 0033 | First-fold workspace KPIs must be rendered in ProductHero | Feature-Specific | **REMOVE** | Only for ProductHero component |
| 0034 | Hero KPI order must be explicit, not payload-driven | Feature-Specific | **REMOVE** | Only for ProductHero component |
| 0035 | Workspace top bars that belong to the shell must be rendered as shell strips | Layout | **REMOVE** | UI layout decision (cosmetic) |
| 0036 | Shell bars inside padded app mains must break out at the route root | Layout | **REMOVE** | UI layout decision (cosmetic) |

---

## Summary

### KEEP (11 lessons)
0001, 0002, 0003, 0004, 0005, 0006, 0007, 0008, 0009, 0013, 0023

These are cross-domain, prevent repeated architectural failures, or enforce critical process rules.

### CONDITIONAL KEEP (2 lessons)
- **0010** — If legacy migration is still active phase
- **0015** — If legacy migration is still active phase

### REMOVE (23 lessons)
0011, 0012, 0014, 0016–0022, 0024–0036

These are cosmetic, too narrow, phase-specific, or one-off tweaks that don't represent permanent architectural knowledge.

---

## Next Steps

1. Move lessons 0011, 0012, 0014, 0016–0022, 0024–0036 to `.brain/lessons/archived/`
2. Review 0010 and 0015: confirm if legacy migration is still active (check tasks/todo.md)
3. Redistribute remaining 11 lessons to domain-specific directories per Distributed Lesson Architecture
4. Update brain-lesson.md skill to reference this framework
5. Rebuild brain.db with remaining lessons only

---

**Decision threshold:** A lesson survives if it would apply to the next feature, the feature after that, and beyond. One-off fixes belong in commit messages or task notes, not in the Brain.
