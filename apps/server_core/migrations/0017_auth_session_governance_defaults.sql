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
  'Initial database-backed default enabling the backend-owned web session surface.'
)
ON CONFLICT (flag_name, scope_type, scope_key, effective_from) DO NOTHING;

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
  'auth.session_idle_timeout_minutes',
  'global',
  'global',
  30.0,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed idle timeout for backend-owned authenticated web sessions.'
)
ON CONFLICT (threshold_name, scope_type, scope_key, effective_from) DO NOTHING;

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
  'auth.session_absolute_timeout_minutes',
  'global',
  'global',
  480.0,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed absolute timeout for backend-owned authenticated web sessions.'
)
ON CONFLICT (threshold_name, scope_type, scope_key, effective_from) DO NOTHING;
