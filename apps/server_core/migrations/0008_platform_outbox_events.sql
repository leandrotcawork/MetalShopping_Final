CREATE TABLE IF NOT EXISTS outbox_events (
  event_id TEXT PRIMARY KEY,
  aggregate_type TEXT NOT NULL,
  aggregate_id TEXT NOT NULL,
  event_name TEXT NOT NULL,
  event_version TEXT NOT NULL,
  tenant_id TEXT NULL,
  trace_id TEXT NULL,
  idempotency_key TEXT NOT NULL,
  payload_json JSONB NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  attempts INTEGER NOT NULL DEFAULT 0,
  available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  published_at TIMESTAMPTZ NULL,
  last_error TEXT NULL,
  CONSTRAINT chk_outbox_events_status CHECK (
    status IN ('pending', 'published', 'failed')
  ),
  CONSTRAINT chk_outbox_events_attempts CHECK (attempts >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_outbox_events_idempotency_key
  ON outbox_events (idempotency_key);

CREATE INDEX IF NOT EXISTS idx_outbox_events_dispatch
  ON outbox_events (status, available_at, created_at);
