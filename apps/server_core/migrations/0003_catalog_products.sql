CREATE TABLE IF NOT EXISTS catalog_products (
  product_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  sku TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_catalog_products_tenant_sku UNIQUE (tenant_id, sku),
  CONSTRAINT chk_catalog_products_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_catalog_products_tenant_id ON catalog_products(tenant_id);
CREATE INDEX IF NOT EXISTS idx_catalog_products_tenant_sku ON catalog_products(tenant_id, sku);

ALTER TABLE catalog_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE catalog_products FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS catalog_products_tenant_isolation ON catalog_products;
CREATE POLICY catalog_products_tenant_isolation
ON catalog_products
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
