package sankhya

import (
	"context"
	"errors"
	"strings"
	"testing"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
)

func TestValidateConnectionAcceptsStructuredConnection(t *testing.T) {
	t.Parallel()

	connector := NewWithRunnerFactory(func(context.Context, erp_runtime.ExtractConnection) (dbsource.QueryRunner, error) {
		return &noopRunner{}, nil
	})
	if err := connector.ValidateConnection(context.Background(), erp_runtime.ExtractConnection{
		Host:              "dummy-host.example",
		Port:              defaultOraclePort,
		ServiceName:       strPtr("PROD"),
		Username:          "user",
		PasswordSecretRef: "erp/sankhya/password",
	}); err != nil {
		t.Fatalf("ValidateConnection returned error: %v", err)
	}
}

func TestValidateConnectionRejectsMissingHost(t *testing.T) {
	t.Parallel()

	connector := NewWithRunnerFactory(func(context.Context, erp_runtime.ExtractConnection) (dbsource.QueryRunner, error) {
		return &noopRunner{}, nil
	})
	err := connector.ValidateConnection(context.Background(), erp_runtime.ExtractConnection{
		Port:              defaultOraclePort,
		ServiceName:       strPtr("PROD"),
		Username:          "user",
		PasswordSecretRef: "erp/sankhya/password",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "host must not be empty") {
		t.Fatalf("expected host validation error, got %v", err)
	}
}

func TestValidateConnectionRejectsMissingPasswordSecretRef(t *testing.T) {
	t.Parallel()

	connector := NewWithRunnerFactory(func(context.Context, erp_runtime.ExtractConnection) (dbsource.QueryRunner, error) {
		return &noopRunner{}, nil
	})
	err := connector.ValidateConnection(context.Background(), erp_runtime.ExtractConnection{
		Host:        "dummy-host.example",
		Port:        defaultOraclePort,
		ServiceName: strPtr("PROD"),
		Username:    "user",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "password_secret_ref must not be empty") {
		t.Fatalf("expected password validation error, got %v", err)
	}
}

func strPtr(v string) *string { return &v }

func TestValidateConnectionUsesRunnerFactory(t *testing.T) {
	t.Parallel()

	called := false
	connector := NewWithRunnerFactory(func(_ context.Context, conn erp_runtime.ExtractConnection) (dbsource.QueryRunner, error) {
		called = true
		if conn.Host != "10.55.10.101" {
			return nil, errors.New("unexpected host")
		}
		return &noopRunner{}, nil
	})

	err := connector.ValidateConnection(context.Background(), erp_runtime.ExtractConnection{
		Kind:              "oracle",
		Host:              "10.55.10.101",
		Port:              1521,
		ServiceName:       strPtr("ORCL"),
		Username:          "leandroth",
		PasswordSecretRef: "erp/sankhya/password",
	})
	if err != nil {
		t.Fatalf("ValidateConnection returned error: %v", err)
	}
	if !called {
		t.Fatal("expected ValidateConnection to call runner factory")
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

type noopRunner struct{}

func (n *noopRunner) Query(context.Context, dbsource.QuerySpec, func(dbsource.RowReader) error) error {
	return nil
}

func (n *noopRunner) Close() error {
	return nil
}
