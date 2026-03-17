# SDK Generation Checklist

## Before editing

- confirm the generation target
- confirm the contract source set
- confirm whether the change belongs in strategy docs, scripts, or generated outputs

## While editing

- keep `contracts/` as the only source of truth
- keep generated outputs downstream only
- keep target ownership explicit
- keep the generation path scriptable

## Final review

- sources come only from `contracts/`
- outputs map to the correct `packages/generated/*` target
- no manual parallel type system is introduced
- script entrypoints remain the canonical workflow
- frontend and worker consumption paths remain clear

