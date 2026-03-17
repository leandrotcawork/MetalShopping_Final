# MetalShopping Final

MetalShopping nasce aqui como uma plataforma empresarial server-first, organizada em monorepo e preparada para evoluir como produto de longo prazo.

## Fase atual

O repositorio esta em modo de planejamento. Antes do desenvolvimento funcional, a base oficial precisa estar congelada em SoT, ADRs, plano de implementacao e arquivos de orientacao para agentes.

## Direcao congelada

- `apps/server_core` e o nucleo canonico da plataforma.
- `apps/*_worker` executam compute, integracao, automacao e entrega assincrona.
- `contracts/` e a fonte de verdade para APIs, eventos e governanca.
- `Postgres` e a base canonica do estado transacional.
- `Go` lidera o core; `Python` continua nos workers durante a transicao.

## Layout do repositorio

- `apps/`: core, workers e clientes thin.
- `contracts/`: APIs, eventos e schemas de governanca.
- `bootstrap/`: seeds iniciais e material de boot.
- `packages/`: UI compartilhada e artefatos gerados.
- `ops/`: docker, k8s, observabilidade, runbooks e provisioning.
- `docs/`: arquitetura, SoT, ADRs, plano e progresso.
- `scripts/`: automacao de scaffold e manutencao do repo.

## Documentos principais

- `ARCHITECTURE.md`: arquitetura oficial da plataforma.
- `docs/PROJECT_SOT.md`: fonte operacional de verdade para a fase atual.
- `docs/IMPLEMENTATION_PLAN.md`: plano de execucao por fases.
- `docs/PROGRESS.md`: status atual e proximos passos.
- `docs/adrs/`: decisoes arquiteturais congeladas.
- `docs/LEGACY_TO_TARGET_MAP.md`: ponte entre o repositorio antigo e o alvo final.
- `scripts/scaffold_architecture.ps1`: script idempotente para materializar a estrutura base.
- `AGENTS.md`: orientacao global para agentes trabalharem com contexto certo e menos desperdicio de tokens.
