package unit

import (
	"testing"

	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/config_registry"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
)

func TestFeatureFlagStoreBuildsScopeValuesFromRecords(t *testing.T) {
	registry := governancebootstrap.NewRegistry()

	values, err := feature_flags.TestOnlyScopeValuesFromRecords(registry, []feature_flags.TestOnlyValueRecord{
		{
			FlagName:  governancebootstrap.CatalogProductCreationEnabledKey,
			ScopeType: config_registry.ScopeGlobal,
			ScopeKey:  "global",
			Value:     true,
		},
		{
			FlagName:  governancebootstrap.CatalogProductCreationEnabledKey,
			ScopeType: config_registry.ScopeTenant,
			ScopeKey:  "tenant-a",
			Value:     false,
		},
	})
	if err != nil {
		t.Fatalf("expected no scope value error, got %v", err)
	}

	scopeValues := values[governancebootstrap.CatalogProductCreationEnabledKey]
	if scopeValues.Global == nil || !*scopeValues.Global {
		t.Fatal("expected global flag value true")
	}
	if got := scopeValues.Tenant["tenant-a"]; got {
		t.Fatalf("expected tenant-a false override, got %v", got)
	}
}
