---
name: metalshopping-sdk-generation
description: Generate MetalShopping SDK and type artifacts from contracts/. Called by $ms for T4. Run after any contract change before writing frontend code.
---

# MetalShopping SDK Generation

## Workflow
1. Confirm the contract change is finalized in `contracts/`
2. Run generation: `./scripts/generate_contract_artifacts.ps1`
3. Verify: `pnpm tsc --noEmit` passes
4. Commit: `chore(sdk): regenerate after <module> contract`

## Rules
- `contracts/` is the only source of truth — never edit generated outputs
- `packages/generated/*` contains generated artifacts only — never edit manually
- Run generation all-or-nothing: do not partially regenerate
- Frontend code is written only after generation succeeds and tsc passes

## Output targets
- `packages/generated/sdk_ts/` — TypeScript SDK clients (web consumption)
- `packages/generated/types_ts/` — shared TypeScript types
- `packages/generated/sdk_py/` — Python models (worker consumption)

## References
- Generation script: `scripts/generate_contract_artifacts.ps1`
- Strategy doc: `docs/SDK_GENERATION_STRATEGY.md`
