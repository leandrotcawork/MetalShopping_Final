CREATE TABLE IF NOT EXISTS erp_integration_instances (
  instance_id    TEXT PRIMARY KEY,
  tenant_id      TEXT NOT NULL,
  connector_type TEXT NOT NULL,
  display_name   TEXT NOT NULL,
  connection_ref TEXT NOT NULL,
  enabled_entities TEXT[] NOT NULL,
  sync_schedule  TEXT NULL,
  status         TEXT NOT NULL DEFAULT 'active',
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_erp_instances_connector_type CHECK (connector_type IN ('sankhya')),
  CONSTRAINT chk_erp_instances_status CHECK (status IN ('active', 'paused', 'disabled')),
  CONSTRAINT chk_erp_instances_display_name CHECK (BTRIM(display_name) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_erp_instances_tenant_active
  ON erp_integration_instances (tenant_id) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_erp_instances_tenant
  ON erp_integration_instances (tenant_id);

ALTER TABLE erp_integration_instances ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_integration_instances FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_integration_instances_tenant_isolation ON erp_integration_instances;
CREATE POLICY erp_integration_instances_tenant_isolation
ON erp_integration_instances
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
