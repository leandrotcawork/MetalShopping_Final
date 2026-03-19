CREATE TABLE IF NOT EXISTS shopping_price_run_requests (
  run_request_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  request_status TEXT NOT NULL,
  input_mode TEXT NOT NULL,
  input_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  requested_by TEXT NOT NULL,
  requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  claimed_at TIMESTAMPTZ NULL,
  started_at TIMESTAMPTZ NULL,
  finished_at TIMESTAMPTZ NULL,
  worker_id TEXT NULL,
  run_id TEXT NULL,
  error_message TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_shopping_price_run_requests_status CHECK (
    request_status IN ('queued', 'claimed', 'running', 'completed', 'failed', 'cancelled')
  ),
  CONSTRAINT chk_shopping_price_run_requests_input_mode CHECK (
    input_mode IN ('xlsx', 'catalog')
  ),
  CONSTRAINT chk_shopping_price_run_requests_requested_by CHECK (BTRIM(requested_by) <> '')
);

CREATE INDEX IF NOT EXISTS idx_shopping_price_run_requests_tenant_status_requested
  ON shopping_price_run_requests (tenant_id, request_status, requested_at DESC);

CREATE INDEX IF NOT EXISTS idx_shopping_price_run_requests_tenant_run
  ON shopping_price_run_requests (tenant_id, run_id);

ALTER TABLE shopping_price_run_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopping_price_run_requests FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS shopping_price_run_requests_tenant_isolation ON shopping_price_run_requests;
CREATE POLICY shopping_price_run_requests_tenant_isolation
ON shopping_price_run_requests
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

