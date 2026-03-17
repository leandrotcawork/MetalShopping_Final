CREATE TABLE IF NOT EXISTS pricing_product_prices (
  price_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  currency_code TEXT NOT NULL,
  price_amount NUMERIC(18,4) NOT NULL,
  cost_basis_amount NUMERIC(18,4) NOT NULL,
  margin_floor_value NUMERIC(9,4) NOT NULL,
  pricing_status TEXT NOT NULL,
  effective_from TIMESTAMPTZ NOT NULL,
  effective_to TIMESTAMPTZ NULL,
  origin_type TEXT NOT NULL,
  origin_ref TEXT NULL,
  reason_code TEXT NOT NULL,
  updated_by TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_pricing_product_prices_product
    FOREIGN KEY (product_id) REFERENCES catalog_products(product_id) ON DELETE CASCADE,
  CONSTRAINT chk_pricing_product_prices_currency_code CHECK (currency_code ~ '^[A-Z]{3}$'),
  CONSTRAINT chk_pricing_product_prices_price_amount CHECK (price_amount >= 0),
  CONSTRAINT chk_pricing_product_prices_cost_basis_amount CHECK (cost_basis_amount >= 0),
  CONSTRAINT chk_pricing_product_prices_margin_floor_value CHECK (margin_floor_value >= 0),
  CONSTRAINT chk_pricing_product_prices_status CHECK (
    pricing_status IN ('draft', 'active', 'inactive')
  ),
  CONSTRAINT chk_pricing_product_prices_origin_type CHECK (
    origin_type IN ('manual', 'policy', 'import')
  ),
  CONSTRAINT chk_pricing_product_prices_reason_code CHECK (
    BTRIM(reason_code) <> ''
  ),
  CONSTRAINT chk_pricing_product_prices_effective_window CHECK (
    effective_to IS NULL OR effective_to > effective_from
  )
);

CREATE INDEX IF NOT EXISTS idx_pricing_product_prices_tenant_product_effective
  ON pricing_product_prices (tenant_id, product_id, effective_from DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_pricing_product_prices_open_window
  ON pricing_product_prices (tenant_id, product_id)
  WHERE effective_to IS NULL;

ALTER TABLE pricing_product_prices ENABLE ROW LEVEL SECURITY;
ALTER TABLE pricing_product_prices FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS pricing_product_prices_tenant_isolation ON pricing_product_prices;
CREATE POLICY pricing_product_prices_tenant_isolation
ON pricing_product_prices
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
