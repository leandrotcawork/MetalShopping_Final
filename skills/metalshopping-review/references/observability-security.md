# Observability and Security Baseline

## Structured logging (every handler)
```go
slog.Info("module_request",
    "action",      "module.get_overview",
    "trace_id",    traceID,
    "result",      result,       // "success" | "not_found" | "internal_error"
    "status",      statusCode,
    "duration_ms", time.Since(startedAt).Milliseconds(),
)
```
Never log: passwords, tokens, session_ids, raw bodies with PII.

## Worker logging
```python
log("worker_start", tenant_id=tenant_id, run_id=run_id, items=n)
log("worker_end",   tenant_id=tenant_id, rows_written=n)
log("worker_error", tenant_id=tenant_id, error=str(exc))
```

## Auth flow (enforced in every handler)
1. JWT validated by `platform/auth` middleware before handler runs
2. Handler: `PrincipalFromContext` → 401 if missing
3. Handler: `TenantFromContext` → 403 if missing
4. Handler calls service with `tenant.ID`

## Security rules
- CSRF validated on all mutating endpoints (POST, PUT, PATCH, DELETE)
- Cookies: HttpOnly + Secure + SameSite=Strict
- JWT never in response body — session cookie only
- Error messages never expose stack traces or internal details
- Credentials and tokens never logged

## Tenancy enforcement layers
1. App layer: `pgdb.BeginTenantTx` sets `app.current_tenant_id`
2. SQL layer: `current_tenant_id()` in every WHERE
3. DB layer: RLS policies as defense-in-depth
4. Worker layer: Python `set_config('app.current_tenant_id',...)` before every write tx
