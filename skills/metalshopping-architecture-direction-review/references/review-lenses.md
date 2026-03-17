# Review Lenses

Use these lenses to keep the review sharp and non-generic.

## 1. Boundary integrity

- Is the code in the right layer: `platform`, `modules`, `shared`, worker, or client?
- Is a business concern leaking into infrastructure or vice versa?
- Is the slice reinforcing the modular monolith instead of creating hidden coupling?

## 2. Canonical ownership

- Is the owner of state explicit?
- Is the canonical model being strengthened or bypassed?
- Is a convenience field being added where a supporting structure should exist instead?

## 3. Multi-tenant and data safety

- Does the slice respect `tenant_id`, runtime tenant context, and future RLS enforcement?
- Would this hold up as more tenants, data volume, and modules arrive?
- Does the data design still support stronger isolation later if needed?

## 4. Governance and contracts

- Should the behavior be governed instead of hardcoded?
- Are API, event, or governance contracts required or drifting?
- Is the implementation still consistent with contract-first rules?

## 5. Product future-fit

- Does this choice make future modules easier or harder:
  analytics
  CRM
  procurement/buying
  market intelligence
  adaptive integration and DB connectivity
- Is the slice solving a local problem at the expense of the long-term platform?

## 6. Operational quality

- Is the work consistent with observability, auth, authz, tenancy, and audit expectations?
- Is the runtime behavior production-grade for this phase?
- Is this a clean phased bootstrap, or a hidden long-term compromise?

## 7. Review verdict

End with one of:

- `aligned`
- `partially aligned`
- `misaligned`

Then give the single next best structural move.
