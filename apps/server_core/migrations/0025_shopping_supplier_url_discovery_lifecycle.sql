-- ADR-0035: Shopping Supplier URL Discovery Lifecycle v1
--
-- Adds deterministic cooldown fields for Playwright URL discovery so the worker can
-- avoid reprocessing NOT_FOUND products on every run.

ALTER TABLE shopping_supplier_product_signals
  ADD COLUMN IF NOT EXISTS next_discovery_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS not_found_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE shopping_supplier_product_signals
  DROP CONSTRAINT IF EXISTS chk_shopping_supplier_product_signals_not_found_count;
ALTER TABLE shopping_supplier_product_signals
  ADD CONSTRAINT chk_shopping_supplier_product_signals_not_found_count CHECK (not_found_count >= 0);

CREATE INDEX IF NOT EXISTS idx_shopping_supplier_product_signals_discovery_eligibility
  ON shopping_supplier_product_signals (tenant_id, supplier_code, next_discovery_at)
  WHERE product_url IS NULL AND manual_override = FALSE;

-- Table grants are already applied in 0024, but keep this idempotent in case of replays.
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE shopping_supplier_product_signals TO metalshopping_app;

