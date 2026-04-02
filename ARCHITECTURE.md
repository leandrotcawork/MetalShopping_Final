# MetalShopping Final Architecture

## Status

## Document boundary

This file owns the stable architecture thesis of MetalShopping.

It does not own:

- day-to-day status tracking
- active execution order details
- task progress or backlog state
- agent precedence rules

Those belong to `docs/PROJECT_SOT.md`, `docs/IMPLEMENTATION_PLAN.md`, and `docs/PROGRESS.md` according to the repository governance hierarchy.

Esta arquitetura esta aprovada como base oficial do repositorio. O objetivo deste documento e fixar a forma da plataforma; os detalhes normativos que nao devem mais ser rediscutidos ficam congelados nos ADRs.

## North Star

MetalShopping deve crescer como plataforma empresarial de longo prazo:

- monorepo
- server-first
- modular monolith no core
- workers especializados para compute, integracao e automacao
- Postgres como estado canonico
- Go no core da plataforma
- Python nos workers analiticos e de integracao durante a transicao
- contratos e governanca explicitos fora dos apps
- clientes thin: web, desktop e admin console

## Principios congelados

1. `apps/server_core` e o nucleo do sistema.
2. Nenhum worker e dono do estado canonico.
3. O core e dono de auth, tenant, governanca, contratos publicos, estado transacional e serving sincrono.
4. Workers existem para compute, ingestao, automacao e entrega assincrona.
5. `contracts/` e a fonte de verdade para APIs, eventos, policies, thresholds e feature flags.
6. Configuracao governada nao pode ficar espalhada por modulo.
7. Historico nao vira modulo transversal.
8. Dominio e plataforma permanecem separados explicitamente.

## Freezes finais obrigatorios

Os seguintes pontos passam a ser regra oficial:

1. Isolamento inicial multitenant: shared DB + `tenant_id` + `RLS`, com opcao futura de isolamento maior para tenants premium ou regulados.
2. Historico pertence a cada dominio; `platform/db/timeseries` e apenas infraestrutura de suporte temporal.
3. Governanca runtime usa `contracts/governance/*` para schema, `bootstrap/seeds/*` para defaults, banco para estado efetivo e `platform/governance/*` para resolucao em runtime.
4. Frontend permanece thin-client e consome SDKs gerados a partir de contratos; BFF separado so entra se aparecer divergencia real.
5. Integracao assincrona usa eventos versionados, outbox/inbox e workers desacoplados; request normal do core nao depende de worker sincronamente.

## Layout aprovado

```text
apps/
  server_core/
  analytics_worker/
  integration_worker/
  automation_worker/
  notifications_worker/
  web/
  desktop/
  admin_console/

contracts/
  api/
  events/
  governance/

bootstrap/
  seeds/

packages/
  feature-auth-session/
  feature-products/
  generated-types/
  platform-sdk/
  ui/
  generated/

ops/
  docker/
  k8s/
  observability/
  runbooks/
  provisioning/

docs/
scripts/
```

## Responsabilidades por bloco

### `apps/server_core`

Backend principal da plataforma. Responsavel por:

- API
- auth e authz
- tenancy
- governanca
- estado transacional
- publicacao de eventos via outbox
- composicao de respostas para web, desktop e admin

### `apps/analytics_worker`

Compute analitico, scoring, projecoes, explainability e publicacao de outputs analiticos.

### `apps/integration_worker`

Conectores externos, ERP, marketplaces, crawlers, import/export e normalizacao.

### `apps/automation_worker`

Triggers, campanhas, acoes automaticas e orquestracao assincrona.

### `apps/notifications_worker`

Entrega de email, SMS, WhatsApp e webhook. O dominio de alertas continua no core.

## Core structure

Todo modulo em `apps/server_core/internal/modules/*` deve seguir o padrao:

```text
domain/
application/
ports/
adapters/
transport/
events/
readmodel/
```

Separacao obrigatoria:

- `internal/platform/`: infraestrutura compartilhada
- `internal/modules/`: negocio
- `internal/shared/`: componentes compartilhados pequenos e neutros

## Modulos do core

- `iam`
- `tenant_admin`
- `catalog`
- `inventory`
- `pricing`
- `sales`
- `suppliers`
- `customers`
- `procurement`
- `market_intelligence`
- `analytics_serving`
- `crm`
- `automation`
- `integrations_control`
- `alerts`

## Governanca

Estrutura obrigatoria:

```text
contracts/governance/
bootstrap/seeds/governance/
apps/server_core/internal/platform/governance/
```

Deve suportar:

- schema formal
- seeds iniciais
- estado governado no banco
- versionamento
- effective dates
- override hierarchy
- auditoria
- explainability
- resolucao consistente entre Go e Python

Hierarquia alvo de resolucao:

- global
- environment
- tenant
- module
- entity/profile
- feature-target

## Comunicacao

Fluxo alvo:

1. `web`, `desktop` e `admin_console` chamam `server_core`.
2. `server_core` resolve auth, tenant e authz.
3. `server_core` le e escreve em Postgres.
4. Mutacoes relevantes publicam eventos via outbox.
5. Workers consomem eventos ou jobs.
6. Workers processam e publicam outputs materializados.
7. `server_core` serve UI e API usando estado canonico, read models e outputs materializados.

Regras:

- worker nao vira dono do produto
- worker nao depende de acesso arbitrario ao banco do core
- request sincrono normal nao pode depender de worker

## Docs relacionados

- `docs/PROJECT_SOT.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/PROGRESS.md`
- `docs/adrs/ADR-0001-architecture-foundation.md`
- `docs/adrs/ADR-0002-tenant-isolation.md`
- `docs/adrs/ADR-0003-history-model.md`
- `docs/adrs/ADR-0004-runtime-governance.md`
- `docs/adrs/ADR-0005-thin-clients-and-generated-sdks.md`
- `docs/adrs/ADR-0006-versioned-async-integration.md`
- `docs/adrs/ADR-0015-login-mvp-closure-governance.md`
- `docs/adrs/ADR-0016-sdk-generated-runtime-boundary.md`
