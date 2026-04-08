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

ALTER TABLE erp_integration_instances
  ADD CONSTRAINT chk_erp_instances_connection_kind CHECK (connection_kind = 'oracle'),
  ADD CONSTRAINT chk_erp_instances_oracle_host CHECK (BTRIM(db_host) <> ''),
  ADD CONSTRAINT chk_erp_instances_oracle_port CHECK (db_port > 0),
  ADD CONSTRAINT chk_erp_instances_oracle_username CHECK (BTRIM(db_username) <> ''),
  ADD CONSTRAINT chk_erp_instances_oracle_password_secret_ref CHECK (BTRIM(db_password_secret_ref) <> ''),
  ADD CONSTRAINT chk_erp_instances_oracle_target CHECK ((db_service_name IS NOT NULL) <> (db_sid IS NOT NULL)),
  ADD CONSTRAINT chk_erp_instances_oracle_connect_timeout CHECK (connect_timeout_seconds IS NULL OR connect_timeout_seconds > 0),
  ADD CONSTRAINT chk_erp_instances_oracle_fetch_batch_size CHECK (fetch_batch_size IS NULL OR fetch_batch_size > 0),
  ADD CONSTRAINT chk_erp_instances_oracle_entity_batch_size CHECK (entity_batch_size IS NULL OR entity_batch_size > 0);

ALTER TABLE erp_integration_instances
  DROP COLUMN connection_ref;
