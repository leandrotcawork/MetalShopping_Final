# Development Guidelines Make It Work First

## Status

Regra ativa para os proximos modulos do produto.

## Filosofia

Ordem obrigatoria:

1. make it work
2. make it clean
3. make it fast

Nao pular etapas.

## Fluxo padrao por modulo

Todo modulo novo segue exatamente:

1. endpoint declarado em `contracts/api/openapi/*.yaml`
2. handler Go implementado no `server_core`
3. SDK gerado via `scripts/generate_contract_artifacts.ps1`
4. pagina React consumindo dados via `@metalshopping/sdk-runtime`

## Criterio de pronto nivel 1

Um modulo esta pronto nivel 1 quando:

- endpoint existe no OpenAPI
- handler retorna dados reais
- SDK gera sem erro
- pagina renderiza dados reais sem erro de console
- `npm run web:typecheck` passa
- `go build ./...` passa
- login e modulos anteriores nao quebram

## Regras operacionais

- sem perfeicao prematura no primeiro corte
- sem refatoracao estrutural sem necessidade real
- sem criar padroes alternativos de integracao front/back
- sem editar artefato gerado manualmente
- sem reabrir modulo fechado nivel 1 sem gatilho real de negocio

## Gates minimos obrigatorios

- `npm run web:typecheck`
- `go build ./...`
- guards de boundary SDK no CI

Testes E2E, performance e cobertura avancada ficam para tranches posteriores.

## Ordem atual de implementacao

1. Home Page
2. Shopping Price
3. Analytics
4. CRM
