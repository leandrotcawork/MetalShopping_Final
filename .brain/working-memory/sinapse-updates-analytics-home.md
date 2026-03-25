---
task_id: analytics-home-review-2026-03-25
created_at: 2026-03-25T03:15:00Z
status: awaiting-approval
---

# Sinapse Update Proposals — Analytics Home Review

## Proposals for Developer Review

### 1. UPDATE: Frontend Type Safety Rule

**Current Sinapse:** cortex/frontend/type-safety.md

**Proposal:** Add explicit rule about `@ts-nocheck`

```markdown
## Anti-Pattern: @ts-nocheck

**Rule:** `@ts-nocheck` is never acceptable in production code.

**Why:** Disables all type checking. Shifts bugs from compile-time to runtime.

**When used:** Usually signals "I don't know how to fix the type errors" or "I'm cutting corners".

**The right move:** Fix the underlying type issues instead.

**Example Anti-Pattern:**
```typescript
// ❌ BAD: Disables type checking for entire file
// @ts-nocheck
const data: Record<string, unknown> = response.data;
```

**Example Correct Pattern:**
```typescript
// ✅ GOOD: Explicitly type the response
type TaxonomyScopeResponse = {
  kpis: { margin_pct: number };
  scope: { level_label: string };
};

const data: TaxonomyScopeResponse = response.data;
if (!data.kpis.margin_pct) throw new Error("Invalid response");
```

**Cost of fixing:** Usually 1-2 hours per file to enable checking + fix revealed errors.
**Cost of ignoring:** Bugs slip to QA + production. Refactoring is dangerous.
```

**Status:** ✅ APPROVED / ❌ REJECTED / 🤔 REVISE

---

### 2. UPDATE: API Response Validation

**Current Sinapse:** cortex/frontend/sdk-patterns.md

**Proposal:** Add section on discriminated unions for API responses

```markdown
## API Response Validation Pattern

**Pattern:** All API responses should use discriminated unions to represent success + error states.

**Why:**
- Single return type covers all cases (no undefined surprises)
- TypeScript narrows type correctly
- Compile-time safety for all branches

**Pattern:**
```typescript
type ApiResult<T> =
  | { ok: true; data: T; error: null }
  | { ok: false; data: null; error: ApiError };

// Use:
const result = await sdk.analytics.getTaxonomy();
if (!result.ok) {
  console.error(result.error.message);
  return;
}
// Now TypeScript knows result.data is TaxonomyV1
```

**Legacy Pattern (Avoid):**
```typescript
// ❌ Data is Record<string, unknown>, no type safety
const result = await api.analytics.workspaceTaxonomyScope({...});
const data = result.data; // Type is unknown
```
```

**Status:** ✅ APPROVED / ❌ REJECTED / 🤔 REVISE

---

### 3. NEW SINAPSE: Migration Checklist

**Proposal:** Create `cortex/frontend/legacy-migration-checklist.md`

```markdown
---
id: legacy-migration-checklist
title: Legacy Code Migration Checklist
region: cortex/frontend
tags: [migration, refactor, typescript, patterns]
weight: 0.65
---

# Legacy Migration Checklist

When migrating legacy code to production-ready:

## Type Safety
- [ ] Remove all `@ts-nocheck` directives
- [ ] Run `npm run web:typecheck` with no errors
- [ ] All API responses are typed (not `Record<string, unknown>`)
- [ ] Event handlers use proper React types

## Data Validation
- [ ] Network responses validated at runtime
- [ ] localStorage/sessionStorage parsed safely
- [ ] No type assertions (`as` keyword) without guards

## API Integration
- [ ] Using SDK methods, not internal API
- [ ] Responses match generated contract types
- [ ] Error cases are handled explicitly

## Component Quality
- [ ] All async components have loading + error + empty states
- [ ] useEffect cleanup patterns correct (disposed flag)
- [ ] No hardcoded values (colors, sizes, text)
- [ ] Reused components from packages/ui

## Testing
- [ ] Unit tests for data transformation
- [ ] Integration tests for API calls
- [ ] Error state tests (network failure, invalid data)
```

**Status:** ✅ APPROVED / ❌ REJECTED / 🤔 REVISE

---

### 4. NEW LESSON: TypeScript Migration Pattern

**Proposal:** Create lesson to be stored in cortex/frontend/lessons/

```markdown
---
id: lesson-analytics-home-types
title: "Lesson: Disabling TypeScript Didn't Help"
domain: frontend
tags: [typescript, migration, type-safety]
severity: high
status: resolved
---

# Lesson: Type Safety in Legacy Migrations

## What Happened

Analytics home migration hit TypeScript errors. Instead of fixing them, added `@ts-nocheck` to 23 files.

## Why It Failed

1. **Compile-time safety lost** — bugs found in QA, not before
2. **Refactoring dangerous** — no warning when schema changes
3. **IDE loses hints** — autocomplete stops working
4. **Onboarding harder** — new engineers can't learn from types

## The Cost

- Initial fix: 10 hours to enable TypeScript + type API layer
- If not fixed: recurring bugs, slower refactors, harder maintenance

## What To Do Instead

1. Run typecheck with errors visible
2. Create proper types for API responses
3. Add runtime validation (zod)
4. Use discriminated unions for success/error

## Pattern That Works

See: cortex/frontend/sdk-patterns.md — API Response Validation Pattern
```

**Status:** ✅ APPROVED / ❌ REJECTED / 🤔 REVISE

---

## Summary

**4 proposals:**
- 2 updates to existing sinapses
- 2 new sinapses (checklist + lesson)

**Total additions:** ~800 lines of documentation

**Impact:** Future migrations will follow type-safe patterns from day 1.

---

**Next:** Developer reviews proposals, marks ✅/❌/🤔 for each. Brain-consolidate will apply approvals.
