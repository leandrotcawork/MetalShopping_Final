package sankhya

import (
	"context"
	"strings"
	"testing"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
)

func TestValidateConnectionParsesHostUserPasswordAndDefaultPort(t *testing.T) {
	t.Parallel()

	connector := New()
	if err := connector.ValidateConnection(context.Background(), "sankhya://leandroth:Leandrocruz04@10.55.10.101?service=PROD"); err != nil {
		t.Fatalf("ValidateConnection returned error: %v", err)
	}

	cfg, err := parseConnectionRef("sankhya://leandroth:Leandrocruz04@10.55.10.101?service=PROD")
	if err != nil {
		t.Fatalf("parseConnectionRef returned error: %v", err)
	}
	if cfg.Host != "10.55.10.101" {
		t.Fatalf("expected host 10.55.10.101, got %q", cfg.Host)
	}
	if cfg.Port != defaultOraclePort {
		t.Fatalf("expected default port %d, got %d", defaultOraclePort, cfg.Port)
	}
	if cfg.Username != "leandroth" || cfg.Password != "Leandrocruz04" {
		t.Fatalf("unexpected credentials parsed: %#v", cfg)
	}
	if cfg.Service != "PROD" {
		t.Fatalf("expected service PROD, got %q", cfg.Service)
	}
}

func TestQueryForEntityReturnsSnapshotSQL(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"products":  "FROM TGFPRO",
		"prices":    "FROM TGFTAB",
		"inventory": "FROM TGFEST",
	}

	for entityName, wantFragment := range cases {
		entity := erp_runtime.EntityType(entityName)
		sql, err := queryForEntity(entity)
		if err != nil {
			t.Fatalf("queryForEntity(%s) returned error: %v", entity, err)
		}
		if !strings.Contains(sql, wantFragment) {
			t.Fatalf("queryForEntity(%s) = %q, want fragment %q", entity, sql, wantFragment)
		}
	}
}
