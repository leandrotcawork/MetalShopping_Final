package unit

import (
	"encoding/json"
	"testing"

	"metalshopping/server_core/internal/platform/governance/config_registry"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
	"metalshopping/server_core/internal/platform/governance/policy_resolver"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
)

func TestFeatureFlagResolverHonorsMostSpecificScope(t *testing.T) {
	registry := config_registry.NewRegistry()
	registry.MustRegister(config_registry.Entry{
		Key:            "catalog.product_creation_enabled",
		Kind:           config_registry.ArtifactFeatureFlag,
		BoundedContext: "catalog",
		ValueType:      config_registry.ValueTypeBool,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeTenant,
			config_registry.ScopeFeatureTarget,
		},
	})

	global := true
	resolver := feature_flags.NewResolver(registry, map[string]feature_flags.ScopeValues{
		"catalog.product_creation_enabled": {
			Global:        &global,
			Tenant:        map[string]bool{"tenant-a": false},
			FeatureTarget: map[string]bool{"beta-rollout": true},
		},
	})

	value, err := resolver.Resolve("catalog.product_creation_enabled", feature_flags.ResolutionContext{
		TenantID:      "tenant-a",
		FeatureTarget: "beta-rollout",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !value {
		t.Fatal("expected feature-target override to win")
	}
}

func TestThresholdResolverHonorsTenantOverEnvironment(t *testing.T) {
	registry := config_registry.NewRegistry()
	registry.MustRegister(config_registry.Entry{
		Key:            "pricing.default_margin_floor",
		Kind:           config_registry.ArtifactThreshold,
		BoundedContext: "pricing",
		ValueType:      config_registry.ValueTypeNumber,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
		},
	})

	global := 10.0
	resolver := threshold_resolver.NewResolver(registry, map[string]threshold_resolver.ScopeValues{
		"pricing.default_margin_floor": {
			Global:      &global,
			Environment: map[string]float64{"local": 12.5},
			Tenant:      map[string]float64{"tenant-a": 15.0},
		},
	})

	value, ok, err := resolver.Resolve("pricing.default_margin_floor", threshold_resolver.ResolutionContext{
		Environment: "local",
		TenantID:    "tenant-a",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok || value != 15.0 {
		t.Fatalf("expected tenant override 15.0, got ok=%v value=%v", ok, value)
	}
}

func TestPolicyResolverReturnsModulePolicy(t *testing.T) {
	registry := config_registry.NewRegistry()
	registry.MustRegister(config_registry.Entry{
		Key:            "iam.admin_role_assignment",
		Kind:           config_registry.ArtifactPolicy,
		BoundedContext: "iam",
		ValueType:      config_registry.ValueTypeJSON,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeTenant,
			config_registry.ScopeModule,
		},
	})

	resolver := policy_resolver.NewResolver(registry, map[string]policy_resolver.ScopeValues{
		"iam.admin_role_assignment": {
			Global: json.RawMessage(`{"allow_self_assignment":false}`),
			Module: map[string]json.RawMessage{
				"iam": json.RawMessage(`{"allow_self_assignment":false,"require_admin_actor":true}`),
			},
		},
	})

	value, ok, err := resolver.Resolve("iam.admin_role_assignment", policy_resolver.ResolutionContext{
		Module: "iam",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatal("expected policy value")
	}
	if string(value) != `{"allow_self_assignment":false,"require_admin_actor":true}` {
		t.Fatalf("unexpected policy payload: %s", string(value))
	}
}
