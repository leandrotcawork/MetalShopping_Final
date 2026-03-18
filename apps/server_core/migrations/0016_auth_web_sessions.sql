CREATE TABLE IF NOT EXISTS auth_web_login_states (
  login_state_id TEXT PRIMARY KEY,
  code_verifier TEXT NOT NULL,
  return_to TEXT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_web_login_states_expires_at
  ON auth_web_login_states (expires_at);

CREATE TABLE IF NOT EXISTS auth_web_sessions (
  session_id TEXT PRIMARY KEY,
  subject_id TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  email TEXT NULL,
  display_name TEXT NULL,
  issued_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  idle_timeout_expires_at TIMESTAMPTZ NOT NULL,
  absolute_timeout_expires_at TIMESTAMPTZ NOT NULL,
  invalidated_at TIMESTAMPTZ NULL,
  invalidation_reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_web_sessions_subject_id
  ON auth_web_sessions (subject_id);

CREATE INDEX IF NOT EXISTS idx_auth_web_sessions_tenant_id
  ON auth_web_sessions (tenant_id);

CREATE INDEX IF NOT EXISTS idx_auth_web_sessions_active
  ON auth_web_sessions (session_id, idle_timeout_expires_at, absolute_timeout_expires_at)
  WHERE invalidated_at IS NULL;
