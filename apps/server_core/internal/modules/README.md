# Modules

`internal/modules/` contem apenas dominio de negocio.

Catalogo inicial congelado:

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

Cada modulo deve respeitar a estrutura:

```text
domain/
application/
ports/
adapters/
transport/
events/
readmodel/
```

## Regra oficial

Os principios e responsabilidades detalhados de modulo estao em `docs/MODULE_STANDARDS.md`.

## Template

Use `internal/modules/_template/` como ponto de partida estrutural para novos modulos.
