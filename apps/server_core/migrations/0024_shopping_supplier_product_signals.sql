CREATE TABLE IF NOT EXISTS shopping_supplier_product_signals (
  tenant_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  supplier_code TEXT NOT NULL,
  product_url TEXT NULL,
  url_status TEXT NOT NULL DEFAULT 'STALE',
  lookup_mode TEXT NOT NULL DEFAULT 'REFERENCE',
  lookup_mode_source TEXT NOT NULL DEFAULT 'INFERRED',
  manual_override BOOLEAN NOT NULL DEFAULT FALSE,
  last_checked_at TIMESTAMPTZ NULL,
  last_success_at TIMESTAMPTZ NULL,
  last_http_status INTEGER NULL,
  last_error_message TEXT NULL,
  created_by TEXT NOT NULL DEFAULT 'system',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, product_id, supplier_code),
  CONSTRAINT fk_shopping_supplier_product_signals_product
    FOREIGN KEY (product_id) REFERENCES catalog_products(product_id) ON DELETE CASCADE,
  CONSTRAINT chk_shopping_supplier_product_signals_supplier_code CHECK (BTRIM(supplier_code) <> ''),
  CONSTRAINT chk_shopping_supplier_product_signals_url_status
    CHECK (url_status IN ('ACTIVE', 'STALE', 'INVALID')),
  CONSTRAINT chk_shopping_supplier_product_signals_lookup_mode
    CHECK (lookup_mode IN ('EAN', 'REFERENCE')),
  CONSTRAINT chk_shopping_supplier_product_signals_lookup_mode_source
    CHECK (lookup_mode_source IN ('INFERRED', 'MANUAL')),
  CONSTRAINT chk_shopping_supplier_product_signals_http_status
    CHECK (last_http_status IS NULL OR last_http_status BETWEEN 100 AND 599),
  CONSTRAINT chk_shopping_supplier_product_signals_timestamps
    CHECK (last_success_at IS NULL OR last_checked_at IS NULL OR last_checked_at >= last_success_at),
  CONSTRAINT chk_shopping_supplier_product_signals_created_by
    CHECK (BTRIM(created_by) <> '')
);

CREATE INDEX IF NOT EXISTS idx_shopping_supplier_product_signals_tenant_supplier
  ON shopping_supplier_product_signals (tenant_id, supplier_code, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_shopping_supplier_product_signals_tenant_url_status
  ON shopping_supplier_product_signals (tenant_id, url_status, updated_at DESC);

ALTER TABLE shopping_supplier_product_signals ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopping_supplier_product_signals FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS shopping_supplier_product_signals_tenant_isolation ON shopping_supplier_product_signals;
CREATE POLICY shopping_supplier_product_signals_tenant_isolation
ON shopping_supplier_product_signals
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE shopping_supplier_product_signals TO metalshopping_app;
