package feature_flags

import (
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/config_registry"
)

type ScopeValues struct {
	Global        *bool
	Environment   map[string]bool
	Tenant        map[string]bool
	Module        map[string]bool
	EntityProfile map[string]bool
	FeatureTarget map[string]bool
}

type ResolutionContext struct {
	Environment   string
	TenantID      string
	Module        string
	EntityProfile string
	FeatureTarget string
}

type Resolver struct {
	registry *config_registry.Registry
	values   map[string]ScopeValues
}

func NewResolver(registry *config_registry.Registry, values map[string]ScopeValues) *Resolver {
	return &Resolver{registry: registry, values: values}
}

func (r *Resolver) Resolve(key string, ctx ResolutionContext) (bool, error) {
	entry, ok := r.registry.Get(key)
	if !ok {
		return false, fmt.Errorf("feature flag not registered: %s", key)
	}
	if entry.Kind != config_registry.ArtifactFeatureFlag {
		return false, fmt.Errorf("governance entry is not a feature flag: %s", key)
	}

	scopeValues, ok := r.values[strings.TrimSpace(key)]
	if !ok {
		return false, nil
	}

	if value, ok := lookup(scopeValues.FeatureTarget, ctx.FeatureTarget); ok {
		return value, nil
	}
	if value, ok := lookup(scopeValues.EntityProfile, ctx.EntityProfile); ok {
		return value, nil
	}
	if value, ok := lookup(scopeValues.Module, ctx.Module); ok {
		return value, nil
	}
	if value, ok := lookup(scopeValues.Tenant, ctx.TenantID); ok {
		return value, nil
	}
	if value, ok := lookup(scopeValues.Environment, ctx.Environment); ok {
		return value, nil
	}
	if scopeValues.Global != nil {
		return *scopeValues.Global, nil
	}
	return false, nil
}

func lookup(values map[string]bool, key string) (bool, bool) {
	if len(values) == 0 {
		return false, false
	}
	value, ok := values[strings.TrimSpace(key)]
	return value, ok
}
