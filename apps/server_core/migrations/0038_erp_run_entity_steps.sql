CREATE TABLE IF NOT EXISTS erp_run_entity_steps (
  step_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  status TEXT NOT NULL,
  batch_ordinal INT NOT NULL DEFAULT 0,
  source_cursor TEXT NULL,
  failure_summary TEXT NULL,
  started_at TIMESTAMPTZ NULL,
  completed_at TIMESTAMPTZ NULL,
  CONSTRAINT fk_erp_run_entity_steps_run FOREIGN KEY (run_id) REFERENCES erp_sync_runs(run_id) ON DELETE CASCADE,
  CONSTRAINT chk_erp_run_entity_steps_entity CHECK (entity_type IN ('products','prices','costs','inventory','sales','purchases','customers','suppliers')),
  CONSTRAINT chk_erp_run_entity_steps_status CHECK (status IN ('running','completed','failed','skipped_due_to_dependency'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_erp_run_entity_steps_tenant_run_entity
  ON erp_run_entity_steps (tenant_id, run_id, entity_type);

CREATE INDEX IF NOT EXISTS idx_erp_run_entity_steps_run
  ON erp_run_entity_steps (run_id, entity_type, status);

ALTER TABLE erp_run_entity_steps ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_run_entity_steps FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_run_entity_steps_tenant_isolation ON erp_run_entity_steps;
CREATE POLICY erp_run_entity_steps_tenant_isolation
ON erp_run_entity_steps
USING (tenant_id = current_tenant_id() OR current_tenant_id() = '*')
WITH CHECK (tenant_id = current_tenant_id());
