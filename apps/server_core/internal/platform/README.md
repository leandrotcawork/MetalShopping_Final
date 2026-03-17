# Platform

`internal/platform/` concentra infraestrutura compartilhada do core, sem carregar regras de negocio.

Blocos reservados:

- `auth`
- `tenancy_runtime`
- `runtime_config`
- `governance`
- `db`
- `messaging`
- `jobs`
- `cache`
- `files`
- `delivery`
- `observability`
- `security`
- `auditlog`

## Regra oficial

As fronteiras entre `platform/`, `modules/` e `shared/` estao detalhadas em `docs/PLATFORM_BOUNDARIES.md`.

## Template

Use `internal/platform/_template_package/` como ponto de partida conceitual para novos pacotes de plataforma.
