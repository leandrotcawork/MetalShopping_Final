package oracle

import (
	"strings"
	"testing"
)

func TestConnectStringRejectsMissingTarget(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Host:     "db.example.internal",
		Port:     1521,
		Username: "erp_user",
		Password: "erp_secret",
	}

	_, err := cfg.ConnectString()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "exactly one of service_name or sid") {
		t.Fatalf("expected target validation error, got %v", err)
	}
}

func TestConnectStringRejectsBothTargets(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Host:        "db.example.internal",
		Port:        1521,
		ServiceName: strPtr("ORCL"),
		SID:         strPtr("XE"),
		Username:    "erp_user",
		Password:    "erp_secret",
	}

	_, err := cfg.ConnectString()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "exactly one of service_name or sid") {
		t.Fatalf("expected target validation error, got %v", err)
	}
}

func TestConnectStringBuildsServiceNameDsn(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Host:              "db.example.internal",
		Port:              1521,
		ServiceName:       strPtr("ORCL"),
		Username:          "erp_user",
		Password:          "erp_secret",
		ConnectTimeoutSec: 12,
	}

	got, err := cfg.ConnectString()
	if err != nil {
		t.Fatalf("ConnectString returned error: %v", err)
	}
	for _, want := range []string{"db.example.internal", "1521", "/ORCL", "erp_user", "erp_secret", "connect_timeout=12"} {
		if !strings.Contains(got, want) {
			t.Fatalf("ConnectString() = %q, missing %q", got, want)
		}
	}
}

func TestConnectStringRejectsMissingHost(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Port:              1521,
		ServiceName:       strPtr("ORCL"),
		Username:          "erp_user",
		Password:          "erp_secret",
		ConnectTimeoutSec: 12,
	}

	_, err := cfg.ConnectString()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "host must not be empty") {
		t.Fatalf("expected host validation error, got %v", err)
	}
}

func TestConnectStringRejectsMissingPort(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Host:        "db.example.internal",
		ServiceName: strPtr("ORCL"),
		Username:    "erp_user",
		Password:    "erp_secret",
	}

	_, err := cfg.ConnectString()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "port must be a positive integer") {
		t.Fatalf("expected port validation error, got %v", err)
	}
}

func strPtr(v string) *string {
	return &v
}
