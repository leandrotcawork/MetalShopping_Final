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
  'auth.web_session_enabled',
  'global',
  'global',
  TRUE,
  TRUE,
  NOW(),
  'bootstrap',
  'Repair migration ensuring the auth web session feature flag is populated in the runtime flag_value column.'
)
ON CONFLICT (flag_name, scope_type, scope_key, effective_from) DO NOTHING;
