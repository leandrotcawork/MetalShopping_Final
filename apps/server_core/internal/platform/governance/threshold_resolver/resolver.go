package threshold_resolver

import (
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/config_registry"
)

type ScopeValues struct {
	Global        *float64
	Environment   map[string]float64
	Tenant        map[string]float64
	Module        map[string]float64
	EntityProfile map[string]float64
}

type ResolutionContext struct {
	Environment   string
	TenantID      string
	Module        string
	EntityProfile string
}

type Resolver struct {
	registry *config_registry.Registry
	values   map[string]ScopeValues
}

func NewResolver(registry *config_registry.Registry, values map[string]ScopeValues) *Resolver {
	return &Resolver{registry: registry, values: values}
}

func (r *Resolver) Resolve(key string, ctx ResolutionContext) (float64, bool, error) {
	entry, ok := r.registry.Get(key)
	if !ok {
		return 0, false, fmt.Errorf("threshold not registered: %s", key)
	}
	if entry.Kind != config_registry.ArtifactThreshold {
		return 0, false, fmt.Errorf("governance entry is not a threshold: %s", key)
	}

	scopeValues, ok := r.values[strings.TrimSpace(key)]
	if !ok {
		return 0, false, nil
	}

	if value, ok := lookup(scopeValues.EntityProfile, ctx.EntityProfile); ok {
		return value, true, nil
	}
	if value, ok := lookup(scopeValues.Module, ctx.Module); ok {
		return value, true, nil
	}
	if value, ok := lookup(scopeValues.Tenant, ctx.TenantID); ok {
		return value, true, nil
	}
	if value, ok := lookup(scopeValues.Environment, ctx.Environment); ok {
		return value, true, nil
	}
	if scopeValues.Global != nil {
		return *scopeValues.Global, true, nil
	}

	return 0, false, nil
}

func lookup(values map[string]float64, key string) (float64, bool) {
	if len(values) == 0 {
		return 0, false
	}
	value, ok := values[strings.TrimSpace(key)]
	return value, ok
}
