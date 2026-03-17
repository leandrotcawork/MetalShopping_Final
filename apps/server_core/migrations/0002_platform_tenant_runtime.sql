CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS TEXT
LANGUAGE sql
STABLE
AS $$
  SELECT NULLIF(current_setting('app.tenant_id', true), '');
$$;

COMMENT ON FUNCTION current_tenant_id() IS
'Returns the tenant identifier currently bound to the Postgres session via app.tenant_id for RLS and tenancy-aware access patterns.';
