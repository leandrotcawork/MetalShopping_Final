# tasks/quality-debt.md
# Known quality issues that don't block delivery but should be addressed.
# Updated by $metalshopping-review when verdict is COMPLIANT BUT WEAK.
# Review at the start of each new domain cycle.

---

## QD-001 — writeJSON duplicated in 7 modules
Pattern: each module defines its own writeJSON and error helper
Target: internal/platform/httputil/ with shared writeJSON + writeError
Priority: fix before next domain cycle
Modules affected: home, catalog, shopping, pricing, inventory, suppliers, iam

## QD-002 — Inconsistent error envelope format
Pattern: homeErrorEnvelope, apiErrorEnvelope, writeShoppingError are different shapes
Target: one shared error format across all modules
Priority: fix together with QD-001 (same PR)

## QD-003 — Untyped response shapes (map[string]any)
Pattern: 16 occurrences of writeJSON(w, 200, map[string]any{...})
Target: typed response structs per endpoint, compiler-verified
Priority: next cycle, module by module

## QD-004 — No unit tests anywhere
Pattern: zero *_test.go files in internal/modules/
Target: domain invariants + service use cases tested with memory repos
Priority: add tests alongside any new feature — do not ship new features without tests

## QD-005 — No memory test doubles
Pattern: no in-memory repo implementations
Target: memory/ alongside every postgres/ adapter
Priority: create when writing QD-004 tests

## QD-006 — Missing pagination on list endpoints
Pattern: handleListProducts, handleListPositions return unbounded results
Target: limit + offset on every list, max page size enforced
Priority: fix before analytics/CRM domain (data volume will be higher)

## QD-007 — shopping handler imports adapters/postgres directly
Pattern: transport layer knows about infrastructure
Target: handler knows only application.Service
Priority: fix during layer migration

## QD-008 — Architecture layer direction (ports/adapters → clean domain)
Pattern: pgdb.BeginTenantTx coupled into adapter layer, domain not pure Go
Target: domain/ pure Go, infrastructure/ implements interfaces, delivery/ translates HTTP
Priority: migrate module by module starting with new modules
