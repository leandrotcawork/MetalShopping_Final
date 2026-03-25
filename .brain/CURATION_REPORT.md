---
name: Lesson Curation Report
description: Final results of the lesson quality audit and distributed architecture migration
date: 2026-03-24
---

# Lesson Curation Report

**Date:** 2026-03-24
**Status:** Complete ✓

## Summary

Curated 36 lessons to 13 high-quality architectural lessons organized in distributed domain-specific directories.

- **Kept:** 13 lessons (36.1%)
- **Archived:** 23 lessons (63.9%)

---

## Kept Lessons (Distributed Architecture)

### Backend Lessons (5)
Location: `.brain/cortex/backend/lessons/`

| # | Title | Severity | Trigger | Reason |
|---|-------|----------|---------|--------|
| 0001 | Tenant-safe DB access is mandatory | Critical | baseline | Cross-domain (all Go), prevents data leaks |
| 0002 | Handlers must fail fast on auth and tenancy | Critical | baseline | Cross-domain (all handlers), security critical |
| 0003 | Outbox must be atomic with writes | Critical | baseline | Cross-domain (any event emission), prevents data loss |
| 0004 | Worker writes require tenant context and idempotency | Critical | baseline | Cross-domain (all workers), prevents data loss |
| 0013 | Observability is part of the contract | High | baseline | Structural (logging at API boundary), enforced widely |

### Frontend Lessons (5)
Location: `.brain/cortex/frontend/lessons/`

| # | Title | Severity | Trigger | Reason |
|---|-------|----------|---------|--------|
| 0005 | Frontend data flow must use platform SDK contracts | Critical | baseline | Cross-domain (all data fetch), ensures type safety |
| 0006 | Reuse design system before adding UI primitives | High | baseline | Structural (component architecture) |
| 0010 | Legacy migration follows parity-first sequencing | High | baseline | Prevents functionality regression (legacy phase) |
| 0015 | Legacy migration must preserve interactive behavior | High | correction | Prevents UX regression during migration |
| 0023 | Feature code must import shared UI from package entrypoint | High | baseline | Enforces module structure, prevents import chaos |

### Cross-Domain Process Lessons (3)
Location: `.brain/lessons/cross-domain/`

| # | Title | Severity | Trigger | Reason |
|---|-------|----------|---------|--------|
| 0007 | Generated artifacts are read-only outputs | High | baseline | Prevents build system breakage |
| 0008 | Completion requires validation + commit | High | baseline | Prevents unvalidated work from being lost |
| 0009 | tasks/todo.md edits must be block-scoped | Medium | baseline | Prevents merge conflicts in todo tracking |

---

## Archived Lessons (23)

Location: `.brain/lessons/archived/`

### Reason Categories

| Category | Count | Example | Why Removed |
|----------|-------|---------|------------|
| Cosmetic Frontend Tweaks | 12 | Hover selector specificity, CSS token redeclaration | Not architectural; styling implementation detail |
| Too Narrow/Phase-Specific | 6 | Legacy snapshot diffing, legacy chart handling | Only relevant during one phase or surface |
| Feature-Specific Details | 3 | ProductHero KPI order, simulator hero metrics | Applies to single component, not cross-domain |
| One-Off Technical Workarounds | 2 | UTF-8 encoding quirk, TypeScript syntax cleanup | Temporary fix; not a reusable pattern |

### Full List

| # | Title | Type | Reason |
|---|-------|------|--------|
| 0011 | Legacy CSS must define safe token fallbacks | Frontend | Too narrow (legacy migration CSS) |
| 0012 | Runtime behavior changes require operational verification | Vague | Too vague; duplicates other rules |
| 0014 | Keep this file high-signal | Meta | Meta-lesson about lessons themselves |
| 0016 | Mock semantics must match UI contract keys | Testing | Testing implementation detail |
| 0017 | Hover parity requires selector specificity on label elements | Cosmetic | CSS selector choice |
| 0018 | Table header hover must bind to explicit label node | Cosmetic | CSS/HTML structure detail |
| 0019 | Portaled UI must redeclare local CSS tokens | Cosmetic | CSS implementation detail |
| 0020 | Header hover must be bound to interactive target only | Cosmetic | CSS interaction detail |
| 0021 | Feature CSS modules need local token baselines | Cosmetic | Styling implementation detail |
| 0022 | Similar surfaces should share the same base fill | Cosmetic | Color/styling choice |
| 0024 | Remove local wrappers after migration to shared UI | Operational | Only relevant during migration |
| 0025 | Delete orphan facade files when usage hits zero | Hygiene | Maintenance task, not a pattern |
| 0026 | Workspace root must provide token fallbacks | Cosmetic | CSS/layout implementation detail |
| 0027 | Second-pass parity must diff against legacy snapshot | Operational | Only for legacy migration |
| 0028 | Do not downgrade legacy charts in parity phase | Operational | Only for legacy migration |
| 0029 | Legacy copy must be normalized to local DTO shapes | Operational | Only for legacy migration |
| 0030 | Preserve UTF-8 when patching legacy-copied frontend files | Workaround | One-off encoding quirk |
| 0031 | Remove ts-nocheck with minimal explicit callback typing | Technical | TypeScript syntax cleanup |
| 0032 | Simulator hero metrics need tolerant alias mapping | Feature | Only for simulator feature |
| 0033 | First-fold workspace KPIs must be rendered in ProductHero | Feature | Only for ProductHero component |
| 0034 | Hero KPI order must be explicit, not payload-driven | Feature | Only for ProductHero component |
| 0035 | Workspace top bars that belong to the shell must be rendered as shell strips | Layout | UI layout decision |
| 0036 | Shell bars inside padded app mains must break out at the route root | Layout | UI layout decision |

---

## What This Means

### Brain Quality Improvement
- **Before:** 36 lessons, mixed quality (architectural + cosmetic + one-offs)
- **After:** 13 lessons, all high-quality (cross-domain + prevents repeated failures)
- **Reduction:** 64% reduction in lesson clutter

### Architectural Clarity
The 13 kept lessons now represent **true architectural knowledge**:
- **Backend:** Tenant safety, handler patterns, event atomicity, observability
- **Frontend:** SDK contracts, design system reuse, legacy migration procedures, module structure
- **Cross-domain:** Build system integrity, completion discipline, collaboration discipline

### For brain-lesson Skill
Going forward, the Learner skill will:
1. Classify lessons into distributed domain-specific directories
2. Apply curation criteria before archiving new lessons
3. Reference only high-quality lessons in ContextMapper
4. Escalate when 3+ similar lessons exist → propose hippocampus convention

### For build_brain_db.py
The brain.db index will be rebuilt to:
- Include only 13 lessons from distributed directories
- Exclude all 23 archived lessons
- Support domain-specific lesson queries (e.g., "show me backend lessons")
- Trigger on next `python build_brain_db.py` run

---

## Next Steps

1. ✓ Define curation framework (LESSON_CURATION_FRAMEWORK.md)
2. ✓ Audit 36 lessons against criteria
3. ✓ Archive 23 low-quality lessons
4. ✓ Distribute 13 kept lessons to domain-specific directories
5. ✓ Update cortex/*/index.md with "Known Pitfalls" sections
6. ✓ Update cortex_registry.md with distributed architecture description
7. ✓ Update brain-lesson.md skill with new workflow
8. **TODO:** Rebuild brain.db with distributed lesson support
9. **TODO:** Update brain-map skill to load lessons from domain-specific directories
10. **TODO:** Test distributed lesson retrieval in brain.db queries

---

## Verification

To verify the curation is complete:

```bash
# Total kept lessons
find .brain/cortex/*/lessons .brain/lessons/cross-domain -name "lesson-*.md" | wc -l
# Should output: 13

# Total archived lessons
find .brain/lessons/archived -name "lesson-*.md" | wc -l
# Should output: 23

# No lessons in flat root
ls -la .brain/lessons/*.md 2>/dev/null | wc -l
# Should output: 0 (no flat lessons)
```

---

**Completion Status:** Phase 1 Complete ✓
**Next Phase:** Update build_brain_db.py to support distributed lesson indexing
