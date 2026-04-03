CREATE TABLE IF NOT EXISTS erp_sync_runs (
  run_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  instance_id TEXT NOT NULL,
  connector_type TEXT NOT NULL,
  run_mode TEXT NOT NULL,
  entity_scope TEXT[] NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  started_at TIMESTAMPTZ NULL,
  completed_at TIMESTAMPTZ NULL,
  promoted_count INT NOT NULL DEFAULT 0,
  warning_count INT NOT NULL DEFAULT 0,
  rejected_count INT NOT NULL DEFAULT 0,
  review_count INT NOT NULL DEFAULT 0,
  failure_summary TEXT NULL,
  cursor_state JSONB NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_erp_sync_runs_instance FOREIGN KEY (instance_id) REFERENCES erp_integration_instances(instance_id) ON DELETE CASCADE,
  CONSTRAINT chk_erp_sync_runs_mode CHECK (run_mode IN ('bulk','incremental','manual_rerun')),
  CONSTRAINT chk_erp_sync_runs_status CHECK (status IN ('pending','running','completed','failed','partial'))
);

CREATE INDEX IF NOT EXISTS idx_erp_sync_runs_tenant_instance
  ON erp_sync_runs (tenant_id, instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_erp_sync_runs_pending
  ON erp_sync_runs (status, created_at)
  WHERE status = 'pending';

ALTER TABLE erp_sync_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_sync_runs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_sync_runs_tenant_isolation ON erp_sync_runs;
CREATE POLICY erp_sync_runs_tenant_isolation
ON erp_sync_runs
USING (tenant_id = current_tenant_id() OR current_tenant_id() = '*')
WITH CHECK (tenant_id = current_tenant_id());
