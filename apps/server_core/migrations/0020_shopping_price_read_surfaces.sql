CREATE TABLE IF NOT EXISTS shopping_price_runs (
  run_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_status TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  finished_at TIMESTAMPTZ NULL,
  processed_items BIGINT NOT NULL DEFAULT 0,
  total_items BIGINT NOT NULL DEFAULT 0,
  notes TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_shopping_price_runs_status CHECK (
    run_status IN ('queued', 'running', 'completed', 'failed')
  ),
  CONSTRAINT chk_shopping_price_runs_processed_items CHECK (processed_items >= 0),
  CONSTRAINT chk_shopping_price_runs_total_items CHECK (total_items >= 0),
  CONSTRAINT chk_shopping_price_runs_time_window CHECK (
    finished_at IS NULL OR finished_at >= started_at
  )
);

CREATE INDEX IF NOT EXISTS idx_shopping_price_runs_tenant_started
  ON shopping_price_runs (tenant_id, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_shopping_price_runs_tenant_status_started
  ON shopping_price_runs (tenant_id, run_status, started_at DESC);

ALTER TABLE shopping_price_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopping_price_runs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS shopping_price_runs_tenant_isolation ON shopping_price_runs;
CREATE POLICY shopping_price_runs_tenant_isolation
ON shopping_price_runs
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

CREATE TABLE IF NOT EXISTS shopping_price_run_items (
  run_item_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  seller_name TEXT NOT NULL,
  channel TEXT NOT NULL,
  observed_price NUMERIC(18,4) NOT NULL,
  currency_code TEXT NOT NULL,
  observed_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_shopping_price_run_items_run
    FOREIGN KEY (run_id) REFERENCES shopping_price_runs(run_id) ON DELETE CASCADE,
  CONSTRAINT fk_shopping_price_run_items_product
    FOREIGN KEY (product_id) REFERENCES catalog_products(product_id) ON DELETE CASCADE,
  CONSTRAINT chk_shopping_price_run_items_seller_name CHECK (BTRIM(seller_name) <> ''),
  CONSTRAINT chk_shopping_price_run_items_channel CHECK (BTRIM(channel) <> ''),
  CONSTRAINT chk_shopping_price_run_items_observed_price CHECK (observed_price >= 0),
  CONSTRAINT chk_shopping_price_run_items_currency CHECK (currency_code ~ '^[A-Z]{3}$')
);

CREATE INDEX IF NOT EXISTS idx_shopping_price_run_items_tenant_run
  ON shopping_price_run_items (tenant_id, run_id, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_shopping_price_run_items_tenant_product
  ON shopping_price_run_items (tenant_id, product_id, observed_at DESC);

ALTER TABLE shopping_price_run_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopping_price_run_items FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS shopping_price_run_items_tenant_isolation ON shopping_price_run_items;
CREATE POLICY shopping_price_run_items_tenant_isolation
ON shopping_price_run_items
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

CREATE TABLE IF NOT EXISTS shopping_price_latest_snapshot (
  snapshot_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  seller_name TEXT NOT NULL,
  channel TEXT NOT NULL,
  observed_price NUMERIC(18,4) NOT NULL,
  currency_code TEXT NOT NULL,
  observed_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_shopping_price_latest_snapshot_run
    FOREIGN KEY (run_id) REFERENCES shopping_price_runs(run_id) ON DELETE CASCADE,
  CONSTRAINT fk_shopping_price_latest_snapshot_product
    FOREIGN KEY (product_id) REFERENCES catalog_products(product_id) ON DELETE CASCADE,
  CONSTRAINT chk_shopping_price_latest_snapshot_seller_name CHECK (BTRIM(seller_name) <> ''),
  CONSTRAINT chk_shopping_price_latest_snapshot_channel CHECK (BTRIM(channel) <> ''),
  CONSTRAINT chk_shopping_price_latest_snapshot_observed_price CHECK (observed_price >= 0),
  CONSTRAINT chk_shopping_price_latest_snapshot_currency CHECK (currency_code ~ '^[A-Z]{3}$')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_shopping_price_latest_snapshot_tenant_product
  ON shopping_price_latest_snapshot (tenant_id, product_id);

CREATE INDEX IF NOT EXISTS idx_shopping_price_latest_snapshot_tenant_observed
  ON shopping_price_latest_snapshot (tenant_id, observed_at DESC);

ALTER TABLE shopping_price_latest_snapshot ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopping_price_latest_snapshot FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS shopping_price_latest_snapshot_tenant_isolation ON shopping_price_latest_snapshot;
CREATE POLICY shopping_price_latest_snapshot_tenant_isolation
ON shopping_price_latest_snapshot
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());
