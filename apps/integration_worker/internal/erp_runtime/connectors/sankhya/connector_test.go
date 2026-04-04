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
	if err := connector.ValidateConnection(context.Background(), "sankhya://user:pass@dummy-host.example?service=PROD"); err != nil {
		t.Fatalf("ValidateConnection returned error: %v", err)
	}

	cfg, err := parseConnectionRef("sankhya://user:pass@dummy-host.example?service=PROD")
	if err != nil {
		t.Fatalf("parseConnectionRef returned error: %v", err)
	}
	if cfg.Host != "dummy-host.example" {
		t.Fatalf("expected host dummy-host.example, got %q", cfg.Host)
	}
	if cfg.Port != defaultOraclePort {
		t.Fatalf("expected default port %d, got %d", defaultOraclePort, cfg.Port)
	}
	if cfg.Username != "user" || cfg.Password != "pass" {
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
		switch entity {
		case erp_runtime.EntityTypeProducts:
			for _, wantColumn := range []string{"CODPROD", "DESCRPROD", "REFERENCIA", "REFFORN"} {
				if !strings.Contains(sql, wantColumn) {
					t.Fatalf("queryForEntity(%s) missing key column %q", entity, wantColumn)
				}
			}
		case erp_runtime.EntityTypePrices:
			for _, wantColumn := range []string{"NUTAB", "CODTAB", "DTVIGOR", "VLRVENDA"} {
				if !strings.Contains(sql, wantColumn) {
					t.Fatalf("queryForEntity(%s) missing key column %q", entity, wantColumn)
				}
			}
		case erp_runtime.EntityTypeInventory:
			for _, wantColumn := range []string{"CODPROD", "CODEMP", "CODLOCAL", "ESTOQUE", "RESERVADO"} {
				if !strings.Contains(sql, wantColumn) {
					t.Fatalf("queryForEntity(%s) missing key column %q", entity, wantColumn)
				}
			}
		}
	}
}
