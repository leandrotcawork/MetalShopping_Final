# server_core

`server_core` e o backend principal do MetalShopping.

Ele e dono de:

- contratos publicos
- auth, authz e tenancy
- governanca
- estado transacional em Postgres
- serving sincrono para web, desktop e admin
- publicacao de eventos via outbox

Todo modulo de negocio em `internal/modules/` deve seguir o padrao:

`domain/`, `application/`, `ports/`, `adapters/`, `transport/`, `events/`, `readmodel/`

