ALTER TABLE shopping_price_run_items
  ADD COLUMN IF NOT EXISTS supplier_code TEXT NOT NULL DEFAULT 'DEFAULT',
  ADD COLUMN IF NOT EXISTS item_status TEXT NOT NULL DEFAULT 'OK',
  ADD COLUMN IF NOT EXISTS product_url TEXT NULL,
  ADD COLUMN IF NOT EXISTS http_status INTEGER NULL,
  ADD COLUMN IF NOT EXISTS elapsed_s NUMERIC(10,3) NULL,
  ADD COLUMN IF NOT EXISTS chosen_seller_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS notes TEXT NULL;

ALTER TABLE shopping_price_run_items
  ADD CONSTRAINT chk_shopping_price_run_items_supplier_code
    CHECK (BTRIM(supplier_code) <> '') NOT VALID;
ALTER TABLE shopping_price_run_items
  VALIDATE CONSTRAINT chk_shopping_price_run_items_supplier_code;

ALTER TABLE shopping_price_run_items
  ADD CONSTRAINT chk_shopping_price_run_items_item_status
    CHECK (item_status IN ('OK', 'NOT_FOUND', 'AMBIGUOUS', 'ERROR')) NOT VALID;
ALTER TABLE shopping_price_run_items
  VALIDATE CONSTRAINT chk_shopping_price_run_items_item_status;

ALTER TABLE shopping_price_run_items
  ADD CONSTRAINT chk_shopping_price_run_items_http_status
    CHECK (http_status IS NULL OR http_status BETWEEN 100 AND 599) NOT VALID;
ALTER TABLE shopping_price_run_items
  VALIDATE CONSTRAINT chk_shopping_price_run_items_http_status;

ALTER TABLE shopping_price_run_items
  ADD CONSTRAINT chk_shopping_price_run_items_elapsed_s
    CHECK (elapsed_s IS NULL OR elapsed_s >= 0) NOT VALID;
ALTER TABLE shopping_price_run_items
  VALIDATE CONSTRAINT chk_shopping_price_run_items_elapsed_s;

CREATE UNIQUE INDEX IF NOT EXISTS uq_shopping_price_run_items_natural_key
  ON shopping_price_run_items (tenant_id, run_id, product_id, supplier_code);

ALTER TABLE shopping_price_latest_snapshot
  ADD COLUMN IF NOT EXISTS supplier_code TEXT NOT NULL DEFAULT 'DEFAULT',
  ADD COLUMN IF NOT EXISTS item_status TEXT NOT NULL DEFAULT 'OK',
  ADD COLUMN IF NOT EXISTS product_url TEXT NULL,
  ADD COLUMN IF NOT EXISTS http_status INTEGER NULL,
  ADD COLUMN IF NOT EXISTS elapsed_s NUMERIC(10,3) NULL,
  ADD COLUMN IF NOT EXISTS chosen_seller_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS notes TEXT NULL;

ALTER TABLE shopping_price_latest_snapshot
  ADD CONSTRAINT chk_shopping_price_latest_snapshot_supplier_code
    CHECK (BTRIM(supplier_code) <> '') NOT VALID;
ALTER TABLE shopping_price_latest_snapshot
  VALIDATE CONSTRAINT chk_shopping_price_latest_snapshot_supplier_code;

ALTER TABLE shopping_price_latest_snapshot
  ADD CONSTRAINT chk_shopping_price_latest_snapshot_item_status
    CHECK (item_status IN ('OK', 'NOT_FOUND', 'AMBIGUOUS', 'ERROR')) NOT VALID;
ALTER TABLE shopping_price_latest_snapshot
  VALIDATE CONSTRAINT chk_shopping_price_latest_snapshot_item_status;

ALTER TABLE shopping_price_latest_snapshot
  ADD CONSTRAINT chk_shopping_price_latest_snapshot_http_status
    CHECK (http_status IS NULL OR http_status BETWEEN 100 AND 599) NOT VALID;
ALTER TABLE shopping_price_latest_snapshot
  VALIDATE CONSTRAINT chk_shopping_price_latest_snapshot_http_status;

ALTER TABLE shopping_price_latest_snapshot
  ADD CONSTRAINT chk_shopping_price_latest_snapshot_elapsed_s
    CHECK (elapsed_s IS NULL OR elapsed_s >= 0) NOT VALID;
ALTER TABLE shopping_price_latest_snapshot
  VALIDATE CONSTRAINT chk_shopping_price_latest_snapshot_elapsed_s;

DROP INDEX IF EXISTS uq_shopping_price_latest_snapshot_tenant_product;
CREATE UNIQUE INDEX IF NOT EXISTS uq_shopping_price_latest_snapshot_tenant_product_supplier
  ON shopping_price_latest_snapshot (tenant_id, product_id, supplier_code);

