CREATE TABLE IF NOT EXISTS governance_threshold_values (
  threshold_name TEXT NOT NULL,
  scope_type TEXT NOT NULL,
  scope_key TEXT NOT NULL,
  threshold_value DOUBLE PRECISION NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  effective_to TIMESTAMPTZ NULL,
  updated_by TEXT NOT NULL,
  reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (threshold_name, scope_type, scope_key, effective_from),
  CONSTRAINT chk_governance_threshold_values_scope_type CHECK (
    scope_type IN ('global', 'environment', 'tenant', 'module', 'entity/profile')
  ),
  CONSTRAINT chk_governance_threshold_values_scope_key CHECK (
    BTRIM(scope_key) <> ''
  ),
  CONSTRAINT chk_governance_threshold_values_effective_window CHECK (
    effective_to IS NULL OR effective_to > effective_from
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_governance_threshold_values_open_window
  ON governance_threshold_values (threshold_name, scope_type, scope_key)
  WHERE is_active = TRUE AND effective_to IS NULL;

INSERT INTO governance_threshold_values (
  threshold_name,
  scope_type,
  scope_key,
  threshold_value,
  is_active,
  effective_from,
  updated_by,
  reason
)
VALUES (
  'catalog.max_description_length',
  'global',
  'global',
  4000,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial maximum description length for canonical catalog products.'
)
ON CONFLICT (threshold_name, scope_type, scope_key, effective_from) DO NOTHING;
