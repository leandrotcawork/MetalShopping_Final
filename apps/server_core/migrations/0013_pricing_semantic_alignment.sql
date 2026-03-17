ALTER TABLE pricing_product_prices
  DROP CONSTRAINT IF EXISTS chk_pricing_product_prices_cost_basis_amount;

ALTER TABLE pricing_product_prices
  DROP CONSTRAINT IF EXISTS chk_pricing_product_prices_margin_floor_value;

ALTER TABLE pricing_product_prices
  RENAME COLUMN cost_basis_amount TO replacement_cost_amount;

ALTER TABLE pricing_product_prices
  ADD COLUMN IF NOT EXISTS average_cost_amount NUMERIC(18,4) NULL;

ALTER TABLE pricing_product_prices
  DROP COLUMN IF EXISTS margin_floor_value;

ALTER TABLE pricing_product_prices
  ADD CONSTRAINT chk_pricing_product_prices_replacement_cost_amount CHECK (replacement_cost_amount >= 0);

ALTER TABLE pricing_product_prices
  ADD CONSTRAINT chk_pricing_product_prices_average_cost_amount CHECK (
    average_cost_amount IS NULL OR average_cost_amount >= 0
  );
