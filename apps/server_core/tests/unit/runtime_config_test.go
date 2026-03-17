package unit

import (
	"os"
	"path/filepath"
	"testing"

	"metalshopping/server_core/internal/platform/runtime_config"
)

func TestLoadDotEnvIfPresentLoadsMissingVariables(t *testing.T) {
	tempDir := t.TempDir()
	dotEnvPath := filepath.Join(tempDir, ".env")
	content := []byte("APP_PORT=8090\nMS_AUTH_STATIC_NAME=Local Admin\n")
	if err := os.WriteFile(dotEnvPath, content, 0o644); err != nil {
		t.Fatalf("write temp dotenv: %v", err)
	}

	unsetRuntimeEnv(t, "APP_PORT", "MS_AUTH_STATIC_NAME")

	if err := runtime_config.LoadDotEnvIfPresent(dotEnvPath); err != nil {
		t.Fatalf("expected dotenv to load, got %v", err)
	}

	if got := os.Getenv("APP_PORT"); got != "8090" {
		t.Fatalf("expected APP_PORT 8090, got %q", got)
	}
	if got := os.Getenv("MS_AUTH_STATIC_NAME"); got != "Local Admin" {
		t.Fatalf("expected auth name, got %q", got)
	}
}

func unsetRuntimeEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}

func TestHTTPAddressFromEnvUsesDefaultPort(t *testing.T) {
	t.Setenv("APP_PORT", "")

	addr, err := runtime_config.HTTPAddressFromEnv()
	if err != nil {
		t.Fatalf("expected default address, got %v", err)
	}
	if addr != ":8080" {
		t.Fatalf("expected :8080, got %q", addr)
	}
}

func TestEnvironmentFromEnvUsesLocalDefault(t *testing.T) {
	t.Setenv("APP_ENV", "")

	if got := runtime_config.EnvironmentFromEnv(); got != "local" {
		t.Fatalf("expected local, got %q", got)
	}
}
