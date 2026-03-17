package unit

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"metalshopping/server_core/internal/platform/db/postgres"
)

func TestLoadConfigFromEnvUsesSecureDefaultsForComponentConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PGHOST", "db.internal")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGDATABASE", "metalshopping")
	t.Setenv("PGUSER", "service_user")
	t.Setenv("PGPASSWORD", "super-secret")
	t.Setenv("PGSSLMODE", "")
	t.Setenv("MS_PG_MAX_OPEN_CONNS", "")
	t.Setenv("MS_PG_MAX_IDLE_CONNS", "")
	t.Setenv("MS_PG_CONN_MAX_LIFETIME_SECONDS", "")
	t.Setenv("MS_PG_CONN_MAX_IDLE_TIME_SECONDS", "")
	t.Setenv("MS_PG_PING_TIMEOUT_SECONDS", "")

	cfg, err := postgres.LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}

	parsed, err := url.Parse(cfg.DSN)
	if err != nil {
		t.Fatalf("expected valid dsn, got error: %v", err)
	}

	if got := parsed.Query().Get("sslmode"); got != "require" {
		t.Fatalf("expected sslmode=require, got %q", got)
	}
	if cfg.MaxOpenConns != 25 {
		t.Fatalf("expected max open conns 25, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 25 {
		t.Fatalf("expected max idle conns 25, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("expected conn max lifetime 30m, got %s", cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != 5*time.Minute {
		t.Fatalf("expected conn max idle time 5m, got %s", cfg.ConnMaxIdleTime)
	}
	if cfg.PingTimeout != 5*time.Second {
		t.Fatalf("expected ping timeout 5s, got %s", cfg.PingTimeout)
	}
}

func TestLoadConfigFromEnvPrefersDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://svc:secret@db.example.com:5432/metalshopping?sslmode=verify-full")
	t.Setenv("PGHOST", "")
	t.Setenv("PGDATABASE", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")

	cfg, err := postgres.LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}

	if cfg.DSN != "postgres://svc:secret@db.example.com:5432/metalshopping?sslmode=verify-full" {
		t.Fatalf("expected DATABASE_URL to win, got %q", cfg.DSN)
	}
}

func TestLoadConfigFromEnvRejectsInvalidPoolConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://svc:secret@db.example.com:5432/metalshopping?sslmode=require")
	t.Setenv("MS_PG_MAX_OPEN_CONNS", "0")

	_, err := postgres.LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid pool config")
	}
	if !strings.Contains(err.Error(), "MS_PG_MAX_OPEN_CONNS") {
		t.Fatalf("expected pool config error, got %v", err)
	}
}

func TestLoadConfigFromEnvRejectsMissingDatabaseSettings(t *testing.T) {
	unsetEnv(t, "DATABASE_URL", "PGHOST", "PGPORT", "PGDATABASE", "PGUSER", "PGPASSWORD", "PGSSLMODE")

	_, err := postgres.LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for missing postgres config")
	}
	if !strings.Contains(err.Error(), "postgres config missing") {
		t.Fatalf("expected missing config error, got %v", err)
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}
