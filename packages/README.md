# Packages

`packages/` concentra UI compartilhada e artefatos gerados a partir dos contratos canonicos.

## Regras de trabalho

- `packages/ui` nao e fonte de regra de negocio
- `packages/feature-*` concentra adaptadores e view models por feature, sem redefinir contratos canonicos
- `packages/generated/*` recebe apenas artefatos gerados
- estrategia oficial de geracao esta em `docs/SDK_GENERATION_STRATEGY.md`
