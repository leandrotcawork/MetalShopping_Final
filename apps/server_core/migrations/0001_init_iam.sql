CREATE TABLE IF NOT EXISTS iam_users (
  user_id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS iam_user_roles (
  user_id TEXT NOT NULL REFERENCES iam_users(user_id) ON DELETE CASCADE,
  role_code TEXT NOT NULL,
  assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  assigned_by TEXT NOT NULL,
  PRIMARY KEY (user_id, role_code),
  CONSTRAINT chk_iam_user_roles_role_code CHECK (
    role_code IN (
      'admin',
      'tenant_admin',
      'catalog_manager',
      'pricing_manager',
      'sales_manager',
      'analyst',
      'automation_owner',
      'viewer'
    )
  )
);

CREATE INDEX IF NOT EXISTS idx_iam_user_roles_role_code ON iam_user_roles(role_code);
