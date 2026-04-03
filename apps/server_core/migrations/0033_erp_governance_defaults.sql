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
  'erp_integrations.integration_enabled',
  'global',
  'global',
  FALSE,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed governance default for ERP integration feature flag (disabled by default).'
)
ON CONFLICT (flag_name, scope_type, scope_key, effective_from) DO NOTHING;

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
  'erp_integrations.auto_promotion',
  'global',
  'global',
  '{"allow_auto_promotion": true}'::jsonb,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed policy enabling auto-promotion of reconciled ERP records.'
)
ON CONFLICT (policy_name, scope_type, scope_key, effective_from) DO NOTHING;
