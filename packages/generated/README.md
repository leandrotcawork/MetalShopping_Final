# Generated

`packages/generated/` existe para artefatos derivados dos contratos canonicos.

- `sdk_ts/`: SDK TypeScript para clientes web e admin
- `sdk_py/`: SDK Python para workers e integracoes
- `types_ts/`: tipos TS compartilhados derivados de schema
- `sdk_ts/` deve sair do OpenAPI Generator oficial, nao de emissao TypeScript artesanal
- o runtime HTTP compartilhado do frontend deve ficar fino e centralizado, consumindo o client gerado

Nenhum artefato aqui deve virar fonte primaria manual.
