CREATE TABLE IF NOT EXISTS erp_raw_records (
  raw_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  connector_type TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  payload_json JSONB NOT NULL,
  payload_hash TEXT NOT NULL,
  source_timestamp TIMESTAMPTZ NULL,
  cursor_value TEXT NULL,
  extracted_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT fk_erp_raw_records_run FOREIGN KEY (run_id) REFERENCES erp_sync_runs(run_id) ON DELETE CASCADE,
  CONSTRAINT chk_erp_raw_records_entity_type CHECK (entity_type IN ('products','prices','costs','inventory','sales','purchases','customers','suppliers'))
);

CREATE INDEX IF NOT EXISTS idx_erp_raw_records_run_entity
  ON erp_raw_records (run_id, entity_type);

CREATE INDEX IF NOT EXISTS idx_erp_raw_records_tenant_source
  ON erp_raw_records (tenant_id, entity_type, source_id);

ALTER TABLE erp_raw_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_raw_records FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_raw_records_tenant_isolation ON erp_raw_records;
CREATE POLICY erp_raw_records_tenant_isolation
ON erp_raw_records
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
