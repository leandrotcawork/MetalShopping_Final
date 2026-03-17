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
  'pricing.default_margin_floor',
  'global',
  'global',
  15.0,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed default pricing margin floor for the first pricing slice.'
)
ON CONFLICT (threshold_name, scope_type, scope_key, effective_from) DO NOTHING;

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
  'pricing.manual_price_override',
  'global',
  'global',
  '{"allow_manual_price_override": true}'::jsonb,
  TRUE,
  NOW(),
  'bootstrap',
  'Initial database-backed policy allowing manual price override in the first pricing slice.'
)
ON CONFLICT (policy_name, scope_type, scope_key, effective_from) DO NOTHING;
