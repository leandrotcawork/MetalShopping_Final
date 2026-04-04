ALTER TABLE pricing_product_prices
  ADD COLUMN IF NOT EXISTS price_context_code TEXT NOT NULL DEFAULT 'default';

ALTER TABLE pricing_product_prices
  ALTER COLUMN price_context_code SET DEFAULT 'default';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_pricing_product_prices_price_context_code'
      AND conrelid = 'pricing_product_prices'::regclass
  ) THEN
    ALTER TABLE pricing_product_prices
      ADD CONSTRAINT chk_pricing_product_prices_price_context_code
      CHECK (BTRIM(price_context_code) <> '');
  END IF;
END $$;

DROP INDEX IF EXISTS idx_pricing_product_prices_tenant_product_effective;
CREATE INDEX IF NOT EXISTS idx_pricing_product_prices_tenant_product_effective
  ON pricing_product_prices (tenant_id, product_id, price_context_code, effective_from DESC);

DROP INDEX IF EXISTS uq_pricing_product_prices_open_window;
CREATE UNIQUE INDEX IF NOT EXISTS uq_pricing_product_prices_open_window
  ON pricing_product_prices (tenant_id, product_id, price_context_code)
  WHERE effective_to IS NULL;

ALTER TABLE inventory_product_positions
  ADD COLUMN IF NOT EXISTS source_company_code TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS source_location_code TEXT NOT NULL DEFAULT '';

ALTER TABLE inventory_product_positions
  ALTER COLUMN source_company_code SET DEFAULT '',
  ALTER COLUMN source_location_code SET DEFAULT '';

ALTER TABLE inventory_product_positions
  DROP CONSTRAINT IF EXISTS chk_inventory_product_positions_on_hand_quantity;

DROP INDEX IF EXISTS idx_inventory_product_positions_tenant_product_effective;
CREATE INDEX IF NOT EXISTS idx_inventory_product_positions_tenant_product_effective
  ON inventory_product_positions (tenant_id, product_id, source_company_code, source_location_code, effective_from DESC);

DROP INDEX IF EXISTS uq_inventory_product_positions_open_window;
CREATE UNIQUE INDEX IF NOT EXISTS uq_inventory_product_positions_open_window
  ON inventory_product_positions (tenant_id, product_id, source_company_code, source_location_code)
  WHERE effective_to IS NULL;

CREATE TABLE IF NOT EXISTS erp_source_price_context_mappings (
  mapping_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  source_system TEXT NOT NULL,
  source_table_code TEXT NOT NULL,
  canonical_context_code TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_erp_source_price_context_mappings_source_system CHECK (BTRIM(source_system) <> ''),
  CONSTRAINT chk_erp_source_price_context_mappings_source_table_code CHECK (BTRIM(source_table_code) <> ''),
  CONSTRAINT chk_erp_source_price_context_mappings_canonical_context_code CHECK (BTRIM(canonical_context_code) <> '')
);

CREATE INDEX IF NOT EXISTS idx_erp_source_price_context_mappings_tenant_source
  ON erp_source_price_context_mappings (tenant_id, source_system, source_table_code, is_active, updated_at DESC);

DROP INDEX IF EXISTS uq_erp_source_price_context_mappings_active_source;
CREATE UNIQUE INDEX IF NOT EXISTS uq_erp_source_price_context_mappings_active_source
  ON erp_source_price_context_mappings (tenant_id, source_system, source_table_code)
  WHERE is_active;

ALTER TABLE erp_source_price_context_mappings ENABLE ROW LEVEL SECURITY;
ALTER TABLE erp_source_price_context_mappings FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS erp_source_price_context_mappings_tenant_isolation ON erp_source_price_context_mappings;
CREATE POLICY erp_source_price_context_mappings_tenant_isolation
ON erp_source_price_context_mappings
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
