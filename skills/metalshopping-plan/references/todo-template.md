# tasks/todo.md

## Feature: <n>
Type: <read-only | write+events | CRUD+governance | scraping>
Events: <yes: names | no>
ADR: <yes: number | no>

## Contracts
- [ ] contracts/api/openapi/<module>_v1.openapi.yaml

## Tasks
- [ ] T1: contract — $metalshopping-openapi-contracts
      commit: "feat(<m>): add OpenAPI contract"
- [ ] T2: Go module — $metalshopping-implement
      commit: "feat(<m>): implement handler and adapter"
- [ ] T3: worker — $metalshopping-implement (scraping only)
      commit: "feat(worker): implement <m> worker"
- [ ] T4: SDK — $metalshopping-sdk-generation
      commit: "chore(sdk): regenerate after <m>"
- [ ] T5: frontend — $metalshopping-frontend
      commit: "feat(<m>): implement React page"
- [ ] T6: ADR — $metalshopping-adr (if needed)
      commit: "docs(adr): ADR-XXXX <title> — verified and closed"

## Acceptance tests (all must pass before [x])
- [ ] go build ./... passes
- [ ] pnpm tsc --noEmit passes
- [ ] GET /api/v1/<route> returns real data (no mock)
- [ ] data visible in browser
- [ ] smoke: <script name>
- [ ] no regression
