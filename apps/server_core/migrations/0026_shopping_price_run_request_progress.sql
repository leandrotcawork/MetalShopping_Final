ALTER TABLE shopping_price_run_requests
  ADD COLUMN IF NOT EXISTS total_items BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS processed_items BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS current_supplier_code TEXT NULL,
  ADD COLUMN IF NOT EXISTS current_product_id TEXT NULL,
  ADD COLUMN IF NOT EXISTS current_product_label TEXT NULL,
  ADD COLUMN IF NOT EXISTS progress_updated_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_shopping_price_run_requests_tenant_progress
  ON shopping_price_run_requests (tenant_id, request_status, progress_updated_at DESC);
