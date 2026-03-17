package unit

import (
	"testing"

	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
)

func TestGovernanceBootstrapRegistersFeatureFlags(t *testing.T) {
	registry := governancebootstrap.NewRegistry()

	entry, ok := registry.Get(governancebootstrap.CatalogProductCreationEnabledKey)
	if !ok {
		t.Fatal("expected catalog product creation flag to be registered")
	}
	if entry.BoundedContext != "catalog" {
		t.Fatalf("expected catalog bounded context, got %q", entry.BoundedContext)
	}
}
