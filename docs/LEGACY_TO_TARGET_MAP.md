# Legacy To Target Map

Este documento define a ponte entre o repositorio antigo e o novo repositorio congelado.

## Objetivo

Migrar por intencao de dominio e contrato, nao por copia cega de pastas.

## Mapeamento inicial

| Legado | Destino principal | Observacao |
| --- | --- | --- |
| `apps/server_api` | `apps/server_core` | Renomeado para refletir papel canonico do backend principal. |
| `apps/server_api/internal/platform/authn` | `apps/server_core/internal/platform/auth` | Consolidar auth e authz na linguagem alvo do novo core. |
| `apps/server_api/internal/platform/config` | `apps/server_core/internal/platform/runtime_config` e `.../governance/*` | Config livre vai para runtime; policy/threshold/flag vira governanca explicita. |
| `apps/server_api/internal/platform/db` | `apps/server_core/internal/platform/db/postgres` e `.../timeseries` | Separar persistencia transacional de suporte temporal. |
| `apps/server_api/internal/modules/iam` | `apps/server_core/internal/modules/iam` | Pode ser o primeiro modulo efetivamente portado. |
| `contexts/analytics` | `apps/analytics_worker` e `apps/server_core/internal/modules/analytics_serving` | Compute pesado fica no worker; serving e read models ficam no core. |
| `contexts/buying` | `apps/server_core/internal/modules/procurement` | Nome interno estabilizado como `procurement`. |
| `contexts/integration` | `apps/integration_worker` e `apps/server_core/internal/modules/integrations_control` | Execucao externa no worker; governanca operacional no core. |
| `drivers/` | `apps/integration_worker/src/connectors`, `.../crawlers`, `.../normalizers` | Drivers deixam de ficar soltos como camada transversal. |
| `db/` | `apps/server_core/migrations` e modulos donos do historico | Historico deve permanecer dentro do dominio certo. |
| `hosts/http_api/` | `apps/server_core` | Transporte oficial passa a nascer dentro do core. |
| `frontend/apps` | `apps/web` | UI web continua thin e orientada ao server core. |
| `frontend/packages` | `packages/ui` e `packages/generated` | Shared UI e artefatos gerados saem da fronteira do app web. |
| `ui/` | `apps/desktop` | Desktop vira cliente fino. |
| `platform/` | `apps/server_core/internal/platform` | Infra compartilhada do produto passa a viver com fronteira explicita do core. |
| `shared/` | `apps/server_core/internal/shared` ou `packages/generated` | Nao manter `shared_types` manual como fonte paralela. |

## Nomes que nao devem ser perpetuados

- `shopping` como nome de dominio final
- `analytics_read` como nome de produto
- `integrations` sem separar controle e execucao
- `notifications` misturando dominio com entrega
- `history` como modulo transversal

## Ordem sugerida de migracao

1. Consolidar o repo novo com a estrutura final.
2. Portar `server_api` para `server_core`, preservando auth, tenancy, db e `iam`.
3. Extrair analytics legado para `analytics_worker` e `analytics_serving`.
4. Mover integracoes e drivers para `integration_worker`.
5. Reposicionar `buying` como `procurement` e repartir o que hoje estiver misturado com pricing, suppliers e inventory.
6. Trazer clientes thin para `apps/web`, `apps/desktop` e `apps/admin_console`.

