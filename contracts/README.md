# Contracts

`contracts/` e a fonte de verdade para APIs, eventos e governanca.

- `api/openapi`: contratos HTTP publicos.
- `api/jsonschema`: schemas compartilhados e validaveis.
- `events/v1`, `events/v2`: envelopes e eventos versionados.
- `governance/*`: policies, thresholds e feature flags.

## Regras de trabalho

- contratos sao escritos aqui, nunca em `apps/` como fonte primaria
- clientes e workers consomem artefatos gerados a partir destes contratos
- convencoes oficiais estao em `docs/CONTRACT_CONVENTIONS.md`
- estrategia de geracao esta em `docs/SDK_GENERATION_STRATEGY.md`
- templates iniciais vivem nas subpastas com prefixo `_template`
