CREATE TABLE IF NOT EXISTS governance_feature_flag_values (
  flag_name TEXT NOT NULL,
  scope_type TEXT NOT NULL,
  scope_key TEXT NOT NULL,
  flag_value BOOLEAN NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  effective_to TIMESTAMPTZ NULL,
  updated_by TEXT NOT NULL,
  reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (flag_name, scope_type, scope_key, effective_from),
  CONSTRAINT chk_governance_feature_flags_scope_type CHECK (
    scope_type IN ('global', 'environment', 'tenant', 'feature-target')
  ),
  CONSTRAINT chk_governance_feature_flags_scope_key CHECK (
    BTRIM(scope_key) <> ''
  ),
  CONSTRAINT chk_governance_feature_flags_effective_window CHECK (
    effective_to IS NULL OR effective_to > effective_from
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_governance_feature_flag_values_open_window
  ON governance_feature_flag_values (flag_name, scope_type, scope_key)
  WHERE is_active = TRUE AND effective_to IS NULL;

CREATE INDEX IF NOT EXISTS idx_governance_feature_flag_values_lookup
  ON governance_feature_flag_values (flag_name, is_active, scope_type, scope_key, effective_from DESC);

INSERT INTO governance_feature_flag_values (
  flag_name,
  scope_type,
  scope_key,
  flag_value,
  is_active,
  effective_from,
  updated_by,
  reason
)
VALUES (
  'catalog.product_creation_enabled',
  'global',
  'global',
  TRUE,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed governance default for catalog product creation.'
)
ON CONFLICT (flag_name, scope_type, scope_key, effective_from) DO NOTHING;
