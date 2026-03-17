CREATE TABLE IF NOT EXISTS governance_policy_values (
  policy_name TEXT NOT NULL,
  scope_type TEXT NOT NULL,
  scope_key TEXT NOT NULL,
  policy_json JSONB NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  effective_to TIMESTAMPTZ NULL,
  updated_by TEXT NOT NULL,
  reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (policy_name, scope_type, scope_key, effective_from),
  CONSTRAINT chk_governance_policy_values_scope_type CHECK (
    scope_type IN ('global', 'environment', 'tenant', 'module')
  ),
  CONSTRAINT chk_governance_policy_values_scope_key CHECK (
    BTRIM(scope_key) <> ''
  ),
  CONSTRAINT chk_governance_policy_values_effective_window CHECK (
    effective_to IS NULL OR effective_to > effective_from
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_governance_policy_values_open_window
  ON governance_policy_values (policy_name, scope_type, scope_key)
  WHERE is_active = TRUE AND effective_to IS NULL;

INSERT INTO governance_policy_values (
  policy_name,
  scope_type,
  scope_key,
  policy_json,
  is_active,
  effective_from,
  updated_by,
  reason
)
VALUES (
  'iam.admin_role_assignment',
  'global',
  'global',
  '{"allow_admin_role_assignment": true}'::jsonb,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial IAM policy allowing admin role assignment while bootstrap auth remains active.'
)
ON CONFLICT (policy_name, scope_type, scope_key, effective_from) DO NOTHING;
