ALTER TABLE erp_raw_records
  ADD COLUMN IF NOT EXISTS batch_ordinal INT NOT NULL DEFAULT 1;

ALTER TABLE erp_staging_records
  ADD COLUMN IF NOT EXISTS batch_ordinal INT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_erp_raw_records_run_entity_batch
  ON erp_raw_records (run_id, entity_type, batch_ordinal);

CREATE INDEX IF NOT EXISTS idx_erp_staging_records_run_entity_batch
  ON erp_staging_records (run_id, entity_type, batch_ordinal);
