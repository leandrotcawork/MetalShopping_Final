-- Allow explicit system-claim sessions (`app.tenant_id='*'`) to read pending
-- run/reconciliation rows across tenants while keeping tenant-scoped writes
-- protected by WITH CHECK (tenant_id = current_tenant_id()).

DROP POLICY IF EXISTS erp_sync_runs_tenant_isolation ON erp_sync_runs;
CREATE POLICY erp_sync_runs_tenant_isolation
ON erp_sync_runs
USING (tenant_id = current_tenant_id() OR current_tenant_id() = '*')
WITH CHECK (tenant_id = current_tenant_id());

DROP POLICY IF EXISTS erp_reconciliation_results_tenant_isolation ON erp_reconciliation_results;
CREATE POLICY erp_reconciliation_results_tenant_isolation
ON erp_reconciliation_results
USING (tenant_id = current_tenant_id() OR current_tenant_id() = '*')
WITH CHECK (tenant_id = current_tenant_id());
