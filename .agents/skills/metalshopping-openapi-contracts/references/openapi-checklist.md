# OpenAPI Checklist

## Before editing
- [ ] Confirmed bounded context and target filename
- [ ] Checked if shared schemas already exist in `contracts/api/jsonschema/`

## While editing
- [ ] OpenAPI version explicit
- [ ] `info.version` explicit
- [ ] `x-metalshopping` metadata present
- [ ] Path versioning: `/api/v1/...`
- [ ] Shared payloads referenced, not duplicated inline

## Final
- [ ] File in `contracts/api/openapi/` with snake_case name
- [ ] No app code being treated as parallel contract source
- [ ] Changes fit generated SDK consumption
- [ ] Breaking change? → version bump required
