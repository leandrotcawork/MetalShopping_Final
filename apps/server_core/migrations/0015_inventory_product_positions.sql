CREATE TABLE IF NOT EXISTS inventory_product_positions (
  position_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  on_hand_quantity NUMERIC(18,4) NOT NULL,
  last_purchase_at TIMESTAMPTZ NULL,
  last_sale_at TIMESTAMPTZ NULL,
  position_status TEXT NOT NULL,
  effective_from TIMESTAMPTZ NOT NULL,
  effective_to TIMESTAMPTZ NULL,
  origin_type TEXT NOT NULL,
  origin_ref TEXT NULL,
  reason_code TEXT NOT NULL,
  updated_by TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_inventory_product_positions_product
    FOREIGN KEY (product_id) REFERENCES catalog_products(product_id) ON DELETE CASCADE,
  CONSTRAINT chk_inventory_product_positions_on_hand_quantity CHECK (on_hand_quantity >= 0),
  CONSTRAINT chk_inventory_product_positions_status CHECK (
    position_status IN ('active', 'inactive')
  ),
  CONSTRAINT chk_inventory_product_positions_origin_type CHECK (
    origin_type IN ('manual', 'import', 'reconciliation')
  ),
  CONSTRAINT chk_inventory_product_positions_reason_code CHECK (
    BTRIM(reason_code) <> ''
  ),
  CONSTRAINT chk_inventory_product_positions_effective_window CHECK (
    effective_to IS NULL OR effective_to > effective_from
  )
);

CREATE INDEX IF NOT EXISTS idx_inventory_product_positions_tenant_product_effective
  ON inventory_product_positions (tenant_id, product_id, effective_from DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_inventory_product_positions_open_window
  ON inventory_product_positions (tenant_id, product_id)
  WHERE effective_to IS NULL;

ALTER TABLE inventory_product_positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_product_positions FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS inventory_product_positions_tenant_isolation ON inventory_product_positions;
CREATE POLICY inventory_product_positions_tenant_isolation
ON inventory_product_positions
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
