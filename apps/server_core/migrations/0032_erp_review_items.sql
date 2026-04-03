CREATE TABLE IF NOT EXISTS erp_review_items (
  review_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  instance_id TEXT NOT NULL,
  connector_type TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  severity TEXT NOT NULL,
  reason_code TEXT NOT NULL,
  problem_summary TEXT NOT NULL,
  raw_id TEXT NOT NULL,
  staging_id TEXT NOT NULL,
  reconciliation_id TEXT NOT NULL,
  staging_snapshot JSONB NULL,
  reconciliation_output JSONB NULL,
  recommended_action TEXT NOT NULL,
  item_status TEXT NOT NULL DEFAULT 'open',
  resolved_at TIMESTAMPTZ NULL,
  resolved_by TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_erp_review_items_instance FOREIGN KEY (instance_id) REFERENCES erp_integration_instances(instance_id) ON DELETE CASCADE,
  CONSTRAINT fk_erp_review_items_run FOREIGN KEY (run_id) REFERENCES erp_sync_runs(run_id) ON DELETE CASCADE,
  CONSTRAINT chk_erp_review_items_severity CHECK (severity IN ('info','warning','error','critical')),
  CONSTRAINT chk_erp_review_items_status CHECK (item_status IN ('open','resolved','dismissed'))
);

CREATE INDEX IF NOT EXISTS idx_erp_review_items_tenant_status
  ON erp_review_items (tenant_id, item_status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_erp_review_items_run
  ON erp_review_items (run_id);

ALTER TABLE erp_review_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_review_items FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_review_items_tenant_isolation ON erp_review_items;
CREATE POLICY erp_review_items_tenant_isolation
ON erp_review_items
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
