# Scripts

`scripts/` concentra automacoes do repositório, especialmente scaffold, validacao e geracao de artefatos derivados de contratos.

## Script inventory

- `scaffold_architecture.ps1`: garante a estrutura base do monorepo
- `generate_contract_artifacts.ps1`: ponto de entrada padrao para geracao de SDKs e tipos
- `validate_contracts.ps1`: ponto de entrada padrao para validacao de contratos
- `sync_login_theme_tokens.ps1`: sincroniza e valida tokens visuais compartilhados entre login React e tema do Keycloak
- `check_sot_doc_drift.ps1`: valida drift de arquitetura entre mudancas estruturais e atualizacao obrigatoria de `PROJECT_SOT` e `PROGRESS` (local por working tree ou CI por diff de branch usando `-BaseRef`)
- `smoke_auth_session_local.ps1`: smoke end-to-end local do fluxo `auth/session` (`login -> me -> refresh/logout com CSRF`) contra backend + Keycloak
- `start_metalshopping_local.ps1`: abre backend (`server_core`) e frontend (`web`) em duas janelas PowerShell locais

## Regras

- scripts nao sao fonte de verdade de contrato
- scripts devem operar a partir de `contracts/`
- qualquer workflow estavel de geracao ou validacao deve passar por aqui
