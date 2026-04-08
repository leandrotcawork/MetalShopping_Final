ALTER TABLE erp_integration_instances
  ADD COLUMN connection_kind TEXT NOT NULL DEFAULT 'oracle',
  ADD COLUMN db_host TEXT NOT NULL DEFAULT '',
  ADD COLUMN db_port INT NOT NULL DEFAULT 1521,
  ADD COLUMN db_service_name TEXT NULL,
  ADD COLUMN db_sid TEXT NULL,
  ADD COLUMN db_username TEXT NOT NULL DEFAULT '',
  ADD COLUMN db_password_secret_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN connect_timeout_seconds INT NULL,
  ADD COLUMN fetch_batch_size INT NULL,
  ADD COLUMN entity_batch_size INT NULL;

-- Preserve the legacy opaque connection_ref until a safe backfill exists for
-- historical rows. New writes already use the structured Oracle config.
