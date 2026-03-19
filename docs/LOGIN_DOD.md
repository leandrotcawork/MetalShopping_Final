# Login Definition Of Done

## Rule

Login is done only when every required gate is green and every remaining unchecked item has an explicit `Decisao: POST-MVP`.

## Contracts And Generation

- [x] auth/session contracts are the source of truth for payload shape  
  Gate: `contract_validation` workflow job + `scripts/validate_contracts.ps1 -Scope all`  
  Evidencia: validacao de contrato passa  
  Data: 2026-03-19

- [x] generated artifacts are up to date and validated with drift check  
  Gate: `sdk_generation_check` workflow job + `scripts/generate_contract_artifacts.ps1 -Target all -Check`  
  Evidencia: check passa  
  Data: 2026-03-19

- [x] guard de imutabilidade de `packages/generated-types` implementado  
  Gate: CI step `Guard - imutabilidade de generated-types`  
  Evidencia: step adicionado ao workflow de PR  
  Data: 2026-03-19

## SDK And Runtime Boundary

- [x] runtime consome generated artifacts somente por package boundary estavel  
  Gate: grep estrutural e imports de workspace  
  Evidencia: sem deep relative import  
  Data: 2026-03-19

- [x] zero deep relative import para internals de generated  
  Gate: CI step `Guard - deep relative import no runtime`  
  Evidencia: guard verde localmente (`rg -n "\.\./.*generated" packages/platform-sdk/src` vazio)  
  Data: 2026-03-19

- [x] zero `as unknown as` nao documentado no runtime  
  Gate: CI step `Guard - as unknown as nao documentado no runtime`  
  Evidencia: guard verde localmente (`rg -n "as unknown as" packages/platform-sdk/src` vazio)  
  Data: 2026-03-19

- [x] zero `fetch(` manual no runtime (`platform-sdk`)  
  Gate: CI step `SDK boundary checks`  
  Evidencia: `grep "fetch(" packages/platform-sdk/src --include="*.ts"` vazio  
  Data: 2026-03-19

- [x] zero rota hardcoded `'/api/'` e `"/api/"` no runtime (`platform-sdk`)  
  Gate: CI step `SDK boundary checks`  
  Evidencia: grep sem matches em `packages/platform-sdk/src`  
  Data: 2026-03-19

- [x] zero marcador de divida `LEGACY-OK`  
  Gate: grep estrutural  
  Evidencia: `grep "LEGACY-OK" packages/platform-sdk/src --include="*.ts"` vazio  
  Data: 2026-03-19

## Backend Auth Session Behavior

- [x] login start e callback backend-owned operacionais  
  Gate: `scripts/smoke_auth_session_local.ps1`  
  Evidencia: fluxo `/login -> callback -> /me` validado  
  Data: 2026-03-19

- [x] `GET /api/v1/auth/session/me` retorna `200` autenticado e `401` sem sessao  
  Gate: smoke local  
  Evidencia: cenarios autenticado e pos-logout validados  
  Data: 2026-03-19

- [x] `POST /api/v1/auth/session/refresh` exige CSRF valido e sessao valida  
  Gate: smoke local  
  Evidencia: sem CSRF `403`, CSRF invalido `403`, CSRF valido `200`  
  Data: 2026-03-19

- [x] `POST /api/v1/auth/session/logout` exige CSRF valido e sessao valida  
  Gate: smoke local  
  Evidencia: logout com CSRF valido `200`  
  Data: 2026-03-19

## Frontend Behavior

> Itens com `Gate: manual` foram verificados na sessao de validacao de 2026-03-19.
> Automacao E2E com Playwright fica no roadmap post-MVP e nao bloqueia o fechamento desta tranche.

- [x] cookie `session_id` com `HttpOnly`, `Secure`, `SameSite` verificado  
  Gate: manual  
  Evidencia: validacao em navegador e fluxo local com sessao backend-owned  
  Data: 2026-03-19  
  Responsavel: Leandro Theodoro

- [x] ausencia de flicker relevante no bootstrap verificada  
  Gate: manual  
  Evidencia: hard refresh em rota protegida sem exposicao de conteudo autenticado antes do redirect  
  Data: 2026-03-19  
  Responsavel: Leandro Theodoro

- [x] redirect pos-login sem loop verificado  
  Gate: manual  
  Evidencia: validacao manual com no maximo um redirect no fluxo  
  Data: 2026-03-19  
  Responsavel: Leandro Theodoro

- [x] fallback/manual login route controlada e nao quebrada  
  Gate: manual  
  Evidencia: `/login?manual=1` permanece operacional  
  Data: 2026-03-19  
  Responsavel: Leandro Theodoro

- [x] logout retorna para estado nao autenticado controlado  
  Gate: manual + smoke local  
  Evidencia: logout concluido e `GET /me` retorna `401`  
  Data: 2026-03-19  
  Responsavel: Leandro Theodoro

- [x] baseline visual React + Keycloak validada nesta tranche  
  Gate: manual (checklist visual)  
  Evidencia: tema aplicado e sincronizado com tokens de login  
  Data: 2026-03-19  
  Responsavel: Leandro Theodoro

## Quality Gates

- [x] `go test ./apps/server_core/...` verde  
  Gate: workflow `backend_tests` + comando local  
  Evidencia: passa  
  Data: 2026-03-19

- [x] `npm run web:typecheck` verde  
  Gate: workflow `web_quality` step `Typecheck web` + comando local  
  Evidencia: passa  
  Data: 2026-03-19

- [x] `npm run web:build` verde  
  Gate: workflow `web_quality` step `Build web` + comando local  
  Evidencia: passa  
  Data: 2026-03-19

- [x] `web:test` integrado ao CI  
  Gate: workflow `web_quality` step `web:test (vitest)`  
  Evidencia: `npm.cmd --workspace @metalshopping/web run test:ci` passou localmente e step existe no workflow  
  Data: 2026-03-19

- [x] SoT drift guard verde  
  Gate: workflow `sot_doc_drift` + `scripts/check_sot_doc_drift.ps1 -BaseRef ...`  
  Evidencia: passa  
  Data: 2026-03-19

- [x] boundary guards (deep import, cast, fetch manual, rota hardcoded) presentes e ativos  
  Gate: workflow `web_quality`  
  Evidencia: steps adicionados e verificacoes locais verdes  
  Data: 2026-03-19

## E2E Minimum Scenarios

- [x] login happy-path operacional  
  Gate: smoke local  
  Evidencia: sucesso ponta a ponta  
  Data: 2026-03-19

- [x] credencial invalida nao cria sessao  
  Gate: smoke local  
  Evidencia: `/me` permanece `401` apos tentativa invalida  
  Data: 2026-03-19

- [x] token CSRF ausente/invalido em mutacao retorna `403`  
  Gate: smoke local  
  Evidencia: validado em `refresh/logout`  
  Data: 2026-03-19

- [x] logout invalida sessao (`/me -> 401`)  
  Gate: smoke local  
  Evidencia: validado  
  Data: 2026-03-19

## Forbidden Patterns

- [x] nenhum token JWT/sessao em `localStorage` ou `sessionStorage`  
  Gate: revisao de codigo frontend auth  
  Evidencia: sem persistencia de token no storage  
  Data: 2026-03-19

- [x] nenhuma chamada auth direta fora do runtime oficial  
  Gate: boundary checks + revisao estrutural  
  Evidencia: sem fetch auth manual nas features/paginas  
  Data: 2026-03-19

## Run Canonico De Fechamento

| Campo | Valor |
|-------|-------|
| Run ID | pendente apos merge |
| Branch | main |
| Commit | pendente apos merge |
| Data | pendente apos merge |
| Jobs verdes | typecheck, go-test, web-build, web-test, boundary-guards, immutability-guard, contract-drift |

- [ ] anexar run canonico remoto de fechamento na branch `main`  
  Decisao: POST-MVP - nao bloqueia fechamento tecnico desta tranche local  
  Motivo: run remoto final so existe apos push/merge do conjunto atual  
  Ticket: https://github.com/<org>/<repo>/issues/POST-MVP-LOGIN-CANONICAL-RUN

## Open Items (Post-MVP Decision Log)

- [ ] automacao E2E browser-level com Playwright para auth bootstrap/redirect/session-expired  
  Decisao: POST-MVP - nao bloqueia fechamento desta tranche  
  Motivo: fluxo MVP validado por smoke backend + validacao manual browser-level  
  Ticket: https://github.com/<org>/<repo>/issues/POST-MVP-AUTH-E2E-PLAYWRIGHT

- [ ] teste de imutabilidade `generated-types` exercitado em PR real sem label `codegen`  
  Decisao: POST-MVP - nao bloqueia fechamento desta tranche  
  Motivo: guard implementado; falta apenas evidencia de execucao em PR remoto  
  Ticket: https://github.com/<org>/<repo>/issues/POST-MVP-CODEGEN-IMMUTABILITY-EXERCISE
