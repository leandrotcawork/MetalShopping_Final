# OpenAPI Checklist

## Before editing

- confirm bounded context
- confirm target file name
- confirm whether shared schemas already exist

## While editing

- keep OpenAPI version explicit
- keep `info.version` explicit
- keep `x-metalshopping` metadata present
- prefer schema references over duplicated inline payloads
- keep path names and operation IDs stable

## Final review

- file lives under `contracts/api/openapi/`
- filename follows lowercase snake_case
- path versioning is explicit
- shared payloads are in `contracts/api/jsonschema/` when reusable
- no app code is being treated as a parallel contract source
- changes fit generated SDK consumption

