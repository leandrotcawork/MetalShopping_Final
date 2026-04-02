CREATE TABLE IF NOT EXISTS erp_staging_records (
  staging_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  raw_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  normalized_json JSONB NOT NULL,
  validation_status TEXT NOT NULL,
  validation_errors JSONB NULL,
  normalized_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT fk_erp_staging_records_run FOREIGN KEY (run_id) REFERENCES erp_sync_runs(run_id) ON DELETE CASCADE,
  CONSTRAINT fk_erp_staging_records_raw FOREIGN KEY (raw_id) REFERENCES erp_raw_records(raw_id) ON DELETE CASCADE,
  CONSTRAINT chk_erp_staging_validation_status CHECK (validation_status IN ('valid','invalid'))
);

CREATE INDEX IF NOT EXISTS idx_erp_staging_records_run_entity
  ON erp_staging_records (run_id, entity_type);

ALTER TABLE erp_staging_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_staging_records FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_staging_records_tenant_isolation ON erp_staging_records;
CREATE POLICY erp_staging_records_tenant_isolation
ON erp_staging_records
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
