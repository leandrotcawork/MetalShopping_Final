# Governance Checklist

## Before editing

- confirm whether the artifact is a policy, threshold, or feature flag
- confirm the bounded context
- confirm whether a companion schema already exists

## While editing

- keep version explicit
- keep status explicit
- keep bounded context explicit
- keep scope hierarchy explicit
- ensure the semantics fit both core and worker consumers

## Final review

- file lives under the correct `contracts/governance/` subfolder
- the artifact type is explicit and stable
- the scope hierarchy matches the accepted runtime model
- schema references are correct when used
- no app code is being treated as a parallel governance source
- the artifact supports auditability and explainability

