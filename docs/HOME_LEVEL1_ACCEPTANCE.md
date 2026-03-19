# Home Level 1 Acceptance

## Status

Closed as `Pronto Nivel 1` in make-it-work-first mode.

## Scope checked

- contract-first endpoint: `GET /api/v1/home/summary`
- backend-owned Go module route in `server_core`
- generated SDK consumption in `apps/web`
- real KPI rendering on Home page (no mock transport path)

## Evidence (2026-03-19)

- `go build ./apps/server_core/...` -> pass
- `go test ./apps/server_core/...` -> pass
- `npm.cmd --workspace @metalshopping/web run typecheck` -> pass
- `npm.cmd --workspace @metalshopping/web run build` -> pass
- `rg -n "fetch\(" packages/platform-sdk/src --glob "*.ts"` -> no matches
- `rg -n "\.\./.*generated" packages/platform-sdk/src/index.ts` -> no matches
- `rg -n "LEGACY-OK|as unknown as" packages/platform-sdk/src --glob "*.ts"` -> no matches

## Notes

- This closure follows the active rule in `docs/DEVELOPMENT_GUIDELINES_MAKE_IT_WORK.md`.
- Next module remains `Shopping Price` after contract/data-surface freeze.
