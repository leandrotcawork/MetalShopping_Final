# web

Cliente web thin. Deve consumir o `server_core` sem carregar regra de negocio canonica.

## Current phase

The first operational surface is `Products`.

`apps/web` owns:

- shell
- routing
- providers
- page composition

It does not own:

- canonical contract definitions
- business-semantic aggregation rules
- manual DTO systems that compete with generated artifacts
