CREATE TABLE IF NOT EXISTS catalog_product_identifiers (
  product_identifier_id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  identifier_type TEXT NOT NULL,
  identifier_value TEXT NOT NULL,
  source_system TEXT NULL,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_catalog_product_identifiers_product
    FOREIGN KEY (product_id) REFERENCES catalog_products(product_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_product_identifiers_tenant_type_value
  ON catalog_product_identifiers (tenant_id, identifier_type, identifier_value);

CREATE INDEX IF NOT EXISTS idx_catalog_product_identifiers_tenant_product
  ON catalog_product_identifiers (tenant_id, product_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_product_identifiers_primary_per_type
  ON catalog_product_identifiers (tenant_id, product_id, identifier_type)
  WHERE is_primary = TRUE;

ALTER TABLE catalog_product_identifiers ENABLE ROW LEVEL SECURITY;
ALTER TABLE catalog_product_identifiers FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS catalog_product_identifiers_tenant_isolation ON catalog_product_identifiers;
CREATE POLICY catalog_product_identifiers_tenant_isolation
ON catalog_product_identifiers
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
