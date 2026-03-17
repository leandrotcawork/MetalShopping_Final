# Scripts

`scripts/` concentra automacoes do repositório, especialmente scaffold, validacao e geracao de artefatos derivados de contratos.

## Script inventory

- `scaffold_architecture.ps1`: garante a estrutura base do monorepo
- `generate_contract_artifacts.ps1`: ponto de entrada padrao para geracao de SDKs e tipos
- `validate_contracts.ps1`: ponto de entrada padrao para validacao de contratos

## Regras

- scripts nao sao fonte de verdade de contrato
- scripts devem operar a partir de `contracts/`
- qualquer workflow estavel de geracao ou validacao deve passar por aqui

