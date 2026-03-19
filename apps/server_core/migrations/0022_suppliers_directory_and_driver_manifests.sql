CREATE TABLE IF NOT EXISTS suppliers_directory (
  supplier_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  supplier_code TEXT NOT NULL,
  supplier_label TEXT NOT NULL,
  execution_kind TEXT NOT NULL,
  lookup_policy TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_suppliers_directory_tenant_code UNIQUE (tenant_id, supplier_code),
  CONSTRAINT chk_suppliers_directory_code CHECK (BTRIM(supplier_code) <> ''),
  CONSTRAINT chk_suppliers_directory_label CHECK (BTRIM(supplier_label) <> ''),
  CONSTRAINT chk_suppliers_directory_execution_kind CHECK (execution_kind IN ('HTTP', 'PLAYWRIGHT')),
  CONSTRAINT chk_suppliers_directory_lookup_policy CHECK (lookup_policy IN ('EAN_FIRST', 'REFERENCE_FIRST'))
);

CREATE INDEX IF NOT EXISTS idx_suppliers_directory_tenant_enabled
  ON suppliers_directory (tenant_id, enabled, supplier_code);

ALTER TABLE suppliers_directory ENABLE ROW LEVEL SECURITY;
ALTER TABLE suppliers_directory FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS suppliers_directory_tenant_isolation ON suppliers_directory;
CREATE POLICY suppliers_directory_tenant_isolation
ON suppliers_directory
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

CREATE TABLE IF NOT EXISTS supplier_driver_manifests (
  manifest_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  supplier_code TEXT NOT NULL,
  version_number INTEGER NOT NULL,
  family TEXT NOT NULL,
  config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  validation_status TEXT NOT NULL,
  validation_errors_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  is_active BOOLEAN NOT NULL DEFAULT FALSE,
  created_by TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_supplier_driver_manifests_tenant_supplier_version UNIQUE (tenant_id, supplier_code, version_number),
  CONSTRAINT fk_supplier_driver_manifests_directory
    FOREIGN KEY (tenant_id, supplier_code)
    REFERENCES suppliers_directory (tenant_id, supplier_code)
    ON DELETE CASCADE,
  CONSTRAINT chk_supplier_driver_manifests_version CHECK (version_number > 0),
  CONSTRAINT chk_supplier_driver_manifests_family CHECK (BTRIM(family) <> ''),
  CONSTRAINT chk_supplier_driver_manifests_validation_status CHECK (
    validation_status IN ('valid', 'invalid', 'pending')
  ),
  CONSTRAINT chk_supplier_driver_manifests_created_by CHECK (BTRIM(created_by) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_driver_manifests_active_per_supplier
  ON supplier_driver_manifests (tenant_id, supplier_code)
  WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_supplier_driver_manifests_tenant_supplier_version
  ON supplier_driver_manifests (tenant_id, supplier_code, version_number DESC);

ALTER TABLE supplier_driver_manifests ENABLE ROW LEVEL SECURITY;
ALTER TABLE supplier_driver_manifests FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS supplier_driver_manifests_tenant_isolation ON supplier_driver_manifests;
CREATE POLICY supplier_driver_manifests_tenant_isolation
ON supplier_driver_manifests
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

