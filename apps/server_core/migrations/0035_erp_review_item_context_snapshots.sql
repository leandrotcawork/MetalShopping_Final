ALTER TABLE erp_review_items
  ADD COLUMN IF NOT EXISTS staging_snapshot JSONB NULL,
  ADD COLUMN IF NOT EXISTS reconciliation_output JSONB NULL;
