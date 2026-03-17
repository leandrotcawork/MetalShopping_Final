package unit

import (
	"context"
	"testing"

	cataloggov "metalshopping/server_core/internal/modules/catalog/adapters/governance"
	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
)

func TestCatalogProductCreationGuardUsesBootstrapDefault(t *testing.T) {
	registry := governancebootstrap.NewRegistry()
	globalEnabled := true
	resolver := feature_flags.NewResolver(registry, map[string]feature_flags.ScopeValues{
		governancebootstrap.CatalogProductCreationEnabledKey: {
			Global: &globalEnabled,
		},
	})

	guard := cataloggov.NewProductCreationGuard(resolver, "local")

	enabled, err := guard.IsProductCreationEnabled(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no guard error, got %v", err)
	}
	if !enabled {
		t.Fatal("expected product creation enabled by default")
	}
}
