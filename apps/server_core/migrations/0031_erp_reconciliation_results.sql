CREATE TABLE IF NOT EXISTS erp_reconciliation_results (
  reconciliation_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  staging_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  canonical_id TEXT NULL,
  action TEXT NOT NULL,
  classification TEXT NOT NULL,
  reason_code TEXT NOT NULL,
  warning_details JSONB NULL,
  reconciled_at TIMESTAMPTZ NOT NULL,
  promotion_status TEXT NOT NULL DEFAULT 'pending',
  CONSTRAINT fk_erp_reconciliation_staging FOREIGN KEY (staging_id) REFERENCES erp_staging_records(staging_id) ON DELETE CASCADE,
  CONSTRAINT chk_erp_reconciliation_action CHECK (action IN ('create','update','skip')),
  CONSTRAINT chk_erp_reconciliation_classification CHECK (classification IN ('promotable','promotable_with_warning','review_required','rejected')),
  CONSTRAINT chk_erp_reconciliation_promotion_status CHECK (promotion_status IN ('pending','promoting','promoted','failed'))
);

CREATE INDEX IF NOT EXISTS idx_erp_reconciliation_run
  ON erp_reconciliation_results (run_id, entity_type);

CREATE INDEX IF NOT EXISTS idx_erp_reconciliation_promotable
  ON erp_reconciliation_results (promotion_status, classification)
  WHERE promotion_status = 'pending'
    AND classification IN ('promotable','promotable_with_warning');

ALTER TABLE erp_reconciliation_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_reconciliation_results FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_reconciliation_results_tenant_isolation ON erp_reconciliation_results;
CREATE POLICY erp_reconciliation_results_tenant_isolation
ON erp_reconciliation_results
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
