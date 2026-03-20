# Governance Contract Checklist

## Before editing
- [ ] Artifact type confirmed: policy | threshold | feature_flag
- [ ] Bounded context confirmed
- [ ] Companion schema checked in `contracts/api/jsonschema/`

## While editing
- [ ] Version explicit
- [ ] Status explicit
- [ ] Bounded context explicit
- [ ] Scope hierarchy explicit (global → env → tenant → module)
- [ ] Semantics work for both Go and Python consumers

## Final
- [ ] File in correct `contracts/governance/` subfolder
- [ ] Artifact type stable and explicit
- [ ] No app code treated as parallel governance source
- [ ] Supports auditability and explainability
