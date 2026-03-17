UPDATE governance_threshold_values
SET
  is_active = FALSE,
  effective_to = COALESCE(effective_to, NOW()),
  updated_at = NOW(),
  updated_by = 'bootstrap',
  reason = 'Deprecated after pricing semantic alignment removed canonical margin floor persistence.'
WHERE threshold_name = 'pricing.default_margin_floor'
  AND is_active = TRUE
  AND effective_to IS NULL;
