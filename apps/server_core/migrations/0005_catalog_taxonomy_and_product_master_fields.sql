ALTER TABLE catalog_products
  ADD COLUMN IF NOT EXISTS brand_name TEXT NULL,
  ADD COLUMN IF NOT EXISTS stock_profile_code TEXT NULL,
  ADD COLUMN IF NOT EXISTS primary_taxonomy_node_id TEXT NULL;

CREATE TABLE IF NOT EXISTS catalog_taxonomy_nodes (
  taxonomy_node_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  name TEXT NOT NULL,
  name_norm TEXT NOT NULL,
  code TEXT NULL,
  parent_taxonomy_node_id TEXT NULL,
  level INTEGER NOT NULL,
  path TEXT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_catalog_taxonomy_nodes_level CHECK (level >= 0),
  CONSTRAINT fk_catalog_taxonomy_nodes_parent
    FOREIGN KEY (parent_taxonomy_node_id) REFERENCES catalog_taxonomy_nodes(taxonomy_node_id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_taxonomy_nodes_tenant_parent_name_norm
  ON catalog_taxonomy_nodes (tenant_id, COALESCE(parent_taxonomy_node_id, ''), name_norm);

CREATE INDEX IF NOT EXISTS idx_catalog_taxonomy_nodes_tenant_parent
  ON catalog_taxonomy_nodes (tenant_id, parent_taxonomy_node_id);

CREATE INDEX IF NOT EXISTS idx_catalog_taxonomy_nodes_tenant_level
  ON catalog_taxonomy_nodes (tenant_id, level);

CREATE TABLE IF NOT EXISTS catalog_taxonomy_level_defs (
  tenant_id TEXT NOT NULL,
  level INTEGER NOT NULL,
  label TEXT NOT NULL,
  short_label TEXT NULL,
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, level),
  CONSTRAINT chk_catalog_taxonomy_level_defs_level CHECK (level >= 0)
);

ALTER TABLE catalog_products
  DROP CONSTRAINT IF EXISTS fk_catalog_products_primary_taxonomy_node;

ALTER TABLE catalog_products
  ADD CONSTRAINT fk_catalog_products_primary_taxonomy_node
  FOREIGN KEY (primary_taxonomy_node_id) REFERENCES catalog_taxonomy_nodes(taxonomy_node_id) ON DELETE SET NULL;

ALTER TABLE catalog_taxonomy_nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE catalog_taxonomy_nodes FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS catalog_taxonomy_nodes_tenant_isolation ON catalog_taxonomy_nodes;
CREATE POLICY catalog_taxonomy_nodes_tenant_isolation
ON catalog_taxonomy_nodes
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

ALTER TABLE catalog_taxonomy_level_defs ENABLE ROW LEVEL SECURITY;
ALTER TABLE catalog_taxonomy_level_defs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS catalog_taxonomy_level_defs_tenant_isolation ON catalog_taxonomy_level_defs;
CREATE POLICY catalog_taxonomy_level_defs_tenant_isolation
ON catalog_taxonomy_level_defs
USING (tenant_id = current_tenant_id())
WITH CHECK (tenant_id = current_tenant_id());

INSERT INTO catalog_taxonomy_level_defs (
  tenant_id,
  level,
  label,
  short_label,
  is_enabled
)
VALUES
  ('bootstrap-local', 0, 'Department', 'Dept', TRUE),
  ('bootstrap-local', 1, 'Category', 'Cat', TRUE),
  ('bootstrap-local', 2, 'Family', 'Fam', TRUE)
ON CONFLICT (tenant_id, level) DO NOTHING;
